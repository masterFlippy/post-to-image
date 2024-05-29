package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/comprehend"
)

type InputEvent struct {
	Detail struct {
		Text  string `json:"text"`
		S3Key string `json:"s3Key"`
		Bedrock bool `json:"bedrock"`
	} `json:"detail"`
}

type OutputEvent struct {
	Prompt string `json:"prompt"`
	S3Key  string `json:"s3Key"`
	Bedrock bool `json:"bedrock"`
}

func getTopSentiment(sentimentScores *comprehend.SentimentScore) string {
	sentiments := map[string]float64{
		"Positive": *sentimentScores.Positive,
		"Negative": *sentimentScores.Negative,
		"Neutral":  *sentimentScores.Neutral,
		"Mixed":    *sentimentScores.Mixed,
	}

	var topSentiment string
	var maxScore float64

	for sentiment, score := range sentiments {
		if score > maxScore {
			maxScore = score
			topSentiment = sentiment
		}
	}

	switch topSentiment {
	case "Positive":
		topSentiment = "happy"
	case "Negative":
		topSentiment = "sad"
	case "Neutral":
		fallthrough
	case "Mixed":
		topSentiment = "neutral"
	}

	return topSentiment
}


func handler(ctx context.Context, event InputEvent) (OutputEvent, error) {
	text := event.Detail.Text
	svc := comprehend.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1"),
	})))

	sentimentParams := &comprehend.DetectSentimentInput{
		Text:         aws.String(text),
		LanguageCode: aws.String("en"),
	}

	sentimentResult, err := svc.DetectSentiment(sentimentParams)
	if err != nil {
		return OutputEvent{}, fmt.Errorf("failed to detect sentiment: %w", err)
	}

	topSentiment := getTopSentiment(sentimentResult.SentimentScore)

	// Limit the text to 400 characters to be sure not to hit the roof on the bedrock side. temp fix, next step is to use a summary function to get the most important parts of the text
	if(len(text) > 400) {
		text = text[:500]
	}

	prompt := fmt.Sprintf("Generate a %s image based on the following text: %s", topSentiment, text)



	outputEvent := OutputEvent{Prompt: prompt, S3Key: event.Detail.S3Key, Bedrock: event.Detail.Bedrock}

	return outputEvent, nil
}

func main() {
	lambda.Start(handler)
}
