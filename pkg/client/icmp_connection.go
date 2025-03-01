package client

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"dinoc2/pkg/crypto"
	"dinoc2/pkg/protocol"
)

// ICMPConnection implements an ICMP connection to the server
type ICMPConnection struct {
	*BaseConnection
	conn       *icmp.PacketConn
	serverIP   net.IP
	sequenceID int
}

// NewICMPConnection creates a new ICMP connection
func NewICMPConnection(serverAddress string, protocolHandler *protocol.ProtocolHandler, sessionID crypto.SessionID) (*ICMPConnection, error) {
	base := NewBaseConnection(serverAddress, protocolHandler, sessionID, ProtocolICMP)

	// Resolve server IP
	ips, err := net.LookupIP(serverAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve server address: %w", err)
	}

	// Find IPv4 address
	var serverIP net.IP
	for _, ip := range ips {
		if ip.To4() != nil {
			serverIP = ip
			break
		}
	}

	if serverIP == nil {
		return nil, fmt.Errorf("no IPv4 address found for server")
	}

	return &ICMPConnection{
		BaseConnection: base,
		serverIP:       serverIP,
		sequenceID:     1,
	}, nil
}

// Connect establishes an ICMP connection to the server
func (c *ICMPConnection) Connect() error {
	// Open ICMP connection
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return fmt.Errorf("failed to listen for ICMP packets: %w", err)
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

// Close closes the ICMP connection
func (c *ICMPConnection) Close() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.connected = false
		return err
	}
	return nil
}

// SendPacket sends a packet over the ICMP connection
func (c *ICMPConnection) SendPacket(packet *protocol.Packet) error {
	if !c.connected || c.conn == nil {
		return fmt.Errorf("not connected")
	}

	// Prepare the packet for sending
	fragments, err := c.protocolHandler.PrepareOutgoingPacket(packet, c.sessionID, true)
	if err != nil {
		return fmt.Errorf("failed to prepare packet: %w", err)
	}

	// Send each fragment as a separate ICMP echo request
	for _, fragment := range fragments {
		// Create ICMP message
		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   os.Getpid() & 0xffff,
				Seq:  c.sequenceID,
				Data: fragment,
			},
		}
		c.sequenceID++

		// Marshal message
		msgBytes, err := msg.Marshal(nil)
		if err != nil {
			return fmt.Errorf("failed to marshal ICMP message: %w", err)
		}

		// Send message
		_, err = c.conn.WriteTo(msgBytes, &net.IPAddr{IP: c.serverIP})
		if err != nil {
			return fmt.Errorf("failed to send ICMP message: %w", err)
		}
	}

	return nil
}

// ReceivePacket receives a packet from the ICMP connection
func (c *ICMPConnection) ReceivePacket() (*protocol.Packet, error) {
	if !c.connected || c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Set read deadline
	err := c.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	if err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Read ICMP message
	buffer := make([]byte, 1500)
	n, peer, err := c.conn.ReadFrom(buffer)
	if err != nil {
		// If timeout, return nil without error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read ICMP message: %w", err)
	}

	// Verify peer
	if peer.String() != c.serverIP.String() {
		// Ignore messages from other sources
		return nil, nil
	}

	// Parse ICMP message
	msg, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), buffer[:n])
	if err != nil {
		return nil, fmt.Errorf("failed to parse ICMP message: %w", err)
	}

	// Verify message type
	if msg.Type != ipv4.ICMPTypeEchoReply {
		// Ignore non-echo-reply messages
		return nil, nil
	}

	// Extract data
	echo, ok := msg.Body.(*icmp.Echo)
	if !ok {
		return nil, fmt.Errorf("unexpected ICMP body type")
	}

	// Process the packet
	packet, err := c.protocolHandler.ProcessIncomingPacket(echo.Data, c.sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to process packet: %w", err)
	}

	return packet, nil
}

// performHandshake performs the initial handshake with the server
func (c *ICMPConnection) performHandshake() error {
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
