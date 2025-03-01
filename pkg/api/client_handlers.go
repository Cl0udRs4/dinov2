package api

import (
	"net/http"
)

// handleListClients handles GET /api/clients
func (r *Router) handleListClients(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// This would typically come from a client manager
	// For now, we'll just return an empty list
	clients := []interface{}{}
	
	writeJSON(w, clients, http.StatusOK)
}

// handleClientTasks handles GET /api/clients/tasks
func (r *Router) handleClientTasks(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	clientID := req.URL.Query().Get("client_id")
	if clientID == "" {
		writeError(w, "Client ID is required", http.StatusBadRequest)
		return
	}
	
	tasks := r.taskManager.ListClientTasks(clientID)
	writeJSON(w, tasks, http.StatusOK)
}
