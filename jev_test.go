package jev_test

import (
	"encoding/json/v2"
	"reflect"
	"testing"

	"gitlab.com/danninx/go-jev"
)

// --- test helpers ---
func assertMarshal(t *testing.T, expected string, v any) {
	t.Helper()
	bytes, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var actualMap, expectedMap map[string]any
	if err := json.Unmarshal(bytes, &actualMap); err != nil {
		t.Fatalf("failed to parse actual JSON output: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &expectedMap); err != nil {
		t.Fatalf("invalid expected JSON string provided to test: %v", err)
	}

	if !reflect.DeepEqual(actualMap, expectedMap) {
		t.Errorf("\ngot:  %s\nwant: %s", string(bytes), expected)
	}
}

func assertUnmarshal[T any](t *testing.T, rawJSON string, expected T) {
	t.Helper()
	var actual T
	if err := json.Unmarshal([]byte(rawJSON), &actual); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("\ngot:  %+v\nwant: %+v", actual, expected)
	}
}

// --- test question marshalling ---
func TestNoulQuestion_MarshalJSON(t *testing.T) {
	expect := `{"type":"noul","instructions":"Is this urgent?"}`
	q := jev.NoulQuestion{
		Instructions: "Is this urgent?",
	}
	assertMarshal(t, expect, q)

	expect = `{"type":"noul","instructions":"Is this urgent?","criteria":{"true":"Explicit urgency","false":"No urgency"}}`
	q = jev.NoulQuestion{
		Instructions: "Is this urgent?",
		Criteria: &struct {
			True  string `json:"true,omitempty"`
			False string `json:"false,omitempty"`
		}{
			True:  "Explicit urgency",
			False: "No urgency",
		},
	}
	assertMarshal(t, expect, q)
}

func TestChoiceQuestion_MarshalJSON(t *testing.T) {
	q := jev.ChoiceQuestion{
		Instructions: "Which department should handle this?",
		Criteria: map[string]any{
			"billing": "Payments, invoicing, refunds",
			"sales":   "Pricing, upgrades",
		},
	}
	expect := `{"type":"choice","instructions":"Which department should handle this?","criteria":{"billing":"Payments, invoicing, refunds","sales":"Pricing, upgrades"}}`
	assertMarshal(t, expect, q)

	q = jev.ChoiceQuestion{
		Instructions: map[string]string{
			"question": "Is the resume a match?",
			"context":  "Refer to candidate profile",
		},
		Criteria: map[string]any{"yes": "Matches", "no": "No match"},
	}
	expect = `{"type":"choice","instructions":{"context":"Refer to candidate profile","question":"Is the resume a match?"},"criteria":{"no":"No match","yes":"Matches"}}`
	assertMarshal(t, expect, q)
}

func TestScoreQuestion_MarshalJSON(t *testing.T) {
	q := jev.ScoreQuestion{
		Instructions: "How frustrated is the customer?",
		Criteria:     []any{"Calm", "Frustrated", "Very angry"},
	}
	assertMarshal(t, `{"type":"score","instructions":"How frustrated is the customer?","criteria":["Calm","Frustrated","Very angry"]}`, q)
}

// --- test request/response json-ing ---
func TestRequest_MarshalJSON(t *testing.T) {
	req := jev.Request{
		State: "My payouts have been failing for 3 days.",
		Model: "jev-latest",
		Questions: map[string]jev.Question{
			"is_urgent": jev.NoulQuestion{
				Instructions: "Does this convey urgency?",
			},
		},
	}
	assertMarshal(t, `{"state":"My payouts have been failing for 3 days.","model":"jev-latest","questions":{"is_urgent":{"type":"noul","instructions":"Does this convey urgency?"}}}`, req)
}

func TestResponse_UnmarshalJSON(t *testing.T) {
	rawResponse := `{
		"model": "jev-1.13.0",
		"answers": {
			"is_urgent": {
				"type": "noul",
				"noul": 0.95
			}
		},
		"usage": {
			"input_tokens": 296,
			"output_tokens": 20
		}
	}`

	noulVal := 0.95
	expected := jev.Response{
		Model: "jev-1.13.0",
		Answers: map[string]jev.Answer{
			"is_urgent": {
				Type: "noul",
				Noul: &noulVal,
			},
		},
		Usage: struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		}{
			InputTokens:  296,
			OutputTokens: 20,
		},
	}

	assertUnmarshal(t, rawResponse, expected)
}
