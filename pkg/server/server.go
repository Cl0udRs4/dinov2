package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"dinoc2/pkg/api"
	"dinoc2/pkg/api/middleware"
	"dinoc2/pkg/listener"
	"dinoc2/pkg/auth"
	"dinoc2/pkg/client"
	"dinoc2/pkg/module/manager"
	"dinoc2/pkg/task"
	
	"golang.org/x/crypto/bcrypt"
)

// APIConfig represents the API configuration
type APIConfig struct {
	Enabled     bool   `json:"enabled"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
	TLSEnabled  bool   `json:"tls_enabled"`
	TLSCertFile string `json:"tls_cert_file,omitempty"`
	TLSKeyFile  string `json:"tls_key_file,omitempty"`
	AuthEnabled bool   `json:"auth_enabled"`
	JWTSecret   string `json:"jwt_secret,omitempty"`
	TokenExpiry int    `json:"token_expiry,omitempty"` // in minutes
}


// ServerConfig represents the server configuration
type ServerConfig struct {
	API       APIConfig `json:"api"`
	UserAuth  auth.UserAuth  `json:"user_auth"`
	Listeners []struct {
		ID       string                 `json:"id"`
		Type     string                 `json:"type"`
		Address  string                 `json:"address"`
		Port     int                    `json:"port"`
		Options  map[string]interface{} `json:"options"`
		Disabled bool                   `json:"disabled,omitempty"`
	} `json:"listeners"`
}

// serverImpl is the actual implementation of the Server
type serverImpl struct {
	listenerManager *listener.Manager
	taskManager     *task.Manager
	mutex           sync.RWMutex
	config          *ServerConfig
}

// Global server state
var serverState *serverImpl

// NewServer creates a new C2 server
func NewServer() *Server {
	// Initialize the server state if it doesn't exist
	if serverState == nil {
		// Initialize with nil client manager, will be replaced in Start()
		serverState = &serverImpl{
			listenerManager: listener.NewManager(nil),
			taskManager:     task.NewManager(),
			config:          &ServerConfig{},
		}
	}
	
	// Return a pointer to the Server struct
	return &Server{}
}

// LoadConfig loads the server configuration from a file
func (s *Server) LoadConfig(configFile string) error {
	// Read the configuration file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read configuration file: %v", err)
	}

	// Parse the configuration
	config := &ServerConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse configuration: %v", err)
	}

	// Store the configuration
	serverState.config = config

	// Initialize user auth if it's empty
	if serverState.config.UserAuth.Username == "" {
		serverState.config.UserAuth = auth.UserAuth{
			Username: "admin",
			Password: "change_this_in_production",
			Role:     "admin",
		}
	}
	
	auth.SetUserAuth(&serverState.config.UserAuth)
	
	// Hash the password if provided in plaintext
	if serverState.config.UserAuth.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(serverState.config.UserAuth.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		serverState.config.UserAuth.PasswordHash = string(hashedPassword)
		serverState.config.UserAuth.Password = "" // Clear plaintext password
		
		// Save the updated configuration with hashed password
		configBytes, err := json.MarshalIndent(serverState.config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		
		if err := os.WriteFile(configFile, configBytes, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
	}

	return nil
}

// Start starts the server
func (s *Server) Start() error {
	// Ensure server state is initialized
	if serverState == nil {
		return fmt.Errorf("server not initialized, call LoadConfig first")
	}
	
	// Initialize module manager
	moduleManager, err := manager.NewModuleManager()
	if err != nil {
		return fmt.Errorf("failed to initialize module manager: %v", err)
	}
	
	// Initialize client manager
	clientManager := client.NewManager()
	
	// Initialize listener manager with client manager
	serverState.listenerManager = listener.NewManager(clientManager)
	
	// Initialize API if enabled
	var apiRouter *api.Router
	var authMiddleware *middleware.AuthMiddleware
	
	if serverState.config.API.Enabled {
		// Create authentication middleware if auth is enabled
		if serverState.config.API.AuthEnabled {
			authConfig := middleware.AuthConfig{
				Enabled:     serverState.config.API.AuthEnabled,
				JWTSecret:   serverState.config.API.JWTSecret,
				TokenExpiry: serverState.config.API.TokenExpiry,
			}
			authMiddleware = middleware.NewAuthMiddleware(authConfig)
		}
		
		// Create API router
		apiRouter = api.NewRouter(serverState.listenerManager, moduleManager, serverState.taskManager, clientManager, authMiddleware)
		
		// Start dedicated API server if configured
		if serverState.config.API.Port > 0 {
			go func() {
				addr := fmt.Sprintf("%s:%d", serverState.config.API.Address, serverState.config.API.Port)
				var err error
				
				log.Printf("Starting API server on %s", addr)
				
				if serverState.config.API.TLSEnabled && serverState.config.API.TLSCertFile != "" && serverState.config.API.TLSKeyFile != "" {
					err = http.ListenAndServeTLS(addr, serverState.config.API.TLSCertFile, serverState.config.API.TLSKeyFile, apiRouter)
				} else {
					err = http.ListenAndServe(addr, apiRouter)
				}
				
				if err != nil && err != http.ErrServerClosed {
					log.Printf("API server error: %v", err)
				}
			}()
		}
	}
	
	// Start all listeners
	for _, listenerConfig := range serverState.config.Listeners {
		// Skip disabled listeners
		if listenerConfig.Disabled {
			log.Printf("Skipping disabled listener %s", listenerConfig.ID)
			continue
		}

		// Convert to listener.ListenerConfig
		config := listener.ListenerConfig{
			Protocol: listenerConfig.Address,
			Address:  listenerConfig.Address,
			Port:     listenerConfig.Port,
			Options:  listenerConfig.Options,
		}
		
		// Pass API router to HTTP and WebSocket listeners if API is enabled
		if serverState.config.API.Enabled && (listenerConfig.Type == string(listener.ListenerTypeHTTP) || listenerConfig.Type == string(listener.ListenerTypeWebSocket)) {
			if config.Options == nil {
				config.Options = make(map[string]interface{})
			}
			config.Options["api_handler"] = apiRouter
		}

		// Create the listener
		listenerType := listener.ListenerType(listenerConfig.Type)
		err := serverState.listenerManager.CreateListener(listenerConfig.ID, listenerType, config)
		if err != nil {
			log.Printf("Failed to create listener %s: %v", listenerConfig.ID, err)
			continue
		}

		// Start the listener
		if err := serverState.listenerManager.StartListener(listenerConfig.ID); err != nil {
			log.Printf("Failed to start listener %s: %v", listenerConfig.ID, err)
			continue
		}

		log.Printf("Started listener %s (%s) on %s:%d", listenerConfig.ID, listenerConfig.Type, listenerConfig.Address, listenerConfig.Port)
	}

	return nil
}

// Shutdown stops the server
func (s *Server) Shutdown() error {
	// Ensure server state is initialized
	if serverState == nil {
		return nil // Nothing to stop
	}
	
	// Stop all listeners
	if err := serverState.listenerManager.StopAll(); err != nil {
		log.Printf("Failed to stop all listeners: %v", err)
	}

	return nil
}

// CreateDefaultConfig creates a default configuration file
func CreateDefaultConfig(outputPath string) error {
// Create a default configuration
	config := &ServerConfig{
		API: APIConfig{
			Enabled:     true,
			Address:     "127.0.0.1",
			Port:        8443,
			TLSEnabled:  false,
			TLSCertFile: "",
			TLSKeyFile:  "",
			AuthEnabled: true,
			JWTSecret:   "change_this_to_a_secure_secret_in_production", // Default secret, should be changed in production
			TokenExpiry: 60, // 1 hour
		},
		UserAuth: auth.UserAuth{
			Username: "admin",
			Password: "change_this_in_production", // Will be hashed during first load
			Role:     "admin",
		},
		Listeners: []struct {
			ID       string                 `json:"id"`
			Type     string                 `json:"type"`
			Address  string                 `json:"address"`
			Port     int                    `json:"port"`
			Options  map[string]interface{} `json:"options"`
			Disabled bool                   `json:"disabled,omitempty"`
		}{
			{
				ID:      "tcp1",
				Type:    string(listener.ListenerTypeTCP),
				Address: "0.0.0.0",
				Port:    8080,
				Options: map[string]interface{}{},
			},
			{
				ID:      "http1",
				Type:    string(listener.ListenerTypeHTTP),
				Address: "0.0.0.0",
				Port:    8000,
				Options: map[string]interface{}{
					"use_http2":  true,
					"enable_api": true,
				},
			},
			{
				ID:      "dns1",
				Type:    string(listener.ListenerTypeDNS),
				Address: "0.0.0.0",
				Port:    5353,
				Options: map[string]interface{}{
					"domain": "c2.example.com",
					"ttl":    60,
				},
			},
			{
				ID:      "ws1",
				Type:    string(listener.ListenerTypeWebSocket),
				Address: "0.0.0.0",
				Port:    8001,
				Options: map[string]interface{}{
					"path": "/ws",
				},
			},
			{
				ID:      "icmp1",
				Type:    string(listener.ListenerTypeICMP),
				Address: "0.0.0.0",
				Port:    0,
				Options: map[string]interface{}{
					"protocol": "icmp",
				},
				Disabled: false,
			},
		},
	}

	// Convert to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %v", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %v", err)
	}

	return nil
}

// GetListenerManager returns the listener manager
func (s *Server) GetListenerManager() *listener.Manager {
	if serverState == nil {
		return nil
	}
	return serverState.listenerManager
}

// GetTaskManager returns the task manager
func (s *Server) GetTaskManager() *task.Manager {
	if serverState == nil {
		return nil
	}
	return serverState.taskManager
}
