package jev_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/danninx/go-jev"
)

func TestNewAPIError_Sentinels(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectSentinel error
	}{
		{"401 Unauthorized", 401, jev.ErrUnauthorized},
		{"422 Unprocessable Entity", 422, jev.ErrUnprocessableEntity},
		{"429 Rate Limited", 429, jev.ErrRateLimited},
		{"529 Overloaded", 529, jev.ErrOverloaded},
		{"500 Internal Server Error (No Sentinel)", 500, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := jev.NewAPIError(tt.statusCode, "test message", nil)

			if tt.expectSentinel != nil {
				if !errors.Is(apiErr, tt.expectSentinel) {
					t.Errorf("expected errors.Is(err, %v) to be true", tt.expectSentinel)
				}
			} else {
				if apiErr.Unwrap() != nil {
					t.Errorf("expected nil wrapped error for status %d, got %v", tt.statusCode, apiErr.Unwrap())
				}
			}
		})
	}
}

func TestAPIError_ErrorsAs(t *testing.T) {
	detail := map[string]any{"field": "questions.is_urgent"}
	err := error(jev.NewAPIError(422, "validation failed", detail))

	var apiErr *jev.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected errors.As to target *jev.APIError")
	}

	if apiErr.StatusCode != 422 {
		t.Errorf("got status %d, want 422", apiErr.StatusCode)
	}

	if apiErr.Message != "validation failed" {
		t.Errorf("got message %q, want 'validation failed'", apiErr.Message)
	}
}

func TestClient_HTTPErrorResponses(t *testing.T) {
	// Mock HTTP server returning API errors
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message": "rate limit exceeded"}`))
		}),
	)
	defer server.Close()

	client, err := jev.NewClient("test-key", server.URL, "typesafe/jev-1.13")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.Evaluate(context.Background(), &jev.Request{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Verify Sentinel
	if !errors.Is(err, jev.ErrRateLimited) {
		t.Errorf("expected error to match ErrRateLimited, got %v", err)
	}

	// Verify Concrete Error Details
	var apiErr *jev.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected errors.As to extract *jev.APIError")
	}
	if apiErr.StatusCode != 429 {
		t.Errorf("got status code %d, want 429", apiErr.StatusCode)
	}
}

func TestClient_NonApplicationErrors(t *testing.T) {
	// Point client to an invalid unreachable address
	client, err := jev.NewClient("test-key", "http://127.0.0.1:1", "typesafe/jev-1.13")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.Evaluate(context.Background(), &jev.Request{})
	if err == nil {
		t.Fatal("expected network error, got nil")
	}

	// 1. Should NOT be wrapped in an *APIError
	var apiErr *jev.APIError
	if errors.As(err, &apiErr) {
		t.Errorf("network failure should not return an *APIError, got %+v", apiErr)
	}

	// 2. Should unwrappable as underlying network error
	var netErr net.Error
	if !errors.As(err, &netErr) {
		t.Errorf("expected error to be a net.Error, got %T: %v", err, err)
	}
}
