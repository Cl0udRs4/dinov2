package test

import (
	"fmt"
	"testing"
	"time"

	"dinoc2/pkg/listener/websocket"
)

func TestWebSocketListener(t *testing.T) {
	// Create a WebSocket listener with test configuration
	config := websocket.WebSocketConfig{
		Address: "127.0.0.1",
		Port:    8444,
		Path:    "/ws",
	}
	
	listener := websocket.NewWebSocketListener(config)
	
	// Start the listener
	err := listener.Start()
	if err != nil {
		t.Fatalf("Failed to start WebSocket listener: %v", err)
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
		t.Fatalf("Failed to stop WebSocket listener: %v", err)
	}
	
	// Verify the listener is stopped
	if status := listener.Status(); status != "stopped" {
		t.Errorf("Expected listener status to be 'stopped', got '%s'", status)
	}
	
	fmt.Println("WebSocket listener test completed successfully")
}
