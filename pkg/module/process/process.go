package process

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"dinoc2/pkg/module"
)

// ProcessModule implements a process management module
type ProcessModule struct {
	name        string
	description string
	mutex       sync.Mutex
	isRunning   bool
	isPaused    bool
	stats       map[string]interface{}
}

// NewProcessModule creates a new process module
func NewProcessModule() module.Module {
	return &ProcessModule{
		name:        "process",
		description: "Process management",
		stats:       make(map[string]interface{}),
	}
}

// Init initializes the module
func (m *ProcessModule) Init(params map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if already running
	if m.isRunning {
		return fmt.Errorf("process module already running")
	}

	m.isRunning = true
	m.stats["operations"] = 0

	return nil
}

// Exec executes a process operation
func (m *ProcessModule) Exec(command string, args ...interface{}) (interface{}, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if running
	if !m.isRunning {
		return nil, fmt.Errorf("process module not running")
	}

	// Check if paused
	if m.isPaused {
		return nil, fmt.Errorf("process module is paused")
	}

	// Update operation count
	m.stats["operations"] = m.stats["operations"].(int) + 1

	// Parse command
	switch command {
	case "list":
		// List processes
		return m.listProcesses()

	case "kill":
		// Kill a process
		if len(args) < 1 {
			return nil, fmt.Errorf("missing process ID")
		}
		
		var pid int
		switch v := args[0].(type) {
		case int:
			pid = v
		case string:
			var err error
			pid, err = strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("invalid process ID: %w", err)
			}
		default:
			return nil, fmt.Errorf("invalid process ID type")
		}
		
		return nil, m.killProcess(pid)

	case "execute":
		// Execute a command
		if len(args) < 1 {
			return nil, fmt.Errorf("missing command")
		}
		cmd, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid command")
		}
		
		// Get command arguments
		var cmdArgs []string
		if len(args) > 1 {
			switch v := args[1].(type) {
			case []string:
				cmdArgs = v
			case string:
				cmdArgs = strings.Fields(v)
			}
		}
		
		return m.executeCommand(cmd, cmdArgs)

	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

// listProcesses lists running processes
func (m *ProcessModule) listProcesses() ([]string, error) {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("tasklist", "/FO", "CSV")
	case "darwin":
		cmd = exec.Command("ps", "-e", "-o", "pid,ppid,user,%cpu,%mem,command")
	default: // Linux and others
		cmd = exec.Command("ps", "-e", "-o", "pid,ppid,user,%cpu,%mem,command")
	}
	
	// Execute command
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}
	
	// Parse output
	lines := strings.Split(string(output), "\n")
	
	// Remove empty lines
	var result []string
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			result = append(result, line)
		}
	}
	
	return result, nil
}

// killProcess kills a process
func (m *ProcessModule) killProcess(pid int) error {
	// Get process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}
	
	// Kill process
	var killErr error
	if runtime.GOOS == "windows" {
		killErr = process.Kill()
	} else {
		// On Unix-like systems, first try SIGTERM
		killErr = process.Signal(syscall.SIGTERM)
		if killErr == nil {
			// Wait a bit for the process to terminate
			// In a real implementation, we would wait for the process to exit
			// For now, just return success
			return nil
		}
		
		// If SIGTERM failed, try SIGKILL
		killErr = process.Kill()
	}
	
	if killErr != nil {
		return fmt.Errorf("failed to kill process: %w", killErr)
	}
	
	return nil
}

// executeCommand executes a command
func (m *ProcessModule) executeCommand(command string, args []string) (string, error) {
	// Create command
	cmd := exec.Command(command, args...)
	
	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command execution failed: %w", err)
	}
	
	return string(output), nil
}

// Shutdown shuts down the module
func (m *ProcessModule) Shutdown() error {
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
func (m *ProcessModule) GetStatus() module.ModuleStatus {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	status := module.ModuleStatus{
		Running: m.isRunning,
		Stats:   m.stats,
	}

	return status
}

// GetCapabilities returns the module capabilities
func (m *ProcessModule) GetCapabilities() []string {
	return []string{
		"list",
		"kill",
		"execute",
	}
}

// Pause temporarily pauses the module's operations
func (m *ProcessModule) Pause() error {
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
func (m *ProcessModule) Resume() error {
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

// init registers the module
func init() {
	module.RegisterModule("process", NewProcessModule)
}
