package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RegisterClientRoutes registers client-related routes
func (r *Router) RegisterClientRoutes() {
	// Register client routes
	r.routes["/api/clients"] = r.handleListClients
	r.routes["/api/clients/tasks"] = r.handleClientTasks
}

// handleListClients handles the request to list all clients
func (r *Router) handleListClients(w http.ResponseWriter, req *http.Request) {
	// Only allow GET requests
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if client manager is available
	if r.clientManager == nil {
		http.Error(w, "Client manager not available", http.StatusInternalServerError)
		return
	}

	fmt.Printf("DEBUG: handleListClients called, client manager type: %T\n", r.clientManager)

	// Get clients from client manager
	clients := r.clientManager.ListClients()
	fmt.Printf("DEBUG: Got %d clients from client manager\n", len(clients))

	// Convert clients to JSON-friendly format
	clientsJSON := make([]map[string]interface{}, 0, len(clients))
	for _, client := range clients {
		clientJSON := map[string]interface{}{
			"id":        string(client.SessionID),
			"address":   client.Address,
			"protocol":  client.Protocol,
			"last_seen": client.LastSeen.Format(time.RFC3339),
		}
		
		// Add additional info if available
		if client.Info != nil {
			for k, v := range client.Info {
				clientJSON[k] = v
			}
		}
		
		clientsJSON = append(clientsJSON, clientJSON)
	}

	// Return clients as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clients": clientsJSON,
		"status": "success",
	})
}

// handleClientTasks handles the request to get tasks for a client
func (r *Router) handleClientTasks(w http.ResponseWriter, req *http.Request) {
	// Only allow GET requests
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get client ID from query parameters
	clientID := req.URL.Query().Get("id")
	if clientID == "" {
		http.Error(w, "Client ID is required", http.StatusBadRequest)
		return
	}

	// Check if task manager is available
	if r.taskManager == nil {
		http.Error(w, "Task manager not available", http.StatusInternalServerError)
		return
	}

	// Get tasks for client
	tasks, err := r.taskManager.GetTasksForClient(clientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return tasks as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tasks": tasks,
		"status": "success",
	})
}
