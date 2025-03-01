package loader

import (
	"dinoc2/pkg/module"
	"fmt"
	"sync"
)

// LoaderFactory manages module loaders
type LoaderFactory struct {
	loaders map[LoaderType]ModuleLoader
	mutex   sync.RWMutex
}

// NewLoaderFactory creates a new loader factory
func NewLoaderFactory() *LoaderFactory {
	return &LoaderFactory{
		loaders: make(map[LoaderType]ModuleLoader),
	}
}

// GetLoader returns a loader of the specified type
func (f *LoaderFactory) GetLoader(loaderType LoaderType) (ModuleLoader, error) {
	f.mutex.RLock()
	loader, exists := f.loaders[loaderType]
	f.mutex.RUnlock()

	if exists {
		return loader, nil
	}

	// Create a new loader
	loader, err := createLoader(loaderType)
	if err != nil {
		return nil, err
	}

	// Store the loader
	f.mutex.Lock()
	f.loaders[loaderType] = loader
	f.mutex.Unlock()

	return loader, nil
}

// createLoader creates a new loader of the specified type
func createLoader(loaderType LoaderType) (ModuleLoader, error) {
	switch loaderType {
	case LoaderTypeNative:
		return NewNativeLoader(), nil
	case LoaderTypePlugin:
		return NewPluginLoader(), nil
	case LoaderTypeDLL:
		return NewDLLLoader(), nil
	case LoaderTypeWasm:
		return NewWasmLoader(), nil
	case LoaderTypeRPC:
		return NewRPCLoader(), nil
	default:
		return nil, fmt.Errorf("unsupported loader type: %s", loaderType)
	}
}

// GetSupportedLoaderTypes returns a list of supported loader types
func (f *LoaderFactory) GetSupportedLoaderTypes() []LoaderType {
	return GetSupportedLoaders()
}

// LoadModule loads a module using the specified loader
func (f *LoaderFactory) LoadModule(loaderType LoaderType, path string) (module.Module, error) {
	loader, err := f.GetLoader(loaderType)
	if err != nil {
		return nil, fmt.Errorf("failed to get loader: %w", err)
	}

	return loader.Load(path)
}

// UnloadModule unloads a module
func (f *LoaderFactory) UnloadModule(loaderType LoaderType, mod module.Module) error {
	loader, err := f.GetLoader(loaderType)
	if err != nil {
		return fmt.Errorf("failed to get loader: %w", err)
	}

	return loader.Unload(mod)
}
