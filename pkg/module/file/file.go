package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"dinoc2/pkg/module"
)

// FileModule implements a file system operations module
type FileModule struct {
	name        string
	description string
	mutex       sync.Mutex
	isRunning   bool
	isPaused    bool
	stats       map[string]interface{}
}

// NewFileModule creates a new file module
func NewFileModule() module.Module {
	return &FileModule{
		name:        "file",
		description: "File system operations",
		stats:       make(map[string]interface{}),
	}
}

// Init initializes the module
func (m *FileModule) Init(params map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if already running
	if m.isRunning {
		return fmt.Errorf("file module already running")
	}

	m.isRunning = true
	m.stats["operations"] = 0

	return nil
}

// Exec executes a file operation
func (m *FileModule) Exec(command string, args ...interface{}) (interface{}, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if running
	if !m.isRunning {
		return nil, fmt.Errorf("file module not running")
	}

	// Check if paused
	if m.isPaused {
		return nil, fmt.Errorf("file module is paused")
	}

	// Update operation count
	m.stats["operations"] = m.stats["operations"].(int) + 1

	// Parse command
	switch command {
	case "list":
		// List files in a directory
		if len(args) < 1 {
			return nil, fmt.Errorf("missing directory path")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid directory path")
		}
		return m.listFiles(path)

	case "read":
		// Read a file
		if len(args) < 1 {
			return nil, fmt.Errorf("missing file path")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid file path")
		}
		return m.readFile(path)

	case "write":
		// Write to a file
		if len(args) < 2 {
			return nil, fmt.Errorf("missing file path or data")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid file path")
		}
		
		var data []byte
		switch v := args[1].(type) {
		case []byte:
			data = v
		case string:
			data = []byte(v)
		default:
			return nil, fmt.Errorf("invalid data type")
		}
		
		return nil, m.writeFile(path, data)

	case "delete":
		// Delete a file
		if len(args) < 1 {
			return nil, fmt.Errorf("missing file path")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid file path")
		}
		return nil, m.deleteFile(path)

	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

// listFiles lists files in a directory
func (m *FileModule) listFiles(path string) ([]string, error) {
	// Open directory
	dir, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open directory: %w", err)
	}
	defer dir.Close()

	// Read directory entries
	entries, err := dir.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Create result
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		// Format entry
		entryType := "F"
		if entry.IsDir() {
			entryType = "D"
		}
		result = append(result, fmt.Sprintf("%s %s %d %s",
			entryType,
			entry.Mode().String(),
			entry.Size(),
			entry.Name()))
	}

	return result, nil
}

// readFile reads a file
func (m *FileModule) readFile(path string) ([]byte, error) {
	// Open file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read file
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// writeFile writes data to a file
func (m *FileModule) writeFile(path string, data []byte) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write data
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// deleteFile deletes a file
func (m *FileModule) deleteFile(path string) error {
	// Delete file
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Shutdown shuts down the module
func (m *FileModule) Shutdown() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if running
	if !m.isRunning {
		return nil
	}

	m.isRunning = false

	return nil
}

// GetStatus returns the module status
func (m *FileModule) GetStatus() module.ModuleStatus {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	status := module.ModuleStatus{
		Running: m.isRunning,
		Stats:   m.stats,
	}

	return status
}

// GetCapabilities returns the module capabilities
func (m *FileModule) GetCapabilities() []string {
	return []string{
		"list",
		"read",
		"write",
		"delete",
	}
}

// Pause temporarily pauses the module's operations
func (m *FileModule) Pause() error {
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
func (m *FileModule) Resume() error {
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
