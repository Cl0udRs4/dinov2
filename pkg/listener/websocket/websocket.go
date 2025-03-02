package websocket

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"dinoc2/pkg/client"
	"dinoc2/pkg/crypto"
	"dinoc2/pkg/protocol"
)

// WebSocketListener implements the Listener interface for WebSocket protocol
type WebSocketListener struct {
	config     WebSocketConfig
	server     *http.Server
	status     string
	statusLock sync.RWMutex
	upgrader   websocket.Upgrader
	clients    map[*websocket.Conn]bool
	clientLock sync.RWMutex
}

// WebSocketConfig holds configuration for the WebSocket listener
type WebSocketConfig struct {
	Address     string
	Port        int
	Path        string
	TLSCertFile string
	TLSKeyFile  string
	Options     map[string]interface{}
}

// NewWebSocketListener creates a new WebSocket listener
func NewWebSocketListener(config WebSocketConfig) *WebSocketListener {
	// Set default path if not provided
	if config.Path == "" {
		config.Path = "/ws"
	}

	return &WebSocketListener{
		config: config,
		status: "stopped",
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Allow all origins for now
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

// Start implements the Listener interface
func (l *WebSocketListener) Start() error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status == "running" {
		return fmt.Errorf("WebSocket listener is already running")
	}

	// Create a new HTTP server for the WebSocket
	addr := fmt.Sprintf("%s:%d", l.config.Address, l.config.Port)
	
	// Create a router/mux for handling requests
	mux := http.NewServeMux()
	
	// Register the WebSocket handler
	mux.HandleFunc(l.config.Path, l.handleWebSocket)
	
	l.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	
	// Start the server in a goroutine
	go func() {
		var err error
		
		// Check if TLS is configured
		if l.config.TLSCertFile != "" && l.config.TLSKeyFile != "" {
			err = l.server.ListenAndServeTLS(l.config.TLSCertFile, l.config.TLSKeyFile)
		} else {
			err = l.server.ListenAndServe()
		}
		
		if err != nil && err != http.ErrServerClosed {
			l.statusLock.Lock()
			l.status = "error"
			l.statusLock.Unlock()
			fmt.Printf("Error starting WebSocket listener: %v\n", err)
		}
	}()
	
	l.status = "running"
	return nil
}

// Stop implements the Listener interface
func (l *WebSocketListener) Stop() error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status != "running" {
		return nil // Already stopped
	}

	// Close all client connections
	l.clientLock.Lock()
	for client := range l.clients {
		client.Close()
		delete(l.clients, client)
	}
	l.clientLock.Unlock()

	// Shutdown the server
	if l.server != nil {
		err := l.server.Close()
		if err != nil {
			l.status = "error"
			return fmt.Errorf("error stopping WebSocket listener: %w", err)
		}
	}

	l.status = "stopped"
	return nil
}

// Status implements the Listener interface
func (l *WebSocketListener) Status() string {
	l.statusLock.RLock()
	defer l.statusLock.RUnlock()
	return l.status
}

// Configure implements the Listener interface
func (l *WebSocketListener) Configure(config interface{}) error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status == "running" {
		return fmt.Errorf("cannot configure a running WebSocket listener")
	}

	wsConfig, ok := config.(WebSocketConfig)
	if !ok {
		return fmt.Errorf("invalid configuration type for WebSocket listener")
	}

	l.config = wsConfig
	return nil
}

// handleWebSocket handles WebSocket connections
func (l *WebSocketListener) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := l.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Error upgrading to WebSocket: %v\n", err)
		return
	}
	
	// Register the client
	l.clientLock.Lock()
	l.clients[conn] = true
	l.clientLock.Unlock()
	
	// Handle the connection in a goroutine
	go l.handleConnection(conn)
}

// handleConnection processes a WebSocket connection
func (l *WebSocketListener) handleConnection(conn *websocket.Conn) {
	defer func() {
		// Unregister the client when the function returns
		l.clientLock.Lock()
		delete(l.clients, conn)
		l.clientLock.Unlock()
		
		// Close the connection
		conn.Close()
	}()
	
	for {
		// Read message from the client
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			// Check if it's a normal close
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}
		
		// Process the message
		go l.processMessage(conn, messageType, message)
	}
}

