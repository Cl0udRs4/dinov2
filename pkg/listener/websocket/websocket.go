package websocket

import (
	"fmt"
	"log"
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
			log.Printf("[WebSocket] Error starting WebSocket listener: %v\n", err)
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
	log.Printf("[WebSocket] New connection request from %s\n", r.RemoteAddr)
	
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := l.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] Error upgrading to WebSocket for %s: %v\n", r.RemoteAddr, err)
		return
	}
	log.Printf("[WebSocket] Successfully upgraded connection to WebSocket for %s\n", conn.RemoteAddr())
	
	// Register the client
	l.clientLock.Lock()
	l.clients[conn] = true
	l.clientLock.Unlock()
	log.Printf("[WebSocket] Registered WebSocket connection for %s\n", conn.RemoteAddr())
	
	// Handle the connection in a goroutine
	go l.handleConnection(conn)
}

// handleConnection processes a WebSocket connection
func (l *WebSocketListener) handleConnection(conn *websocket.Conn) {
	log.Printf("[WebSocket] Starting to handle connection from %s\n", conn.RemoteAddr())
	
	defer func() {
		// Unregister the client when the function returns
		l.clientLock.Lock()
		delete(l.clients, conn)
		l.clientLock.Unlock()
		log.Printf("[WebSocket] Unregistered WebSocket connection for %s\n", conn.RemoteAddr())
		
		// Close the connection
		conn.Close()
		log.Printf("[WebSocket] Closed WebSocket connection for %s\n", conn.RemoteAddr())
	}()
	
	for {
		// Read message from the client
		log.Printf("[WebSocket] Waiting for message from %s\n", conn.RemoteAddr())
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			// Check if it's a normal close
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] Unexpected error for %s: %v\n", conn.RemoteAddr(), err)
			} else {
				log.Printf("[WebSocket] Connection closed for %s: %v\n", conn.RemoteAddr(), err)
			}
			break
		}
		log.Printf("[WebSocket] Received message of type %d with %d bytes from %s\n", 
			messageType, len(message), conn.RemoteAddr())
		
		// Process the message
		go l.processMessage(conn, messageType, message)
	}
}

