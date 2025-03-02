package listener

import (
	"dinoc2/pkg/listener/dns"
	"dinoc2/pkg/listener/http"
	"dinoc2/pkg/listener/icmp"
	"dinoc2/pkg/listener/websocket"
	"fmt"
	"log"
	"sync"
)

// ListenerConfig represents the configuration for a listener
type ListenerConfig struct {
	ID      string
	Type    string
	Address string
	Port    int
	Options map[string]interface{}
}

// Listener interface defines methods that all listeners must implement
type Listener interface {
	Start() error
	Stop() error
	SetClientManager(interface{})
}

// Manager manages multiple listeners
type Manager struct {
	listeners     map[string]Listener
	clientManager interface{}
	mutex         sync.RWMutex
}

// NewManager creates a new listener manager
func NewManager() *Manager {
	return &Manager{
		listeners: make(map[string]Listener),
	}
}

// SetClientManager sets the client manager for all listeners
func (m *Manager) SetClientManager(clientManager interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.clientManager = clientManager
	
	// Pass client manager to all existing listeners
	for _, listener := range m.listeners {
		listener.SetClientManager(clientManager)
	}
}

// AddListener adds a new listener
func (m *Manager) AddListener(config map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Extract listener ID
	id, ok := config["id"].(string)
	if !ok {
		return fmt.Errorf("invalid configuration: listener ID is required")
	}
	
	// Check if listener already exists
	if _, exists := m.listeners[id]; exists {
		return fmt.Errorf("listener with ID %s already exists", id)
	}
	
	// Extract listener type
	listenerType, ok := config["type"].(string)
	if !ok {
		return fmt.Errorf("invalid configuration: listener type is required")
	}
	
	// Extract address and port
	address, ok := config["address"].(string)
	if !ok {
		return fmt.Errorf("invalid configuration: listener address is required")
	}
	
	portFloat, ok := config["port"].(float64)
	if !ok {
		return fmt.Errorf("invalid configuration: listener port is required")
	}
	port := int(portFloat)
	
	// Extract options
	options, ok := config["options"].(map[string]interface{})
	if !ok {
		options = make(map[string]interface{})
	}
	
	// Create listener configuration
	listenerConfig := ListenerConfig{
		ID:      id,
		Type:    listenerType,
		Address: address,
		Port:    port,
		Options: options,
	}
	
	// Create listener based on type
	var listener Listener
	var err error
	
	switch listenerType {
	case "tcp":
		listener, err = NewTCPListener(listenerConfig)
	case "http":
		listener, err = http.NewHTTPListener(config)
	case "websocket":
		listener, err = websocket.NewWebSocketListener(config)
	case "dns":
		listener, err = dns.NewDNSListener(config)
	case "icmp":
		listener, err = icmp.NewICMPListener(config)
	default:
		return fmt.Errorf("unsupported listener type: %s", listenerType)
	}
	
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	
	// Pass client manager to listener
	if m.clientManager != nil {
		listener.SetClientManager(m.clientManager)
	}
	
	// Add listener to map
	m.listeners[id] = listener
	
	return nil
}

// RemoveListener removes a listener
func (m *Manager) RemoveListener(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Check if listener exists
	listener, exists := m.listeners[id]
	if !exists {
		return fmt.Errorf("listener with ID %s does not exist", id)
	}
	
	// Stop listener if running
	err := listener.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop listener: %w", err)
	}
	
	// Remove listener from map
	delete(m.listeners, id)
	
	return nil
}

// GetListener gets a listener by ID
func (m *Manager) GetListener(id string) (Listener, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Check if listener exists
	listener, exists := m.listeners[id]
	if !exists {
		return nil, fmt.Errorf("listener with ID %s does not exist", id)
	}
	
	return listener, nil
}

// StartAll starts all listeners
func (m *Manager) StartAll() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Start all listeners
	for id, listener := range m.listeners {
		err := listener.Start()
		if err != nil {
			return fmt.Errorf("failed to start listener %s: %w", id, err)
		}
		
		log.Printf("Started listener %s (%s) on %s:%d", id, m.getListenerType(listener), m.getListenerAddress(listener), m.getListenerPort(listener))
	}
	
	return nil
}

// StopAll stops all listeners
func (m *Manager) StopAll() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Stop all listeners
	for id, listener := range m.listeners {
		err := listener.Stop()
		if err != nil {
			return fmt.Errorf("failed to stop listener %s: %w", id, err)
		}
	}
	
	return nil
}

// getListenerType gets the type of a listener
func (m *Manager) getListenerType(listener Listener) string {
	switch listener.(type) {
	case *TCPListener:
		return "tcp"
	case *http.HTTPListener:
		return "http"
	case *websocket.WebSocketListener:
		return "websocket"
	case *dns.DNSListener:
		return "dns"
	case *icmp.ICMPListener:
		return "icmp"
	default:
		return "unknown"
	}
}

// getListenerAddress gets the address of a listener
func (m *Manager) getListenerAddress(listener Listener) string {
	switch l := listener.(type) {
	case *TCPListener:
		return l.address
	default:
		return "unknown"
	}
}

// getListenerPort gets the port of a listener
func (m *Manager) getListenerPort(listener Listener) int {
	switch l := listener.(type) {
	case *TCPListener:
		return l.port
	default:
		return 0
	}
}
