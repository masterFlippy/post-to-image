package main

import (
	"context"
	"fmt"
	"log"
	"strings"

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

	keyPhraseParams := &comprehend.DetectKeyPhrasesInput{
		Text:         aws.String(text),
		LanguageCode: aws.String("en"),
	}

	sentimentResult, err := svc.DetectSentiment(sentimentParams)
	if err != nil {
		return OutputEvent{}, fmt.Errorf("failed to detect sentiment: %w", err)
	}

	keyPhraseResult, err := svc.DetectKeyPhrases(keyPhraseParams)
	if err != nil {
		return OutputEvent{}, fmt.Errorf("failed to process text: %w", err)
	}

	var keyPhrases []string
	topSentiment := getTopSentiment(sentimentResult.SentimentScore)
	keyPhrases = append(keyPhrases, topSentiment)
	for _, phrase := range keyPhraseResult.KeyPhrases {
		keyPhrases = append(keyPhrases, *phrase.Text)
	}

	summary := strings.Join(keyPhrases, " ")

	prompt := fmt.Sprintf("Generate a image based on the following key words: %s",  summary)
	log.Printf("Prompt: %s", prompt)
	outputEvent := OutputEvent{Prompt: prompt, S3Key: event.Detail.S3Key, Bedrock: event.Detail.Bedrock}

	return outputEvent, nil
}

func main() {
	lambda.Start(handler)
}
