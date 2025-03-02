package listener

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ListenerStatus represents the current status of a listener
type ListenerStatus string

const (
	StatusRunning  ListenerStatus = "running"
	StatusStopped  ListenerStatus = "stopped"
	StatusError    ListenerStatus = "error"
	StatusUnknown  ListenerStatus = "unknown"
)

// ListenerConfig holds configuration for a listener
type ListenerConfig struct {
	Protocol string
	Address  string
	Port     int
	Options  map[string]interface{}
}

// ListenerStats holds statistics for a listener
type ListenerStats struct {
	StartTime      time.Time
	ConnectionsIn  int64
	ConnectionsOut int64
	BytesReceived  int64
	BytesSent      int64
	LastError      string
	LastErrorTime  time.Time
}

// Listener interface defines methods that all listener types must implement
type Listener interface {
	Start() error
	Stop() error
	Status() ListenerStatus
	Configure(config ListenerConfig) error
}

// Manager handles multiple listeners
type Manager struct {
	listeners    map[string]Listener
	listenerType map[string]ListenerType
	stats        map[string]*ListenerStats
	mutex        sync.RWMutex
	monitorStop  chan struct{}
	clientManager interface{} // Client manager for registering clients
}

// NewManager creates a new listener manager
func NewManager(clientManager interface{}) *Manager {
	manager := &Manager{
		listeners:    make(map[string]Listener),
		listenerType: make(map[string]ListenerType),
		stats:        make(map[string]*ListenerStats),
		monitorStop:  make(chan struct{}),
		clientManager: clientManager,
	}
	
	// Start the health monitor
	go manager.monitorHealth()
	
	return manager
}

// CreateListener creates a new listener with the specified type and configuration
func (m *Manager) CreateListener(id string, listenerType ListenerType, config ListenerConfig) error {
	// Validate the configuration
	if err := ValidateListenerConfig(listenerType, config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Add client manager to the listener options
	if config.Options == nil {
		config.Options = make(map[string]interface{})
	}
	config.Options["client_manager"] = m.clientManager
	
	// Create the listener
	listener, err := CreateListener(listenerType, config)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	
	// Add the listener to the manager
	if err := m.AddListener(id, listener); err != nil {
		return err
	}
	
	// Store the listener type
	m.mutex.Lock()
	m.listenerType[id] = listenerType
	m.stats[id] = &ListenerStats{}
	m.mutex.Unlock()
	
	return nil
}

// AddListener adds a new listener to the manager
func (m *Manager) AddListener(id string, listener Listener) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.listeners[id]; exists {
		return errors.New("listener with this ID already exists")
	}

	m.listeners[id] = listener
	return nil
}

// RemoveListener removes a listener from the manager
func (m *Manager) RemoveListener(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if listener, exists := m.listeners[id]; exists {
		if listener.Status() == StatusRunning {
			return errors.New("cannot remove a running listener, stop it first")
		}
		delete(m.listeners, id)
		delete(m.listenerType, id)
		delete(m.stats, id)
		return nil
	}

	return errors.New("listener not found")
}

// StartListener starts a specific listener
func (m *Manager) StartListener(id string) error {
	m.mutex.RLock()
	listener, exists := m.listeners[id]
	m.mutex.RUnlock()

	if !exists {
		return errors.New("listener not found")
	}

	err := listener.Start()
	if err == nil {
		// Update stats
		m.mutex.Lock()
		m.stats[id].StartTime = time.Now()
		m.mutex.Unlock()
	} else {
		// Update error stats
		m.mutex.Lock()
		m.stats[id].LastError = err.Error()
		m.stats[id].LastErrorTime = time.Now()
		m.mutex.Unlock()
	}
	
	return err
}

// StopListener stops a specific listener
func (m *Manager) StopListener(id string) error {
	m.mutex.RLock()
	listener, exists := m.listeners[id]
	m.mutex.RUnlock()

	if !exists {
		return errors.New("listener not found")
	}

	return listener.Stop()
}

// GetStatus returns the status of a specific listener
func (m *Manager) GetStatus(id string) (ListenerStatus, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if listener, exists := m.listeners[id]; exists {
		return listener.Status(), nil
	}

	return StatusUnknown, errors.New("listener not found")
}

// GetStats returns the statistics for a specific listener
func (m *Manager) GetStats(id string) (*ListenerStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if stats, exists := m.stats[id]; exists {
		return stats, nil
	}

	return nil, errors.New("listener not found")
}

// ListListeners returns a list of all listener IDs and their statuses
func (m *Manager) ListListeners() map[string]ListenerStatus {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]ListenerStatus)
	for id, listener := range m.listeners {
		result[id] = listener.Status()
	}

	return result
}

// GetListenerType returns the type of a specific listener
func (m *Manager) GetListenerType(id string) (ListenerType, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if listenerType, exists := m.listenerType[id]; exists {
		return listenerType, nil
	}

	return "", errors.New("listener not found")
}

// StopAll stops all running listeners
func (m *Manager) StopAll() error {
	m.mutex.RLock()
	listeners := make([]string, 0, len(m.listeners))
	for id := range m.listeners {
		listeners = append(listeners, id)
	}
	m.mutex.RUnlock()

	var lastErr error
	for _, id := range listeners {
		if err := m.StopListener(id); err != nil {
			lastErr = fmt.Errorf("failed to stop listener %s: %w", id, err)
		}
	}

	return lastErr
}

// Shutdown stops all listeners and shuts down the manager
func (m *Manager) Shutdown() error {
	// Stop the health monitor
	close(m.monitorStop)
	
	// Stop all listeners
	return m.StopAll()
}

// monitorHealth periodically checks the health of all listeners
func (m *Manager) monitorHealth() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkListenerHealth()
		case <-m.monitorStop:
			return
		}
	}
}

// checkListenerHealth checks the health of all listeners and attempts to recover if needed
func (m *Manager) checkListenerHealth() {
	m.mutex.RLock()
	listeners := make(map[string]Listener)
	for id, listener := range m.listeners {
		listeners[id] = listener
	}
	m.mutex.RUnlock()

	for id, listener := range listeners {
		status := listener.Status()
		
		// If the listener is in an error state, attempt to restart it
		if status == StatusError {
			fmt.Printf("Listener %s is in error state, attempting to restart\n", id)
			
			// Stop the listener
			if err := listener.Stop(); err != nil {
				fmt.Printf("Error stopping listener %s: %v\n", id, err)
				continue
			}
			
			// Wait a moment before restarting
			time.Sleep(1 * time.Second)
			
			// Start the listener
			if err := listener.Start(); err != nil {
				fmt.Printf("Error restarting listener %s: %v\n", id, err)
				
				// Update error stats
				m.mutex.Lock()
				m.stats[id].LastError = err.Error()
				m.stats[id].LastErrorTime = time.Now()
				m.mutex.Unlock()
			} else {
				fmt.Printf("Successfully restarted listener %s\n", id)
				
				// Update stats
				m.mutex.Lock()
				m.stats[id].StartTime = time.Now()
				m.mutex.Unlock()
			}
		}
	}
}
