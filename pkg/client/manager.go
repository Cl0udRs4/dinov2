package client

import (
	"errors"
	"fmt"
	"sync"
)

// Manager handles client connections and management
type Manager struct {
	clients     map[string]*Client
	clientMutex sync.RWMutex
}

// NewManager creates a new client manager
func NewManager() *Manager {
	fmt.Println("DEBUG: Creating new client manager")
	return &Manager{
		clients: make(map[string]*Client),
	}
}

// RegisterClient registers a client with the manager
func (m *Manager) RegisterClient(client *Client) string {
	m.clientMutex.Lock()
	defer m.clientMutex.Unlock()
	
	fmt.Printf("DEBUG: RegisterClient called with client: %+v\n", client)
	
	// Use GetSessionID() method instead of directly accessing SessionID field
	clientID := client.GetSessionID()
	fmt.Printf("DEBUG: Using clientID: %s\n", clientID)
	
	m.clients[clientID] = client
	fmt.Printf("DEBUG: Client registered, total clients: %d\n", len(m.clients))
	
	return clientID
}

// UnregisterClient removes a client from the manager
func (m *Manager) UnregisterClient(clientID string) error {
	m.clientMutex.Lock()
	defer m.clientMutex.Unlock()
	
	if _, exists := m.clients[clientID]; !exists {
		return errors.New("client not found")
	}
	
	delete(m.clients, clientID)
	return nil
}

// GetClient retrieves a client by ID
func (m *Manager) GetClient(clientID string) (*Client, error) {
	m.clientMutex.RLock()
	defer m.clientMutex.RUnlock()
	
	client, exists := m.clients[clientID]
	if !exists {
		return nil, errors.New("client not found")
	}
	
	return client, nil
}

// ListClients returns a list of all connected clients
func (m *Manager) ListClients() []*Client {
	m.clientMutex.RLock()
	defer m.clientMutex.RUnlock()
	
	fmt.Printf("DEBUG: ListClients called, number of clients: %d\n", len(m.clients))
	for id, _ := range m.clients {
		fmt.Printf("DEBUG: Client ID: %s\n", id)
	}
	
	clients := make([]*Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	
	return clients
}

// SwitchProtocol switches a client to a different protocol
func (m *Manager) SwitchProtocol(clientID string, protocol string) error {
	client, err := m.GetClient(clientID)
	if err != nil {
		return err
	}
	
	// Convert protocol string to ProtocolType
	protocolType := ProtocolType(protocol)
	
	// Check if the protocol is supported
	supported := false
	for _, p := range client.config.Protocols {
		if p == protocolType {
			supported = true
			break
		}
	}
	
	if !supported {
		return fmt.Errorf("protocol %s not supported by client", protocol)
	}
	
	// Switch the protocol
	return client.HandleProtocolSwitchCommand(protocol)
}