// processMessage processes a WebSocket message
func (l *WebSocketListener) processMessage(conn *websocket.Conn, messageType int, message []byte) {
	fmt.Printf("Received WebSocket message from %s\n", conn.RemoteAddr())
	
	// Create a protocol handler for processing the data
	protocolHandler := protocol.NewProtocolHandler()
	
	// Generate a session ID based on the connection address
	sessionID := crypto.SessionID(conn.RemoteAddr().String())
	
	// Decode the packet to get the encryption algorithm
	packet, err := protocol.DecodePacket(message)
	if err != nil {
		fmt.Printf("Error decoding WebSocket packet data: %v\n", err)
		// Echo the message back for invalid packets
		conn.WriteMessage(messageType, message)
		return
	}
	
	// Determine the encryption algorithm from the packet header
	var encAlgorithm string
	var cryptoAlgorithm crypto.Algorithm
	switch packet.Header.EncAlgorithm {
	case protocol.EncryptionAlgorithmAES:
		encAlgorithm = "aes"
		cryptoAlgorithm = crypto.AlgorithmAES
	case protocol.EncryptionAlgorithmChacha20:
		encAlgorithm = "chacha20"
		cryptoAlgorithm = crypto.AlgorithmChacha20
	default:
		encAlgorithm = "aes" // Default to AES if not specified
		cryptoAlgorithm = crypto.AlgorithmAES
	}
	
	fmt.Printf("Detected encryption algorithm for WebSocket connection: %s\n", encAlgorithm)
	
	// Create a session with the detected encryption algorithm
	err = protocolHandler.CreateSession(sessionID, cryptoAlgorithm)
	if err != nil {
		fmt.Printf("Error creating session for WebSocket data with %s: %v\n", encAlgorithm, err)
		return
	}
	
	// Get the client manager from the options if available
	if l.config.Options != nil {
		if clientManager, ok := l.config.Options["client_manager"]; ok {
			// Create a new client with the detected encryption algorithm
			config := client.DefaultConfig()
			config.ServerAddress = fmt.Sprintf("%s:%d", l.config.Address, l.config.Port)
			config.EncryptionAlg = encAlgorithm
			
			newClient, err := client.NewClient(config)
			if err != nil {
				fmt.Printf("Error creating client: %v\n", err)
			} else {
				// Register the client with the client manager
				if cm, ok := clientManager.(interface{ RegisterClient(*client.Client) string }); ok {
					clientID := cm.RegisterClient(newClient)
					fmt.Printf("Registered WebSocket client with ID %s using %s encryption\n", clientID, encAlgorithm)
				}
			}
		}
	}
	
	// Handle packet based on type
	var responseData []byte
	
	switch packet.Header.Type {
	case protocol.PacketTypeKeyExchange:
		fmt.Printf("Received key exchange from %s via WebSocket\n", conn.RemoteAddr())
		// Create a response packet with the same session ID
		responsePacket := protocol.NewPacket(protocol.PacketTypeKeyExchange, []byte(string(sessionID)))
		responseData = protocol.EncodePacket(responsePacket)
		
	case protocol.PacketTypeHeartbeat:
		fmt.Printf("Received heartbeat from %s via WebSocket\n", conn.RemoteAddr())
		// Create a heartbeat response
		responsePacket := protocol.NewPacket(protocol.PacketTypeHeartbeat, []byte("pong"))
		responseData = protocol.EncodePacket(responsePacket)
		
	default:
		fmt.Printf("Received packet type %d from %s via WebSocket\n", packet.Header.Type, conn.RemoteAddr())
		// Echo back the packet for other types
		responseData = protocol.EncodePacket(packet)
	}
	
	// Clean up
	protocolHandler.RemoveSession(sessionID)
	
	// Send the response
	err = conn.WriteMessage(messageType, responseData)
	if err != nil {
		fmt.Printf("Error writing WebSocket message: %v\n", err)
	}
}

// Broadcast sends a message to all connected clients
func (l *WebSocketListener) Broadcast(message []byte) {
	l.clientLock.RLock()
	defer l.clientLock.RUnlock()
	
	for client := range l.clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			fmt.Printf("Error broadcasting to client: %v\n", err)
			client.Close()
		}
	}
}
