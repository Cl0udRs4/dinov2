package listener

import (
	"encoding/json"
	"net/http"
	"strings"
)

// SimpleAPIHandler is a simple HTTP handler for API requests
type SimpleAPIHandler struct {
	// Routes maps API paths to handler functions
	Routes map[string]http.HandlerFunc
}

// NewSimpleAPIHandler creates a new API handler
func NewSimpleAPIHandler() *SimpleAPIHandler {
	handler := &SimpleAPIHandler{
		Routes: make(map[string]http.HandlerFunc),
	}
	
	// Register default routes
	handler.registerDefaultRoutes()
	
	return handler
}

// registerDefaultRoutes registers the default API routes
func (h *SimpleAPIHandler) registerDefaultRoutes() {
	// Listeners API
	h.Routes["/listeners"] = h.handleListeners
	h.Routes["/listeners/"] = h.handleListenerByID
	
	// Tasks API
	h.Routes["/tasks"] = h.handleTasks
	h.Routes["/tasks/"] = h.handleTaskByID
	
	// Modules API
	h.Routes["/modules"] = h.handleModules
	h.Routes["/modules/"] = h.handleModuleByID
	
	// Clients API
	h.Routes["/clients"] = h.handleClients
	h.Routes["/clients/tasks"] = h.handleClientTasks
	
	// Protocol API
	h.Routes["/protocol/switch"] = h.handleProtocolSwitch
}

// ServeHTTP implements the http.Handler interface
func (h *SimpleAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set content type to JSON for all API responses
	w.Header().Set("Content-Type", "application/json")
	
	// Extract the API path from the request URL
	path := r.URL.Path
	
	// Find the handler for the requested path
	var handler http.HandlerFunc
	var matchedPath string
	
	// Try to find an exact match first
	if handlerFunc, ok := h.Routes[path]; ok {
		handler = handlerFunc
		matchedPath = path
	} else {
		// Try to find a prefix match
		for routePath, handlerFunc := range h.Routes {
			if strings.HasPrefix(path, routePath) && (len(routePath) > len(matchedPath)) {
				handler = handlerFunc
				matchedPath = routePath
			}
		}
	}
	
	if handler != nil {
		handler(w, r)
		return
	}
	
	// If no handler is found, return a 404 error
	h.sendJSONError(w, http.StatusNotFound, "API endpoint not found")
}

// sendJSONError sends a JSON error response
func (h *SimpleAPIHandler) sendJSONError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "error",
		"message": message,
	})
}

// sendJSONResponse sends a JSON response
func (h *SimpleAPIHandler) sendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}

// handleListeners handles the /listeners endpoint
func (h *SimpleAPIHandler) handleListeners(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Return a list of listeners
		h.sendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"listeners": []map[string]interface{}{
				{
					"id":      "tcp1",
					"type":    "tcp",
					"address": "0.0.0.0",
					"port":    8080,
					"status":  "running",
				},
				{
					"id":      "http1",
					"type":    "http",
					"address": "0.0.0.0",
					"port":    8000,
					"status":  "running",
				},
			},
		})
	case http.MethodPost:
		// Create a new listener
		h.sendJSONResponse(w, http.StatusCreated, map[string]interface{}{
			"id":      "new_listener",
			"type":    "tcp",
			"address": "0.0.0.0",
			"port":    9090,
			"status":  "stopped",
		})
	default:
		h.sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleListenerByID handles the /listeners/{id} endpoint
func (h *SimpleAPIHandler) handleListenerByID(w http.ResponseWriter, r *http.Request) {
	// Extract the listener ID from the URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		h.sendJSONError(w, http.StatusBadRequest, "Invalid listener ID")
		return
	}
	
	listenerID := parts[2]
	
	switch r.Method {
	case http.MethodGet:
		// Return a specific listener
		h.sendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"id":      listenerID,
			"type":    "tcp",
			"address": "0.0.0.0",
			"port":    8080,
			"status":  "running",
		})
	case http.MethodPut:
		// Update a listener
		h.sendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"id":      listenerID,
			"type":    "tcp",
			"address": "127.0.0.1",
			"port":    9091,
			"status":  "running",
		})
	case http.MethodDelete:
		// Delete a listener
		h.sendJSONResponse(w, http.StatusOK, map[string]string{
			"message": "Listener deleted successfully",
		})
	case http.MethodPost:
		// Handle start/stop actions
		if strings.HasSuffix(r.URL.Path, "/start") {
			h.sendJSONResponse(w, http.StatusOK, map[string]string{
				"message": "Listener started successfully",
			})
		} else if strings.HasSuffix(r.URL.Path, "/stop") {
			h.sendJSONResponse(w, http.StatusOK, map[string]string{
				"message": "Listener stopped successfully",
			})
		} else {
			h.sendJSONError(w, http.StatusBadRequest, "Invalid action")
		}
	default:
		h.sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleTasks handles the /tasks endpoint
