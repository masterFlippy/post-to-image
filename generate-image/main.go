package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
)

type InputEventDetail struct {
	Prompt string `json:"prompt"`
}

type OutputEventDetail struct {
	Url string `json:"prompt"`
}

type DalleRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Size   string `json:"size"`
	N      int    `json:"n"`
}

type ResponseObject struct {
	Data []struct {
		Url string `json:"url"`
	} `json:"data"`
}

func handler(request events.CloudWatchEvent) error {
	openaiApiKey := os.Getenv("OPENAI_API_KEY")
	if openaiApiKey == "" {
		log.Println("Error: OPENAI_API_KEY environment variable is not set")
		return fmt.Errorf("missing OPENAI_API_KEY")
	}

	eventBusName := os.Getenv("EVENT_BUS_NAME")
	if eventBusName == "" {
		log.Println("Error: EVENT_BUS_NAME environment variable is not set")
		return fmt.Errorf("missing EVENT_BUS_NAME")
	}

	var inputEvent InputEventDetail
	err := json.Unmarshal((request.Detail), &inputEvent)
	if err != nil {
		log.Println("Error unmarshalling request:", err)
		return err
	}

	requestBody := DalleRequest{
		Model:  "dall-e-3",
		Prompt: inputEvent.Prompt,
		Size:   "1024x1024",
		N:      1,
	}

	log.Println("Sending request to OpenAI API with data:", requestBody)
	payload, err := json.Marshal(requestBody)
	if err != nil {
		log.Println("Error marshalling data:", err)
		return err
	}
	log.Println("Payload:", string(payload))
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(payload))
	if err != nil {
		log.Println("Error creating request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", openaiApiKey))

	log.Println("Sending request to:", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
		return err
	}
	log.Println("Received response with status:", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Non-OK HTTP status: %s\nResponse body: %s\n", resp.Status, string(bodyBytes))
		return fmt.Errorf("non-OK HTTP status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return err
	}
	var respData ResponseObject
	err = json.Unmarshal(body, &respData)

	if err != nil {
		log.Println("Error unmarshalling response:", err)
		return err
	}

	if len(respData.Data) == 0 {
		log.Println("No data received in the response")
		return fmt.Errorf("no data in the response")
	}

	imageURL := respData.Data[0].Url

	eventDetail := OutputEventDetail{Url: imageURL}

	detailJson, err := json.Marshal(eventDetail)
	if err != nil {
		log.Println("Error marshalling event detail:", err)
		return err
	}

	sesh := session.Must(session.NewSession())
	eb := eventbridge.New(sesh)

	outputEvent := &eventbridge.PutEventsRequestEntry{
		Source:       aws.String("generate_image_function"),
		DetailType:   aws.String("imageGenerated"),
		Detail:       aws.String(string(detailJson)),
		EventBusName: aws.String(eventBusName),
	}

	_, err = eb.PutEvents(&eventbridge.PutEventsInput{
		Entries: []*eventbridge.PutEventsRequestEntry{outputEvent},
	})

	if err != nil {
		log.Println("Error putting event to EventBridge:", err)
		return err
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
