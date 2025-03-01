package isolation

import (
	"dinoc2/pkg/module"
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

// IsolatedModule wraps a module with safety boundaries
type IsolatedModule struct {
	module     module.Module
	name       string
	mutex      sync.Mutex
	timeout    time.Duration
	lastError  error
	lastStatus module.ModuleStatus
}

// NewIsolatedModule creates a new isolated module
func NewIsolatedModule(mod module.Module, name string, timeout time.Duration) *IsolatedModule {
	return &IsolatedModule{
		module:  mod,
		name:    name,
		timeout: timeout,
	}
}

// Init initializes the module with safety boundaries
func (m *IsolatedModule) Init(params map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Create a channel to receive the result
	resultChan := make(chan error, 1)
	
	// Execute the init function in a goroutine with panic recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- fmt.Errorf("module init panicked: %v\n%s", r, debug.Stack())
			}
		}()
		
		resultChan <- m.module.Init(params)
	}()
	
	// Wait for the result with timeout
	select {
	case err := <-resultChan:
		m.lastError = err
		return err
	case <-time.After(m.timeout):
		m.lastError = fmt.Errorf("module init timed out after %v", m.timeout)
		return m.lastError
	}
}

// Exec executes a command on the module with safety boundaries
func (m *IsolatedModule) Exec(command string, args ...interface{}) (interface{}, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Create channels to receive the result
	resultChan := make(chan interface{}, 1)
	errChan := make(chan error, 1)
	
	// Execute the exec function in a goroutine with panic recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("module exec panicked: %v\n%s", r, debug.Stack())
			}
		}()
		
		result, err := m.module.Exec(command, args...)
		resultChan <- result
		errChan <- err
	}()
	
	// Wait for the result with timeout
	select {
	case result := <-resultChan:
		err := <-errChan
		m.lastError = err
		return result, err
	case <-time.After(m.timeout):
		m.lastError = fmt.Errorf("module exec timed out after %v", m.timeout)
		return nil, m.lastError
	}
}

// Shutdown shuts down the module with safety boundaries
func (m *IsolatedModule) Shutdown() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Create a channel to receive the result
	resultChan := make(chan error, 1)
	
	// Execute the shutdown function in a goroutine with panic recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- fmt.Errorf("module shutdown panicked: %v\n%s", r, debug.Stack())
			}
		}()
		
		resultChan <- m.module.Shutdown()
	}()
	
	// Wait for the result with timeout
	select {
	case err := <-resultChan:
		m.lastError = err
		return err
	case <-time.After(m.timeout):
		m.lastError = fmt.Errorf("module shutdown timed out after %v", m.timeout)
		return m.lastError
	}
}

// GetStatus returns the module status
func (m *IsolatedModule) GetStatus() module.ModuleStatus {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Create channels to receive the result
	resultChan := make(chan module.ModuleStatus, 1)
	
	// Execute the getStatus function in a goroutine with panic recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Return an error status on panic
				resultChan <- module.ModuleStatus{
					Running: false,
					Error:   fmt.Errorf("module getStatus panicked: %v", r),
					Stats:   make(map[string]interface{}),
				}
			}
		}()
		
		resultChan <- m.module.GetStatus()
	}()
	
	// Wait for the result with timeout
	select {
	case status := <-resultChan:
		m.lastStatus = status
		return status
	case <-time.After(m.timeout):
		// Return an error status on timeout
		errorStatus := module.ModuleStatus{
			Running: false,
			Error:   fmt.Errorf("module getStatus timed out after %v", m.timeout),
			Stats:   make(map[string]interface{}),
		}
		m.lastStatus = errorStatus
		return errorStatus
	}
}

// GetCapabilities returns the module capabilities
func (m *IsolatedModule) GetCapabilities() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Create a channel to receive the result
	resultChan := make(chan []string, 1)
	
	// Execute the getCapabilities function in a goroutine with panic recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Return empty capabilities on panic
				resultChan <- []string{}
			}
		}()
		
		resultChan <- m.module.GetCapabilities()
	}()
	
	// Wait for the result with timeout
	select {
	case capabilities := <-resultChan:
		return capabilities
	case <-time.After(m.timeout):
		// Return empty capabilities on timeout
		return []string{}
	}
}

// Pause pauses the module with safety boundaries
func (m *IsolatedModule) Pause() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Create a channel to receive the result
	resultChan := make(chan error, 1)
	
	// Execute the pause function in a goroutine with panic recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- fmt.Errorf("module pause panicked: %v\n%s", r, debug.Stack())
			}
		}()
		
		resultChan <- m.module.Pause()
	}()
	
	// Wait for the result with timeout
	select {
	case err := <-resultChan:
		m.lastError = err
		return err
	case <-time.After(m.timeout):
		m.lastError = fmt.Errorf("module pause timed out after %v", m.timeout)
		return m.lastError
	}
}

// Resume resumes the module with safety boundaries
func (m *IsolatedModule) Resume() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Create a channel to receive the result
	resultChan := make(chan error, 1)
	
	// Execute the resume function in a goroutine with panic recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- fmt.Errorf("module resume panicked: %v\n%s", r, debug.Stack())
			}
		}()
		
		resultChan <- m.module.Resume()
	}()
	
	// Wait for the result with timeout
	select {
	case err := <-resultChan:
		m.lastError = err
		return err
	case <-time.After(m.timeout):
		m.lastError = fmt.Errorf("module resume timed out after %v", m.timeout)
		return m.lastError
	}
}

// GetLastError returns the last error that occurred
func (m *IsolatedModule) GetLastError() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	return m.lastError
}
