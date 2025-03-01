package adapter

import (
	"dinoc2/pkg/module"
	"fmt"
	"net/rpc"
	"sync"
)

// RPCModule implements the Module interface for RPC-based modules
type RPCModule struct {
	client     *rpc.Client
	name       string
	mutex      sync.Mutex
	isRunning  bool
	isPaused   bool
	lastStatus module.ModuleStatus
}

// NewRPCModule creates a new RPC module
func NewRPCModule(client *rpc.Client, name string) *RPCModule {
	return &RPCModule{
		client:    client,
		name:      name,
		isRunning: false,
		isPaused:  false,
		lastStatus: module.ModuleStatus{
			Running: false,
			Error:   nil,
			Stats:   make(map[string]interface{}),
		},
	}
}

// Init initializes the module
func (m *RPCModule) Init(params map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.isRunning {
		return fmt.Errorf("module already running")
	}
	
	// Call remote Init method
	var reply bool
	err := m.client.Call(m.name+".Init", params, &reply)
	if err != nil {
		return fmt.Errorf("RPC Init failed: %w", err)
	}
	
	if !reply {
		return fmt.Errorf("module initialization failed")
	}
	
	m.isRunning = true
	
	// Update status
	err = m.updateStatus()
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	
	return nil
}

// Exec executes a command on the module
func (m *RPCModule) Exec(command string, args ...interface{}) (interface{}, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if !m.isRunning {
		return nil, fmt.Errorf("module not running")
	}
	
	if m.isPaused {
		return nil, fmt.Errorf("module is paused")
	}
	
	// Prepare arguments
	execArgs := struct {
		Command string
		Args    []interface{}
	}{
		Command: command,
		Args:    args,
	}
	
	// Call remote Exec method
	var reply interface{}
	err := m.client.Call(m.name+".Exec", execArgs, &reply)
	if err != nil {
		return nil, fmt.Errorf("RPC Exec failed: %w", err)
	}
	
	// Update status
	err = m.updateStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}
	
	return reply, nil
}

// Shutdown shuts down the module
func (m *RPCModule) Shutdown() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if !m.isRunning {
		return nil
	}
	
	// Call remote Shutdown method
	var reply bool
	err := m.client.Call(m.name+".Shutdown", struct{}{}, &reply)
	if err != nil {
		return fmt.Errorf("RPC Shutdown failed: %w", err)
	}
	
	m.isRunning = false
	m.isPaused = false
	
	// Update status
	err = m.updateStatus()
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	
	return nil
}

// GetStatus returns the module status
func (m *RPCModule) GetStatus() module.ModuleStatus {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Update status
	_ = m.updateStatus()
	
	return m.lastStatus
}

// GetCapabilities returns the module capabilities
func (m *RPCModule) GetCapabilities() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Call remote GetCapabilities method
	var reply []string
	err := m.client.Call(m.name+".GetCapabilities", struct{}{}, &reply)
	if err != nil {
		// Return empty capabilities on error
		return []string{}
	}
	
	return reply
}

// Pause pauses the module
func (m *RPCModule) Pause() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if !m.isRunning {
		return fmt.Errorf("module not running")
	}
	
	if m.isPaused {
		return nil // Already paused
	}
	
	// Call remote Pause method
	var reply bool
	err := m.client.Call(m.name+".Pause", struct{}{}, &reply)
	if err != nil {
		return fmt.Errorf("RPC Pause failed: %w", err)
	}
	
	if !reply {
		return fmt.Errorf("module pause failed")
	}
	
	m.isPaused = true
	
	// Update status
	err = m.updateStatus()
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	
	return nil
}

// Resume resumes the module
func (m *RPCModule) Resume() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if !m.isRunning {
		return fmt.Errorf("module not running")
	}
	
	if !m.isPaused {
		return nil // Not paused
	}
	
	// Call remote Resume method
	var reply bool
	err := m.client.Call(m.name+".Resume", struct{}{}, &reply)
	if err != nil {
		return fmt.Errorf("RPC Resume failed: %w", err)
	}
	
	if !reply {
		return fmt.Errorf("module resume failed")
	}
	
	m.isPaused = false
	
	// Update status
	err = m.updateStatus()
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	
	return nil
}

// updateStatus updates the module status
func (m *RPCModule) updateStatus() error {
	// Call remote GetStatus method
	var reply module.ModuleStatus
	err := m.client.Call(m.name+".GetStatus", struct{}{}, &reply)
	if err != nil {
		return fmt.Errorf("RPC GetStatus failed: %w", err)
	}
	
	m.lastStatus = reply
	m.isRunning = reply.Running
	
	return nil
}
