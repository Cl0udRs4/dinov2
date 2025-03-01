package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ClientHandler handles client API endpoints
type ClientHandler struct {
	*Handler
	clientManager interface{} // Replace with actual client manager type
}

// NewClientHandler creates a new client handler
func NewClientHandler(base *Handler, clientManager interface{}) *ClientHandler {
	return &ClientHandler{
		Handler:      base,
		clientManager: clientManager,
	}
}

// RegisterRoutes registers the client API routes
func (h *ClientHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/clients", h.AuthMiddleware(h.handleClients))
	mux.HandleFunc("/api/v1/clients/", h.AuthMiddleware(h.handleClient))
}

// handleClients handles GET requests to /api/v1/clients
func (h *ClientHandler) handleClients(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getClients(w, r)
	default:
		h.RespondError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
	}
}

// handleClient handles GET and DELETE requests to /api/v1/clients/{id}
func (h *ClientHandler) handleClient(w http.ResponseWriter, r *http.Request) {
	// Extract the client ID from the URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		h.RespondError(w, http.StatusBadRequest, fmt.Errorf("invalid URL"))
		return
	}
	id := parts[4]

	// Check if this is a task endpoint
	if len(parts) >= 6 && parts[5] == "tasks" {
		h.handleClientTasks(w, r, id)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getClient(w, r, id)
	case http.MethodDelete:
		h.disconnectClient(w, r, id)
	default:
		h.RespondError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
	}
}

// handleClientTasks handles POST requests to /api/v1/clients/{id}/tasks
func (h *ClientHandler) handleClientTasks(w http.ResponseWriter, r *http.Request, id string) {
	switch r.Method {
	case http.MethodPost:
		h.sendTask(w, r, id)
	default:
		h.RespondError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
	}
}

// getClients returns a list of all connected clients
func (h *ClientHandler) getClients(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would get the clients from the client manager
	clients := []map[string]interface{}{
		{
			"id":        "client1",
			"ip":        "192.168.1.100",
			"hostname":  "workstation1",
			"os":        "Windows 10",
			"connected": true,
		},
	}

	h.RespondJSON(w, http.StatusOK, clients, "Clients retrieved successfully")
}

// getClient returns a specific client
func (h *ClientHandler) getClient(w http.ResponseWriter, r *http.Request, id string) {
	// In a real implementation, this would get the client from the client manager
	client := map[string]interface{}{
		"id":        id,
		"ip":        "192.168.1.100",
		"hostname":  "workstation1",
		"os":        "Windows 10",
		"connected": true,
	}

	h.RespondJSON(w, http.StatusOK, client, "Client retrieved successfully")
}

// disconnectClient disconnects a specific client
func (h *ClientHandler) disconnectClient(w http.ResponseWriter, r *http.Request, id string) {
	// In a real implementation, this would disconnect the client
	h.RespondJSON(w, http.StatusOK, nil, "Client disconnected successfully")
}

// sendTask sends a task to a specific client
func (h *ClientHandler) sendTask(w http.ResponseWriter, r *http.Request, id string) {
	var task map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// In a real implementation, this would send the task to the client
	h.RespondJSON(w, http.StatusOK, nil, "Task sent successfully")
}
