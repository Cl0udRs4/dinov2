package api

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Server represents the API server
type Server struct {
	config        map[string]interface{}
	router        *Router
	httpServer    *http.Server
	isRunning     bool
}

// NewServer creates a new API server
func NewServer(config map[string]interface{}) (*Server, error) {
	// Create router
	router := NewRouter(config)

	// Extract address and port
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
	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", address, port),
		Handler: router,
	}
	
	// Configure TLS if enabled
	tlsEnabled, ok := config["tls_enabled"].(bool)
	if ok && tlsEnabled {
		// Load TLS configuration
		certFile, ok := config["cert_file"].(string)
		if !ok {
			return nil, fmt.Errorf("TLS enabled but cert_file not specified")
		}
		
		keyFile, ok := config["key_file"].(string)
		if !ok {
			return nil, fmt.Errorf("TLS enabled but key_file not specified")
		}
		
		// Create TLS configuration
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		
		// Load certificate
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
		}
		
		tlsConfig.Certificates = []tls.Certificate{cert}
		httpServer.TLSConfig = tlsConfig
	}

	return &Server{
		config:     config,
		router:     router,
		httpServer: httpServer,
		isRunning:  false,
	}, nil
}

// SetClientManager sets the client manager for the API server
func (s *Server) SetClientManager(clientManager interface{}) {
	s.router.SetClientManager(clientManager)
}

// SetTaskManager sets the task manager for the API server
func (s *Server) SetTaskManager(taskManager interface{}) {
	s.router.SetTaskManager(taskManager)
}

// Start starts the API server
func (s *Server) Start() error {
	if s.isRunning {
		return fmt.Errorf("API server is already running")
	}

	// Extract address and port for logging
	address, ok := s.config["address"].(string)
	if !ok {
		address = "127.0.0.1"
	}
	
	portFloat, ok := s.config["port"].(float64)
	if !ok {
		portFloat = 8443
	}
	port := int(portFloat)

	// Start HTTP server in a goroutine
	go func() {
		// Check if TLS is enabled
		tlsEnabled, ok := s.config["tls_enabled"].(bool)
		if ok && tlsEnabled {
			// Start HTTPS server
			certFile, _ := s.config["cert_file"].(string)
			keyFile, _ := s.config["key_file"].(string)
			
			err := s.httpServer.ListenAndServeTLS(certFile, keyFile)
			if err != nil && err != http.ErrServerClosed {
				log.Printf("API server error: %v", err)
			}
		} else {
			// Start HTTP server
			err := s.httpServer.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Printf("API server error: %v", err)
			}
		}
	}()

	s.isRunning = true
	log.Printf("Starting API server on %s:%d", address, port)

	return nil
}

// Stop stops the API server
func (s *Server) Stop() error {
	if !s.isRunning {
		return nil
	}

	// Close HTTP server
	err := s.httpServer.Close()
	if err != nil {
		return fmt.Errorf("failed to close API server: %w", err)
	}

	s.isRunning = false
	return nil
}
