package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dinoc2/pkg/protocol"
)

func main() {
	// Parse command line flags
	serverAddr := flag.String("server", "", "C2 server address")
	protocolList := flag.String("protocol", "tcp", "Comma-separated list of protocols to use (tcp,dns,icmp)")
	flag.Parse()

	if *serverAddr == "" {
		fmt.Println("Error: Server address is required")
		flag.Usage()
		os.Exit(1)
	}

	// TODO: Parse protocol list and initialize client connections

	fmt.Println("C2 Client started. Connected to server:", *serverAddr)
	fmt.Println("Using protocols:", *protocolList)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Simple heartbeat for now
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Wait for termination signal or heartbeat
	for {
		select {
		case <-ticker.C:
			fmt.Println("Heartbeat sent")
			// TODO: Implement actual heartbeat mechanism
		case <-sigChan:
			fmt.Println("\nShutting down client...")
			// TODO: Implement proper cleanup
			fmt.Println("Client shutdown complete.")
			return
		}
	}
}
