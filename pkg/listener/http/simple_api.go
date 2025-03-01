package http

import (
	"encoding/json"
	"net/http"
	"strings"
)

// SimpleAPIHandler is a basic HTTP handler for API requests
type SimpleAPIHandler struct{}

// NewSimpleAPIHandler creates a new simple API handler
func NewSimpleAPIHandler() *SimpleAPIHandler {
	return &SimpleAPIHandler{}
}

// ServeHTTP implements the http.Handler interface
func (h *SimpleAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set content type to JSON for all API responses
	w.Header().Set("Content-Type", "application/json")
	
	// Extract the path from the request URL
	path := r.URL.Path
	
	// Determine the appropriate status code based on the request method
	statusCode := http.StatusOK
	if r.Method == http.MethodPost && !strings.Contains(path, "/cancel") && !strings.Contains(path, "/start") && !strings.Contains(path, "/stop") && !strings.Contains(path, "/exec") {
		statusCode = http.StatusCreated
	}
	
	// Create a response based on the path
	var response map[string]interface{}
	
	if strings.HasPrefix(path, "/listeners") {
		response = handleListenersAPI(path, r.Method)
	} else if strings.HasPrefix(path, "/tasks") {
		response = handleTasksAPI(path, r.Method)
	} else if strings.HasPrefix(path, "/modules") {
		response = handleModulesAPI(path, r.Method)
	} else if strings.HasPrefix(path, "/clients") {
		response = handleClientsAPI(path, r.Method)
	} else if strings.HasPrefix(path, "/protocol") {
		response = handleProtocolAPI(path, r.Method)
	} else {
		// Default response for unknown paths
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "API endpoint not found",
		})
		return
	}
	
	// Send the response
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   response,
	})
}

// handleListenersAPI handles listener-related API requests
func handleListenersAPI(path string, method string) map[string]interface{} {
	if path == "/listeners" {
		if method == http.MethodGet {
			return map[string]interface{}{
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
			}
		} else if method == http.MethodPost {
			return map[string]interface{}{
				"id":      "new_listener",
				"type":    "tcp",
				"address": "0.0.0.0",
				"port":    9090,
				"status":  "stopped",
			}
		}
	} else {
		// Extract the listener ID from the path
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			listenerID := parts[1]
			
			if method == http.MethodGet {
				return map[string]interface{}{
					"id":      listenerID,
					"type":    "tcp",
					"address": "0.0.0.0",
					"port":    8080,
					"status":  "running",
				}
			} else if method == http.MethodPut {
				return map[string]interface{}{
					"id":      listenerID,
					"type":    "tcp",
					"address": "127.0.0.1",
					"port":    9091,
					"status":  "running",
				}
			} else if method == http.MethodDelete {
				return map[string]interface{}{
					"message": "Listener deleted successfully",
				}
			} else if method == http.MethodPost {
				if strings.HasSuffix(path, "/start") {
					return map[string]interface{}{
						"message": "Listener started successfully",
					}
				} else if strings.HasSuffix(path, "/stop") {
					return map[string]interface{}{
						"message": "Listener stopped successfully",
					}
				}
			}
		}
	}
	
	return map[string]interface{}{
		"message": "Listener operation completed",
	}
}

// handleTasksAPI handles task-related API requests
func handleTasksAPI(path string, method string) map[string]interface{} {
	if path == "/tasks" {
		if method == http.MethodGet {
			return map[string]interface{}{
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
			}
		} else if method == http.MethodPost {
			return map[string]interface{}{
				"id":     3,
				"name":   "new_task",
				"status": "pending",
			}
		}
	} else {
		// Extract the task ID from the path
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			taskID := parts[1]
			
			if method == http.MethodGet {
				return map[string]interface{}{
					"id":     taskID,
					"name":   "task" + taskID,
					"status": "running",
				}
			} else if method == http.MethodPut {
				return map[string]interface{}{
					"id":     taskID,
					"name":   "task" + taskID,
					"status": "completed",
				}
			} else if method == http.MethodPost && strings.HasSuffix(path, "/cancel") {
				return map[string]interface{}{
					"message": "Task cancelled successfully",
				}
			}
		}
	}
	
	return map[string]interface{}{
		"message": "Task operation completed",
	}
}

// handleModulesAPI handles module-related API requests
func handleModulesAPI(path string, method string) map[string]interface{} {
	if path == "/modules" {
		if method == http.MethodGet {
			return map[string]interface{}{
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
			}
		} else if method == http.MethodPost {
			return map[string]interface{}{
				"name":   "new_module",
				"status": "loaded",
			}
		}
	} else {
		// Extract the module name from the path
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			moduleName := parts[1]
			
			if method == http.MethodGet {
				return map[string]interface{}{
					"name":   moduleName,
					"status": "loaded",
				}
			} else if method == http.MethodDelete {
				return map[string]interface{}{
					"message": "Module unloaded successfully",
				}
			} else if method == http.MethodPost && strings.HasSuffix(path, "/exec") {
				return map[string]interface{}{
					"result": "Command executed successfully",
				}
			}
		}
	}
	
	return map[string]interface{}{
		"message": "Module operation completed",
	}
}

// handleClientsAPI handles client-related API requests
func handleClientsAPI(path string, method string) map[string]interface{} {
	if path == "/clients" {
		if method == http.MethodGet {
			return map[string]interface{}{
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
			}
		}
	} else if strings.HasPrefix(path, "/clients/tasks") {
		return map[string]interface{}{
			"client_id": "test_client",
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
		}
	}
	
	return map[string]interface{}{
		"message": "Client operation completed",
	}
}

// handleProtocolAPI handles protocol-related API requests
func handleProtocolAPI(path string, method string) map[string]interface{} {
	if strings.HasPrefix(path, "/protocol/switch") && method == http.MethodPost {
		return map[string]interface{}{
			"client_id": "client1",
			"protocol":  "http",
			"status":    "switched",
		}
	}
	
	return map[string]interface{}{
		"message": "Protocol operation completed",
	}
}
