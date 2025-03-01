package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"
)

// MemoryProtectionOptions configures memory protection behavior
type MemoryProtectionOptions struct {
	EnableEncryption     bool
	EnableObfuscation    bool
	EnableIntegrityChecks bool
	EnableCanaries       bool
	EncryptionAlgorithm  string
	CanarySize           int
	IntegrityCheckInterval int
}

// DefaultMemoryProtectionOptions returns default memory protection options
func DefaultMemoryProtectionOptions() MemoryProtectionOptions {
	return MemoryProtectionOptions{
		EnableEncryption:     true,
		EnableObfuscation:    true,
		EnableIntegrityChecks: true,
		EnableCanaries:       true,
		EncryptionAlgorithm:  "aes-256-gcm",
		CanarySize:           16,
		IntegrityCheckInterval: 60, // seconds
	}
}

// MemoryProtection implements memory protection techniques
type MemoryProtection struct {
	options     MemoryProtectionOptions
	mutex       sync.RWMutex
	protectedData map[string]*ProtectedMemory
	canaries    map[uintptr][]byte
	stopChan    chan struct{}
}

// ProtectedMemory represents a protected memory region
type ProtectedMemory struct {
	Key         []byte
	IV          []byte
	Ciphertext  []byte
	Plaintext   []byte
	Checksum    []byte
	IsEncrypted bool
}

// NewMemoryProtection creates a new memory protection with the specified options
func NewMemoryProtection(options MemoryProtectionOptions) *MemoryProtection {
	mp := &MemoryProtection{
		options:     options,
		protectedData: make(map[string]*ProtectedMemory),
		canaries:    make(map[uintptr][]byte),
		stopChan:    make(chan struct{}),
	}

	// Start integrity check goroutine if enabled
	if options.EnableIntegrityChecks {
		go mp.integrityCheckLoop()
	}

	return mp
}

// Protect encrypts and protects sensitive data in memory
func (m *MemoryProtection) Protect(id string, data []byte) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Create a new protected memory entry
	protected := &ProtectedMemory{
		Plaintext: make([]byte, len(data)),
		IsEncrypted: false,
	}

	// Copy the plaintext data
	copy(protected.Plaintext, data)

	// Encrypt the data if encryption is enabled
	if m.options.EnableEncryption {
		// Generate encryption key
		key := make([]byte, 32) // 256-bit key
		if _, err := io.ReadFull(rand.Reader, key); err != nil {
			return fmt.Errorf("failed to generate encryption key: %w", err)
		}
		protected.Key = key

		// Generate IV
		iv := make([]byte, 12) // 96-bit IV for GCM
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return fmt.Errorf("failed to generate IV: %w", err)
		}
		protected.IV = iv

		// Create cipher
		block, err := aes.NewCipher(key)
		if err != nil {
			return fmt.Errorf("failed to create cipher: %w", err)
		}

		// Create GCM mode
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return fmt.Errorf("failed to create GCM: %w", err)
		}

		// Encrypt the data
		protected.Ciphertext = gcm.Seal(nil, iv, protected.Plaintext, nil)
		protected.IsEncrypted = true

		// Clear the plaintext from memory
		for i := range protected.Plaintext {
			protected.Plaintext[i] = 0
		}
	}

	// Calculate checksum for integrity checks
	if m.options.EnableIntegrityChecks {
		var dataToHash []byte
		if protected.IsEncrypted {
			dataToHash = protected.Ciphertext
		} else {
			dataToHash = protected.Plaintext
		}
		protected.Checksum = m.calculateChecksum(dataToHash)
	}

	// Store the protected data
	m.protectedData[id] = protected

	return nil
}

