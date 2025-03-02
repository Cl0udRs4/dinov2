package listener

import (
	"fmt"
	"net"
	"time"
)

// TCPListener implements a TCP listener for the C2 server
type TCPListener struct {
	config        map[string]interface{}
	address       string
	port          int
	listener      net.Listener
	clientManager interface{}
	isRunning     bool
}

// NewTCPListener creates a new TCP listener
func NewTCPListener(config ListenerConfig) (*TCPListener, error) {
	// Extract address and port
	address := config.Address
	port := config.Port
	
	if address == "" {
		return nil, fmt.Errorf("invalid configuration: listener address is required")
	}
	
	if port <= 0 {
		return nil, fmt.Errorf("invalid configuration: listener port is required")
	}

	return &TCPListener{
		config:    config.Options,
		address:   address,
		port:      port,
		isRunning: false,
	}, nil
}

// SetClientManager sets the client manager for the listener
func (l *TCPListener) SetClientManager(cm interface{}) {
	l.clientManager = cm
	fmt.Printf("DEBUG: TCP listener client manager set: %T\n", cm)
}

// Start starts the TCP listener
func (l *TCPListener) Start() error {
	if l.isRunning {
		return fmt.Errorf("listener is already running")
	}

	// Create TCP listener
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", l.address, l.port))
	if err != nil {
		return fmt.Errorf("failed to create TCP listener: %w", err)
	}

	l.listener = listener
	l.isRunning = true

	fmt.Printf("TCP listener started on %s:%d\n", l.address, l.port)

	// Start accepting connections
	go l.acceptConnections()

	return nil
}

// Stop stops the TCP listener
func (l *TCPListener) Stop() error {
	if !l.isRunning {
		return nil
	}

	// Close listener
	if l.listener != nil {
		err := l.listener.Close()
		if err != nil {
			return fmt.Errorf("failed to close TCP listener: %w", err)
		}
	}

	l.isRunning = false
	fmt.Printf("TCP listener stopped\n")

	return nil
}

// acceptConnections accepts incoming TCP connections
func (l *TCPListener) acceptConnections() {
	for l.isRunning {
		// Accept connection
		conn, err := l.listener.Accept()
		if err != nil {
			if l.isRunning {
				fmt.Printf("Failed to accept connection: %v\n", err)
			}
			continue
		}

		// Handle connection in a new goroutine
		go l.handleConnection(conn)
	}
}

// handleConnection handles a TCP connection
func (l *TCPListener) handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Printf("New connection from %s\n", conn.RemoteAddr())

	// Register client with client manager if available
	if l.clientManager != nil {
		fmt.Printf("DEBUG: Client manager type: %T\n", l.clientManager)
		
		// Create a simple client object with the connection information
		client := struct {
			Address string
			ID      string
		}{
			Address: conn.RemoteAddr().String(),
			ID:      fmt.Sprintf("client-%d", time.Now().UnixNano()),
		}
		
		// Register client using type assertion
		if cm, ok := l.clientManager.(interface{ RegisterClient(interface{}) string }); ok {
			clientID := cm.RegisterClient(client)
			fmt.Printf("Registered client with ID %s\n", clientID)
		} else {
			fmt.Printf("Client manager does not implement RegisterClient method\n")
		}
	} else {
		fmt.Printf("No client manager available\n")
	}

	// Simple echo server for testing
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Connection closed: %v\n", err)
			break
		}

		_, err = conn.Write(buffer[:n])
		if err != nil {
			fmt.Printf("Failed to send response: %v\n", err)
			break
		}
	}
}
