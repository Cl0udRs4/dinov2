package registry

import (
	"dinoc2/pkg/module"
	"fmt"
	"sync"
)

// ModuleInfo contains information about a registered module
type ModuleInfo struct {
	Name         string
	Description  string
	Version      string
	Author       string
	Capabilities []string
	Platforms    []string
}

// RegistryManager manages module registration and discovery
type RegistryManager struct {
	modules map[string]ModuleInfo
	mutex   sync.RWMutex
}

// NewRegistryManager creates a new registry manager
func NewRegistryManager() *RegistryManager {
	return &RegistryManager{
		modules: make(map[string]ModuleInfo),
	}
}

// RegisterModule registers a module with the registry
func (r *RegistryManager) RegisterModule(info ModuleInfo) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.modules[info.Name]; exists {
		return fmt.Errorf("module %s already registered", info.Name)
	}

	r.modules[info.Name] = info
	return nil
}

// UnregisterModule unregisters a module from the registry
func (r *RegistryManager) UnregisterModule(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.modules[name]; !exists {
		return fmt.Errorf("module %s not registered", name)
	}

	delete(r.modules, name)
	return nil
}

// GetModuleInfo returns information about a registered module
func (r *RegistryManager) GetModuleInfo(name string) (ModuleInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	info, exists := r.modules[name]
	if !exists {
		return ModuleInfo{}, fmt.Errorf("module %s not registered", name)
	}

	return info, nil
}

// ListModules returns a list of registered modules
func (r *RegistryManager) ListModules() map[string]ModuleInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Create a copy of the modules map
	result := make(map[string]ModuleInfo, len(r.modules))
	for name, info := range r.modules {
		result[name] = info
	}

	return result
}

// FindModulesByCapability returns a list of modules that have the specified capability
func (r *RegistryManager) FindModulesByCapability(capability string) []ModuleInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []ModuleInfo
	for _, info := range r.modules {
		for _, cap := range info.Capabilities {
			if cap == capability {
				result = append(result, info)
				break
			}
		}
	}

	return result
}

// FindModulesByPlatform returns a list of modules that support the specified platform
func (r *RegistryManager) FindModulesByPlatform(platform string) []ModuleInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []ModuleInfo
	for _, info := range r.modules {
		for _, plat := range info.Platforms {
			if plat == platform {
				result = append(result, info)
				break
			}
		}
	}

	return result
}
