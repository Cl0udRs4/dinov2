package server

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"dinoc2/pkg/api"
	"dinoc2/pkg/listener"
	"dinoc2/pkg/module/manager"
	"dinoc2/pkg/security"
)

// ServerConfig represents the server configuration
type ListenerConfig struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Address string                 `json:"address"`
	Port    int                    `json:"port"`
	Enabled bool                   `json:"enabled"`
	Options map[string]interface{} `json:"options"`
}

type ServerConfig struct {
	Listeners []ListenerConfig `json:"listeners"`
	APIServer APIServerConfig   `json:"api_server"`
}

// APIServerConfig represents the API server configuration
type APIServerConfig struct {
	Enabled bool   `json:"enabled"`
	Address string `json:"address"`
	Port    int    `json:"port"`
}

// Server represents the C2 server
type Server struct {
	config          ServerConfig
	listenerManager *listener.Manager
	moduleManager   *manager.ModuleManager
	apiServer       *api.Server
	authenticator   *security.Authenticator
}

// NewServer creates a new server
func NewServer(config ServerConfig) *Server {
	// Create listener manager
	listenerManager := listener.NewManager()

	// Create module manager
	moduleManager, err := manager.NewModuleManager()
	if err != nil {
		log.Printf("Warning: Failed to initialize module manager: %v", err)
		moduleManager = nil
	}

	// Create authenticator
	defaultAuthOptions := security.DefaultAuthenticationOptions{}
	authenticator := security.NewAuthenticator(defaultAuthOptions.DefaultAuthenticationOptions())


	return &Server{
		config:          config,
		listenerManager: listenerManager,
		moduleManager:   moduleManager,
		authenticator:   authenticator,
	}
}

// Start starts the server
func (s *Server) Start() error {
	// Start all enabled listeners
	for _, lc := range s.config.Listeners {
		if !lc.Enabled {
			log.Printf("Listener %s is disabled, skipping", lc.ID)
			continue
		}

		// Convert ListenerConfig to listener.Config
		listenerConfig := listener.Config{
			ID:      lc.ID,
			Type:    lc.Type,
			Address: lc.Address,
			Port:    lc.Port,
			Enabled: lc.Enabled,
			Options: lc.Options,
		}

		// Create listener
		l, err := s.listenerManager.CreateListener(listenerConfig)
		if err != nil {
			log.Printf("Error creating listener %s: %v", lc.ID, err)
			continue
		}

		// Start listener
		if err := l.Start(); err != nil {
			log.Printf("Error starting listener %s: %v", lc.ID, err)
			continue
		}

		log.Printf("Listener %s started successfully", lc.ID)
	}

	// Start API server if enabled
	if s.config.APIServer.Enabled {
		if err := s.StartAPIServer(s.config.APIServer.Address, s.config.APIServer.Port); err != nil {
			return fmt.Errorf("failed to start API server: %w", err)
		}
	}

	return nil
}

// StartAPIServer starts the API server
func (s *Server) StartAPIServer(address string, port int) error {
	if s.apiServer == nil {
		// Create API server
		s.apiServer = api.NewServer(s.authenticator, s.listenerManager, s.moduleManager, nil, "./modules")
	}

	// Start API server in a goroutine
	go func() {
		if err := s.apiServer.Start(address, port); err != nil {
			log.Printf("Error starting API server: %v", err)
		}
	}()

	log.Printf("API server started on %s:%d", address, port)
	return nil
}

// Stop stops the server
func (s *Server) Stop() error {
	// Stop all listeners
	if err := s.listenerManager.StopAll(); err != nil {
		log.Printf("Error stopping listeners: %v", err)
	}

	return nil
}

// WaitForInterrupt waits for an interrupt signal
func (s *Server) WaitForInterrupt() {
	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for a signal
	<-sigChan

	// Stop the server
	log.Println("Interrupt received, shutting down...")
	if err := s.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}
}
