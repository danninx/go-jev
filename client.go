package jev

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"time"
)

// --- Jev Client ---
type Jev struct {
	APIKey   string
	Endpoint string
	Client   *http.Client
}

func NewClient(apiKey string, endpoint string) (*Jev, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	j := &Jev{
		APIKey:   apiKey,
		Endpoint: endpoint,
		Client:   client,
	}

	return j, nil
}

func (j *Jev) Evaluate(ctx context.Context, request *Request) (*Response, error) {
	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, j.Endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+j.APIKey)

	resp, err := j.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errDetail struct {
			Message string `json:"message"`
			Error   any    `json:"error"`
		}
		_ = json.Unmarshal(respBody, &errDetail)

		msg := errDetail.Message
		if msg == "" {
			msg = string(respBody)
		}

		return nil, NewAPIError(resp.StatusCode, msg, errDetail.Error)
	}

	var evaluateResponse Response
	if err := json.Unmarshal(respBody, &evaluateResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &evaluateResponse, nil
}
