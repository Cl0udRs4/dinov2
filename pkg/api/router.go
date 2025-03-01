package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	
	"dinoc2/pkg/listener"
	"dinoc2/pkg/module/manager"
	"dinoc2/pkg/task"
)

// Router handles HTTP API routing
type Router struct {
	listenerManager *listener.Manager
	moduleManager   *manager.Manager
	taskManager     *task.Manager
	routes          map[string]http.HandlerFunc
}

// NewRouter creates a new API router
func NewRouter(listenerManager *listener.Manager, moduleManager *manager.Manager, taskManager *task.Manager) *Router {
	r := &Router{
		listenerManager: listenerManager,
		moduleManager:   moduleManager,
		taskManager:     taskManager,
		routes:          make(map[string]http.HandlerFunc),
	}
	
	// Register routes
	r.registerRoutes()
	
	return r
}

// registerRoutes registers all API routes
func (r *Router) registerRoutes() {
	// Listener routes
	r.routes["/api/listeners"] = r.handleListListeners
	r.routes["/api/listeners/create"] = r.handleCreateListener
	r.routes["/api/listeners/delete"] = r.handleDeleteListener
	r.routes["/api/listeners/status"] = r.handleListenerStatus
	
	// Task routes
	r.routes["/api/tasks"] = r.handleListTasks
	r.routes["/api/tasks/create"] = r.handleCreateTask
	r.routes["/api/tasks/status"] = r.handleTaskStatus
	
	// Module routes
	r.routes["/api/modules"] = r.handleListModules
	r.routes["/api/modules/load"] = r.handleLoadModule
	r.routes["/api/modules/exec"] = r.handleExecModule
	
	// Client routes
	r.routes["/api/clients"] = r.handleListClients
	r.routes["/api/clients/tasks"] = r.handleClientTasks
	
	// Protocol switching routes
	r.routes["/api/protocol/switch"] = r.handleProtocolSwitch
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Set common headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Server", "Microsoft-IIS/10.0")
	
	// Find handler for the requested path
	for path, handler := range r.routes {
		if strings.HasPrefix(req.URL.Path, path) {
			handler(w, req)
			return
		}
	}
	
	// No handler found, return 404
	http.NotFound(w, req)
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func writeError(w http.ResponseWriter, message string, statusCode int) {
	writeJSON(w, map[string]string{"error": message}, statusCode)
}
