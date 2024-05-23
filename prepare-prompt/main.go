package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
)

type Request struct {
	Text string `json:"text"`
}

type EventDetail struct {
	Prompt string `json:"prompt"`
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	eventBusName := os.Getenv("EVENT_BUS_NAME")
	if eventBusName == "" {
		log.Println("Error: EVENT_BUS_NAME environment variable is not set")
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintln("missing EVENT_BUS_NAME"),
		}, nil
	}
	var req Request
	err := json.Unmarshal([]byte(request.Body), &req)
	if err != nil {
		log.Println("Error unmarshalling request:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf("Invalid request: %s", err.Error()),
		}, nil
	}

	prompt := fmt.Sprintf("Generate an image based on the following text: %s", req.Text)
	eventDetail := EventDetail{Prompt: prompt}

	detailJson, err := json.Marshal(eventDetail)
	if err != nil {
		log.Println("Error marshalling event detail:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf("Error preparing event detail: %s", err.Error()),
		}, nil
	}

	sesh := session.Must(session.NewSession())
	eb := eventbridge.New(sesh)

	event := &eventbridge.PutEventsRequestEntry{
		Source:       aws.String("prepare_prompt_function"),
		DetailType:   aws.String("promptSubmitted"),
		Detail:       aws.String(string(detailJson)),
		EventBusName: aws.String(eventBusName),
	}

	_, err = eb.PutEvents(&eventbridge.PutEventsInput{
		Entries: []*eventbridge.PutEventsRequestEntry{event},
	})

	if err != nil {
		log.Println("Error putting event to EventBridge:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf("Error sending event to EventBridge: %s", err.Error()),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       `{"message": "Event sent to EventBridge"}`,
	}, nil
}

func main() {
	lambda.Start(handler)
}
