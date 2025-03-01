package loader

import (
	"dinoc2/pkg/module"
	"errors"
	"fmt"
	"sync"
)

// RPCLoader implements a loader for RPC-based modules
type RPCLoader struct {
	clients     map[string]interface{} // RPC client
	modules     map[string]module.Module
	moduleInfos map[string]module.ModuleInfo
	mutex       sync.RWMutex
}

// NewRPCLoader creates a new RPC loader
func NewRPCLoader() *RPCLoader {
	return &RPCLoader{
		clients:     make(map[string]interface{}),
		modules:     make(map[string]module.Module),
		moduleInfos: make(map[string]module.ModuleInfo),
	}
}

// Load loads a module from an RPC endpoint
func (l *RPCLoader) Load(endpoint string) (module.Module, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// Check if module is already loaded
	if mod, exists := l.modules[endpoint]; exists {
		return mod, nil
	}
	
	// In a real implementation, we would establish an RPC connection to the endpoint
	// and create a proxy module that forwards calls to the remote module
	// For now, we'll just return an error
	return nil, errors.New("RPC loading not implemented")
}

// Unload unloads a module
func (l *RPCLoader) Unload(mod module.Module) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// Find endpoint
	var endpoint string
	for ep, m := range l.modules {
		if m == mod {
			endpoint = ep
			break
		}
	}
	
	if endpoint == "" {
		return errors.New("module not found")
	}
	
	// Shutdown module
	err := mod.Shutdown()
	if err != nil {
		return fmt.Errorf("failed to shutdown module: %w", err)
	}
	
	// Remove client and module
	delete(l.clients, endpoint)
	delete(l.modules, endpoint)
	
	return nil
}

// GetType returns the loader type
func (l *RPCLoader) GetType() LoaderType {
	return LoaderTypeRPC
}

// IsSupported returns true if the loader is supported on the current platform
func (l *RPCLoader) IsSupported() bool {
	return true // RPC is supported on all platforms
}
