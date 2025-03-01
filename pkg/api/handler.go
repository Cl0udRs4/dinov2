package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Time    time.Time   `json:"time"`
}

// Handler is the base handler for all API endpoints
type Handler struct {
	authenticator interface{} // Will be replaced with actual authenticator type
}

// NewHandler creates a new API handler
func NewHandler(authenticator interface{}) *Handler {
	return &Handler{
		authenticator: authenticator,
	}
}

// RespondJSON sends a JSON response
func (h *Handler) RespondJSON(w http.ResponseWriter, statusCode int, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Success: statusCode >= 200 && statusCode < 300,
		Message: message,
		Data:    data,
		Time:    time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding response: %v\n", err)
	}
}

// RespondError sends an error response
func (h *Handler) RespondError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Success: false,
		Error:   err.Error(),
		Time:    time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding error response: %v\n", err)
	}
}

// AuthMiddleware is a middleware that checks for authentication
func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			h.RespondError(w, http.StatusUnauthorized, fmt.Errorf("authorization header is required"))
			return
		}

		// Validate the token
		// In a real implementation, this would validate the token against the authenticator
		// For now, we'll just check if it's not empty
		if authHeader == "" {
			h.RespondError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
			return
		}

		// Call the next handler
		next(w, r)
	}
}
