package api

import (
	"encoding/json"
	"net/http"
	"time"
)

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

// CreateSuccessResponse creates a new success response
func CreateSuccessResponse(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
		Time:    time.Now(),
	}
}

// CreateErrorResponse creates a new error response
func CreateErrorResponse(err string) Response {
	return Response{
		Success: false,
		Error:   err,
		Time:    time.Now(),
	}
}
