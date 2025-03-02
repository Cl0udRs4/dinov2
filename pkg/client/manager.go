package client

import (
	"dinoc2/pkg/crypto"
	"fmt"
	"sync"
	"time"
)

// Protocol represents a communication protocol
type Protocol string

// Communication protocols
const (
	ProtocolTCP       Protocol = "tcp"
	ProtocolHTTP      Protocol = "http"
	ProtocolWebSocket Protocol = "websocket"
	ProtocolDNS       Protocol = "dns"
	ProtocolICMP      Protocol = "icmp"
)

// Client represents a connected client
type Client struct {
	SessionID crypto.SessionID
	Address   string
	Protocol  Protocol
	LastSeen  time.Time
	Info      map[string]interface{}
}

// NewClient creates a new client
func NewClient(sessionID crypto.SessionID, address string, protocol Protocol) *Client {
	return &Client{
		SessionID: sessionID,
		Address:   address,
		Protocol:  protocol,
		LastSeen:  time.Now(),
		Info:      make(map[string]interface{}),
	}
}

// UpdateLastSeen updates the last seen timestamp
func (c *Client) UpdateLastSeen() {
	c.LastSeen = time.Now()
}

// Manager manages clients
type Manager struct {
	clients map[string]*Client
	mutex   sync.RWMutex
}

// NewManager creates a new client manager
func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*Client),
	}
}

// RegisterClient registers a client
func (m *Manager) RegisterClient(client *Client) string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Use session ID as client ID
	clientID := string(client.SessionID)
	
	// Check if client already exists
	if existingClient, exists := m.clients[clientID]; exists {
		// Update last seen timestamp
		existingClient.UpdateLastSeen()
		fmt.Printf("Updated existing client with ID %s\n", clientID)
		return clientID
	}
	
	// Add client to map
	m.clients[clientID] = client
	fmt.Printf("Registered new client with ID %s\n", clientID)
	
	return clientID
}

// UnregisterClient unregisters a client
func (m *Manager) UnregisterClient(clientID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Check if client exists
	if _, exists := m.clients[clientID]; !exists {
		return
	}
	
	// Remove client from map
	delete(m.clients, clientID)
	fmt.Printf("Unregistered client with ID %s\n", clientID)
}

// GetClient gets a client by ID
func (m *Manager) GetClient(clientID string) (*Client, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Check if client exists
	client, exists := m.clients[clientID]
	if !exists {
		return nil, fmt.Errorf("client not found")
	}
	
	return client, nil
}

// ListClients lists all clients
func (m *Manager) ListClients() []*Client {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	fmt.Printf("DEBUG: ListClients called, number of clients: %d\n", len(m.clients))
	
	// Create list of clients
	clients := make([]*Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	
	return clients
}

// CleanupClients removes inactive clients
func (m *Manager) CleanupClients(maxAge time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Get current time
	now := time.Now()
	
	// Check each client
	for id, client := range m.clients {
		// Check if client has been inactive for too long
		if now.Sub(client.LastSeen) > maxAge {
			// Remove client
			delete(m.clients, id)
			fmt.Printf("Removed inactive client with ID %s\n", id)
		}
	}
}
