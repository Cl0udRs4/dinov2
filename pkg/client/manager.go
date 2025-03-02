package client

import (
	"errors"
	"fmt"
	"sync"
	"time"
)
package client

import (
	"errors"
	"fmt"
	"sync"
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
	
	fmt.Printf("DEBUG: RegisterClient called with client: %+v\n", client)
	
	var clientID string
	
	// Try to get client ID based on type
	switch c := client.(type) {
	case *Client:
		clientID = c.GetSessionID()
	case struct {
		Address string
		ID      string
	}:
		clientID = c.ID
	default:
		clientID = fmt.Sprintf("unknown-%d", time.Now().UnixNano())
	}
	
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
func (m *Manager) ListClients() []*Client {
	m.clientMutex.RLock()
	defer m.clientMutex.RUnlock()
	
	fmt.Printf("DEBUG: ListClients called, number of clients: %d\n", len(m.clients))
	for id, _ := range m.clients {
		fmt.Printf("DEBUG: Client ID: %s\n", id)
	}
	
	clients := make([]*Client, 0)
	for _, client := range m.clients {
		if c, ok := client.(*Client); ok {
			clients = append(clients, c)
		} else {
			fmt.Printf("DEBUG: Client is not of type *Client: %T\n", client)
		}
	}
	
	return clients
}
