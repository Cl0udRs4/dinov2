package http

import (
	"encoding/json"
	"net/http"
	"strings"
)

// APIHandler is a simple HTTP handler for API requests
type APIHandler struct {
	// Routes maps API paths to handler functions
	Routes map[string]http.HandlerFunc
}

// NewAPIHandler creates a new API handler
func NewAPIHandler() *APIHandler {
	return &APIHandler{
		Routes: make(map[string]http.HandlerFunc),
	}
}

// RegisterRoute registers a new route with the API handler
func (h *APIHandler) RegisterRoute(path string, handler http.HandlerFunc) {
	h.Routes[path] = handler
}

// ServeHTTP implements the http.Handler interface
func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set content type to JSON for all API responses
	w.Header().Set("Content-Type", "application/json")
	
	// Extract the API path from the request URL
	path := strings.TrimPrefix(r.URL.Path, "/api")
	
	// Find the handler for the requested path
	for routePath, handler := range h.Routes {
		if strings.HasPrefix(path, routePath) {
			handler(w, r)
			return
		}
	}
	
	// If no handler is found, return a 404 error
	h.sendJSONError(w, http.StatusNotFound, "API endpoint not found")
}

// sendJSONError sends a JSON error response
func (h *APIHandler) sendJSONError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "error",
		"message": message,
	})
}

// sendJSONResponse sends a JSON response
func (h *APIHandler) sendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}
