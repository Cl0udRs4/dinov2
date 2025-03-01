package manager

import (
	"dinoc2/pkg/module"
	"dinoc2/pkg/module/loader"
	"fmt"
	"sync"
)

// ModuleManager manages module loading, initialization, and lifecycle
type ModuleManager struct {
	loaders       map[loader.LoaderType]loader.ModuleLoader
	loadedModules map[string]module.Module
	moduleInfo    map[string]ModuleInfo
	mutex         sync.RWMutex
}

// ModuleInfo contains information about a loaded module
type ModuleInfo struct {
	Name       string
	Path       string
	LoaderType loader.LoaderType
	Status     module.ModuleStatus
}

// NewModuleManager creates a new module manager
func NewModuleManager() (*ModuleManager, error) {
	manager := &ModuleManager{
		loaders:       make(map[loader.LoaderType]loader.ModuleLoader),
		loadedModules: make(map[string]module.Module),
		moduleInfo:    make(map[string]ModuleInfo),
	}
	
	// Initialize loaders
	supportedLoaders := loader.GetSupportedLoaders()
	for _, loaderType := range supportedLoaders {
		l, err := loader.GetLoader(loaderType)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize loader %s: %w", loaderType, err)
		}
		
		manager.loaders[loaderType] = l
	}
	
	return manager, nil
}

// LoadModule loads a module using the specified loader
func (m *ModuleManager) LoadModule(name, path string, loaderType loader.LoaderType) (module.Module, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Check if module is already loaded
	if mod, exists := m.loadedModules[name]; exists {
		return mod, nil
	}
	
	// Get loader
	l, exists := m.loaders[loaderType]
	if !exists {
		return nil, fmt.Errorf("loader %s not found", loaderType)
	}
	
	// Load module
	mod, err := l.Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load module: %w", err)
	}
	
	// Store module
	m.loadedModules[name] = mod
	m.moduleInfo[name] = ModuleInfo{
		Name:       name,
		Path:       path,
		LoaderType: loaderType,
		Status:     mod.GetStatus(),
	}
	
	return mod, nil
}

// InitModule initializes a loaded module
func (m *ModuleManager) InitModule(name string, params map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Get module
	mod, exists := m.loadedModules[name]
	if !exists {
		return fmt.Errorf("module %s not found", name)
	}
	
	// Initialize module
	err := mod.Init(params)
	if err != nil {
		return fmt.Errorf("failed to initialize module: %w", err)
	}
	
	// Update status
	info := m.moduleInfo[name]
	info.Status = mod.GetStatus()
	m.moduleInfo[name] = info
	
	return nil
}

// UnloadModule unloads a module
func (m *ModuleManager) UnloadModule(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Get module
	mod, exists := m.loadedModules[name]
	if !exists {
		return fmt.Errorf("module %s not found", name)
	}
	
	// Get loader
	info := m.moduleInfo[name]
	l, exists := m.loaders[info.LoaderType]
	if !exists {
		return fmt.Errorf("loader %s not found", info.LoaderType)
	}
	
	// Unload module
	err := l.Unload(mod)
	if err != nil {
		return fmt.Errorf("failed to unload module: %w", err)
	}
	
	// Remove module
	delete(m.loadedModules, name)
	delete(m.moduleInfo, name)
	
	return nil
}

// GetModule returns a loaded module
func (m *ModuleManager) GetModule(name string) (module.Module, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	mod, exists := m.loadedModules[name]
	if !exists {
		return nil, fmt.Errorf("module %s not found", name)
	}
	
	return mod, nil
}

// GetModuleInfo returns information about a loaded module
func (m *ModuleManager) GetModuleInfo(name string) (ModuleInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	info, exists := m.moduleInfo[name]
	if !exists {
		return ModuleInfo{}, fmt.Errorf("module %s not found", name)
	}
	
	return info, nil
}

// ListModules returns a list of loaded modules
func (m *ModuleManager) ListModules() map[string]ModuleInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Create a copy of the module info map
	result := make(map[string]ModuleInfo, len(m.moduleInfo))
	for name, info := range m.moduleInfo {
		result[name] = info
	}
	
	return result
}

// ExecModule executes a command on a module
func (m *ModuleManager) ExecModule(name, command string, args ...interface{}) (interface{}, error) {
	m.mutex.RLock()
	mod, exists := m.loadedModules[name]
	m.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("module %s not found", name)
	}
	
	// Execute command
	result, err := mod.Exec(command, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}
	
	// Update status
	m.mutex.Lock()
	info := m.moduleInfo[name]
	info.Status = mod.GetStatus()
	m.moduleInfo[name] = info
	m.mutex.Unlock()
	
	return result, nil
}

// PauseModule pauses a module
func (m *ModuleManager) PauseModule(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	mod, exists := m.loadedModules[name]
	if !exists {
		return fmt.Errorf("module %s not found", name)
	}
	
	// Pause module
	err := mod.Pause()
	if err != nil {
		return fmt.Errorf("failed to pause module: %w", err)
	}
	
	// Update status
	info := m.moduleInfo[name]
	info.Status = mod.GetStatus()
	m.moduleInfo[name] = info
	
	return nil
}

// ResumeModule resumes a paused module
func (m *ModuleManager) ResumeModule(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	mod, exists := m.loadedModules[name]
	if !exists {
		return fmt.Errorf("module %s not found", name)
	}
	
	// Resume module
	err := mod.Resume()
	if err != nil {
		return fmt.Errorf("failed to resume module: %w", err)
	}
	
	// Update status
	info := m.moduleInfo[name]
	info.Status = mod.GetStatus()
	m.moduleInfo[name] = info
	
	return nil
}

// ShutdownModule shuts down a module
func (m *ModuleManager) ShutdownModule(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	mod, exists := m.loadedModules[name]
	if !exists {
		return fmt.Errorf("module %s not found", name)
	}
	
	// Shutdown module
	err := mod.Shutdown()
	if err != nil {
		return fmt.Errorf("failed to shutdown module: %w", err)
	}
	
	// Update status
	info := m.moduleInfo[name]
	info.Status = mod.GetStatus()
	m.moduleInfo[name] = info
	
	return nil
}

// ShutdownAllModules shuts down all modules
func (m *ModuleManager) ShutdownAllModules() []error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	var errors []error
	
	// Shutdown all modules
	for name, mod := range m.loadedModules {
		err := mod.Shutdown()
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to shutdown module %s: %w", name, err))
		}
		
		// Update status
		info := m.moduleInfo[name]
		info.Status = mod.GetStatus()
		m.moduleInfo[name] = info
	}
	
	return errors
}
