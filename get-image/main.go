package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)


func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		log.Println("Error: BUCKET_NAME environment variable is not set")
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:      "missing BUCKET_NAME",
		},nil
	}

	queryParams := request.QueryStringParameters
	s3Key, exists := queryParams["s3Key"]
	if !exists {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Missing query parameter 's3Key'",
		},nil
	}

	sesh := session.Must(session.NewSession())

	s3Client := s3.New(sesh)


	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(s3Key),
	}

	_, err := s3Client.HeadObject(input)
	if err != nil {
		log.Printf(`{"url": "https://%s.s3.eu-north-1.amazonaws.com/%s"}`, bucketName, s3Key)
		log.Println("param:", s3Key)
		log.Println("Error checking if object exists:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       "Object does not exist",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf(`{"url": "https://%s.s3.eu-north-1.amazonaws.com/%s"}`, bucketName, s3Key),
	}, nil
}

func main() {
	lambda.Start(handler)
}
