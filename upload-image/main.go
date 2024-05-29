package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Event struct {
	Url   string `json:"url"`
	S3Key string `json:"s3Key"`
}

func handler(ctx context.Context, event Event)  error {
	// Grabbing the name of the bucket from the environment variable
	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		log.Println("Error: BUCKET_NAME environment variable is not set")
		return fmt.Errorf("missing BUCKET_NAME")
	}

	region := os.Getenv("REGION")
	if region == "" {
		log.Println("Error: REGION environment variable is not set")
		return fmt.Errorf("missing REGION")
	}

	imageResp, err := http.Get(event.Url)
	if err != nil {
		log.Println("Error downloading image:", err)
		return err
	}
	defer imageResp.Body.Close()

  	// Downloading image from provided URL
	if imageResp.StatusCode != http.StatusOK {
		log.Printf("Error: received non-OK HTTP status when downloading image: %s\n", imageResp.Status)
		return fmt.Errorf("non-OK HTTP status when downloading image: %s", imageResp.Status)
	}

	imageData, err := io.ReadAll(imageResp.Body)
	if err != nil {
		log.Println("Error reading image data:", err)
		return err
	}
	// Creating a session for S3
	sesh := session.Must(session.NewSession())

	s3Client := s3.New(sesh)

	objectKey := event.S3Key
 	 // Uploading image to S3
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
