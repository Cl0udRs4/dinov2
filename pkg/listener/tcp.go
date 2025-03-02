package listener

import (
	"dinoc2/pkg/client"
	"dinoc2/pkg/crypto"
	"dinoc2/pkg/protocol"
	"fmt"
	"log"
	"net"
	"sync"
)

// TCPListener implements a TCP listener for the C2 server
type TCPListener struct {
	id            string
	address       string
	port          int
	listener      net.Listener
	connections   map[net.Conn]bool
	connMutex     sync.RWMutex
	clientManager interface{}
	isRunning     bool
}

// NewTCPListener creates a new TCP listener
func NewTCPListener(config ListenerConfig) (*TCPListener, error) {
	return &TCPListener{
		id:          config.ID,
		address:     config.Address,
		port:        config.Port,
		connections: make(map[net.Conn]bool),
		isRunning:   false,
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
	
	// Start accepting connections in a goroutine
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
	
	// Close all connections
	l.connMutex.Lock()
	for conn := range l.connections {
		conn.Close()
	}
	l.connections = make(map[net.Conn]bool)
	l.connMutex.Unlock()
	
	l.isRunning = false
	
	return nil
}

// acceptConnections accepts incoming TCP connections
func (l *TCPListener) acceptConnections() {
	for l.isRunning {
		// Accept connection
		conn, err := l.listener.Accept()
		if err != nil {
			if l.isRunning {
				log.Printf("Failed to accept TCP connection: %v", err)
			}
			continue
		}
		
		// Add connection to map
		l.connMutex.Lock()
		l.connections[conn] = true
		l.connMutex.Unlock()
		
		// Handle connection in a goroutine
		go l.handleConnection(conn)
	}
}

// handleConnection handles a TCP connection
func (l *TCPListener) handleConnection(conn net.Conn) {
	// Defer connection cleanup
	defer func() {
		// Close connection
		conn.Close()
		
		// Remove connection from map
		l.connMutex.Lock()
		delete(l.connections, conn)
		l.connMutex.Unlock()
	}()
	
	// Create a protocol handler for processing the data
	protocolHandler := protocol.NewProtocolHandler()
	
	// Generate a unique session ID
	sessionID := crypto.GenerateSessionID()
	
	// Create a session with AES encryption (default, will be updated based on client handshake)
	err := protocolHandler.CreateSession(sessionID, crypto.AlgorithmAES)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		return
	}
	
	// Create a new client with the connection information
	newClient := client.NewClient(sessionID, conn.RemoteAddr().String(), client.ProtocolTCP)
	
	// Register client with client manager if available
	if l.clientManager != nil {
		fmt.Printf("DEBUG: Client manager type: %T\n", l.clientManager)
		
		// Try to register client using type assertion
		if cm, ok := l.clientManager.(*client.Manager); ok {
			clientID := cm.RegisterClient(newClient)
			fmt.Printf("Registered client with ID %s\n", clientID)
		} else {
			fmt.Printf("Client manager does not implement RegisterClient method or is not of type *client.Manager\n")
		}
	} else {
		fmt.Printf("No client manager available\n")
	}
	
	// Buffer for reading data
	buffer := make([]byte, 4096)
	
	// Read loop
	for {
		// Read data
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("TCP read error: %v", err)
			break
		}
		
		// Process data
		data := buffer[:n]
		
		// Process the incoming packet
		packet, err := protocol.DecodePacket(data)
		if err != nil {
			log.Printf("Failed to decode packet: %v", err)
			continue
		}
		
		// Update client last seen timestamp
		newClient.UpdateLastSeen()
		
		// Prepare a response packet
		responsePacket := &protocol.Packet{
			Header: protocol.PacketHeader{
				Version:      protocol.ProtocolVersion,
				EncAlgorithm: protocol.EncryptionAlgorithmNone,
				Type:         protocol.PacketTypeResponse,
				TaskID:       packet.Header.TaskID,
				Checksum:     0, // Will be calculated during encoding
			},
			Data: []byte("OK"),
		}
		
		// Encode response packet
		responseData, err := protocol.EncodePacket(responsePacket)
		if err != nil {
			log.Printf("Failed to encode response packet: %v", err)
			continue
		}
		
		// Send response
		_, err = conn.Write(responseData)
		if err != nil {
			log.Printf("TCP write error: %v", err)
			break
		}
	}
}
