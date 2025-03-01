package loader

import (
	"dinoc2/pkg/module"
	"errors"
	"fmt"
	"sync"
)

// WasmLoader implements a loader for WebAssembly modules
type WasmLoader struct {
	instances   map[string]interface{} // WebAssembly instance
	modules     map[string]module.Module
	moduleInfos map[string]module.ModuleInfo
	mutex       sync.RWMutex
}

// NewWasmLoader creates a new WebAssembly loader
func NewWasmLoader() *WasmLoader {
	return &WasmLoader{
		instances:   make(map[string]interface{}),
		modules:     make(map[string]module.Module),
		moduleInfos: make(map[string]module.ModuleInfo),
	}
}

// Load loads a module from a WebAssembly file
func (l *WasmLoader) Load(path string) (module.Module, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// Check if module is already loaded
	if mod, exists := l.modules[path]; exists {
		return mod, nil
	}
	
	// In a real implementation, we would use a WebAssembly runtime like Wasmer or Wasmtime
	// to load and execute the WebAssembly module
	// For now, we'll just return an error
	return nil, errors.New("WebAssembly loading not implemented")
}

// Unload unloads a module
func (l *WasmLoader) Unload(mod module.Module) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// Find module path
	var modulePath string
	for path, m := range l.modules {
		if m == mod {
			modulePath = path
			break
		}
	}
	
	if modulePath == "" {
		return errors.New("module not found")
	}
	
	// Shutdown module
	err := mod.Shutdown()
	if err != nil {
		return fmt.Errorf("failed to shutdown module: %w", err)
	}
	
	// Remove instance and module
	delete(l.instances, modulePath)
	delete(l.modules, modulePath)
	
	return nil
}

// GetType returns the loader type
func (l *WasmLoader) GetType() LoaderType {
	return LoaderTypeWasm
}

// IsSupported returns true if the loader is supported on the current platform
func (l *WasmLoader) IsSupported() bool {
	return true // WebAssembly is supported on all platforms
}
