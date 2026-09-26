package jev

import (
	"encoding/json/v2"
)

type Question interface {
	GetType() string
	json.Marshaler
}

// --- Noul Question ---
type NoulQuestion struct {
	Instructions any `json:"instructions"`
	Criteria *struct{
		True string `json:"true,omitempty"`
		False string `json:"false,omitempty"`
	}  `json:"criteria,omitempty"`
}

func (q NoulQuestion) GetType() string { return "noul" }

func (q NoulQuestion) MarshalJSON() ([]byte, error) {
	type Alias NoulQuestion
	return json.Marshal(&struct {
		Type string `json:"type"`
		Alias
	}{
		Type: q.GetType(),
		Alias: (Alias)(q),
	})
}

// --- Choice Question ---
type ChoiceQuestion struct {
	Instructions any `json:"instructions"`
	Criteria map[string]any `json:"criteria"`
}

func (q ChoiceQuestion) GetType() string { return "choice" }

func (q ChoiceQuestion) MarshalJSON() ([]byte, error) {
	type Alias ChoiceQuestion
	return json.Marshal(&struct {
		Type string `json:"type"`
		Alias
	}{
		Type: q.GetType(),
		Alias: (Alias)(q),
	})
}

// --- Score Question ---
type ScoreQuestion struct {
	Instructions any `json:"instructions"`	
	Criteria []any `json:"criteria"`
}

func (q ScoreQuestion) GetType() string { return "score" }

func (q ScoreQuestion) MarshalJSON() ([]byte, error) {
	type Alias ScoreQuestion
	return json.Marshal(&struct {
		Type string `json:"type"`
		Alias
	}{
		Type: q.GetType(),
		Alias: (Alias)(q),
	})
}

// --- Request ---
type Request struct {
	State any `json:"state"`
	Model string `json:"model"`
	Questions map[string]Question `json:"questions"`
}

// --- Response ---
type Response struct {
	Model string `json:"model"`
	Answers map[string]Answer `json:"answers"`
	Usage struct{
		InputTokens int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type Answer struct {
	Type string `json:"type"`
	Noul          *float64            `json:"noul,omitempty"`
	Choice        *string             `json:"choice,omitempty"`
	Score         *float64            `json:"score,omitempty"`
	Legend        map[string]string  `json:"legend,omitempty"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
	Confidence    *float64            `json:"confidence,omitempty"`
}

