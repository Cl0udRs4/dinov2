package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	
	"dinoc2/pkg/server"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "", "Path to configuration file")
	createConfig := flag.Bool("create-config", false, "Create a default configuration file")
	configOutput := flag.String("config-output", "config.json", "Output path for the default configuration file")
	flag.Parse()

	// Create default configuration if requested
	if *createConfig {
		outputPath := *configOutput
		if !filepath.IsAbs(outputPath) {
			// Convert to absolute path if relative
			cwd, err := os.Getwd()
			if err != nil {
				log.Fatalf("Failed to get current working directory: %v", err)
			}
			outputPath = filepath.Join(cwd, outputPath)
		}

		fmt.Printf("Creating default configuration file at: %s\n", outputPath)
		if err := server.CreateDefaultConfig(outputPath); err != nil {
			log.Fatalf("Failed to create default configuration: %v", err)
		}
		fmt.Println("Default configuration created successfully.")
		return
	}

	// Create a new server
	srv := server.NewServer()

	// Load configuration from file if provided
	if *configFile != "" {
		fmt.Println("Loading configuration from:", *configFile)
		if err := srv.LoadConfig(*configFile); err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
	} else {
		log.Println("No configuration file provided. Use -config to specify a configuration file or -create-config to create a default one.")
		log.Println("Starting with no listeners.")
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the server
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	fmt.Println("C2 Server started. Press Ctrl+C to exit.")

	// Wait for termination signal
	<-sigChan
	fmt.Println("\nShutting down server...")

	// Perform clean shutdown
	if err := srv.Shutdown(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	fmt.Println("Server shutdown complete.")
}
