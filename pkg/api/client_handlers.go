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
	
	// Get clients from client manager
	clients := r.clientManager.ListClients()
	
	// Convert to client info for response
	clientInfos := make([]map[string]interface{}, 0, len(clients))
	for _, client := range clients {
		clientInfos = append(clientInfos, map[string]interface{}{
			"id":                 client.GetSessionID(),
			"protocol":           client.GetCurrentProtocol(),
			"state":              getStateString(int(client.GetState())),
			"encryption_algorithm": client.GetEncryptionAlgorithm(),
			"last_heartbeat":     client.GetLastHeartbeat().Format("2006-01-02 15:04:05"),
		})
	}
	
	writeJSON(w, map[string]interface{}{
		"status":  "success",
		"clients": clientInfos,
	}, http.StatusOK)
}

// getStateString converts a ConnectionState to a string
func getStateString(state int) string {
	switch state {
	case 0:
		return "disconnected"
	case 1:
		return "connecting"
	case 2:
		return "connected"
	case 3:
		return "reconnecting"
	case 4:
		return "switching_protocol"
	default:
		return "unknown"
	}
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
