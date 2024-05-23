package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type EventDetail struct {
	ImageUrl string `json:"prompt"`
}


func handler(request events.CloudWatchEvent) error  {
	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		log.Println("Error: BUCKET_NAME environment variable is not set")
		return fmt.Errorf("missing BUCKET_NAME")
	}

	var event EventDetail
	err := json.Unmarshal((request.Detail), &event)
	if err != nil {
		log.Println("Error unmarshalling request:", err)
		return err
		}

	imageResp, err := http.Get(event.ImageUrl)
	if err != nil {
		log.Println("Error downloading image:", err)
		return err
	}
	defer imageResp.Body.Close()

	if imageResp.StatusCode != http.StatusOK {
		log.Printf("Error: received non-OK HTTP status when downloading image: %s\n", imageResp.Status)
		return fmt.Errorf("non-OK HTTP status when downloading image: %s", imageResp.Status)
	}

	imageData, err := io.ReadAll(imageResp.Body)
	if err != nil {
		log.Println("Error reading image data:", err)
		return err
	}

	sesh := session.Must(session.NewSession())
	
	s3Client := s3.New(sesh)

	objectKey := fmt.Sprintf("generated-images/%s", filepath.Base(event.ImageUrl))
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(imageData),
		ContentType: aws.String(http.DetectContentType(imageData)),
	})
	if err != nil {
		log.Println("Error uploading image to S3:", err)
		return err
	}

	log.Println("Successfully uploaded image to S3:", objectKey)

	return nil
}

func main() {
	lambda.Start(handler)
}
