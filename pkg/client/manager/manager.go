package manager

import (
	"sync"
	"time"
)

var (
	globalClientManager     *ClientManager
	globalClientManagerLock sync.RWMutex
)

// ClientInfo represents information about a connected client
type ClientInfo struct {
	ID          string    `json:"id"`
	Address     string    `json:"address"`
	Protocols   []string  `json:"protocols"`
	ConnectedAt time.Time `json:"connected_at"`
}

// ClientManager manages connected clients
type ClientManager struct {
	clients map[string]*ClientInfo
	mutex   sync.RWMutex
}

// NewClientManager creates a new client manager
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[string]*ClientInfo),
	}
}

// RegisterClient registers a new client
func (m *ClientManager) RegisterClient(id string, address string, protocol string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if client already exists
	if client, exists := m.clients[id]; exists {
		// Update existing client
		protocolExists := false
		for _, p := range client.Protocols {
			if p == protocol {
				protocolExists = true
				break
			}
		}
		if !protocolExists {
			client.Protocols = append(client.Protocols, protocol)
		}
	} else {
		// Create new client
		m.clients[id] = &ClientInfo{
			ID:          id,
			Address:     address,
			Protocols:   []string{protocol},
			ConnectedAt: time.Now(),
		}
	}
}

// UnregisterClient unregisters a client
func (m *ClientManager) UnregisterClient(id string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.clients, id)
}

// GetClient returns a client by ID
func (m *ClientManager) GetClient(id string) *ClientInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.clients[id]
}

// ListClients returns a list of all clients
func (m *ClientManager) ListClients() []*ClientInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	clients := make([]*ClientInfo, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}

	return clients
}

// SetGlobalClientManager sets the global client manager
func SetGlobalClientManager(cm *ClientManager) {
	globalClientManagerLock.Lock()
	defer globalClientManagerLock.Unlock()
	globalClientManager = cm
}

// GetGlobalClientManager returns the global client manager
func GetGlobalClientManager() *ClientManager {
	globalClientManagerLock.RLock()
	defer globalClientManagerLock.RUnlock()
	return globalClientManager
}
