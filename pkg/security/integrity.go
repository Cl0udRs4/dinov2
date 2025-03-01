package security

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// IntegrityOptions configures integrity checking behavior
type IntegrityOptions struct {
	EnableSelfCheck      bool
	EnableFileCheck      bool
	EnableMemoryCheck    bool
	CheckInterval        time.Duration
	CriticalFiles        []string
	ExpectedHashes       map[string]string
	IntegrityViolationCallback func(violation string)
}

// DefaultIntegrityOptions returns default integrity checking options
func DefaultIntegrityOptions() IntegrityOptions {
	return IntegrityOptions{
		EnableSelfCheck:   true,
		EnableFileCheck:   true,
		EnableMemoryCheck: true,
		CheckInterval:     5 * time.Minute,
		CriticalFiles:     []string{},
		ExpectedHashes:    make(map[string]string),
		IntegrityViolationCallback: nil,
	}
}

// IntegrityChecker implements runtime integrity checking
type IntegrityChecker struct {
	options     IntegrityOptions
	mutex       sync.RWMutex
	stopChan    chan struct{}
	violations  []string
}

// NewIntegrityChecker creates a new integrity checker with the specified options
func NewIntegrityChecker(options IntegrityOptions) *IntegrityChecker {
	return &IntegrityChecker{
		options:    options,
		stopChan:   make(chan struct{}),
		violations: []string{},
	}
}

// Start starts the integrity checker
func (i *IntegrityChecker) Start() {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// Start integrity check goroutine
	go i.checkLoop()
}

// Stop stops the integrity checker
func (i *IntegrityChecker) Stop() {
	close(i.stopChan)
}

// AddCriticalFile adds a critical file to be checked
func (i *IntegrityChecker) AddCriticalFile(path string) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// Add file to critical files
	i.options.CriticalFiles = append(i.options.CriticalFiles, path)

	// Calculate and store hash
	hash, err := i.calculateFileHash(path)
	if err != nil {
		return fmt.Errorf("failed to calculate hash: %w", err)
	}

	i.options.ExpectedHashes[path] = hash

	return nil
}

// GetViolations returns a list of integrity violations
func (i *IntegrityChecker) GetViolations() []string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	// Create a copy of the violations slice
	violations := make([]string, len(i.violations))
	copy(violations, i.violations)

	return violations
}

// CheckIntegrity performs an integrity check
func (i *IntegrityChecker) CheckIntegrity() bool {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// Clear previous violations
	i.violations = []string{}

	// Check self integrity
	if i.options.EnableSelfCheck {
		if !i.checkSelfIntegrity() {
			i.violations = append(i.violations, "self integrity check failed")
		}
	}

	// Check file integrity
	if i.options.EnableFileCheck {
		if !i.checkFileIntegrity() {
			i.violations = append(i.violations, "file integrity check failed")
		}
	}

	// Check memory integrity
	if i.options.EnableMemoryCheck {
		if !i.checkMemoryIntegrity() {
			i.violations = append(i.violations, "memory integrity check failed")
		}
	}

	// Call violation callback if there are violations
	if len(i.violations) > 0 && i.options.IntegrityViolationCallback != nil {
		for _, violation := range i.violations {
			i.options.IntegrityViolationCallback(violation)
		}
	}

	return len(i.violations) == 0
}

// checkLoop runs periodic integrity checks
func (i *IntegrityChecker) checkLoop() {
	ticker := time.NewTicker(i.options.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			i.CheckIntegrity()
		case <-i.stopChan:
			return
		}
	}
}

