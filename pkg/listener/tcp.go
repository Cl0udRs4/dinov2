package listener

import (
	"fmt"
	"net"
	"sync"
	
	"dinoc2/pkg/client"
	"dinoc2/pkg/crypto"
	"dinoc2/pkg/protocol"
)

// TCPListener implements the Listener interface for TCP protocol
type TCPListener struct {
	config     ListenerConfig
	listener   net.Listener
	status     ListenerStatus
	statusLock sync.RWMutex
	stopChan   chan struct{}
}

// NewTCPListener creates a new TCP listener
func NewTCPListener(config ListenerConfig) *TCPListener {
	return &TCPListener{
		config:   config,
		status:   StatusStopped,
		stopChan: make(chan struct{}),
	}
}

// Start implements the Listener interface
func (l *TCPListener) Start() error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status == StatusRunning {
		return fmt.Errorf("listener is already running")
	}

	addr := fmt.Sprintf("%s:%d", l.config.Address, l.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		l.status = StatusError
		return fmt.Errorf("failed to start TCP listener on %s: %w", addr, err)
	}

	l.listener = listener
	l.status = StatusRunning
	l.stopChan = make(chan struct{})

	// Start accepting connections in a goroutine
	go l.acceptConnections()

	return nil
}

// Stop implements the Listener interface
func (l *TCPListener) Stop() error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status != StatusRunning {
		return nil // Already stopped
	}

	// Signal the accept loop to stop
	close(l.stopChan)

	// Close the listener
	if l.listener != nil {
		err := l.listener.Close()
		if err != nil {
			l.status = StatusError
			return fmt.Errorf("error closing TCP listener: %w", err)
		}
	}

	l.status = StatusStopped
	return nil
}

// Status implements the Listener interface
func (l *TCPListener) Status() ListenerStatus {
	l.statusLock.RLock()
	defer l.statusLock.RUnlock()
	return l.status
}

// UpdateStats updates the listener statistics
func (l *TCPListener) UpdateStats(stats map[string]interface{}) {
	// This method can be used to update statistics from the connection handler
	// For example, incrementing connection count, bytes sent/received, etc.
}

// Configure implements the Listener interface
func (l *TCPListener) Configure(config ListenerConfig) error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status == StatusRunning {
		return fmt.Errorf("cannot configure a running listener")
	}

	l.config = config
	return nil
}

// acceptConnections handles incoming TCP connections
func (l *TCPListener) acceptConnections() {
	for {
		select {
		case <-l.stopChan:
			return
		default:
			conn, err := l.listener.Accept()
			if err != nil {
				select {
				case <-l.stopChan:
					// Listener was closed intentionally, not an error
					return
				default:
					// Actual error occurred
					l.statusLock.Lock()
					l.status = StatusError
					l.statusLock.Unlock()
					fmt.Printf("Error accepting connection: %v\n", err)
					return
				}
			}

			// Handle the connection in a new goroutine
			go l.handleConnection(conn)
		}
	}
}