// processMessage processes a WebSocket message
func (l *WebSocketListener) processMessage(conn *websocket.Conn, messageType int, message []byte) {
	log.Printf("[WebSocket] Processing message from %s\n", conn.RemoteAddr())
	
	// Create a protocol handler for processing the data
	protocolHandler := protocol.NewProtocolHandler()
	
	// Generate a session ID based on the connection address
	sessionID := crypto.SessionID(conn.RemoteAddr().String())
	log.Printf("[WebSocket] Generated session ID: %s for connection from %s\n", sessionID, conn.RemoteAddr())
	
	// Decode the packet to get the encryption algorithm
	packet, err := protocol.DecodePacket(message)
	if err != nil {
		log.Printf("[WebSocket] Error decoding WebSocket packet data from %s: %v\n", conn.RemoteAddr(), err)
		// Echo the message back for invalid packets
		conn.WriteMessage(messageType, message)
		return
	}
	log.Printf("[WebSocket] Successfully decoded packet of type %d from %s\n", packet.Header.Type, conn.RemoteAddr())
	
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
		log.Printf("[WebSocket] Warning: Unknown encryption algorithm %d from %s, defaulting to AES\n", 
			packet.Header.EncAlgorithm, conn.RemoteAddr())
		encAlgorithm = "aes" // Default to AES if not specified
		cryptoAlgorithm = crypto.AlgorithmAES
	}
	
	log.Printf("[WebSocket] Detected encryption algorithm: %s from %s\n", encAlgorithm, conn.RemoteAddr())
	
	// Create a session with the detected encryption algorithm
	err = protocolHandler.CreateSession(sessionID, cryptoAlgorithm)
	if err != nil {
		log.Printf("[WebSocket] Error creating session for WebSocket data with %s for %s: %v\n", 
			encAlgorithm, conn.RemoteAddr(), err)
		return
	}
	log.Printf("[WebSocket] Successfully created session with encryption algorithm: %s for %s\n", 
		encAlgorithm, conn.RemoteAddr())
	
	// Get the client manager from the options if available
	if l.config.Options != nil {
		if clientManager, ok := l.config.Options["client_manager"]; ok {
			log.Printf("[WebSocket] Retrieved client manager from listener options for %s\n", conn.RemoteAddr())
			
			// Check if client manager is nil
			if clientManager == nil {
				log.Printf("[WebSocket] Error: Client manager is nil for %s\n", conn.RemoteAddr())
			} else {
				// Create a new client with the detected encryption algorithm
				config := client.DefaultConfig()
				config.ServerAddress = fmt.Sprintf("%s:%d", l.config.Address, l.config.Port)
				config.EncryptionAlg = encAlgorithm
				config.RemoteAddress = conn.RemoteAddr().String()
				
				log.Printf("[WebSocket] Creating new client with server address %s and encryption %s for %s\n", 
					config.ServerAddress, encAlgorithm, conn.RemoteAddr())
				newClient, err := client.NewClient(config)
				if err != nil {
					log.Printf("[WebSocket] Error creating client for %s: %v\n", conn.RemoteAddr(), err)
				} else {
					// Set the session ID for the client
					newClient.SetSessionID(string(sessionID))
					log.Printf("[WebSocket] Created client with encryption algorithm: %s and session ID: %s for %s\n", 
						encAlgorithm, sessionID, conn.RemoteAddr())
					
					// Register the client with the client manager
					if cm, ok := clientManager.(interface{ RegisterClient(*client.Client) string }); ok {
						clientID := cm.RegisterClient(newClient)
						if clientID == "" {
							log.Printf("[WebSocket] Error: Failed to register client for %s (empty client ID returned)\n", conn.RemoteAddr())
						} else {
							log.Printf("[WebSocket] Successfully registered client with ID %s using %s encryption for %s\n", 
								clientID, encAlgorithm, conn.RemoteAddr())
						}
					} else {
						log.Printf("[WebSocket] Error: Client manager does not implement RegisterClient method for %s\n", conn.RemoteAddr())
					}
				}
			}
		} else {
			log.Printf("[WebSocket] Error: Client manager not found in listener options for %s\n", conn.RemoteAddr())
		}
	} else {
		log.Printf("[WebSocket] Error: No options configured for WebSocket listener for %s\n", conn.RemoteAddr())
	}
	
	// Handle packet based on type
	var responseData []byte
	
	switch packet.Header.Type {
	case protocol.PacketTypeKeyExchange:
		log.Printf("[WebSocket] Received key exchange from %s\n", conn.RemoteAddr())
		// Create a response packet with the same session ID
		responsePacket := protocol.NewPacket(protocol.PacketTypeKeyExchange, []byte(string(sessionID)))
		responseData = protocol.EncodePacket(responsePacket)
		log.Printf("[WebSocket] Created key exchange response for %s\n", conn.RemoteAddr())
		
	case protocol.PacketTypeHeartbeat:
		log.Printf("[WebSocket] Received heartbeat from %s\n", conn.RemoteAddr())
		// Create a heartbeat response
		responsePacket := protocol.NewPacket(protocol.PacketTypeHeartbeat, []byte("pong"))
		responseData = protocol.EncodePacket(responsePacket)
		log.Printf("[WebSocket] Created heartbeat response for %s\n", conn.RemoteAddr())
		
	default:
		log.Printf("[WebSocket] Received packet type %d from %s\n", packet.Header.Type, conn.RemoteAddr())
		// Echo back the packet for other types
		responseData = protocol.EncodePacket(packet)
		log.Printf("[WebSocket] Created echo response for packet type %d for %s\n", packet.Header.Type, conn.RemoteAddr())
	}
	
	// Clean up
	protocolHandler.RemoveSession(sessionID)
	log.Printf("[WebSocket] Cleaned up session for %s\n", conn.RemoteAddr())
	
	// Send the response
	err = conn.WriteMessage(messageType, responseData)
	if err != nil {
		log.Printf("[WebSocket] Error writing WebSocket message to %s: %v\n", conn.RemoteAddr(), err)
	} else {
		log.Printf("[WebSocket] Successfully sent response to %s\n", conn.RemoteAddr())
	}
}

// Broadcast sends a message to all connected clients
func (l *WebSocketListener) Broadcast(message []byte) {
	l.clientLock.RLock()
	defer l.clientLock.RUnlock()
	
	log.Printf("[WebSocket] Broadcasting message of %d bytes to %d clients\n", len(message), len(l.clients))
	
	for client := range l.clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("[WebSocket] Error broadcasting to client %s: %v\n", client.RemoteAddr(), err)
			client.Close()
			log.Printf("[WebSocket] Closed connection to client %s due to broadcast error\n", client.RemoteAddr())
		} else {
			log.Printf("[WebSocket] Successfully sent broadcast message to client %s\n", client.RemoteAddr())
		}
	}
}
