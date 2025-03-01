package server

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"dinoc2/pkg/listener"
	"dinoc2/pkg/task"
)

// ServerConfig represents the server configuration
type ServerConfig struct {
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
		serverState = &serverImpl{
			listenerManager: listener.NewManager(),
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

	return nil
}

// Start starts the server
func (s *Server) Start() error {
	// Ensure server state is initialized
	if serverState == nil {
		return fmt.Errorf("server not initialized, call LoadConfig first")
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
