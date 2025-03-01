package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"dinoc2/pkg/crypto"
	"dinoc2/pkg/protocol"
)

// HTTPConnection implements an HTTP connection to the server
type HTTPConnection struct {
	*BaseConnection
	client    *http.Client
	userAgent string
}

// NewHTTPConnection creates a new HTTP connection
func NewHTTPConnection(serverAddress string, protocolHandler *protocol.ProtocolHandler, sessionID crypto.SessionID) (*HTTPConnection, error) {
	base := NewBaseConnection(serverAddress, protocolHandler, sessionID, ProtocolHTTP)

	// Create HTTP client with reasonable timeouts
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
			DisableCompression:  true, // Avoid compression to maintain packet integrity
		},
	}

	return &HTTPConnection{
		BaseConnection: base,
		client:         client,
		userAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	}, nil
}

// Connect establishes an HTTP connection to the server
func (c *HTTPConnection) Connect() error {
	// For HTTP, we don't maintain a persistent connection
	// Instead, we'll verify the server is reachable
	
	// Create a test request
	req, err := http.NewRequest("GET", c.serverAddress, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Connection", "keep-alive")

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	c.connected = true

	// Perform initial handshake
	err = c.performHandshake()
	if err != nil {
		c.Close()
		return fmt.Errorf("handshake failed: %w", err)
	}

	return nil
}

// Close closes the HTTP connection
func (c *HTTPConnection) Close() error {
	// For HTTP, we don't maintain a persistent connection
	c.connected = false
	return nil
}

// SendPacket sends a packet over the HTTP connection
func (c *HTTPConnection) SendPacket(packet *protocol.Packet) error {
	if !c.connected {
		return fmt.Errorf("not connected")
	}

	// Prepare the packet for sending
	fragments, err := c.protocolHandler.PrepareOutgoingPacket(packet, c.sessionID, true)
	if err != nil {
		return fmt.Errorf("failed to prepare packet: %w", err)
	}

	// Send each fragment as a separate HTTP request
	for _, fragment := range fragments {
		// Create HTTP request
		req, err := http.NewRequest("POST", c.serverAddress+"/data", bytes.NewReader(fragment))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("X-Session-ID", string(c.sessionID))

		// Send request
		resp, err := c.client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	}

	return nil
}

// ReceivePacket receives a packet from the HTTP connection
func (c *HTTPConnection) ReceivePacket() (*protocol.Packet, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected")
	}

	// Create HTTP request to poll for data
	req, err := http.NewRequest("GET", c.serverAddress+"/poll", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("X-Session-ID", string(c.sessionID))

	// Send request with short timeout
	c.client.Timeout = 100 * time.Millisecond
	resp, err := c.client.Do(req)
	if err != nil {
		// If timeout, return nil without error
		if err, ok := err.(interface{ Timeout() bool }); ok && err.Timeout() {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Reset timeout
	c.client.Timeout = 30 * time.Second

	// Check response status
	if resp.StatusCode == http.StatusNoContent {
		// No data available
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	// Process the packet
	packet, err := c.protocolHandler.ProcessIncomingPacket(data, c.sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to process packet: %w", err)
	}

	return packet, nil
}

// performHandshake performs the initial handshake with the server
func (c *HTTPConnection) performHandshake() error {
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
