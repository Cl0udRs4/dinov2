package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// Response represents an API response
type Response struct {
	Success   bool        `json:"success"`
	Error     string      `json:"error,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(data interface{}) Response {
	return Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err string) Response {
	return Response{
		Success:   false,
		Error:     err,
		Timestamp: time.Now(),
	}
}

// WriteJSON writes a JSON response to the response writer
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Set status code
	w.WriteHeader(statusCode)

	// Write response
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
