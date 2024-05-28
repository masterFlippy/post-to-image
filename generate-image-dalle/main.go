package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

type InputEvent struct {
	Prompt string `json:"prompt"`
	S3Key  string `json:"s3Key"`
}

type OutputEvent struct {
	Url   string `json:"url"`
	S3Key string `json:"s3Key"`
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

func handler(ctx context.Context, event InputEvent) (OutputEvent, error) {
	openaiApiKey := os.Getenv("OPENAI_API_KEY")
	if openaiApiKey == "" {
		log.Println("Error: OPENAI_API_KEY environment variable is not set")
		return OutputEvent{}, fmt.Errorf("missing OPENAI_API_KEY")
	}

	requestBody := DalleRequest{
		Model:  "dall-e-3",
		Prompt: event.Prompt,
		Size:   "1024x1024",
		N:      1,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		log.Println("Error marshalling data:", err)
		return OutputEvent{}, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(payload))
	if err != nil {
		log.Println("Error creating request:", err)
		return OutputEvent{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", openaiApiKey))

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
		return OutputEvent{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return OutputEvent{}, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Non-OK HTTP status: %s\nResponse body: %s\n", resp.Status, string(body))
		return OutputEvent{}, fmt.Errorf("non-OK HTTP status: %s", resp.Status)
	}

	var respData ResponseObject
	err = json.Unmarshal(body, &respData)
	if err != nil {
		log.Println("Error unmarshalling response:", err)
		return OutputEvent{}, err
	}

	if len(respData.Data) == 0 {
		log.Println("No data received in the response")
		return OutputEvent{}, fmt.Errorf("no data in the response")
	}

	imageURL := respData.Data[0].Url
	outputEvent := OutputEvent{Url: imageURL, S3Key: event.S3Key}

	return outputEvent, nil
}

func main() {
	lambda.Start(handler)
}
