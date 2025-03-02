package client

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Manager handles client connections and management
type Manager struct {
	clients     map[string]interface{}
	clientMutex sync.RWMutex
}

// NewManager creates a new client manager
func NewManager() *Manager {
	fmt.Println("DEBUG: Creating new client manager")
	return &Manager{
		clients: make(map[string]interface{}),
	}
}

// RegisterClient registers a client with the manager
func (m *Manager) RegisterClient(client interface{}) string {
	m.clientMutex.Lock()
	defer m.clientMutex.Unlock()
	
	fmt.Printf("DEBUG: RegisterClient called with client type: %T\n", client)
	
	var clientID string
	
	// Try to get client ID based on type
	switch c := client.(type) {
	case *Client:
		clientID = string(c.sessionID)
		fmt.Printf("DEBUG: Client has sessionID: %s\n", clientID)
	default:
		// For simple client objects, generate a unique ID
		clientID = fmt.Sprintf("client-%d", time.Now().UnixNano())
		fmt.Printf("DEBUG: Generated new clientID: %s\n", clientID)
	}
	
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
func (m *Manager) GetClient(clientID string) (interface{}, error) {
	m.clientMutex.RLock()
	defer m.clientMutex.RUnlock()
	
	client, exists := m.clients[clientID]
	if !exists {
		return nil, errors.New("client not found")
	}
	
	return client, nil
}

// ListClients returns a list of all connected clients
func (m *Manager) ListClients() []map[string]interface{} {
	m.clientMutex.RLock()
	defer m.clientMutex.RUnlock()
	
	fmt.Printf("DEBUG: ListClients called, number of clients: %d\n", len(m.clients))
	for id := range m.clients {
		fmt.Printf("DEBUG: Client ID: %s\n", id)
	}
	
	clients := make([]map[string]interface{}, 0, len(m.clients))
	for id, client := range m.clients {
		clientInfo := map[string]interface{}{
			"id": id,
		}
		
		// Add additional info based on client type
		switch c := client.(type) {
		case *Client:
			clientInfo["type"] = "full_client"
			clientInfo["address"] = c.config.ServerAddress
			clientInfo["protocol"] = c.currentProtocol
		default:
			clientInfo["type"] = "simple_client"
		}
		
		clients = append(clients, clientInfo)
	}
	
	return clients
}