// Access decrypts and provides access to protected data
func (m *MemoryProtection) Access(id string) ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Get the protected data
	protected, exists := m.protectedData[id]
	if !exists {
		return nil, fmt.Errorf("protected data not found: %s", id)
	}

	// If the data is not encrypted, return the plaintext
	if !protected.IsEncrypted {
		return protected.Plaintext, nil
	}

	// Decrypt the data
	block, err := aes.NewCipher(protected.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt the data
	plaintext, err := gcm.Open(nil, protected.IV, protected.Ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// Remove securely removes protected data from memory
func (m *MemoryProtection) Remove(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Get the protected data
	protected, exists := m.protectedData[id]
	if !exists {
		return fmt.Errorf("protected data not found: %s", id)
	}

	// Securely clear all sensitive data
	if protected.Key != nil {
		for i := range protected.Key {
			protected.Key[i] = 0
		}
	}

	if protected.IV != nil {
		for i := range protected.IV {
			protected.IV[i] = 0
		}
	}

	if protected.Plaintext != nil {
		for i := range protected.Plaintext {
			protected.Plaintext[i] = 0
		}
	}

	if protected.Ciphertext != nil {
		for i := range protected.Ciphertext {
			protected.Ciphertext[i] = 0
		}
	}

	if protected.Checksum != nil {
		for i := range protected.Checksum {
			protected.Checksum[i] = 0
		}
	}

	// Remove the entry
	delete(m.protectedData, id)

	// Force garbage collection to clean up memory
	runtime.GC()

	return nil
}

// AddCanary adds a memory canary at the specified address
func (m *MemoryProtection) AddCanary(addr uintptr) {
	if !m.options.EnableCanaries {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Generate random canary value
	canary := make([]byte, m.options.CanarySize)
	if _, err := io.ReadFull(rand.Reader, canary); err != nil {
		return
	}

	// Store the canary
	m.canaries[addr] = canary

	// In a real implementation, this would write the canary to the specified address
	// This is a placeholder since direct memory manipulation is unsafe in Go
}

// CheckCanaries checks all memory canaries for tampering
func (m *MemoryProtection) CheckCanaries() bool {
	if !m.options.EnableCanaries {
		return true
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// In a real implementation, this would check each canary against its stored value
	// This is a placeholder since direct memory manipulation is unsafe in Go
	return true
}

// calculateChecksum calculates a checksum for integrity checks
func (m *MemoryProtection) calculateChecksum(data []byte) []byte {
	// Use SHA-256 for a cryptographic hash
	hash := sha256.Sum256(data)
	return hash[:]
}

// VerifyIntegrity verifies the integrity of all protected data
func (m *MemoryProtection) VerifyIntegrity() bool {
	if !m.options.EnableIntegrityChecks {
		return true
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for id, protected := range m.protectedData {
		// Calculate current checksum
		var dataToHash []byte
		if protected.IsEncrypted {
			dataToHash = protected.Ciphertext
		} else {
			dataToHash = protected.Plaintext
		}
		currentChecksum := m.calculateChecksum(dataToHash)

		// Compare with stored checksum
		if !compareChecksums(currentChecksum, protected.Checksum) {
			fmt.Printf("Integrity check failed for %s\n", id)
			return false
		}
	}

	return true
}

// compareChecksums compares two checksums for equality
func compareChecksums(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	// Constant-time comparison to prevent timing attacks
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

// integrityCheckLoop runs periodic integrity checks
func (m *MemoryProtection) integrityCheckLoop() {
	ticker := time.NewTicker(time.Duration(m.options.IntegrityCheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !m.VerifyIntegrity() {
				// Handle integrity violation
				// In a real implementation, this might trigger an alert or shutdown
				fmt.Println("Memory integrity violation detected!")
			}
		case <-m.stopChan:
			return
		}
	}
}

// Stop stops the memory protection
func (m *MemoryProtection) Stop() {
	close(m.stopChan)

	// Securely clear all protected data
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for id := range m.protectedData {
		m.Remove(id)
	}
}

// ObfuscateMemory applies obfuscation techniques to memory
func (m *MemoryProtection) ObfuscateMemory() {
	if !m.options.EnableObfuscation {
		return
	}

	// In a real implementation, this would:
	// - Apply code obfuscation techniques
	// - Implement control flow obfuscation
	// - Use instruction substitution
	// - Apply data transformation

	// Force garbage collection to make memory analysis harder
	runtime.GC()
}
