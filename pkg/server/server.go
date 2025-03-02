package server

import (
	"dinoc2/pkg/api"
	"dinoc2/pkg/client"
	"dinoc2/pkg/listener"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Server represents the C2 server
type Server struct {
	config         map[string]interface{}
	listenerManager *listener.Manager
	clientManager   *client.Manager
	apiServer       *api.Server
	isRunning       bool
}

// NewServer creates a new server instance
func NewServer(configFile string) (*Server, error) {
	// Load configuration
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create listener manager
	listenerManager := listener.NewManager()

	// Create API server if enabled
	var apiServer *api.Server
	if apiConfig, ok := config["api"].(map[string]interface{}); ok {
		if enabled, ok := apiConfig["enabled"].(bool); ok && enabled {
			apiServer, err = api.NewServer(apiConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to create API server: %w", err)
			}
			
			// Pass user auth to API server
			if userAuth, ok := config["user_auth"].(map[string]interface{}); ok {
				fmt.Printf("Setting user auth: %+v\n", userAuth)
				apiServer.SetUserAuth(userAuth)
			}
		}
	}

	return &Server{
		config:         config,
		listenerManager: listenerManager,
		apiServer:       apiServer,
		isRunning:       false,
	}, nil
}

// SetClientManager sets the client manager for the server
func (s *Server) SetClientManager(clientManager *client.Manager) {
	s.clientManager = clientManager
	
	// Pass client manager to listener manager
	s.listenerManager.SetClientManager(clientManager)
	
	// Pass client manager to API server if available
	if s.apiServer != nil {
		s.apiServer.SetClientManager(clientManager)
	}
}

// Start starts the server
func (s *Server) Start() error {
	if s.isRunning {
		return fmt.Errorf("server is already running")
	}

	// Start API server if available
	if s.apiServer != nil {
		err := s.apiServer.Start()
		if err != nil {
			return fmt.Errorf("failed to start API server: %w", err)
		}
	}

	// Start listeners
	if listeners, ok := s.config["listeners"].([]interface{}); ok {
		for _, listenerConfig := range listeners {
			if listenerConfigMap, ok := listenerConfig.(map[string]interface{}); ok {
				err := s.listenerManager.AddListener(listenerConfigMap)
				if err != nil {
					log.Printf("Failed to add listener: %v", err)
					continue
				}
			}
		}
	}

	// Start all listeners
	err := s.listenerManager.StartAll()
	if err != nil {
		return fmt.Errorf("failed to start listeners: %w", err)
	}

	s.isRunning = true
	return nil
}

// Stop stops the server
func (s *Server) Stop() error {
	if !s.isRunning {
		return nil
	}

	// Stop all listeners
	err := s.listenerManager.StopAll()
	if err != nil {
		return fmt.Errorf("failed to stop listeners: %w", err)
	}

	// Stop API server if available
	if s.apiServer != nil {
		err := s.apiServer.Stop()
		if err != nil {
			return fmt.Errorf("failed to stop API server: %w", err)
		}
	}

	s.isRunning = false
	return nil
}

// loadConfig loads the server configuration from a file
func loadConfig(configFile string) (map[string]interface{}, error) {
	// Read configuration file
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Parse JSON configuration
	var config map[string]interface{}
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	return config, nil
}
