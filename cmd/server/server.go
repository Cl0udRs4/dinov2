package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	
	"dinoc2/pkg/listener"
	"dinoc2/pkg/module/loader"
	"dinoc2/pkg/module/manager"
)

// ServerConfig holds the configuration for the C2 server
type ServerConfig struct {
	Listeners []ListenerConfig `json:"listeners"`
}

// ListenerConfig holds the configuration for a listener
type ListenerConfig struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Address  string                 `json:"address"`
	Port     int                    `json:"port"`
	Options  map[string]interface{} `json:"options"`
	Disabled bool                   `json:"disabled"`
}

// Server represents the C2 server
type Server struct {
	config          ServerConfig
	listenerManager *listener.Manager
	moduleManager   *manager.ModuleManager
}

// NewServer creates a new C2 server
func NewServer() *Server {
	// Create module manager
	moduleManager, err := manager.NewModuleManager()
	if err != nil {
		log.Printf("Warning: Failed to initialize module manager: %v", err)
		moduleManager = nil
	}

	return &Server{
		listenerManager: listener.NewManager(),
		moduleManager:   moduleManager,
	}
}

// LoadConfig loads the server configuration from a file
func (s *Server) LoadConfig(configFile string) error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config ServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	s.config = config
	return nil
}

// Start starts the server and all enabled listeners
func (s *Server) Start() error {
	// Start all enabled listeners
	for _, lc := range s.config.Listeners {
		if lc.Disabled {
			log.Printf("Listener %s is disabled, skipping", lc.ID)
			continue
		}

		// Convert listener type string to ListenerType
		var listenerType listener.ListenerType
		switch lc.Type {
		case "tcp":
			listenerType = listener.ListenerTypeTCP
		case "dns":
			listenerType = listener.ListenerTypeDNS
		case "icmp":
			listenerType = listener.ListenerTypeICMP
		case "http":
			listenerType = listener.ListenerTypeHTTP
		case "websocket":
			listenerType = listener.ListenerTypeWebSocket
		default:
			return fmt.Errorf("unsupported listener type: %s", lc.Type)
		}

		// Create listener configuration
		config := listener.ListenerConfig{
			Address:  lc.Address,
			Port:     lc.Port,
			Options:  lc.Options,
		}

		// Create and start the listener
		err := s.listenerManager.CreateListener(lc.ID, listenerType, config)
		if err != nil {
			log.Printf("Failed to create listener %s: %v", lc.ID, err)
			continue
		}

		err = s.listenerManager.StartListener(lc.ID)
		if err != nil {
			log.Printf("Failed to start listener %s: %v", lc.ID, err)
			continue
		}

		log.Printf("Started listener %s (%s) on %s:%d", lc.ID, lc.Type, lc.Address, lc.Port)
	}

	return nil
}

// Stop stops the server and all listeners
func (s *Server) Stop() error {
	return s.listenerManager.StopAll()
}

// Shutdown performs a clean shutdown of the server
func (s *Server) Shutdown() error {
	// Shutdown module manager if available
	if s.moduleManager != nil {
		errors := s.moduleManager.ShutdownAllModules()
		if len(errors) > 0 {
			for _, err := range errors {
				log.Printf("Error shutting down module: %v", err)
			}
		}
	}

	// Shutdown listener manager
	return s.listenerManager.Shutdown()
}

// LoadModule loads a module using the specified loader
func (s *Server) LoadModule(name, path string, loaderType string) error {
	if s.moduleManager == nil {
		return fmt.Errorf("module manager not initialized")
	}

	// Convert loader type string to LoaderType
	var lt loader.LoaderType
	switch loaderType {
	case "native":
		lt = loader.LoaderTypeNative
	case "plugin":
		lt = loader.LoaderTypePlugin
	case "dll":
		lt = loader.LoaderTypeDLL
	case "wasm":
		lt = loader.LoaderTypeWasm
	case "rpc":
		lt = loader.LoaderTypeRPC
	default:
		return fmt.Errorf("unsupported loader type: %s", loaderType)
	}

	// Load module
	_, err := s.moduleManager.LoadModule(name, path, lt)
	if err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}

	return nil
}

// InitModule initializes a loaded module
func (s *Server) InitModule(name string, params map[string]interface{}) error {
	if s.moduleManager == nil {
		return fmt.Errorf("module manager not initialized")
	}

	return s.moduleManager.InitModule(name, params)
}

// ExecModule executes a command on a module
func (s *Server) ExecModule(name, command string, args ...interface{}) (interface{}, error) {
	if s.moduleManager == nil {
		return nil, fmt.Errorf("module manager not initialized")
	}

	return s.moduleManager.ExecModule(name, command, args...)
}

// UnloadModule unloads a module
func (s *Server) UnloadModule(name string) error {
	if s.moduleManager == nil {
		return fmt.Errorf("module manager not initialized")
	}

	return s.moduleManager.UnloadModule(name)
}

// ListModules returns a list of loaded modules
func (s *Server) ListModules() map[string]manager.ModuleInfo {
	if s.moduleManager == nil {
		return make(map[string]manager.ModuleInfo)
	}

	return s.moduleManager.ListModules()
}

// CreateDefaultConfig creates a default configuration file
func CreateDefaultConfig(filename string) error {
	config := ServerConfig{
		Listeners: []ListenerConfig{
			{
				ID:      "tcp1",
				Type:    "tcp",
				Address: "0.0.0.0",
				Port:    8080,
				Options: map[string]interface{}{},
			},
			{
				ID:      "tcp2",
				Type:    "tcp",
				Address: "0.0.0.0",
				Port:    8081,
				Options: map[string]interface{}{},
			},
			{
				ID:      "dns1",
				Type:    "dns",
				Address: "0.0.0.0",
				Port:    53,
				Options: map[string]interface{}{
					"domain": "c2.example.com",
					"ttl":    60,
				},
			},
			{
				ID:      "http1",
				Type:    "http",
				Address: "0.0.0.0",
				Port:    8443,
				Options: map[string]interface{}{
					"use_http2":    true,
					"tls_cert_file": "certs/server.crt",
					"tls_key_file":  "certs/server.key",
				},
			},
			{
				ID:      "ws1",
				Type:    "websocket",
				Address: "0.0.0.0",
				Port:    8444,
				Options: map[string]interface{}{
					"path":         "/ws",
					"tls_cert_file": "certs/server.crt",
					"tls_key_file":  "certs/server.key",
				},
			},
			{
				ID:      "icmp1",
				Type:    "icmp",
				Address: "0.0.0.0",
				Port:    0, // ICMP doesn't use ports
				Options: map[string]interface{}{
					"protocol": "icmp",
				},
				Disabled: true, // Disabled by default as it requires root privileges
			},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
