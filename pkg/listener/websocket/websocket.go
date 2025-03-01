package websocket

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
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
	// In a real implementation, this would pass the message to the protocol layer
	fmt.Printf("Received WebSocket message from %s: %s\n", conn.RemoteAddr(), string(message))
	
	// Echo the message back for now
	err := conn.WriteMessage(messageType, message)
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
