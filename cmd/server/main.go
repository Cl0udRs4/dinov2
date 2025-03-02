package main

import (
	"dinoc2/pkg/client"
	"dinoc2/pkg/server"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()
	
	// Load configuration
	fmt.Printf("Loading configuration from: %s\n", *configFile)
	
	// Create server
	server, err := server.NewServer(*configFile)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	
	// Create client manager
	clientManager := client.NewManager()
	
	// Set client manager for server
	server.SetClientManager(clientManager)
	
	// Start server
	err = server.Start()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	
	// Wait for interrupt signal
	fmt.Println("C2 Server started. Press Ctrl+C to exit.")
	
	// Create channel for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	// Wait for signal
	<-sigChan
	
	// Stop server
	err = server.Stop()
	if err != nil {
		log.Fatalf("Failed to stop server: %v", err)
	}
	
	fmt.Println("Server stopped.")
}
