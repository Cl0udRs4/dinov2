package websocket

import (
	"dinoc2/pkg/client"
	"dinoc2/pkg/crypto"
	"dinoc2/pkg/protocol"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketListener implements a WebSocket listener for the C2 server
type WebSocketListener struct {
	config        map[string]interface{}
	address       string
	port          int
	path          string
	server        *http.Server
	upgrader      websocket.Upgrader
	clientManager interface{}
	isRunning     bool
}

// NewWebSocketListener creates a new WebSocket listener
func NewWebSocketListener(config map[string]interface{}) (*WebSocketListener, error) {
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
	
	// Extract path
	path, ok := options["path"].(string)
	if !ok {
		path = "/ws"
	}

	return &WebSocketListener{
		config:    options,
		address:   address,
		port:      port,
		path:      path,
		upgrader:  websocket.Upgrader{},
		isRunning: false,
	}, nil
}

// SetClientManager sets the client manager for the listener
func (l *WebSocketListener) SetClientManager(cm interface{}) {
	l.clientManager = cm
	fmt.Printf("DEBUG: WebSocket listener client manager set: %T\n", cm)
}

// Start starts the WebSocket listener
func (l *WebSocketListener) Start() error {
	if l.isRunning {
		return fmt.Errorf("listener is already running")
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc(l.path, l.handleWebSocket)
	
	l.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", l.address, l.port),
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		err := l.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("WebSocket server error: %v\n", err)
		}
	}()

	l.isRunning = true
	fmt.Printf("WebSocket listener started on %s:%d%s\n", l.address, l.port, l.path)

	return nil
}

// Stop stops the WebSocket listener
func (l *WebSocketListener) Stop() error {
	if !l.isRunning {
		return nil
	}

	// Close server
	if l.server != nil {
		err := l.server.Close()
		if err != nil {
			return fmt.Errorf("failed to close WebSocket server: %w", err)
		}
	}

	l.isRunning = false
	fmt.Printf("WebSocket listener stopped\n")

	return nil
}

// handleWebSocket handles a WebSocket connection
func (l *WebSocketListener) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := l.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Failed to upgrade connection: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Printf("New WebSocket connection from %s\n", conn.RemoteAddr())

	// Create a protocol handler for processing the data
	protocolHandler := protocol.NewProtocolHandler()
	
	// Generate a unique session ID
	sessionID := crypto.GenerateSessionID()
	
	// Create a new client with the connection information
	newClient := client.NewClient(sessionID, conn.RemoteAddr().String(), client.ProtocolWebSocket)
	
	// Register client with client manager if available
	if l.clientManager != nil {
		fmt.Printf("DEBUG: WebSocket client manager type: %T\n", l.clientManager)
		
		// Try to register client using type assertion
		if cm, ok := l.clientManager.(*client.Manager); ok {
			clientID := cm.RegisterClient(newClient)
			fmt.Printf("Registered WebSocket client with ID %s\n", clientID)
		} else {
			fmt.Printf("Client manager does not implement RegisterClient method or is not of type *client.Manager\n")
		}
	} else {
		fmt.Printf("No client manager available\n")
	}

	// Read messages from the WebSocket
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("WebSocket connection closed: %v\n", err)
			break
		}

		// Only process binary messages
		if messageType != websocket.BinaryMessage {
			fmt.Printf("Received non-binary message, ignoring\n")
			continue
		}

		// Decode the packet
		packet, err := protocol.DecodePacket(message)
		if err != nil {
			fmt.Printf("Failed to decode packet: %v\n", err)
			continue
		}

		// Create a session with the encryption algorithm if not already created
		if !protocolHandler.HasSession(sessionID) {
			err = protocolHandler.CreateSession(sessionID, packet.Algorithm)
			if err != nil {
				fmt.Printf("Failed to create session: %v\n", err)
				continue
			}
		}

		// Process the packet
		response, err := protocolHandler.ProcessPacket(packet)
		if err != nil {
			fmt.Printf("Failed to process packet: %v\n", err)
			continue
		}

		// Send response
		err = conn.WriteMessage(websocket.BinaryMessage, response)
		if err != nil {
			fmt.Printf("Failed to send response: %v\n", err)
			break
		}
	}
}
