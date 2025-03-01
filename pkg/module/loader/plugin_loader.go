package loader

import (
	"dinoc2/pkg/module"
	"errors"
	"fmt"
	"plugin"
	"runtime"
	"sync"
)

// PluginLoader implements a loader for Go plugins (Linux only)
type PluginLoader struct {
	plugins     map[string]*plugin.Plugin
	modules     map[string]module.Module
	moduleInfos map[string]module.ModuleInfo
	mutex       sync.RWMutex
}

// NewPluginLoader creates a new plugin loader
func NewPluginLoader() *PluginLoader {
	return &PluginLoader{
		plugins:     make(map[string]*plugin.Plugin),
		modules:     make(map[string]module.Module),
		moduleInfos: make(map[string]module.ModuleInfo),
	}
}

// Load loads a module from a plugin file
func (l *PluginLoader) Load(path string) (module.Module, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// Check if plugin is already loaded
	if mod, exists := l.modules[path]; exists {
		return mod, nil
	}
	
	// Load plugin
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}
	
	// Look up module factory symbol
	factorySym, err := p.Lookup("NewModule")
	if err != nil {
		return nil, fmt.Errorf("failed to find NewModule symbol: %w", err)
	}
	
	// Convert symbol to factory function
	factory, ok := factorySym.(func() module.Module)
	if !ok {
		return nil, errors.New("NewModule symbol is not a module factory function")
	}
	
	// Create module
	mod := factory()
	if mod == nil {
		return nil, errors.New("module factory returned nil")
	}
	
	// Store plugin and module
	l.plugins[path] = p
	l.modules[path] = mod
	
	return mod, nil
}

// Unload unloads a module
func (l *PluginLoader) Unload(mod module.Module) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// Find plugin path
	var pluginPath string
	for path, m := range l.modules {
		if m == mod {
			pluginPath = path
			break
		}
	}
	
	if pluginPath == "" {
		return errors.New("module not found")
	}
	
	// Shutdown module
	err := mod.Shutdown()
	if err != nil {
		return fmt.Errorf("failed to shutdown module: %w", err)
	}
	
	// Remove plugin and module
	delete(l.plugins, pluginPath)
	delete(l.modules, pluginPath)
	
	return nil
}

// GetType returns the loader type
func (l *PluginLoader) GetType() LoaderType {
	return LoaderTypePlugin
}

// IsSupported returns true if the loader is supported on the current platform
func (l *PluginLoader) IsSupported() bool {
	return runtime.GOOS == "linux"
}
