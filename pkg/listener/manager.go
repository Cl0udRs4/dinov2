package listener

import (
	"errors"
	"fmt"
	"sync"
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

// Listener interface defines methods that all listener types must implement
type Listener interface {
	Start() error
	Stop() error
	Status() ListenerStatus
	Configure(config ListenerConfig) error
}

// Manager handles multiple listeners
type Manager struct {
	listeners map[string]Listener
	mutex     sync.RWMutex
}

// NewManager creates a new listener manager
func NewManager() *Manager {
	return &Manager{
		listeners: make(map[string]Listener),
	}
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

	return listener.Start()
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
