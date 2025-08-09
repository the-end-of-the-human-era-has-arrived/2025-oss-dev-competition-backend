package api

import (
	"encoding/json"
	"net/http"
)

type HTTPResponse struct {
	StatusCode int    `json:"statusCode"`
	StatusText string `json:"statusText"`
	Message    string `json:"message"`
}

func ResponseStatusCode(w http.ResponseWriter, statusCode int, msg string) error {
	httpResp := &HTTPResponse{
		StatusCode: statusCode,
		StatusText: http.StatusText(statusCode),
		Message:    msg,
	}

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(httpResp); err != nil {
		return NewError(http.StatusInternalServerError, WithError(err))
	}
	return nil
}

func ResponseJSON(w http.ResponseWriter, v any) error {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return NewError(http.StatusInternalServerError, WithError(err))
	}
	return nil
}
