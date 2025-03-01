package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"dinoc2/pkg/listener"
)

// ListenerRequest represents a request to create a listener
type ListenerRequest struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Address string                 `json:"address"`
	Port    int                    `json:"port"`
	Options map[string]interface{} `json:"options"`
}

// handleListListeners handles GET /api/listeners
func (r *Router) handleListListeners(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	listeners := r.listenerManager.ListListeners()
	writeJSON(w, listeners, http.StatusOK)
}

// handleCreateListener handles POST /api/listeners/create
func (r *Router) handleCreateListener(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var listenerReq ListenerRequest
	if err := json.NewDecoder(req.Body).Decode(&listenerReq); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Create listener config
	config := listener.ListenerConfig{
		ID:      listenerReq.ID,
		Type:    listenerReq.Type,
		Address: listenerReq.Address,
		Port:    listenerReq.Port,
		Options: listenerReq.Options,
	}
	
	// Create and start listener
	err := r.listenerManager.CreateListener(config)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	writeJSON(w, map[string]string{"status": "success", "message": "Listener created"}, http.StatusOK)
}

// handleDeleteListener handles POST /api/listeners/delete
func (r *Router) handleDeleteListener(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var request struct {
		ID string `json:"id"`
	}
	
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Delete listener
	err := r.listenerManager.RemoveListener(request.ID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	writeJSON(w, map[string]string{"status": "success", "message": "Listener deleted"}, http.StatusOK)
}

// handleListenerStatus handles GET /api/listeners/status
func (r *Router) handleListenerStatus(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	id := req.URL.Query().Get("id")
	if id == "" {
		writeError(w, "Listener ID is required", http.StatusBadRequest)
		return
	}
	
	status, err := r.listenerManager.GetListenerStatus(id)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	writeJSON(w, status, http.StatusOK)
}
