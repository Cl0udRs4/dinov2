package test

import (
	"fmt"
	"testing"
	"time"

	"dinoc2/pkg/listener/http"
)

func TestHTTPListener(t *testing.T) {
	// Create an HTTP listener with test configuration
	config := http.HTTPConfig{
		Address:  "127.0.0.1",
		Port:     8443,
		UseHTTP2: true,
	}
	
	listener := http.NewHTTPListener(config)
	
	// Start the listener
	err := listener.Start()
	if err != nil {
		t.Fatalf("Failed to start HTTP listener: %v", err)
	}
	
	// Verify the listener is running
	if status := listener.Status(); status != "running" {
		t.Errorf("Expected listener status to be 'running', got '%s'", status)
	}
	
	// Give it a moment to initialize
	time.Sleep(1 * time.Second)
	
	// Stop the listener
	err = listener.Stop()
	if err != nil {
		t.Fatalf("Failed to stop HTTP listener: %v", err)
	}
	
	// Verify the listener is stopped
	if status := listener.Status(); status != "stopped" {
		t.Errorf("Expected listener status to be 'stopped', got '%s'", status)
	}
	
	fmt.Println("HTTP listener test completed successfully")
}
