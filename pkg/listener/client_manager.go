package listener

import (
	"dinoc2/pkg/client/manager"
	"sync"
)

var (
	clientManager     *manager.ClientManager
	clientManagerLock sync.RWMutex
)

// SetClientManager sets the client manager for the listener package
func SetClientManager(cm *manager.ClientManager) {
	clientManagerLock.Lock()
	defer clientManagerLock.Unlock()
	clientManager = cm
}

// GetClientManager returns the client manager for the listener package
func GetClientManager() *manager.ClientManager {
	clientManagerLock.RLock()
	defer clientManagerLock.RUnlock()
	return clientManager
}
