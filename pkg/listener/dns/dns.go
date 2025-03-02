package dns

import (
	"dinoc2/pkg/client"
	"dinoc2/pkg/crypto"
	"dinoc2/pkg/protocol"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// DNSListener implements a DNS listener for the C2 server
type DNSListener struct {
	config        map[string]interface{}
	address       string
	port          int
	server        *dns.Server
	clientManager interface{}
	isRunning     bool
}

// NewDNSListener creates a new DNS listener
func NewDNSListener(config map[string]interface{}) (*DNSListener, error) {
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

	return &DNSListener{
		config:    options,
		address:   address,
		port:      port,
		isRunning: false,
	}, nil
}

// SetClientManager sets the client manager for the listener
func (l *DNSListener) SetClientManager(cm interface{}) {
	l.clientManager = cm
	fmt.Printf("DEBUG: DNS listener client manager set: %T\n", cm)
}

// Start starts the DNS listener
func (l *DNSListener) Start() error {
	if l.isRunning {
		return fmt.Errorf("listener is already running")
	}

	// Create DNS server
	dns.HandleFunc(".", l.handleDNSRequest)
	
	l.server = &dns.Server{
		Addr: fmt.Sprintf("%s:%d", l.address, l.port),
		Net:  "udp",
	}

	// Start server in a goroutine
	go func() {
		err := l.server.ListenAndServe()
		if err != nil {
			fmt.Printf("DNS server error: %v\n", err)
		}
	}()

	l.isRunning = true
	fmt.Printf("DNS listener started on %s:%d\n", l.address, l.port)

	return nil
}

// Stop stops the DNS listener
func (l *DNSListener) Stop() error {
	if !l.isRunning {
		return nil
	}

	// Close server
	if l.server != nil {
		err := l.server.Shutdown()
		if err != nil {
			return fmt.Errorf("failed to close DNS server: %w", err)
		}
	}

	l.isRunning = false
	fmt.Printf("DNS listener stopped\n")

	return nil
}

// handleDNSRequest handles a DNS request
func (l *DNSListener) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	// Create response message
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	// Process each question
	for _, q := range r.Question {
		switch q.Qtype {
		case dns.TypeTXT:
			// Extract data from the domain name
			data := l.extractDataFromDomain(q.Name)
			if data == "" {
				continue
			}

			// Process the data
			response := l.processDNSData(data, w.RemoteAddr())

			// Add TXT record to response
			rr := &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				Txt: []string{response},
			}
			m.Answer = append(m.Answer, rr)
		}
	}

	// Send response
	err := w.WriteMsg(m)
	if err != nil {
		fmt.Printf("Failed to send DNS response: %v\n", err)
	}
}

// extractDataFromDomain extracts data from a domain name
func (l *DNSListener) extractDataFromDomain(domain string) string {
	// Remove trailing dot
	domain = strings.TrimSuffix(domain, ".")
	
	// Split domain into parts
	parts := strings.Split(domain, ".")
	
	// Check if domain has enough parts
	if len(parts) < 3 {
		return ""
	}
	
	// Extract data from subdomain
	return parts[0]
}

// processDNSData processes data from a DNS request
func (l *DNSListener) processDNSData(data string, addr net.Addr) string {
	// Create a protocol handler for processing the data
	protocolHandler := protocol.NewProtocolHandler()
	
	// Generate a unique session ID
	sessionID := crypto.GenerateSessionID()
	
	// Create a new client with the connection information
	newClient := client.NewClient(sessionID, addr.String(), client.ProtocolDNS)
	
	// Register client with client manager if available
	if l.clientManager != nil {
		fmt.Printf("DEBUG: DNS client manager type: %T\n", l.clientManager)
		
		// Try to register client using type assertion
		if cm, ok := l.clientManager.(*client.Manager); ok {
			clientID := cm.RegisterClient(newClient)
			fmt.Printf("Registered DNS client with ID %s\n", clientID)
		} else {
			fmt.Printf("Client manager does not implement RegisterClient method or is not of type *client.Manager\n")
		}
	} else {
		fmt.Printf("No client manager available\n")
	}

	// Simple echo response for testing
	return "Echo: " + data
}
