package loader

import (
	"dinoc2/pkg/module"
	"errors"
	"fmt"
	"sync"
)

// NativeLoader implements a loader for native Go modules
type NativeLoader struct {
	modules     map[string]module.Module
	moduleInfos map[string]module.ModuleInfo
	mutex       sync.RWMutex
}

// NewNativeLoader creates a new native loader
func NewNativeLoader() *NativeLoader {
	return &NativeLoader{
		modules:     make(map[string]module.Module),
		moduleInfos: make(map[string]module.ModuleInfo),
	}
}

// Load loads a module from the registry
func (l *NativeLoader) Load(name string) (module.Module, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// Check if module is already loaded
	if mod, exists := l.modules[name]; exists {
		return mod, nil
	}
	
	// Create module from factory
	factory, err := module.GetModuleFactory(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get module factory: %w", err)
	}
	
	mod := factory()
	if mod == nil {
		return nil, errors.New("module factory returned nil")
	}
	
	// Store module
	l.modules[name] = mod
	
	return mod, nil
}

// Unload unloads a module
func (l *NativeLoader) Unload(mod module.Module) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// Find module name
	var moduleName string
	for name, m := range l.modules {
		if m == mod {
			moduleName = name
			break
		}
	}
	
	if moduleName == "" {
		return errors.New("module not found")
	}
	
	// Shutdown module
	err := mod.Shutdown()
	if err != nil {
		return fmt.Errorf("failed to shutdown module: %w", err)
	}
	
	// Remove module
	delete(l.modules, moduleName)
	
	return nil
}

// GetType returns the loader type
func (l *NativeLoader) GetType() LoaderType {
	return LoaderTypeNative
}

// IsSupported returns true if the loader is supported on the current platform
func (l *NativeLoader) IsSupported() bool {
	return true // Native loader is always supported
}
