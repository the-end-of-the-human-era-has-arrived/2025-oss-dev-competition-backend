package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type HTTPResponse struct {
	StatusCode int    `json:"statusCode"`
	StatusText string `json:"statusText"`
	Message    string `json:"message"`
}

func ResponseStatusCode(ctx context.Context, w http.ResponseWriter, statusCode int, msg string) error {
	httpResp := &HTTPResponse{
		StatusCode: statusCode,
		StatusText: http.StatusText(statusCode),
		Message:    msg,
	}

	requestID := ctx.Value(RequestIDKey{}).(uuid.UUID)
	log.Printf("requestID: %s, result: %+v", requestID.String(), *httpResp)
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(httpResp); err != nil {
		return NewError(http.StatusInternalServerError, WithError(err))
	}
	return nil
}

func ResponseJSON(ctx context.Context, w http.ResponseWriter, v any) error {
	requestID := ctx.Value(RequestIDKey{}).(uuid.UUID)
	log.Printf("requestID: %s, result: %+v", requestID.String(), v)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return NewError(http.StatusInternalServerError, WithError(err))
	}
	return nil
}
