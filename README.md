# go-jev

A go package for interacting with the TypeSafe Jev API. Also compatible with OpenRouter Decision endpoints

## Summary

We kept running into scam bots in a discord server I'm in, and we thought it would be a cool use case for using Jev for some level of auto-moderation

Package `cmd/jev` contains a little script to run against messages using the Jev API

Sample inputs and outputs are in the `tests` folder


## Installation

```
go get github.com/danninx/go-jev
```

## Quickstart
Make a client using `jev.NewClient` to handle `jev.Requests`:

```go
apiKey := os.Getenv("OPENROUTER_API_KEY")
client, err := jev.NewClient(apiKey, "https://openrouter.ai/api/alpha/decisions", "typesafe/jev-1.13")
if err != nil {
	log.Fatalf("Failed to create client: %v", err)
}

req := &jev.Request{
		State: "The customer reported that their payout failed twice, but support hasn't responded in 48 hours.",
		Model: "jev-latest",
		Questions: map[string]jev.Question{
			"is_urgent": jev.NoulQuestion{
				Instructions: "Is this issue time-sensitive or severe?",
				Criteria: jev.NoulCriteria{
					True:  "Direct financial impact or explicit SLA breach",
					False: "General inquiry or low priority",
				},
			},
			"department": jev.ChoiceQuestion{
				Instructions: "Which team should handle this ticket?",
				Criteria: map[string]any{
					"billing": "Payment processing, payouts, or refund issues",
					"tech":    "System bugs or technical errors",
				},
			},
			"frustration_level": jev.ScoreQuestion{
				Instructions: "Rate customer frustration level based on context.",
				Criteria:     []any{"Low", "Medium", "High"},
			},
		},
	}

ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
defer cancel()

resp, err := client.Evaluate(ctx, req)
if err != nil {
	var apiErr *jev.APIError // package provides typed error handling for API errors
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

fmt.Println("Answers:")
for qID, answer := range resp.Answers {
	fmt.Printf("  - [%s] Type: %s\n", qID, answer.Type)
	if answer.Noul != nil {
		fmt.Printf("    Score: %.2f\n", *answer.Noul)
	}
	if answer.Choice != nil {
		fmt.Printf("    Selected: %s\n", *answer.Choice)
	}
	if answer.Score != nil {
		fmt.Printf("    Score: %.2f\n", *answer.Score)
	}
}
```
