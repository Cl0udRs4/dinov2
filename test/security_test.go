package test

import (
	"bytes"
	"dinoc2/pkg/security"
	"fmt"
	"testing"
	"time"
)

func TestSecurityFeatures(t *testing.T) {
	// Create security manager
	options := security.DefaultSecurityManagerOptions()
	options.SecurityViolationCallback = func(violation string) {
		fmt.Printf("Security violation detected: %s\n", violation)
	}

	manager := security.NewSecurityManager(options)
	err := manager.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize security manager: %v", err)
	}

	// Test memory protection
	testData := []byte("sensitive data")
	err = manager.ProtectData("test", testData)
	if err != nil {
		t.Fatalf("Failed to protect data: %v", err)
	}

	// Access protected data
	data, err := manager.AccessProtectedData("test")
	if err != nil {
		t.Fatalf("Failed to access protected data: %v", err)
	}

	// Verify data
	if string(data) != string(testData) {
		t.Fatalf("Protected data mismatch: expected %s, got %s", string(testData), string(data))
	}

	// Remove protected data
	err = manager.RemoveProtectedData("test")
	if err != nil {
		t.Fatalf("Failed to remove protected data: %v", err)
	}

	// Test traffic obfuscation
	originalData := []byte("test message")
	obfuscatedData, err := manager.ObfuscateTraffic(originalData)
	if err != nil {
		t.Fatalf("Failed to obfuscate traffic: %v", err)
	}

	// Deobfuscate traffic
	deobfuscatedData, err := manager.DeobfuscateTraffic(obfuscatedData)
	if err != nil {
		t.Fatalf("Failed to deobfuscate traffic: %v", err)
	}

	// Verify data - trim any null bytes or extra padding that might be present
	trimmedData := bytes.TrimRight(deobfuscatedData, "\x00")
	if string(trimmedData) != string(originalData) {
		t.Fatalf("Deobfuscated data mismatch: expected %s, got %s", string(originalData), string(deobfuscatedData))
	}

	// Run security checks
	result := manager.RunSecurityChecks()
	fmt.Printf("Security checks result: %v\n", result)

	// Get violations
	violations := manager.GetViolations()
	fmt.Printf("Security violations: %v\n", violations)

	// Start security monitoring
	manager.StartSecurityMonitoring(5 * time.Second)

	// Wait for a moment to allow monitoring to run
	time.Sleep(1 * time.Second)

	// Shutdown security manager
	manager.Shutdown()

	fmt.Println("Security features test completed successfully")
}
