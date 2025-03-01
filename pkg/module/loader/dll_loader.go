package loader

import (
	"dinoc2/pkg/module"
	"errors"
	"fmt"
	"runtime"
	"sync"
)

// DLLLoader implements a loader for DLL modules (Windows only)
type DLLLoader struct {
	handles     map[string]uintptr
	modules     map[string]module.Module
	moduleInfos map[string]module.ModuleInfo
	mutex       sync.RWMutex
}

// NewDLLLoader creates a new DLL loader
func NewDLLLoader() *DLLLoader {
	return &DLLLoader{
		handles:     make(map[string]uintptr),
		modules:     make(map[string]module.Module),
		moduleInfos: make(map[string]module.ModuleInfo),
	}
}

// Load loads a module from a DLL file
func (l *DLLLoader) Load(path string) (module.Module, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// Check if DLL is already loaded
	if mod, exists := l.modules[path]; exists {
		return mod, nil
	}
	
	// On non-Windows platforms, return an error
	if runtime.GOOS != "windows" {
		return nil, errors.New("DLL loader is only supported on Windows")
	}
	
	// In a real implementation, we would use syscall.LoadLibrary to load the DLL
	// and syscall.GetProcAddress to get function pointers
	// For now, we'll just return an error
	return nil, errors.New("DLL loading not implemented")
}

// Unload unloads a module
func (l *DLLLoader) Unload(mod module.Module) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// Find DLL path
	var dllPath string
	for path, m := range l.modules {
		if m == mod {
			dllPath = path
			break
		}
	}
	
	if dllPath == "" {
		return errors.New("module not found")
	}
	
	// Shutdown module
	err := mod.Shutdown()
	if err != nil {
		return fmt.Errorf("failed to shutdown module: %w", err)
	}
	
	// In a real implementation, we would use syscall.FreeLibrary to unload the DLL
	
	// Remove DLL and module
	delete(l.handles, dllPath)
	delete(l.modules, dllPath)
	
	return nil
}

// GetType returns the loader type
func (l *DLLLoader) GetType() LoaderType {
	return LoaderTypeDLL
}

// IsSupported returns true if the loader is supported on the current platform
func (l *DLLLoader) IsSupported() bool {
	return runtime.GOOS == "windows"
}
