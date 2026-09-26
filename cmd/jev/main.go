package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"gitlab.com/danninx/go-jev"
)

const OPENROUTER_ENDPOINT = "https://openrouter.ai/api/alpha/decisions"

func newTestRequest(state string) *jev.Request {
	return &jev.Request{
		State: state,
		Model: "typesafe/jev-1.13",
		Questions: map[string]jev.Question{
			"is_scam": jev.NoulQuestion{
				Instructions: "Is this message likely a scam?",
				Criteria: jev.NoulCriteria{
					True:  "Gives away thousands of dollars in high-value electronics, urges users to contact off-platform via SMS/phone, or uses unsolicited channel-wide mentions (@everyone).",
					False: "Legitimate item giveaway or local trade without off-platform redirect.",
				},
			},
		},
	}
}

func main() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: OPENROUTER_API_KEY environment variable is not set.")
	}

	client, err := jev.NewClient(apiKey, OPENROUTER_ENDPOINT)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	filePath := flag.String("file", "", "Path to file with test message")
	flag.Parse()

	if filePath == nil || *filePath == "" {
		fmt.Println("No file provided.")
		os.Exit(0)
	}

	content, err := os.ReadFile(*filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading file:", err)
		os.Exit(1)
	}

	fmt.Println("Sending evaluation request to TypeSafe Jev API...")
	req := newTestRequest(string(content))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := client.Evaluate(ctx, req)
	if err != nil {
		var apiErr *jev.APIError
		if errors.As(err, &apiErr) {
			fmt.Printf("\nAPI Error Returned (Status %d):\n", apiErr.StatusCode)
			fmt.Printf("Message: %s\n", apiErr.Message)
			if apiErr.Detail != nil {
				fmt.Printf("Detail: %+v\n", apiErr.Detail)
			}
			os.Exit(1)
		}

		log.Fatalf("Unexpected client error: %v", err)
	}

	fmt.Printf("\nEvaluated Model: %s\n", resp.Model)
	fmt.Printf("Tokens Used: Input=%d, Output=%d\n\n", resp.Usage.InputTokens, resp.Usage.OutputTokens)

	fmt.Println("Answers:")
	for qID, answer := range resp.Answers {
		fmt.Printf("  - [%s] Type: %s\n", qID, answer.Type)

		if answer.Noul != nil {
			fmt.Printf("    Noul Probability: %.4f\n", *answer.Noul)
		}
		if answer.Choice != nil {
			fmt.Printf("    Selected Choice: %s\n", *answer.Choice)
		}
		if answer.Score != nil {
			fmt.Printf("    Weighted Score: %.2f\n", *answer.Score)
		}
		if answer.Confidence != nil {
			fmt.Printf("    Confidence: %.2f\n", *answer.Confidence)
		}

		if len(answer.Probabilities) > 0 {
			fmt.Println("    Probabilities:")
			for option, prob := range answer.Probabilities {
				legendText := ""
				if text, ok := answer.Legend[option]; ok {
					legendText = fmt.Sprintf(" (%s)", text)
				}
				fmt.Printf("      * %s%s: %.4f (%.1f%%)\n", option, legendText, prob, prob*100)
			}
		}
		fmt.Println()
	}
}
