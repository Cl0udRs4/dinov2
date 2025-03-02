package api

import (
	"dinoc2/pkg/client"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// Router handles API requests
type Router struct {
	config       map[string]interface{}
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
		
		// Validate token
		_, err := r.validateToken(token)
		if err != nil {
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
	userAuth, ok := r.config["user_auth"].(map[string]interface{})
	if !ok {
		http.Error(w, "Authentication not configured", http.StatusInternalServerError)
		return
	}
	
	// Get username and password hash
	username, ok := userAuth["username"].(string)
	if !ok {
		http.Error(w, "Username not configured", http.StatusInternalServerError)
		return
	}
	
	// Check if password is provided directly (for testing)
	password, ok := userAuth["password"].(string)
	if ok {
		fmt.Printf("Validating credentials: username=%s, password=%s\n", loginRequest.Username, loginRequest.Password)
		
		// Check username and password
		if loginRequest.Username != username || loginRequest.Password != password {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
	} else {
		// Get password hash
		passwordHash, ok := userAuth["password_hash"].(string)
		if !ok {
			http.Error(w, "Password hash not configured", http.StatusInternalServerError)
			return
		}
		
		fmt.Printf("Stored username: %s, passwordHash: %s\n", username, passwordHash)
		
		// Check username
		if loginRequest.Username != username {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		
		// Check password
		err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(loginRequest.Password))
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
	}
	
	// Get user role
	role, ok := userAuth["role"].(string)
	if !ok {
		role = "user"
	}
	
	fmt.Printf("Credentials valid, returning role: %s\n", role)
	
	// Generate JWT token
	token, err := r.generateToken(loginRequest.Username, role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	
	// Return token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"role":  role,
	})
}

// generateToken generates a JWT token
func (r *Router) generateToken(username, role string) (string, error) {
	// Get JWT secret
	jwtSecret, ok := r.config["jwt_secret"].(string)
	if !ok {
		return "", fmt.Errorf("JWT secret not configured")
	}
	
	// Get token expiry
	tokenExpiryFloat, ok := r.config["token_expiry"].(float64)
	if !ok {
		tokenExpiryFloat = 60
	}
	tokenExpiry := time.Duration(tokenExpiryFloat) * time.Minute
	
	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"role":     role,
		"iss":      "dinoc2",
		"sub":      username,
		"exp":      time.Now().Add(tokenExpiry).Unix(),
		"nbf":      time.Now().Unix(),
		"iat":      time.Now().Unix(),
	})
	
	// Sign token
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	
	return tokenString, nil
}

// validateToken validates a JWT token
func (r *Router) validateToken(tokenString string) (jwt.MapClaims, error) {
	// Get JWT secret
	jwtSecret, ok := r.config["jwt_secret"].(string)
	if !ok {
		return nil, fmt.Errorf("JWT secret not configured")
	}
	
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		
		return []byte(jwtSecret), nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	
	// Validate token
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	
	// Get claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	
	return claims, nil
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