// checkSelfIntegrity checks the integrity of the running executable
func (i *IntegrityChecker) checkSelfIntegrity() bool {
	// Get the path to the current executable
	exePath, err := os.Executable()
	if err != nil {
		i.violations = append(i.violations, fmt.Sprintf("failed to get executable path: %v", err))
		return false
	}

	// Check if the executable is in the critical files
	_, exists := i.options.ExpectedHashes[exePath]
	if !exists {
		// Calculate and store hash
		hash, err := i.calculateFileHash(exePath)
		if err != nil {
			i.violations = append(i.violations, fmt.Sprintf("failed to calculate executable hash: %v", err))
			return false
		}

		i.options.ExpectedHashes[exePath] = hash
		return true
	}

	// Calculate current hash
	currentHash, err := i.calculateFileHash(exePath)
	if err != nil {
		i.violations = append(i.violations, fmt.Sprintf("failed to calculate executable hash: %v", err))
		return false
	}

	// Compare with expected hash
	if currentHash != i.options.ExpectedHashes[exePath] {
		i.violations = append(i.violations, fmt.Sprintf("executable hash mismatch: %s", exePath))
		return false
	}

	return true
}

// checkFileIntegrity checks the integrity of critical files
func (i *IntegrityChecker) checkFileIntegrity() bool {
	allValid := true

	for _, path := range i.options.CriticalFiles {
		// Calculate current hash
		currentHash, err := i.calculateFileHash(path)
		if err != nil {
			i.violations = append(i.violations, fmt.Sprintf("failed to calculate file hash: %v", err))
			allValid = false
			continue
		}

		// Compare with expected hash
		expectedHash, exists := i.options.ExpectedHashes[path]
		if !exists {
			// Store hash if not exists
			i.options.ExpectedHashes[path] = currentHash
			continue
		}

		if currentHash != expectedHash {
			i.violations = append(i.violations, fmt.Sprintf("file hash mismatch: %s", path))
			allValid = false
		}
	}

	return allValid
}

// checkMemoryIntegrity checks the integrity of memory
func (i *IntegrityChecker) checkMemoryIntegrity() bool {
	// In a real implementation, this would check for:
	// - Memory tampering
	// - Code injection
	// - Hook detection
	// - Breakpoint detection

	// Force garbage collection to clean up memory
	runtime.GC()

	return true
}

// calculateFileHash calculates the SHA-256 hash of a file
func (i *IntegrityChecker) calculateFileHash(path string) (string, error) {
	// Open file
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Calculate hash
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	// Return hex-encoded hash
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// VerifyFileIntegrity verifies the integrity of a file
func (i *IntegrityChecker) VerifyFileIntegrity(path string) (bool, error) {
	i.mutex.RLock()
	expectedHash, exists := i.options.ExpectedHashes[path]
	i.mutex.RUnlock()

	if !exists {
		return false, fmt.Errorf("file not in integrity database")
	}

	// Calculate current hash
	currentHash, err := i.calculateFileHash(path)
	if err != nil {
		return false, fmt.Errorf("failed to calculate hash: %w", err)
	}

	return currentHash == expectedHash, nil
}

// UpdateFileHash updates the expected hash for a file
func (i *IntegrityChecker) UpdateFileHash(path string) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// Calculate hash
	hash, err := i.calculateFileHash(path)
	if err != nil {
		return fmt.Errorf("failed to calculate hash: %w", err)
	}

	// Update hash
	i.options.ExpectedHashes[path] = hash

	return nil
}

// SaveIntegrityDatabase saves the integrity database to a file
func (i *IntegrityChecker) SaveIntegrityDatabase(path string) error {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write critical files
	for _, criticalFile := range i.options.CriticalFiles {
		hash, exists := i.options.ExpectedHashes[criticalFile]
		if !exists {
			continue
		}

		_, err := fmt.Fprintf(file, "%s:%s\n", criticalFile, hash)
		if err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	return nil
}

// LoadIntegrityDatabase loads the integrity database from a file
func (i *IntegrityChecker) LoadIntegrityDatabase(path string) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// Open file
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Parse line
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		filePath := parts[0]
		hash := parts[1]

		// Add to critical files if not already present
		found := false
		for _, criticalFile := range i.options.CriticalFiles {
			if criticalFile == filePath {
				found = true
				break
			}
		}

		if !found {
			i.options.CriticalFiles = append(i.options.CriticalFiles, filePath)
		}

		// Store hash
		i.options.ExpectedHashes[filePath] = hash
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	return nil
}