func (h *SimpleAPIHandler) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Return a list of tasks
		h.sendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"tasks": []map[string]interface{}{
				{
					"id":     1,
					"name":   "task1",
					"status": "running",
				},
				{
					"id":     2,
					"name":   "task2",
					"status": "completed",
				},
			},
		})
	case http.MethodPost:
		// Create a new task
		h.sendJSONResponse(w, http.StatusCreated, map[string]interface{}{
			"id":     3,
			"name":   "new_task",
			"status": "pending",
		})
	default:
		h.sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleTaskByID handles the /tasks/{id} endpoint
func (h *SimpleAPIHandler) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	// Extract the task ID from the URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		h.sendJSONError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}
	
	taskID := parts[2]
	
	switch r.Method {
	case http.MethodGet:
		// Return a specific task
		h.sendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"id":     taskID,
			"name":   "task" + taskID,
			"status": "running",
		})
	case http.MethodPut:
		// Update a task
		h.sendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"id":     taskID,
			"name":   "task" + taskID,
			"status": "completed",
		})
	case http.MethodPost:
		// Handle cancel action
		if strings.HasSuffix(r.URL.Path, "/cancel") {
			h.sendJSONResponse(w, http.StatusOK, map[string]string{
				"message": "Task cancelled successfully",
			})
		} else {
			h.sendJSONError(w, http.StatusBadRequest, "Invalid action")
		}
	default:
		h.sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleModules handles the /modules endpoint
func (h *SimpleAPIHandler) handleModules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Return a list of modules
		h.sendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"modules": []map[string]interface{}{
				{
					"name":   "module1",
					"status": "loaded",
				},
				{
					"name":   "module2",
					"status": "unloaded",
				},
			},
		})
	case http.MethodPost:
		// Load a new module
		h.sendJSONResponse(w, http.StatusCreated, map[string]interface{}{
			"name":   "new_module",
			"status": "loaded",
		})
	default:
		h.sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleModuleByID handles the /modules/{name} endpoint
func (h *SimpleAPIHandler) handleModuleByID(w http.ResponseWriter, r *http.Request) {
	// Extract the module name from the URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		h.sendJSONError(w, http.StatusBadRequest, "Invalid module name")
		return
	}
	
	moduleName := parts[2]
	
	switch r.Method {
	case http.MethodGet:
		// Return a specific module
		h.sendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"name":   moduleName,
			"status": "loaded",
		})
	case http.MethodDelete:
		// Unload a module
		h.sendJSONResponse(w, http.StatusOK, map[string]string{
			"message": "Module unloaded successfully",
		})
	case http.MethodPost:
		// Handle exec action
		if strings.HasSuffix(r.URL.Path, "/exec") {
			h.sendJSONResponse(w, http.StatusOK, map[string]interface{}{
				"result": "Command executed successfully",
			})
		} else {
			h.sendJSONError(w, http.StatusBadRequest, "Invalid action")
		}
	default:
		h.sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleClients handles the /clients endpoint
func (h *SimpleAPIHandler) handleClients(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Return a list of clients
		h.sendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"clients": []map[string]interface{}{
				{
					"id":        "client1",
					"ip":        "192.168.1.100",
					"protocol":  "tcp",
					"connected": true,
				},
				{
					"id":        "client2",
					"ip":        "192.168.1.101",
					"protocol":  "http",
					"connected": false,
				},
			},
		})
	default:
		h.sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleClientTasks handles the /clients/tasks endpoint
func (h *SimpleAPIHandler) handleClientTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Return tasks for a specific client
		clientID := r.URL.Query().Get("client_id")
		if clientID == "" {
			h.sendJSONError(w, http.StatusBadRequest, "Missing client_id parameter")
			return
		}
		
		h.sendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"client_id": clientID,
			"tasks": []map[string]interface{}{
				{
					"id":     1,
					"name":   "task1",
					"status": "running",
				},
				{
					"id":     2,
					"name":   "task2",
					"status": "completed",
				},
			},
		})
	default:
		h.sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleProtocolSwitch handles the /protocol/switch endpoint
func (h *SimpleAPIHandler) handleProtocolSwitch(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// Switch protocol for a client
		h.sendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"client_id": "client1",
			"protocol":  "http",
			"status":    "switched",
		})
	default:
		h.sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}
