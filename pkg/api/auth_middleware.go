package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

// AuthMiddleware is a middleware that checks for authentication
type AuthMiddleware struct {
	authenticator interface{} // Will be replaced with actual authenticator type
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authenticator interface{}) *AuthMiddleware {
	return &AuthMiddleware{
		authenticator: authenticator,
	}
}

// Middleware wraps a handler with authentication middleware
func (m *AuthMiddleware) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No authorization header, return 401
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Response{
				Success: false,
				Error:   "Unauthorized",
			})
			return
		}

		// Check if the authorization header is valid
		// Format: Bearer <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid authorization header, return 401
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Response{
				Success: false,
				Error:   "Invalid authorization header",
			})
			return
		}

		// Get the token
		token := parts[1]

		// Validate the token
		// In a real implementation, this would validate the token with the authenticator
		// For now, we'll just check if the token is not empty
		if token == "" {
			// Invalid token, return 401
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Response{
				Success: false,
				Error:   "Invalid token",
			})
			return
		}

		// Token is valid, call the next handler
		next(w, r)
	}
}
