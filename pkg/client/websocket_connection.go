package client

import (
	"net"
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/websocket"

	"dinoc2/pkg/crypto"
	"dinoc2/pkg/protocol"
)

// WebSocketConnection implements a WebSocket connection to the server
type WebSocketConnection struct {
	*BaseConnection
	conn *websocket.Conn
}

// NewWebSocketConnection creates a new WebSocket connection
func NewWebSocketConnection(serverAddress string, protocolHandler *protocol.ProtocolHandler, sessionID crypto.SessionID) (*WebSocketConnection, error) {
	base := NewBaseConnection(serverAddress, protocolHandler, sessionID, ProtocolWebSocket)
	return &WebSocketConnection{
		BaseConnection: base,
	}, nil
}

// Connect establishes a WebSocket connection to the server
func (c *WebSocketConnection) Connect() error {
	// Parse server address
	u, err := url.Parse(c.serverAddress)
	if err != nil {
		return fmt.Errorf("invalid server address: %w", err)
	}

	// Ensure WebSocket scheme
	if u.Scheme != "ws" && u.Scheme != "wss" {
		u.Scheme = "ws"
	}

	// Add session ID to URL
	q := u.Query()
	q.Set("session", string(c.sessionID))
	u.RawQuery = q.Encode()

	// Connect to WebSocket server
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	conn, _, err := dialer.Dial(u.String(), nil)
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

// Close closes the WebSocket connection
func (c *WebSocketConnection) Close() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.connected = false
		return err
	}
	return nil
}

// SendPacket sends a packet over the WebSocket connection
func (c *WebSocketConnection) SendPacket(packet *protocol.Packet) error {
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
		err := c.conn.WriteMessage(websocket.BinaryMessage, fragment)
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	}

	return nil
}

// ReceivePacket receives a packet from the WebSocket connection
func (c *WebSocketConnection) ReceivePacket() (*protocol.Packet, error) {
	if !c.connected || c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Set read deadline
	err := c.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	if err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Read message
	messageType, data, err := c.conn.ReadMessage()
	if err != nil {
		// If timeout, return nil without error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	// Verify message type
	if messageType != websocket.BinaryMessage {
		return nil, fmt.Errorf("unexpected message type: %d", messageType)
	}

	// Process the packet
	packet, err := c.protocolHandler.ProcessIncomingPacket(data, c.sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to process packet: %w", err)
	}

	return packet, nil
}

// performHandshake performs the initial handshake with the server
func (c *WebSocketConnection) performHandshake() error {
	// Create handshake packet
	handshake := protocol.NewPacket(protocol.PacketTypeKeyExchange, []byte(string(c.sessionID)))

	// Send handshake packet
	err := c.SendPacket(handshake)
	if err != nil {
		return fmt.Errorf("failed to send handshake: %w", err)
	}

	// Wait for response
	for i := 0; i < 5; i++ {
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
