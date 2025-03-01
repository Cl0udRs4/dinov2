package api

import (
	"net/http"
)

// handleDocs serves the API documentation
func (r *Router) handleDocs(w http.ResponseWriter, req *http.Request) {
	// For now, just return a simple JSON response with API information
	// In a real implementation, this would serve Swagger UI or similar
	apiInfo := map[string]interface{}{
		"name":        "DinoC2 API",
		"version":     "1.0.0",
		"description": "API for the DinoC2 Command and Control framework",
		"auth": map[string]interface{}{
			"type": "JWT",
			"endpoints": []map[string]string{
				{"path": "/api/auth/login", "method": "POST", "description": "Login to get a JWT token"},
				{"path": "/api/auth/refresh", "method": "POST", "description": "Refresh a JWT token"},
			},
		},
		"endpoints": []map[string]interface{}{
			{
				"path": "/api/listeners", 
				"method": "GET", 
				"description": "List all listeners",
				"auth_required": true,
				"params": []interface{}{},
				"response": "Array of listener objects",
			},
			{
				"path": "/api/listeners/create", 
				"method": "POST", 
				"description": "Create a new listener",
				"auth_required": true,
				"params": []string{"id", "type", "address", "port", "options"},
				"response": "Success or error message",
			},
			{
				"path": "/api/listeners/delete", 
				"method": "DELETE", 
				"description": "Delete a listener",
				"auth_required": true,
				"params": []string{"id"},
				"response": "Success or error message",
			},
			{
				"path": "/api/listeners/status", 
				"method": "GET", 
				"description": "Get listener status",
				"auth_required": true,
				"params": []string{"id"},
				"response": "Listener status object",
			},
			{
				"path": "/api/tasks", 
				"method": "GET", 
				"description": "List all tasks",
				"auth_required": true,
				"params": []interface{}{},
				"response": "Array of task objects",
			},
			{
				"path": "/api/tasks/create", 
				"method": "POST", 
				"description": "Create a new task",
				"auth_required": true,
				"params": []string{"client_id", "type", "data"},
				"response": "Task ID and status",
			},
			{
				"path": "/api/tasks/status", 
				"method": "GET", 
				"description": "Get task status",
				"auth_required": true,
				"params": []string{"id"},
				"response": "Task status object",
			},
			{
				"path": "/api/modules", 
				"method": "GET", 
				"description": "List all modules",
				"auth_required": true,
				"params": []interface{}{},
				"response": "Array of module objects",
			},
			{
				"path": "/api/modules/load", 
				"method": "POST", 
				"description": "Load a module",
				"auth_required": true,
				"params": []string{"name", "path"},
				"response": "Success or error message",
			},
			{
				"path": "/api/modules/exec", 
				"method": "POST", 
				"description": "Execute a module",
				"auth_required": true,
				"params": []string{"name", "params"},
				"response": "Module execution result",
			},
			{
				"path": "/api/clients", 
				"method": "GET", 
				"description": "List all clients",
				"auth_required": true,
				"params": []interface{}{},
				"response": "Array of client objects",
			},
			{
				"path": "/api/clients/tasks", 
				"method": "GET", 
				"description": "Get client tasks",
				"auth_required": true,
				"params": []string{"id"},
				"response": "Array of task objects for the client",
			},
			{
				"path": "/api/protocol/switch", 
				"method": "POST", 
				"description": "Switch protocol",
				"auth_required": true,
				"params": []string{"client_id", "protocol"},
				"response": "Success or error message",
			},
			{
				"path": "/api/auth/login", 
				"method": "POST", 
				"description": "Login to get a JWT token",
				"auth_required": false,
				"params": []string{"username", "password"},
				"response": "JWT token",
			},
			{
				"path": "/api/auth/refresh", 
				"method": "POST", 
				"description": "Refresh a JWT token",
				"auth_required": false,
				"params": []interface{}{},
				"response": "New JWT token",
			},
			{
				"path": "/api/docs", 
				"method": "GET", 
				"description": "API documentation",
				"auth_required": false,
				"params": []interface{}{},
				"response": "API documentation object",
			},
		},
		"config": map[string]interface{}{
			"api": map[string]interface{}{
				"enabled": "boolean - Whether the API is enabled",
				"address": "string - The address to bind the API server to",
				"port": "integer - The port to listen on",
				"tls_enabled": "boolean - Whether to use TLS",
				"tls_cert_file": "string - The path to the TLS certificate file",
				"tls_key_file": "string - The path to the TLS key file",
				"auth_enabled": "boolean - Whether authentication is enabled",
				"jwt_secret": "string - The secret key for JWT token generation",
				"token_expiry": "integer - The token expiry time in minutes",
			},
		},
	}
	
	writeJSON(w, apiInfo, http.StatusOK)
}
