package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ListenerHandler handles listener API endpoints
type ListenerHandler struct {
	*Handler
	manager interface{} // Will be replaced with actual listener manager type
}

// NewListenerHandler creates a new listener handler
func NewListenerHandler(base *Handler, manager interface{}) *ListenerHandler {
	return &ListenerHandler{
		Handler: base,
		manager: manager,
	}
}

// RegisterRoutes registers the listener API routes
func (h *ListenerHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/listeners", h.AuthMiddleware(h.handleListeners))
	mux.HandleFunc("/api/v1/listeners/", h.AuthMiddleware(h.handleListener))
}

// handleListeners handles GET and POST requests to /api/v1/listeners
func (h *ListenerHandler) handleListeners(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getListeners(w, r)
	case http.MethodPost:
		h.createListener(w, r)
	default:
		h.RespondError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
	}
}

// handleListener handles GET, PUT, and DELETE requests to /api/v1/listeners/{id}
func (h *ListenerHandler) handleListener(w http.ResponseWriter, r *http.Request) {
	// Extract the listener ID from the URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		h.RespondError(w, http.StatusBadRequest, fmt.Errorf("invalid URL"))
		return
	}
	id := parts[4]

	switch r.Method {
	case http.MethodGet:
		h.getListener(w, r, id)
	case http.MethodPut:
		h.updateListener(w, r, id)
	case http.MethodDelete:
		h.deleteListener(w, r, id)
	default:
		h.RespondError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
	}
}

// getListeners returns a list of all listeners
func (h *ListenerHandler) getListeners(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would get the listeners from the listener manager
	listeners := []map[string]interface{}{
		{
			"id":      "listener1",
			"type":    "tcp",
			"address": "0.0.0.0",
			"port":    8080,
			"status":  "running",
		},
	}

	h.RespondJSON(w, http.StatusOK, listeners, "Listeners retrieved successfully")
}

// createListener creates a new listener
func (h *ListenerHandler) createListener(w http.ResponseWriter, r *http.Request) {
	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// In a real implementation, this would create a listener using the listener manager
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": "listener1"}, "Listener created successfully")
}

// getListener returns a specific listener
func (h *ListenerHandler) getListener(w http.ResponseWriter, r *http.Request, id string) {
	// In a real implementation, this would get the listener from the listener manager
	listener := map[string]interface{}{
		"id":      id,
		"type":    "tcp",
		"address": "0.0.0.0",
		"port":    8080,
		"status":  "running",
	}

	h.RespondJSON(w, http.StatusOK, listener, "Listener retrieved successfully")
}

// updateListener updates a specific listener
func (h *ListenerHandler) updateListener(w http.ResponseWriter, r *http.Request, id string) {
	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// In a real implementation, this would update the listener using the listener manager
	h.RespondJSON(w, http.StatusOK, nil, "Listener updated successfully")
}

// deleteListener deletes a specific listener
func (h *ListenerHandler) deleteListener(w http.ResponseWriter, r *http.Request, id string) {
	// In a real implementation, this would delete the listener using the listener manager
	h.RespondJSON(w, http.StatusOK, nil, "Listener deleted successfully")
}
