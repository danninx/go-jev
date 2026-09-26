package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gitlab.com/danninx/go-jev"
)

const OPENROUTER_ENDPOINT = "https://openrouter.ai/api/alpha/decisions"

var TEST_REQUEST = jev.Request{
	State: `@everyone Hey guys! I'm offering out my CANON EOS 1500D, a 18-55mm, a 17-85mm, a 50mm f/1.8, and a Kogan Horizon camera for free. They're in original box, nearly brand new condition. I bought them Christmas 2025, used only for test shoots. I'm offering it out because I just got a drone and want someone who needs it to have it. If you're interested, Text me on +[PHONE]`,
	Model: "typesafe/jev-1.13",
	Questions: map[string]jev.Question{
		"is_scam": jev.NoulQuestion{
			Instructions: "Is this message likely a scam?",
		},
		"scam_level": jev.ScoreQuestion{
			Instructions: "Rate the likelihood that this message is a scam.",
			Criteria:     []any{"Low", "Medium", "High"},
		},
	},
}

var TEST_REQUEST_2 = jev.Request{
	State: `@everyone Giving away my Canon EOS R7 Mirrorless Camera (Body Only), Hybrid Camera, 32.5 Megapixel (APS-C) CMOS Sensor, 4K Video, for Sports, Action, Content Creators, Vlogging Camera, Black Comes with extra lens Dm  if interested`,
	Model: "typesafe/jev-1.13",
	Questions: map[string]jev.Question{
		"is_scam": jev.NoulQuestion{
			Instructions: "Is this message likely a scam?",
		},
		"scam_level": jev.ScoreQuestion{
			Instructions: "Rate the likelihood that this message is a scam.",
			Criteria:     []any{"Low", "Medium", "High"},
		},
	},
}

func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: jev [test1|test2|request]")
		os.Exit(1)
	}

	subcommand := args[1]
	if subcommand != "test1" && subcommand != "test2" && subcommand != "request" {
		fmt.Fprintln(os.Stderr, "usage: jev [test1|test2|request]")
		os.Exit(1)
	}

	switch subcommand {
	case "test1":
		testJevEvaluate(TEST_REQUEST)
	case "test2":
		testJevEvaluate(TEST_REQUEST_2)
	case "request":
		getJevRequest()
	}
}

func getJevRequest() {
	b, err := json.MarshalIndent(TEST_REQUEST, "", "    ")
	if err != nil {
		fmt.Fprintln(os.Stdout, "error marshaling test request", err)
		os.Exit(1)
	}

	fmt.Println("test request:")

	fmt.Println(string(b))
}

func testJevEvaluate(req jev.Request) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: OPENROUTER_API_KEY environment variable is not set.")
	}

	client, err := jev.NewClient(apiKey, OPENROUTER_ENDPOINT)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	fmt.Println("Sending evaluation request to TypeSafe Jev API...")
	resp, err := client.Evaluate(ctx, &req)
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

	fmt.Printf("\nSuccess! Evaluated Model: %s\n", resp.Model)
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

