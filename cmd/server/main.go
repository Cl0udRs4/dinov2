package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"dinoc2/pkg/server"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.json", "Path to configuration file")
	apiEnabled := flag.Bool("api", true, "Enable API server")
	apiAddress := flag.String("api-address", "0.0.0.0", "API server address")
	apiPort := flag.Int("api-port", 8081, "API server port")
	flag.Parse()

	// Load configuration
	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set API server configuration
	config.APIServer.Enabled = *apiEnabled
	config.APIServer.Address = *apiAddress
	config.APIServer.Port = *apiPort

	// Create server
	s := server.NewServer(config)

	// Start server
	if err := s.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait for interrupt signal
	s.WaitForInterrupt()
}

// loadConfig loads the server configuration from a file
func loadConfig(path string) (server.ServerConfig, error) {
	var config server.ServerConfig

	// Check if the configuration file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create a default configuration
		config = server.ServerConfig{
			Listeners: []server.ListenerConfig{
				{
					ID:      "tcp-listener",
					Type:    "tcp",
					Address: "0.0.0.0",
					Port:    8080,
					Enabled: true,
				},
			},
			APIServer: server.APIServerConfig{
				Enabled: true,
				Address: "0.0.0.0",
				Port:    8081,
			},
		}

		// Save the default configuration
		if err := saveConfig(path, config); err != nil {
			return config, fmt.Errorf("failed to save default configuration: %w", err)
		}

		log.Printf("Created default configuration at %s", path)
		return config, nil
	}

	// Read the configuration file
	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Parse the configuration
	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	return config, nil
}

// saveConfig saves the server configuration to a file
func saveConfig(path string, config server.ServerConfig) error {
	// Marshal the configuration
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write the configuration file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}
