package builder

import (
	"dinoc2/pkg/module/registry"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"
)

// ModuleBuilder helps with module creation and packaging
type ModuleBuilder struct {
	outputDir string
	templates map[string]string
}

// NewModuleBuilder creates a new module builder
func NewModuleBuilder(outputDir string) *ModuleBuilder {
	return &ModuleBuilder{
		outputDir: outputDir,
		templates: make(map[string]string),
	}
}

// RegisterTemplate registers a template for module generation
func (b *ModuleBuilder) RegisterTemplate(name, templateStr string) {
	b.templates[name] = templateStr
}

// CreateModule creates a new module from a template
func (b *ModuleBuilder) CreateModule(templateName, moduleName, packageName string, info registry.ModuleInfo) error {
	// Check if template exists
	templateStr, exists := b.templates[templateName]
	if !exists {
		return fmt.Errorf("template %s not found", templateName)
	}

	// Create output directory
	moduleDir := filepath.Join(b.outputDir, moduleName)
	err := os.MkdirAll(moduleDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create module directory: %w", err)
	}

	// Parse template
	tmpl, err := template.New(templateName).Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create module file
	moduleFile := filepath.Join(moduleDir, moduleName+".go")
	file, err := os.Create(moduleFile)
	if err != nil {
		return fmt.Errorf("failed to create module file: %w", err)
	}
	defer file.Close()

	// Execute template
	err = tmpl.Execute(file, struct {
		PackageName  string
		ModuleName   string
		ModuleInfo   registry.ModuleInfo
	}{
		PackageName:  packageName,
		ModuleName:   moduleName,
		ModuleInfo:   info,
	})
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// BuildModule builds a module as a plugin
func (b *ModuleBuilder) BuildModule(moduleName string) error {
	moduleDir := filepath.Join(b.outputDir, moduleName)
	outputFile := filepath.Join(moduleDir, moduleName+".so")

	// Check if we're on Linux (plugins only supported on Linux)
	if runtime.GOOS != "linux" {
		return fmt.Errorf("plugin building is only supported on Linux")
	}

	// Build the module as a plugin
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", outputFile, ".")
	cmd.Dir = moduleDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to build module: %w", err)
	}

	return nil
}

// GetDefaultTemplate returns the default module template
func GetDefaultTemplate() string {
	return `package {{.PackageName}}

import (
	"dinoc2/pkg/module"
	"fmt"
	"sync"
)

// {{.ModuleName}}Module implements a {{.ModuleName}} module
type {{.ModuleName}}Module struct {
	name        string
	description string
	mutex       sync.Mutex
	isRunning   bool
	isPaused    bool
}

// New{{.ModuleName}}Module creates a new {{.ModuleName}} module
func New{{.ModuleName}}Module() module.Module {
	return &{{.ModuleName}}Module{
		name:        "{{.ModuleName}}",
		description: "{{.ModuleInfo.Description}}",
	}
}

// Init initializes the module
func (m *{{.ModuleName}}Module) Init(params map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if already running
	if m.isRunning {
		return fmt.Errorf("module already running")
	}

	// Initialize the module
	// Initialize resources based on parameters
	if params != nil {
		// Process initialization parameters
		if name, ok := params["name"].(string); ok && name != "" {
			m.name = name
		}
		
		if desc, ok := params["description"].(string); ok && desc != "" {
			m.description = desc
		}
		
		// Initialize any additional resources needed by the module
		if err := m.initializeResources(params); err != nil {
			return fmt.Errorf("failed to initialize resources: %w", err)
		}
	}
	
	// Log initialization
	fmt.Printf("Module %s initialized successfully\n", m.name)
	
	m.isRunning = true
	return nil
}

// initializeResources initializes resources needed by the module
func (m *{{.ModuleName}}Module) initializeResources(params map[string]interface{}) error {
	// Initialize any resources needed by the module
	// This could include:
	// - Opening files
	// - Establishing network connections
	// - Allocating memory
	// - Loading configuration
	
	return nil
}

// Exec executes a command on the module
func (m *{{.ModuleName}}Module) Exec(command string, args ...interface{}) (interface{}, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if running
	if !m.isRunning {
		return nil, fmt.Errorf("module not running")
	}

	// Check if paused
	if m.isPaused {
		return nil, fmt.Errorf("module is paused")
	}

	// Execute the command
	switch command {
	case "help":
		return "Available commands: help", nil
	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

// Shutdown shuts down the module
func (m *{{.ModuleName}}Module) Shutdown() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if running
	if !m.isRunning {
		return nil
	}

	// Shutdown the module
	// Release resources
	if err := m.releaseResources(); err != nil {
		return fmt.Errorf("failed to release resources: %w", err)
	}
	
	// Log shutdown
	fmt.Printf("Module %s shutdown successfully\n", m.name)
	
	m.isRunning = false
	return nil
}

// releaseResources releases resources used by the module
func (m *{{.ModuleName}}Module) releaseResources() error {
	// Release any resources used by the module
	// This could include:
	// - Closing files
	// - Closing network connections
	// - Freeing memory
	// - Saving state
	
	return nil
}

// GetStatus returns the module status
func (m *{{.ModuleName}}Module) GetStatus() module.ModuleStatus {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	status := module.ModuleStatus{
		Running: m.isRunning,
		Stats: map[string]interface{}{
			"paused": m.isPaused,
		},
	}

	return status
}

// GetCapabilities returns the module capabilities
func (m *{{.ModuleName}}Module) GetCapabilities() []string {
	return []string{
		{{range .ModuleInfo.Capabilities}}"{{.}}",
		{{end}}
	}
}

// Pause temporarily pauses the module's operations
func (m *{{.ModuleName}}Module) Pause() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isRunning {
		return fmt.Errorf("module not running")
	}

	if m.isPaused {
		return nil // Already paused
	}

	m.isPaused = true
	return nil
}

// Resume resumes the module's operations after a pause
func (m *{{.ModuleName}}Module) Resume() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isRunning {
		return fmt.Errorf("module not running")
	}

	if !m.isPaused {
		return nil // Not paused
	}

	m.isPaused = false
	return nil
}

// This is required for Go plugins
var New{{.ModuleName}} = New{{.ModuleName}}Module
`
}
