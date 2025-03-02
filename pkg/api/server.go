package api

import (
	"fmt"
	"log"
	"net/http"
)

// Server represents the API server
type Server struct {
	config       map[string]interface{}
	router       *Router
	server       *http.Server
	isRunning    bool
}

// NewServer creates a new API server
func NewServer(config map[string]interface{}) (*Server, error) {
	// Create router
	router := NewRouter(config)
	
	// Get address and port
	address, ok := config["address"].(string)
	if !ok {
		address = "127.0.0.1"
	}
	
	portFloat, ok := config["port"].(float64)
	if !ok {
		portFloat = 8443
	}
	port := int(portFloat)
	
	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", address, port),
		Handler: router,
	}
	
	return &Server{
		config:    config,
		router:    router,
		server:    server,
		isRunning: false,
	}, nil
}

// SetClientManager sets the client manager for the API server
func (s *Server) SetClientManager(clientManager interface{}) {
	s.router.SetClientManager(clientManager)
}

// SetUserAuth sets the user authentication for the API server
func (s *Server) SetUserAuth(userAuth map[string]interface{}) {
	s.router.SetUserAuth(userAuth)
}

// Start starts the API server
func (s *Server) Start() error {
	if s.isRunning {
		return fmt.Errorf("API server is already running")
	}
	
	// Start HTTP server in a goroutine
	go func() {
		log.Printf("Starting API server on %s", s.server.Addr)
		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Printf("API server error: %v", err)
		}
	}()
	
	s.isRunning = true
	return nil
}

// Stop stops the API server
func (s *Server) Stop() error {
	if !s.isRunning {
		return nil
	}
	
	// Shutdown HTTP server
	err := s.server.Close()
	if err != nil {
		return fmt.Errorf("failed to stop API server: %w", err)
	}
	
	s.isRunning = false
	return nil
}
