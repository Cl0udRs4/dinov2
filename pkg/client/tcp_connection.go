package client

import (
	"fmt"
	"net"
	"time"

	"dinoc2/pkg/crypto"
	"dinoc2/pkg/protocol"
)

// TCPConnection implements a TCP connection to the server
type TCPConnection struct {
	*BaseConnection
	conn net.Conn
}

// NewTCPConnection creates a new TCP connection
func NewTCPConnection(serverAddress string, protocolHandler *protocol.ProtocolHandler, sessionID crypto.SessionID) (*TCPConnection, error) {
	base := NewBaseConnection(serverAddress, protocolHandler, sessionID, ProtocolTCP)
	return &TCPConnection{
		BaseConnection: base,
	}, nil
}

// Connect establishes a TCP connection to the server
func (c *TCPConnection) Connect() error {
	// Parse server address
	if c.serverAddress == "" {
		return fmt.Errorf("server address is empty")
	}

	// Establish TCP connection
	conn, err := net.DialTimeout("tcp", c.serverAddress, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	c.conn = conn
	c.connected = true

	// Perform initial handshake
	err = c.performHandshake()
	if err != nil {
		c.Close()
		return fmt.Errorf("handshake failed: %w", err)
	}

	return nil
}

// Close closes the TCP connection
func (c *TCPConnection) Close() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.connected = false
		return err
	}
	return nil
}

// SendPacket sends a packet over the TCP connection
func (c *TCPConnection) SendPacket(packet *protocol.Packet) error {
	if !c.connected || c.conn == nil {
		return fmt.Errorf("not connected")
	}

	// Prepare the packet for sending
	fragments, err := c.protocolHandler.PrepareOutgoingPacket(packet, c.sessionID, true)
	if err != nil {
		return fmt.Errorf("failed to prepare packet: %w", err)
	}

	// Send each fragment
	for _, fragment := range fragments {
		// Add length prefix to each fragment
		length := uint16(len(fragment))
		lengthBytes := []byte{byte(length >> 8), byte(length)}

		// Send length prefix
		_, err = c.conn.Write(lengthBytes)
		if err != nil {
			return fmt.Errorf("failed to send length prefix: %w", err)
		}

		// Send fragment
		_, err = c.conn.Write(fragment)
		if err != nil {
			return fmt.Errorf("failed to send fragment: %w", err)
		}
	}

	return nil
}

// ReceivePacket receives a packet from the TCP connection
func (c *TCPConnection) ReceivePacket() (*protocol.Packet, error) {
	if !c.connected || c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Set read deadline
	err := c.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	if err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Read length prefix
	lengthBytes := make([]byte, 2)
	_, err = c.conn.Read(lengthBytes)
	if err != nil {
		// If timeout, return nil without error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read length prefix: %w", err)
	}

	// Parse length
	length := uint16(lengthBytes[0])<<8 | uint16(lengthBytes[1])

	// Read packet data
	data := make([]byte, length)
	_, err = c.conn.Read(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet data: %w", err)
	}

	// Process the packet
	packet, err := c.protocolHandler.ProcessIncomingPacket(data, c.sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to process packet: %w", err)
	}

	return packet, nil
}

// performHandshake performs the initial handshake with the server
func (c *TCPConnection) performHandshake() error {
	// Create handshake packet
	handshake := protocol.NewPacket(protocol.PacketTypeKeyExchange, []byte(string(c.sessionID)))

	// Send handshake packet
	err := c.SendPacket(handshake)
	if err != nil {
		return fmt.Errorf("failed to send handshake: %w", err)
	}

	// Wait for response
	for i := 0; i < 5; i++ {
		// Set read deadline
		err = c.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		if err != nil {
			return fmt.Errorf("failed to set read deadline: %w", err)
		}

		// Receive response
		response, err := c.ReceivePacket()
		if err != nil {
			return fmt.Errorf("failed to receive handshake response: %w", err)
		}

		if response == nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Verify response
		if response.Header.Type != protocol.PacketTypeKeyExchange {
			return fmt.Errorf("unexpected response type: %d", response.Header.Type)
		}

		// Handshake successful
		return nil
	}

	return fmt.Errorf("handshake timed out")
}
