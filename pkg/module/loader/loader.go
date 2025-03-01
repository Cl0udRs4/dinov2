package loader

import (
	"dinoc2/pkg/module"
	"errors"
	"fmt"
	"runtime"
)

// LoaderType represents the type of module loader
type LoaderType string

const (
	// LoaderTypeNative represents a native Go module loader
	LoaderTypeNative LoaderType = "native"
	
	// LoaderTypePlugin represents a Go plugin loader (Linux)
	LoaderTypePlugin LoaderType = "plugin"
	
	// LoaderTypeDLL represents a DLL loader (Windows)
	LoaderTypeDLL LoaderType = "dll"
	
	// LoaderTypeWasm represents a WebAssembly loader
	LoaderTypeWasm LoaderType = "wasm"
	
	// LoaderTypeRPC represents an RPC-based loader
	LoaderTypeRPC LoaderType = "rpc"
)

// ModuleLoader defines the interface for module loaders
type ModuleLoader interface {
	// Load loads a module from the specified path
	Load(path string) (module.Module, error)
	
	// Unload unloads a module
	Unload(module module.Module) error
	
	// GetType returns the loader type
	GetType() LoaderType
	
	// IsSupported returns true if the loader is supported on the current platform
	IsSupported() bool
}

// GetSupportedLoaders returns a list of supported loaders for the current platform
func GetSupportedLoaders() []LoaderType {
	loaders := []LoaderType{LoaderTypeNative} // Native is always supported
	
	// Add platform-specific loaders
	switch runtime.GOOS {
	case "linux":
		loaders = append(loaders, LoaderTypePlugin)
	case "windows":
		loaders = append(loaders, LoaderTypeDLL)
	}
	
	// Add cross-platform loaders
	loaders = append(loaders, LoaderTypeWasm, LoaderTypeRPC)
	
	return loaders
}

// GetLoader returns a loader of the specified type
func GetLoader(loaderType LoaderType) (ModuleLoader, error) {
	switch loaderType {
	case LoaderTypeNative:
		return NewNativeLoader(), nil
	case LoaderTypePlugin:
		if runtime.GOOS != "linux" {
			return nil, errors.New("plugin loader is only supported on Linux")
		}
		return NewPluginLoader(), nil
	case LoaderTypeDLL:
		if runtime.GOOS != "windows" {
			return nil, errors.New("DLL loader is only supported on Windows")
		}
		return NewDLLLoader(), nil
	case LoaderTypeWasm:
		return NewWasmLoader(), nil
	case LoaderTypeRPC:
		return NewRPCLoader(), nil
	default:
		return nil, fmt.Errorf("unsupported loader type: %s", loaderType)
	}
}
