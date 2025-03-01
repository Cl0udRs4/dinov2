package test

import (
	"fmt"
	"testing"
	"time"

	"dinoc2/pkg/listener/dns"
)

func TestDNSListener(t *testing.T) {
	// Create a DNS listener with test configuration
	config := dns.DNSConfig{
		Address: "127.0.0.1",
		Port:    5353,
		Domain:  "test.example.com",
		TTL:     300,
	}
	
	listener := dns.NewDNSListener(config)
	
	// Start the listener
	err := listener.Start()
	if err != nil {
		t.Fatalf("Failed to start DNS listener: %v", err)
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
		t.Fatalf("Failed to stop DNS listener: %v", err)
	}
	
	// Verify the listener is stopped
	if status := listener.Status(); status != "stopped" {
		t.Errorf("Expected listener status to be 'stopped', got '%s'", status)
	}
	
	fmt.Println("DNS listener test completed successfully")
}
