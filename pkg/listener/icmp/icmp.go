package icmp

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"dinoc2/pkg/crypto"
	"dinoc2/pkg/protocol"
)

// ICMPListener implements the Listener interface for ICMP protocol
type ICMPListener struct {
	config     ICMPConfig
	conn       *icmp.PacketConn
	status     string
	statusLock sync.RWMutex
	stopChan   chan struct{}
}

// ICMPConfig holds configuration for the ICMP listener
type ICMPConfig struct {
	ListenAddress string
	Protocol      string // "icmp" or "udp"
	Options       map[string]interface{}
}

// NewICMPListener creates a new ICMP listener
func NewICMPListener(config ICMPConfig) *ICMPListener {
	// Set default values if not provided
	if config.Protocol == "" {
		// Default to privileged ICMP (requires root)
		config.Protocol = "icmp"
	}
	if config.ListenAddress == "" {
		config.ListenAddress = "0.0.0.0"
	}

	return &ICMPListener{
		config:   config,
		status:   "stopped",
		stopChan: make(chan struct{}),
	}
}

// Start implements the Listener interface
func (l *ICMPListener) Start() error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status == "running" {
		return fmt.Errorf("ICMP listener is already running")
	}

	// Create ICMP connection
	// Determine the correct network string format based on protocol
	var network string
	if l.config.Protocol == "icmp" {
		// For privileged raw ICMP endpoints
		network = "ip4:1" // Use protocol number 1 for ICMP
	} else {
		// For non-privileged UDP-based ICMP
		network = "udp4"
	}
	conn, err := icmp.ListenPacket(network, l.config.ListenAddress)
	if err != nil {
		l.status = "error"
		return fmt.Errorf("failed to start ICMP listener: %w", err)
	}

	l.conn = conn
	l.status = "running"
	l.stopChan = make(chan struct{})

	// Start listening for ICMP packets in a goroutine
	go l.listenForPackets()

	return nil
}

// Stop implements the Listener interface
func (l *ICMPListener) Stop() error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status != "running" {
		return nil // Already stopped
	}

	// Signal the listen goroutine to stop
	close(l.stopChan)

	// Create a new stop channel for future use
	l.stopChan = make(chan struct{})

	// Close the connection
	if l.conn != nil {
		err := l.conn.Close()
		if err != nil {
			l.status = "error"
			return fmt.Errorf("error closing ICMP listener: %w", err)
		}
	}

	l.status = "stopped"
	return nil
}

// Status implements the Listener interface
func (l *ICMPListener) Status() string {
	l.statusLock.RLock()
	defer l.statusLock.RUnlock()
	return l.status
}

// Configure implements the Listener interface
func (l *ICMPListener) Configure(config interface{}) error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status == "running" {
		return fmt.Errorf("cannot configure a running ICMP listener")
	}

	icmpConfig, ok := config.(ICMPConfig)
	if !ok {
		return fmt.Errorf("invalid configuration type for ICMP listener")
	}

	l.config = icmpConfig
	return nil
}

// listenForPackets listens for incoming ICMP packets
func (l *ICMPListener) listenForPackets() {
	buffer := make([]byte, 1500) // Standard MTU size

	for {
		select {
		case <-l.stopChan:
			return
		default:
			// Set read deadline to allow for checking the stop channel
			l.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))

			n, addr, err := l.conn.ReadFrom(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// This is just a timeout, continue
					continue
				}

				// Check if we're stopping
				select {
				case <-l.stopChan:
					return
				default:
					// Actual error occurred
					l.statusLock.Lock()
					l.status = "error"
					l.statusLock.Unlock()
					fmt.Printf("Error reading ICMP packet: %v\n", err)
					return
				}
			}

			// Process the packet
			go l.processPacket(buffer[:n], addr)
		}
	}
}

// processPacket processes an ICMP packet
func (l *ICMPListener) processPacket(packet []byte, addr net.Addr) {
	// Parse the ICMP message
	msg, err := icmp.ParseMessage(ipv4.ICMPTypeEcho.Protocol(), packet)
	if err != nil {
		fmt.Printf("Error parsing ICMP message: %v\n", err)
		return
	}

	// Check if it's an echo request
	if msg.Type == ipv4.ICMPTypeEcho {
		echo := msg.Body.(*icmp.Echo)

		// Extract data from the echo request
		data := echo.Data
		
		// Create a protocol handler for processing the data
		protocolHandler := protocol.NewProtocolHandler()
		
		// Generate a session ID based on the connection address and echo ID
		sessionID := crypto.SessionID(fmt.Sprintf("%s-%d", addr.String(), echo.ID))
		
		// Create a session with AES encryption (default)
		err := protocolHandler.CreateSession(sessionID, crypto.AlgorithmAES)
		if err != nil {
			fmt.Printf("Error creating session for ICMP data: %v\n", err)
			return
		}
		
		// Process the data if it's long enough to be a valid packet
		// HeaderSize is 12 bytes based on protocol/encoder.go
		if len(data) > 12 {
			// Decode the packet
			packet, err := protocol.DecodePacket(data)
			if err != nil {
				fmt.Printf("Error decoding ICMP packet data: %v\n", err)
			} else {
				// Handle packet based on type
				switch packet.Header.Type {
				case protocol.PacketTypeKeyExchange:
					fmt.Printf("Received key exchange from %s via ICMP\n", addr)
				case protocol.PacketTypeHeartbeat:
					fmt.Printf("Received heartbeat from %s via ICMP\n", addr)
				default:
					fmt.Printf("Received packet type %d from %s via ICMP\n", packet.Header.Type, addr)
				}
			}
		} else {
			fmt.Printf("Received ICMP echo request from %s (data too short for protocol)\n", addr)
		}
		
		// Clean up
		protocolHandler.RemoveSession(sessionID)
		
		// Send an echo reply
		l.sendEchoReply(addr, echo.ID, echo.Seq, echo.Data)
	}
}

// sendEchoReply sends an ICMP echo reply
func (l *ICMPListener) sendEchoReply(addr net.Addr, id, seq int, data []byte) {
	// Create an echo reply message
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEchoReply,
		Code: 0,
		Body: &icmp.Echo{
			ID:   id,
			Seq:  seq,
			Data: data,
		},
	}

	// Marshal the message
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		fmt.Printf("Error marshaling ICMP echo reply: %v\n", err)
		return
	}

	// Send the reply
	_, err = l.conn.WriteTo(msgBytes, addr)
	if err != nil {
		fmt.Printf("Error sending ICMP echo reply: %v\n", err)
	}
}

// RequiresPrivileges returns true if the listener requires elevated privileges
func (l *ICMPListener) RequiresPrivileges() bool {
	return l.config.Protocol == "icmp"
}

// CheckPrivileges checks if the process has the necessary privileges
func CheckPrivileges() bool {
	// On Unix systems, check if we're running as root
	return os.Geteuid() == 0
}
