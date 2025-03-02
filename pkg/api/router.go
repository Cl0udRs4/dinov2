package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	
	"dinoc2/pkg/api/middleware"
	clientmanager "dinoc2/pkg/client/manager"
	"dinoc2/pkg/listener"
	modulemanager "dinoc2/pkg/module/manager"
	"dinoc2/pkg/task"
)

// Router handles HTTP API routing
type Router struct {
	listenerManager *listener.Manager
	moduleManager   *modulemanager.ModuleManager
	taskManager     *task.Manager
	clientManager   *clientmanager.ClientManager
	routes          map[string]http.HandlerFunc
	authMiddleware  *middleware.AuthMiddleware
}

// NewRouter creates a new API router
func NewRouter(listenerManager *listener.Manager, moduleManager *modulemanager.ModuleManager, taskManager *task.Manager, clientManager *clientmanager.ClientManager, authMiddleware *middleware.AuthMiddleware) *Router {
	r := &Router{
		listenerManager: listenerManager,
		moduleManager:   moduleManager,
		taskManager:     taskManager,
		clientManager:   clientManager,
		routes:          make(map[string]http.HandlerFunc),
		authMiddleware:  authMiddleware,
	}
	
	// Register routes
	r.registerRoutes()
	
	// Register authentication routes if middleware is provided
	if authMiddleware != nil {
		r.RegisterAuthRoutes(authMiddleware)
	}
	
	return r
}

// NewRouterWithoutAuth creates a new API router without authentication
// This is for backward compatibility
func NewRouterWithoutAuth(listenerManager *listener.Manager, moduleManager *modulemanager.ModuleManager, taskManager *task.Manager, clientManager *clientmanager.ClientManager) *Router {
	return NewRouter(listenerManager, moduleManager, taskManager, clientManager, nil)
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
	
	// Documentation route
	r.routes["/api/docs"] = r.handleDocs
	
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
	
	// Skip authentication for login and refresh endpoints
	if req.URL.Path == "/api/auth/login" || req.URL.Path == "/api/auth/refresh" {
		// Find handler for the requested path
		for path, handler := range r.routes {
			if strings.HasPrefix(req.URL.Path, path) {
				handler(w, req)
				return
			}
		}
		
		// No handler found, return 404
		http.NotFound(w, req)
		return
	}
	
	// Apply authentication middleware if available
	if r.authMiddleware != nil {
		// Skip authentication for login, refresh, and docs endpoints
		if req.URL.Path == "/api/auth/login" || req.URL.Path == "/api/auth/refresh" || req.URL.Path == "/api/docs" {
			// Allow these endpoints without authentication
		} else {
			// Get the Authorization header
			authHeader := req.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, "Authorization header required", http.StatusUnauthorized)
				return
			}
			
			// Check if the Authorization header has the correct format
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeError(w, "Invalid authorization format", http.StatusUnauthorized)
				return
			}
			
			// Extract the token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			
			// Validate the token
			_, claims, err := r.authMiddleware.ValidateToken(tokenString)
			if err != nil {
				writeError(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
				return
			}
			
			// Add claims to the request context
			ctx := context.WithValue(req.Context(), "claims", claims)
			req = req.WithContext(ctx)
		}
	}
	
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
