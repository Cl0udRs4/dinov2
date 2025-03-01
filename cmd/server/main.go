package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"dinoc2/pkg/listener"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// TODO: Load configuration from file if provided
	if *configFile != "" {
		fmt.Println("Loading configuration from:", *configFile)
		// Implementation will be added later
	}

	// Create a new listener manager
	manager := listener.NewManager()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// TODO: Start web interface and other server components

	fmt.Println("C2 Server started. Press Ctrl+C to exit.")

	// Wait for termination signal
	<-sigChan
	fmt.Println("\nShutting down server...")

	// Cleanup and stop all listeners
	if err := manager.StopAll(); err != nil {
		log.Printf("Error stopping listeners: %v", err)
	}

	fmt.Println("Server shutdown complete.")
}
