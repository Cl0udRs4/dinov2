package icmp

import (
	"dinoc2/pkg/client"
	"dinoc2/pkg/crypto"
	"dinoc2/pkg/protocol"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// ICMPListener implements an ICMP listener for the C2 server
type ICMPListener struct {
	config        map[string]interface{}
	address       string
	port          int
	conn          *icmp.PacketConn
	clientManager interface{}
	isRunning     bool
}

// NewICMPListener creates a new ICMP listener
func NewICMPListener(config map[string]interface{}) (*ICMPListener, error) {
	// Extract address and port
	address, ok := config["address"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid configuration: listener address is required")
	}
	
	portFloat, ok := config["port"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid configuration: listener port is required")
	}
	port := int(portFloat)
	
	// Extract options
	options, ok := config["options"].(map[string]interface{})
	if !ok {
		options = make(map[string]interface{})
	}

	return &ICMPListener{
		config:    options,
		address:   address,
		port:      port,
		isRunning: false,
	}, nil
}

// SetClientManager sets the client manager for the listener
func (l *ICMPListener) SetClientManager(cm interface{}) {
	l.clientManager = cm
	fmt.Printf("DEBUG: ICMP listener client manager set: %T\n", cm)
}

// Start starts the ICMP listener
func (l *ICMPListener) Start() error {
	if l.isRunning {
		return fmt.Errorf("listener is already running")
	}

	// Create ICMP connection
	conn, err := icmp.ListenPacket("ip4:icmp", l.address)
	if err != nil {
		return fmt.Errorf("failed to create ICMP listener: %w", err)
	}

	l.conn = conn
	l.isRunning = true

	fmt.Printf("ICMP listener started on %s\n", l.address)

	// Start listening for ICMP packets
	go l.listenForPackets()

	return nil
}

// Stop stops the ICMP listener
func (l *ICMPListener) Stop() error {
	if !l.isRunning {
		return nil
	}

	// Close connection
	if l.conn != nil {
		err := l.conn.Close()
		if err != nil {
			return fmt.Errorf("failed to close ICMP connection: %w", err)
		}
	}

	l.isRunning = false
	fmt.Printf("ICMP listener stopped\n")

	return nil
}

// listenForPackets listens for ICMP packets
func (l *ICMPListener) listenForPackets() {
	buffer := make([]byte, 1500)

	for l.isRunning {
		n, addr, err := l.conn.ReadFrom(buffer)
		if err != nil {
			if l.isRunning {
				fmt.Printf("Failed to read ICMP packet: %v\n", err)
			}
			continue
		}

		// Process packet in a new goroutine
		go l.processPacket(buffer[:n], addr)
	}
}

// processPacket processes an ICMP packet
func (l *ICMPListener) processPacket(data []byte, addr net.Addr) {
	// Parse ICMP message
	msg, err := icmp.ParseMessage(ipv4.ICMPTypeEcho.Protocol(), data)
	if err != nil {
		fmt.Printf("Failed to parse ICMP message: %v\n", err)
		return
	}

	// Only process echo requests
	if msg.Type != ipv4.ICMPTypeEcho {
		return
	}

	// Extract echo data
	echo, ok := msg.Body.(*icmp.Echo)
	if !ok {
		fmt.Printf("Failed to extract echo data\n")
		return
	}

	// Create a protocol handler for processing the data
	protocolHandler := protocol.NewProtocolHandler()
	
	// Generate a unique session ID
	sessionID := crypto.GenerateSessionID()
	
	// Create a new client with the connection information
	newClient := client.NewClient(sessionID, addr.String(), client.ProtocolICMP)
	
	// Register client with client manager if available
	if l.clientManager != nil {
		fmt.Printf("DEBUG: ICMP client manager type: %T\n", l.clientManager)
		
		// Try to register client using type assertion
		if cm, ok := l.clientManager.(*client.Manager); ok {
			clientID := cm.RegisterClient(newClient)
			fmt.Printf("Registered ICMP client with ID %s\n", clientID)
		} else {
			fmt.Printf("Client manager does not implement RegisterClient method or is not of type *client.Manager\n")
		}
	} else {
		fmt.Printf("No client manager available\n")
	}

	// Create echo reply
	reply := &icmp.Message{
		Type: ipv4.ICMPTypeEchoReply,
		Code: 0,
		Body: &icmp.Echo{
			ID:   echo.ID,
			Seq:  echo.Seq,
			Data: echo.Data,
		},
	}

	// Marshal reply
	replyBytes, err := reply.Marshal(nil)
	if err != nil {
		fmt.Printf("Failed to marshal ICMP reply: %v\n", err)
		return
	}

	// Send reply
	_, err = l.conn.WriteTo(replyBytes, addr)
	if err != nil {
		fmt.Printf("Failed to send ICMP reply: %v\n", err)
	}
}
