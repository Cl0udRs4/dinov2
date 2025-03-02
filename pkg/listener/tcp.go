package listener

import (
	"fmt"
	"log"
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

	log.Printf("[TCP] New connection from %s\n", conn.RemoteAddr())

	// Create a simple protocol handler for this connection
	protocolHandler := protocol.NewProtocolHandler()
	
	// Generate a session ID based on the connection address
	sessionID := crypto.SessionID(conn.RemoteAddr().String())
	log.Printf("[TCP] Generated session ID: %s for connection from %s\n", sessionID, conn.RemoteAddr())
	
	// Create a session with AES encryption (default, will be updated based on client handshake)
	err := protocolHandler.CreateSession(sessionID, crypto.AlgorithmAES)
	if err != nil {
		log.Printf("[TCP] Error creating session: %v for %s\n", err, conn.RemoteAddr())
		return
	}
	log.Printf("[TCP] Created initial session with AES encryption for %s\n", conn.RemoteAddr())
	
	// Read the first packet to determine the encryption algorithm
	buffer := make([]byte, 4096)
	log.Printf("[TCP] Waiting for data from %s\n", conn.RemoteAddr())
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("[TCP] Error reading from connection %s: %v\n", conn.RemoteAddr(), err)
		return
	}
	log.Printf("[TCP] Received %d bytes from %s\n", n, conn.RemoteAddr())
	
	// Decode the packet to get the encryption algorithm
	packet, err := protocol.DecodePacket(buffer[:n])
	if err != nil {
		log.Printf("[TCP] Error decoding packet from %s: %v\n", conn.RemoteAddr(), err)
		return
	}
	
	log.Printf("[TCP] Received packet with encryption algorithm: %d from %s\n", packet.Header.EncAlgorithm, conn.RemoteAddr())
	
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
		log.Printf("[TCP] Warning: Unknown encryption algorithm %d from %s, defaulting to AES\n", 
			packet.Header.EncAlgorithm, conn.RemoteAddr())
		encAlgorithm = "aes" // Default to AES if not specified
		cryptoAlgorithm = crypto.AlgorithmAES
	}
	
	log.Printf("[TCP] Detected encryption algorithm: %s from %s\n", encAlgorithm, conn.RemoteAddr())
	
	// Always update the session with the detected encryption algorithm to ensure consistency
	// Remove the existing session
	protocolHandler.RemoveSession(sessionID)
	log.Printf("[TCP] Removed existing session for %s\n", conn.RemoteAddr())
	
	// Create a new session with the detected encryption algorithm
	err = protocolHandler.CreateSession(sessionID, cryptoAlgorithm)
	if err != nil {
		log.Printf("[TCP] Error creating session with %s for %s: %v\n", 
			encAlgorithm, conn.RemoteAddr(), err)
		return
	}
	
	log.Printf("[TCP] Successfully created session with encryption algorithm: %s for %s\n", 
		encAlgorithm, conn.RemoteAddr())
	
	// Get the client manager from the listener manager
	if clientManager, ok := l.config.Options["client_manager"]; ok {
		log.Printf("[TCP] Retrieved client manager from listener options for %s\n", conn.RemoteAddr())
		
		// Check if client manager is nil
		if clientManager == nil {
			log.Printf("[TCP] Error: Client manager is nil for %s\n", conn.RemoteAddr())
			return
		}
		
		// Create a new client with the detected encryption algorithm
		config := client.DefaultConfig()
		config.ServerAddress = l.config.Address
		config.EncryptionAlg = encAlgorithm
		
		// Set remote address in client config
		config.RemoteAddress = conn.RemoteAddr().String()
		
		log.Printf("[TCP] Creating new client with server address %s and encryption %s for %s\n", 
			l.config.Address, encAlgorithm, conn.RemoteAddr())
		newClient, err := client.NewClient(config)
		if err != nil {
			log.Printf("[TCP] Error creating client for %s: %v\n", conn.RemoteAddr(), err)
			return
		}
		
		// Set the session ID for the client
		newClient.SetSessionID(string(sessionID))
		log.Printf("[TCP] Created client with encryption algorithm: %s and session ID: %s for %s\n", 
			encAlgorithm, sessionID, conn.RemoteAddr())
		
		// Register the client with the client manager
		if cm, ok := clientManager.(interface{ RegisterClient(*client.Client) string }); ok {
			clientID := cm.RegisterClient(newClient)
			if clientID == "" {
				log.Printf("[TCP] Error: Failed to register client for %s (empty client ID returned)\n", conn.RemoteAddr())
				return
			}
			
			log.Printf("[TCP] Successfully registered client with ID %s using %s encryption for %s\n", 
				clientID, encAlgorithm, conn.RemoteAddr())
			
			// Store the client ID for later use
			clientIDStr := clientID
			
			// Process the initial packet
			processedPacket, err := protocolHandler.ProcessIncomingPacket(buffer[:n], sessionID)
			if err != nil {
				log.Printf("[TCP] Error processing initial packet from %s: %v\n", conn.RemoteAddr(), err)
				return
			}
			
			log.Printf("[TCP] Successfully processed initial packet from %s\n", conn.RemoteAddr())
			
			// Handle the packet based on its type
			switch processedPacket.Header.Type {
			case protocol.PacketTypeHeartbeat:
				log.Printf("[TCP] Received heartbeat from client %s (%s)\n", clientIDStr, conn.RemoteAddr())
				// Send heartbeat response
				heartbeatResponse := protocol.NewPacket(protocol.PacketTypeHeartbeat, []byte("heartbeat-response"))
				responseData, err := protocolHandler.PrepareOutgoingPacket(heartbeatResponse, sessionID, true)
				if err != nil {
					log.Printf("[TCP] Error preparing heartbeat response for %s: %v\n", conn.RemoteAddr(), err)
					return
				}
				
				log.Printf("[TCP] Sending heartbeat response to %s\n", conn.RemoteAddr())
				
				// Send the response
				for _, fragment := range responseData {
					_, err = conn.Write(fragment)
					if err != nil {
						log.Printf("[TCP] Error sending heartbeat response to %s: %v\n", conn.RemoteAddr(), err)
						return
					}
				}
				
				log.Printf("[TCP] Successfully sent heartbeat response to %s\n", conn.RemoteAddr())
			default:
				log.Printf("[TCP] Received packet of type %d from client %s (%s)\n", 
					processedPacket.Header.Type, clientIDStr, conn.RemoteAddr())
			}
		} else {
			log.Printf("[TCP] Error: Client manager does not implement RegisterClient method for %s\n", conn.RemoteAddr())
		}
	} else {
		log.Printf("[TCP] Error: Client manager not found in listener options for %s\n", conn.RemoteAddr())
	}

	// Handle communication loop
	for {
		// Read length prefix
		lengthBytes := make([]byte, 2)
		log.Printf("[TCP] Waiting for data from %s\n", conn.RemoteAddr())
		_, err := conn.Read(lengthBytes)
		if err != nil {
			log.Printf("[TCP] Connection closed from %s: %v\n", conn.RemoteAddr(), err)
			break
		}

		// Parse length
		length := uint16(lengthBytes[0])<<8 | uint16(lengthBytes[1])
		log.Printf("[TCP] Received packet with length %d from %s\n", length, conn.RemoteAddr())

		// Read packet data
		data := make([]byte, length)
		_, err = conn.Read(data)
		if err != nil {
			log.Printf("[TCP] Error reading packet data from %s: %v\n", conn.RemoteAddr(), err)
			break
		}
		log.Printf("[TCP] Successfully read %d bytes of packet data from %s\n", len(data), conn.RemoteAddr())

		// Decode the packet
		packet, err := protocol.DecodePacket(data)
		if err != nil {
			log.Printf("[TCP] Error decoding packet from %s: %v\n", conn.RemoteAddr(), err)
			continue
		}
		log.Printf("[TCP] Successfully decoded packet of type %d from %s\n", packet.Header.Type, conn.RemoteAddr())

		// Handle packet based on type
		var responsePacket *protocol.Packet
		
		switch packet.Header.Type {
		case protocol.PacketTypeKeyExchange:
			// Handle key exchange (handshake)
			log.Printf("[TCP] Received key exchange from %s\n", conn.RemoteAddr())
			
			// Create a response packet with the same session ID
			responsePacket = protocol.NewPacket(protocol.PacketTypeKeyExchange, []byte(string(sessionID)))
			log.Printf("[TCP] Created key exchange response for %s\n", conn.RemoteAddr())
			
		case protocol.PacketTypeHeartbeat:
			// Handle heartbeat
			log.Printf("[TCP] Received heartbeat from %s\n", conn.RemoteAddr())
			
			// Create a heartbeat response
			responsePacket = protocol.NewPacket(protocol.PacketTypeHeartbeat, []byte("pong"))
			log.Printf("[TCP] Created heartbeat response for %s\n", conn.RemoteAddr())
			
		default:
			// For now, just echo back the packet for other types
			log.Printf("[TCP] Received packet type %d from %s\n", packet.Header.Type, conn.RemoteAddr())
			responsePacket = packet
		}

		// Encode the response packet
		responseData := protocol.EncodePacket(responsePacket)
		log.Printf("[TCP] Encoded response packet of type %d for %s\n", responsePacket.Header.Type, conn.RemoteAddr())
		
		// Send length prefix
		responseLengthBytes := []byte{byte(len(responseData) >> 8), byte(len(responseData))}
		_, err = conn.Write(responseLengthBytes)
		if err != nil {
			log.Printf("[TCP] Error sending length prefix to %s: %v\n", conn.RemoteAddr(), err)
			break
		}

		// Send response data
		_, err = conn.Write(responseData)
		if err != nil {
			log.Printf("[TCP] Error sending response to %s: %v\n", conn.RemoteAddr(), err)
			break
		}

		log.Printf("[TCP] Successfully sent response to %s\n", conn.RemoteAddr())
	}
	
	// Clean up
	protocolHandler.RemoveSession(sessionID)
	log.Printf("[TCP] Cleaned up session for %s\n", conn.RemoteAddr())
}
