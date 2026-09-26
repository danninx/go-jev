# go-jev

A go package for interacting with the TypeSafe Jev API. Also compatible with OpenRouter Decision endpoints.

## Summary

We kept running into scam bots in a discord server I'm in, and we thought it would be a cool use case for using Jev for some level of auto-moderation

Main import path is a `jev` client package, initial test in a static input is under `cmd/jev/`

Input:
```json
{
    "state": "@everyone Hey guys! I'm offering out my CANON EOS 1500D, a 18-55mm, a 17-85mm, a 50mm f/1.8, and a Kogan Horizon camera for free. They're in original box, nearly brand new condition. I bought them Christmas 2025, used only for test shoots. I'm offering it out because I just got a drone and want someone who needs it to have it. If you're interested, Text me on +[PHONE]",
    "model": "jev-latest",
    "questions": {
        "is_scam": {
            "type": "noul",
            "instructions": "Is this message likely a scam?"
        },
        "scam_level": {
            "type": "score",
            "instructions": "Rate the likelihood that this message is a scam.",
            "criteria": [
                "Low",
                "Medium",
                "High"
            ]
        }
    }
}
```

Sample output (do note output *is* non-deterministic):
```
Tokens Used: Input=433, Output=37

Answers:
  - [is_scam] Type: noul
    Noul Probability: 0.7300

  - [scam_level] Type: score
    Weighted Score: 1.68
    Confidence: 0.52
    Probabilities:
      * 1 (Medium): 0.2300 (23.0%)
      * 2 (High): 0.7300 (73.0%)
      * 0 (Low): 0.0400 (4.0%)
```

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
				Criteria: &struct {
					True  string `json:"true,omitempty"`
					False string `json:"false,omitempty"`
				}{
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
