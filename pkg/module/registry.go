package module

import (
	"fmt"
	"sync"
)

// ModuleFactory is a function that creates a new module
type ModuleFactory func() Module

// moduleRegistry is a registry of module factories
var moduleRegistry = struct {
	factories map[string]ModuleFactory
	mutex     sync.RWMutex
}{
	factories: make(map[string]ModuleFactory),
}

// RegisterModule registers a module factory
func RegisterModule(name string, factory ModuleFactory) {
	moduleRegistry.mutex.Lock()
	defer moduleRegistry.mutex.Unlock()

	if _, exists := moduleRegistry.factories[name]; exists {
		// Module already registered, just log a warning
		fmt.Printf("Warning: Module '%s' already registered, overwriting\n", name)
	}

	moduleRegistry.factories[name] = factory
}

// GetModuleFactory returns a module factory by name
func GetModuleFactory(name string) (ModuleFactory, error) {
	moduleRegistry.mutex.RLock()
	defer moduleRegistry.mutex.RUnlock()

	factory, exists := moduleRegistry.factories[name]
	if !exists {
		return nil, fmt.Errorf("module '%s' not found", name)
	}

	return factory, nil
}

// CreateModule creates a new module by name
func CreateModule(name string) (Module, error) {
	factory, err := GetModuleFactory(name)
	if err != nil {
		return nil, err
	}

	return factory(), nil
}

// ListModules returns a list of registered module names
func ListModules() []string {
	moduleRegistry.mutex.RLock()
	defer moduleRegistry.mutex.RUnlock()

	modules := make([]string, 0, len(moduleRegistry.factories))
	for name := range moduleRegistry.factories {
		modules = append(modules, name)
	}

	return modules
}
