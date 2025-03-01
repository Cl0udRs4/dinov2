package module

import "errors"

// ModuleStatus represents the current status of a module
type ModuleStatus struct {
	Running bool
	Error   error
	Stats   map[string]interface{}
}

// Module interface defines methods that all modules must implement
type Module interface {
	// Init initializes the module with the given parameters
	Init(params map[string]interface{}) error
	
	// Exec executes a command on the module
	Exec(command string, args ...interface{}) (interface{}, error)
	
	// Shutdown cleanly shuts down the module
	Shutdown() error
	
	// GetStatus returns the current status of the module
	GetStatus() ModuleStatus
	
	// GetCapabilities returns a list of capabilities supported by the module
	GetCapabilities() []string
	
	// Pause temporarily pauses the module's operations
	Pause() error
	
	// Resume resumes the module's operations after a pause
	Resume() error
}

// ModuleType represents the type of module
type ModuleType string

const (
	ModuleTypeShell ModuleType = "shell"
	ModuleTypeFile  ModuleType = "file"
	ModuleTypeInfo  ModuleType = "info"
	ModuleTypeProxy ModuleType = "proxy"
)

// ModuleInfo contains metadata about a module
type ModuleInfo struct {
	Name         string
	Type         ModuleType
	Version      string
	Author       string
	Description  string
	Dependencies []string
	Platforms    []string
}

// Registry manages available modules
type Registry struct {
	modules map[string]Module
	info    map[string]ModuleInfo
}

// NewRegistry creates a new module registry
func NewRegistry() *Registry {
	return &Registry{
		modules: make(map[string]Module),
		info:    make(map[string]ModuleInfo),
	}
}

// Register adds a module to the registry
func (r *Registry) Register(name string, module Module, info ModuleInfo) error {
	if _, exists := r.modules[name]; exists {
		return errors.New("module already registered")
	}
	
	r.modules[name] = module
	r.info[name] = info
	return nil
}

// Get retrieves a module from the registry
func (r *Registry) Get(name string) (Module, error) {
	module, exists := r.modules[name]
	if !exists {
		return nil, errors.New("module not found")
	}
	
	return module, nil
}

// GetInfo retrieves module info from the registry
func (r *Registry) GetInfo(name string) (ModuleInfo, error) {
	info, exists := r.info[name]
	if !exists {
		return ModuleInfo{}, errors.New("module not found")
	}
	
	return info, nil
}

// List returns a list of all registered modules
func (r *Registry) List() map[string]ModuleInfo {
	return r.info
}

// Unregister removes a module from the registry
func (r *Registry) Unregister(name string) error {
	if _, exists := r.modules[name]; !exists {
		return errors.New("module not found")
	}
	
	delete(r.modules, name)
	delete(r.info, name)
	return nil
}