// handleConnection processes a client connection
func (l *TCPListener) handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Printf("New connection from %s\n", conn.RemoteAddr())

	// Create a simple protocol handler for this connection
	protocolHandler := protocol.NewProtocolHandler()
	
	// Generate a session ID based on the connection address
	sessionID := crypto.SessionID(conn.RemoteAddr().String())
	
	// Create a session with AES encryption (default, will be updated based on client handshake)
	err := protocolHandler.CreateSession(sessionID, crypto.AlgorithmAES)
	if err != nil {
		fmt.Printf("Error creating session: %v\n", err)
		return
	}
	
	// Read the first packet to determine the encryption algorithm
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("Error reading from connection: %v\n", err)
		return
	}
	
	// Decode the packet to get the encryption algorithm
	packet, err := protocol.DecodePacket(buffer[:n])
	if err != nil {
		fmt.Printf("Error decoding packet: %v\n", err)
		return
	}
	
	fmt.Printf("Received packet with encryption algorithm: %d\n", packet.Header.EncAlgorithm)
	
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
		fmt.Printf("Warning: Unknown encryption algorithm %d, defaulting to AES\n", packet.Header.EncAlgorithm)
		encAlgorithm = "aes" // Default to AES if not specified
		cryptoAlgorithm = crypto.AlgorithmAES
	}
	
	fmt.Printf("Detected encryption algorithm: %s\n", encAlgorithm)
	
	// Always update the session with the detected encryption algorithm to ensure consistency
	// Remove the existing session
	protocolHandler.RemoveSession(sessionID)
	
	// Create a new session with the detected encryption algorithm
	err = protocolHandler.CreateSession(sessionID, cryptoAlgorithm)
	if err != nil {
		fmt.Printf("Error creating session with %s: %v\n", encAlgorithm, err)
		return
	}
	
	fmt.Printf("Successfully created session with encryption algorithm: %s\n", encAlgorithm)
	
	// Get the client manager from the listener manager
	if clientManager, ok := l.config.Options["client_manager"]; ok {
		// Create a new client with the detected encryption algorithm
		config := client.DefaultConfig()
		config.ServerAddress = l.config.Address
		config.EncryptionAlg = encAlgorithm
		
		newClient, err := client.NewClient(config)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			return
		}
		
		// Set the session ID using reflection to avoid accessing unexported fields
		// This is a temporary solution until the client package is updated
		fmt.Printf("Created client with encryption algorithm: %s\n", encAlgorithm)
		
		// Register the client with the client manager
		if cm, ok := clientManager.(interface{ RegisterClient(*client.Client) string }); ok {
			clientID := cm.RegisterClient(newClient)
			fmt.Printf("Registered client with ID %s using %s encryption\n", clientID, encAlgorithm)
			
			// Store the client ID for later use
			clientIDStr := clientID
			
			// Process the initial packet
			processedPacket, err := protocolHandler.ProcessIncomingPacket(buffer[:n], sessionID)
			if err != nil {
				fmt.Printf("Error processing initial packet: %v\n", err)
				return
			}
			
			// Handle the packet based on its type
			switch processedPacket.Header.Type {
			case protocol.PacketTypeHeartbeat:
				fmt.Printf("Received heartbeat from client %s\n", clientIDStr)
				// Send heartbeat response
				heartbeatResponse := protocol.NewPacket(protocol.PacketTypeHeartbeat, []byte("heartbeat-response"))
				responseData, err := protocolHandler.PrepareOutgoingPacket(heartbeatResponse, sessionID, true)
				if err != nil {
					fmt.Printf("Error preparing heartbeat response: %v\n", err)
					return
				}
				
				// Send the response
				for _, fragment := range responseData {
					_, err = conn.Write(fragment)
					if err != nil {
						fmt.Printf("Error sending heartbeat response: %v\n", err)
						return
					}
				}
			default:
				fmt.Printf("Received packet of type %d from client %s\n", processedPacket.Header.Type, clientIDStr)
			}
		} else {
			fmt.Printf("Client manager does not implement RegisterClient method\n")
		}
	} else {
		fmt.Printf("Client manager not found in listener options\n")
	}

	// Handle communication loop
	for {
		// Read length prefix
		lengthBytes := make([]byte, 2)
		_, err := conn.Read(lengthBytes)
		if err != nil {
			fmt.Printf("Connection closed: %v\n", err)
			break
		}

		// Parse length
		length := uint16(lengthBytes[0])<<8 | uint16(lengthBytes[1])

		// Read packet data
		data := make([]byte, length)
		_, err = conn.Read(data)
		if err != nil {
			fmt.Printf("Error reading packet data: %v\n", err)
			break
		}

		// Decode the packet
		packet, err := protocol.DecodePacket(data)
		if err != nil {
			fmt.Printf("Error decoding packet: %v\n", err)
			continue
		}

		// Handle packet based on type
		var responsePacket *protocol.Packet
		
		switch packet.Header.Type {
		case protocol.PacketTypeKeyExchange:
			// Handle key exchange (handshake)
			fmt.Printf("Received key exchange from %s\n", conn.RemoteAddr())
			
			// Create a response packet with the same session ID
			responsePacket = protocol.NewPacket(protocol.PacketTypeKeyExchange, []byte(string(sessionID)))
			
		case protocol.PacketTypeHeartbeat:
			// Handle heartbeat
			fmt.Printf("Received heartbeat from %s\n", conn.RemoteAddr())
			
			// Create a heartbeat response
			responsePacket = protocol.NewPacket(protocol.PacketTypeHeartbeat, []byte("pong"))
			
		default:
			// For now, just echo back the packet for other types
			fmt.Printf("Received packet type %d from %s\n", packet.Header.Type, conn.RemoteAddr())
			responsePacket = packet
		}

		// Encode the response packet
		responseData := protocol.EncodePacket(responsePacket)
		
		// Send length prefix
		responseLengthBytes := []byte{byte(len(responseData) >> 8), byte(len(responseData))}
		_, err = conn.Write(responseLengthBytes)
		if err != nil {
			fmt.Printf("Error sending length prefix: %v\n", err)
			break
		}

		// Send response data
		_, err = conn.Write(responseData)
		if err != nil {
			fmt.Printf("Error sending response: %v\n", err)
			break
		}

		fmt.Printf("Sent response to %s\n", conn.RemoteAddr())
	}
	
	// Clean up
	protocolHandler.RemoveSession(sessionID)
}
