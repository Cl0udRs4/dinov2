package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ProtocolSwitchRequest represents a request to switch protocols
type ProtocolSwitchRequest struct {
	ClientID string `json:"client_id"`
	Protocol string `json:"protocol"`
}

// handleProtocolSwitch handles POST /api/protocol/switch
func (r *Router) handleProtocolSwitch(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var switchReq ProtocolSwitchRequest
	if err := json.NewDecoder(req.Body).Decode(&switchReq); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Get the client from the client manager
	client, err := r.clientManager.GetClient(switchReq.ClientID)
	if err != nil {
		writeError(w, fmt.Sprintf("Client not found: %v", err), http.StatusNotFound)
		return
	}
	
	// Send protocol switch command to the client
	err = client.SwitchProtocol(switchReq.Protocol)
	if err != nil {
		writeError(w, fmt.Sprintf("Failed to switch protocol: %v", err), http.StatusInternalServerError)
		return
	}
	
	writeJSON(w, map[string]string{
		"status": "success",
		"message": "Protocol switch initiated",
	}, http.StatusOK)
}
