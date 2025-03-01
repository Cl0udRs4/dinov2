package test

import (
	"fmt"
	"testing"
	"time"

	"dinoc2/pkg/listener/icmp"
)

func TestICMPListener(t *testing.T) {
	// Skip test if not running as root
	if !icmp.CheckPrivileges() {
		t.Skip("Skipping ICMP listener test: requires root privileges")
	}
	
	// Create an ICMP listener with test configuration
	config := icmp.ICMPConfig{
		ListenAddress: "127.0.0.1",
	}
	
	listener := icmp.NewICMPListener(config)
	
	// Start the listener
	err := listener.Start()
	if err != nil {
		t.Fatalf("Failed to start ICMP listener: %v", err)
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
		t.Fatalf("Failed to stop ICMP listener: %v", err)
	}
	
	// Verify the listener is stopped
	if status := listener.Status(); status != "stopped" {
		t.Errorf("Expected listener status to be 'stopped', got '%s'", status)
	}
	
	fmt.Println("ICMP listener test completed successfully")
}
