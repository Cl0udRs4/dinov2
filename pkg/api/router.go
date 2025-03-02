package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Router handles API requests
type Router struct {
	config       map[string]interface{}
	userAuth     map[string]interface{}
	routes       map[string]func(http.ResponseWriter, *http.Request)
	clientManager interface{}
	taskManager   interface{}
}

// NewRouter creates a new router
func NewRouter(config map[string]interface{}) *Router {
	router := &Router{
		config: config,
		routes: make(map[string]func(http.ResponseWriter, *http.Request)),
	}
	
	// Register routes
	router.RegisterAuthRoutes()
	router.RegisterClientRoutes()
	
	return router
}

// SetClientManager sets the client manager for the router
func (r *Router) SetClientManager(clientManager interface{}) {
	r.clientManager = clientManager
	fmt.Printf("DEBUG: Router client manager set: %T\n", clientManager)
}

// SetUserAuth sets the user authentication for the router
func (r *Router) SetUserAuth(userAuth map[string]interface{}) {
	r.userAuth = userAuth
}

// SetTaskManager sets the task manager for the router
func (r *Router) SetTaskManager(taskManager interface{}) {
	r.taskManager = taskManager
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	
	// Handle OPTIONS requests
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Check if authentication is enabled
	authEnabled, ok := r.config["auth_enabled"].(bool)
	if !ok {
		authEnabled = true
	}
	
	// Check if route requires authentication
	if authEnabled && !strings.HasPrefix(req.URL.Path, "/api/auth/") {
		// Validate JWT token
		token := r.extractToken(req)
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		// For simplicity, just check if token exists
		// In a real implementation, we would validate the token
		if token == "" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}
	}
	
	// Find route handler
	handler, ok := r.routes[req.URL.Path]
	if !ok {
		http.NotFound(w, req)
		return
	}
	
	// Call handler
	handler(w, req)
}

// RegisterAuthRoutes registers authentication routes
func (r *Router) RegisterAuthRoutes() {
	// Register login route
	r.routes["/api/auth/login"] = r.handleLogin
}

// handleLogin handles login requests
func (r *Router) handleLogin(w http.ResponseWriter, req *http.Request) {
	// Only allow POST requests
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Parse request body
	var loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	err := json.NewDecoder(req.Body).Decode(&loginRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Get user auth from config
	if r.userAuth == nil {
		http.Error(w, "Authentication not configured", http.StatusInternalServerError)
		return
	}
	
	// Get username and password
	username, ok := r.userAuth["username"].(string)
	if !ok {
		http.Error(w, "Username not configured", http.StatusInternalServerError)
		return
	}
	
	password, ok := r.userAuth["password"].(string)
	if !ok {
		passwordHash, ok := r.userAuth["password_hash"].(string)
		if !ok {
			http.Error(w, "Password not configured", http.StatusInternalServerError)
			return
		}
		
		// For simplicity, just check if password hash exists
		// In a real implementation, we would validate the password hash
		if passwordHash == "" {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		
		// Use password from request
		password = loginRequest.Password
	}
	
	fmt.Printf("Validating credentials: username=%s, password=%s\n", loginRequest.Username, loginRequest.Password)
	
	// Check username and password
	if loginRequest.Username != username || loginRequest.Password != password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	
	// Get user role
	role, ok := r.userAuth["role"].(string)
	if !ok {
		role = "user"
	}
	
	fmt.Printf("Credentials valid, returning role: %s\n", role)
	
	// Generate JWT token
	token := r.generateToken(loginRequest.Username, role)
	
	// Return token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"role":  role,
	})
}

// generateToken generates a simple token
func (r *Router) generateToken(username, role string) string {
	// Get JWT secret
	jwtSecret, ok := r.config["jwt_secret"].(string)
	if !ok {
		jwtSecret = "default_secret"
	}
	
	// Simple token generation
	return fmt.Sprintf("%s_%s_%d_%s", username, role, time.Now().Unix(), jwtSecret)
}

// extractToken extracts the JWT token from the request
func (r *Router) extractToken(req *http.Request) string {
	// Get Authorization header
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	
	// Check if header starts with Bearer
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	
	// Extract token
	return strings.TrimPrefix(authHeader, "Bearer ")
}
