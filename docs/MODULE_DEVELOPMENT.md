# DinoC2 Module Development Guide

This guide provides instructions for developing custom modules for the DinoC2 system.

## Table of Contents

1. [Module System Overview](#module-system-overview)
2. [Module Interface](#module-interface)
3. [Module Types](#module-types)
4. [Development Environment](#development-environment)
5. [Creating a Basic Module](#creating-a-basic-module)
6. [Module Loading Mechanisms](#module-loading-mechanisms)
7. [Module Isolation](#module-isolation)
8. [Module Security](#module-security)
9. [Testing Modules](#testing-modules)
10. [Module Distribution](#module-distribution)
11. [Best Practices](#best-practices)

## Module System Overview

The DinoC2 module system is designed to be flexible, extensible, and secure. Modules are self-contained units of functionality that can be loaded and executed by the client. The module system supports various loading mechanisms, isolation techniques, and security features.

### Key Components

1. **Module Interface**: Defines the standard interface that all modules must implement
2. **Module Registry**: Manages module registration and discovery
3. **Module Loader**: Handles loading modules using different mechanisms
4. **Module Manager**: Coordinates module execution and lifecycle
5. **Module Isolation**: Provides isolation between modules and the main client process

### Module Lifecycle

1. **Registration**: Modules are registered with the module registry
2. **Loading**: Modules are loaded using the appropriate loader
3. **Initialization**: Modules are initialized with configuration parameters
4. **Execution**: Modules execute their functionality
5. **Shutdown**: Modules clean up resources and shut down

## Module Interface

All modules must implement the `Module` interface defined in `pkg/module/module.go`:

```go
// Module defines the interface for a module
type Module interface {
    // Init initializes the module with the given parameters
    Init(params map[string]interface{}) error
    
    // Exec executes the module with the given command and arguments
    Exec(command string, args ...interface{}) (interface{}, error)
    
    // Shutdown cleans up resources and shuts down the module
    Shutdown() error
    
    // GetStatus returns the current status of the module
    GetStatus() ModuleStatus
    
    // GetCapabilities returns the capabilities of the module
    GetCapabilities() []string
}

// ModuleStatus represents the status of a module
type ModuleStatus struct {
    Name        string
    Version     string
    Description string
    Author      string
    Running     bool
    Error       string
}
```

## Module Types

DinoC2 supports several types of modules:

1. **Native Modules**: Built-in modules compiled into the client
2. **Plugin Modules**: Dynamically loaded Go plugins
3. **WebAssembly Modules**: Modules compiled to WebAssembly
4. **DLL Modules**: Windows-specific dynamic libraries
5. **RPC Modules**: Modules that communicate via RPC

## Development Environment

### Prerequisites

- Go 1.23.5 or higher
- Git
- Make
- GCC or equivalent C compiler (for CGO support)
- WebAssembly toolchain (for WebAssembly modules)

### Setting Up the Development Environment

1. Clone the DinoC2 repository:

```bash
git clone https://github.com/Cl0udRs4/dinov2.git
cd dinov2
```

2. Install dependencies:

```bash
make setup
```

3. Build the development tools:

```bash
make tools
```

## Creating a Basic Module

### Module Structure

A basic module consists of the following files:

```
pkg/module/mymodule/
├── mymodule.go       # Main module implementation
└── mymodule_test.go  # Module tests
```

### Module Implementation

Here's an example of a basic module implementation:

```go
package mymodule

import (
    "fmt"
    "dinoc2/pkg/module"
)

// MyModule implements the Module interface
type MyModule struct {
    name        string
    version     string
    description string
    author      string
    running     bool
    error       string
    config      map[string]interface{}
}

// NewMyModule creates a new instance of MyModule
func NewMyModule() module.Module {
    return &MyModule{
        name:        "mymodule",
        version:     "1.0.0",
        description: "My custom module",
        author:      "Your Name",
        running:     false,
        error:       "",
    }
}

// Init initializes the module with the given parameters
func (m *MyModule) Init(params map[string]interface{}) error {
    m.config = params
    return nil
}

// Exec executes the module with the given command and arguments
func (m *MyModule) Exec(command string, args ...interface{}) (interface{}, error) {
    m.running = true
    defer func() { m.running = false }()
    
    switch command {
    case "hello":
        name := "World"
        if len(args) > 0 {
            if n, ok := args[0].(string); ok {
                name = n
            }
        }
        return fmt.Sprintf("Hello, %s!", name), nil
    default:
        m.error = fmt.Sprintf("unknown command: %s", command)
        return nil, fmt.Errorf(m.error)
    }
}

// Shutdown cleans up resources and shuts down the module
func (m *MyModule) Shutdown() error {
    m.running = false
    return nil
}

// GetStatus returns the current status of the module
func (m *MyModule) GetStatus() module.ModuleStatus {
    return module.ModuleStatus{
        Name:        m.name,
        Version:     m.version,
        Description: m.description,
        Author:      m.author,
        Running:     m.running,
        Error:       m.error,
    }
}

// GetCapabilities returns the capabilities of the module
func (m *MyModule) GetCapabilities() []string {
    return []string{"hello"}
}

// Register registers the module with the module registry
func init() {
    module.RegisterModule("mymodule", NewMyModule)
}
```

### Module Registration

Modules must be registered with the module registry to be discoverable. There are two ways to register a module:

1. Using the `init()` function (as shown in the example above)
2. Explicitly calling `module.RegisterModule()` in your code

## Module Loading Mechanisms

DinoC2 supports several module loading mechanisms:

### Native Modules

Native modules are compiled directly into the client binary. They are the simplest and most efficient type of module.

To create a native module:

1. Implement the `Module` interface
2. Register the module with the module registry
3. Import the module package in the client code

### Plugin Modules

Plugin modules are dynamically loaded Go plugins. They provide more flexibility than native modules but are only supported on certain platforms (mainly Linux).

To create a plugin module:

1. Implement the `Module` interface
2. Build the module as a Go plugin:

```bash
go build -buildmode=plugin -o mymodule.so pkg/module/mymodule/mymodule.go
```

3. Load the plugin at runtime:

```go
loader := loader.NewPluginLoader()
module, err := loader.Load("mymodule", "/path/to/mymodule.so")
```

### WebAssembly Modules

WebAssembly modules are compiled to WebAssembly and can run on any platform that supports WebAssembly.

To create a WebAssembly module:

1. Implement the `Module` interface
2. Build the module as a WebAssembly module:

```bash
GOOS=js GOARCH=wasm go build -o mymodule.wasm pkg/module/mymodule/mymodule.go
```

3. Load the WebAssembly module at runtime:

```go
loader := loader.NewWasmLoader()
module, err := loader.Load("mymodule", "/path/to/mymodule.wasm")
```

### DLL Modules

DLL modules are Windows-specific dynamic libraries.

To create a DLL module:

1. Implement the `Module` interface
2. Build the module as a DLL:

```bash
go build -buildmode=c-shared -o mymodule.dll pkg/module/mymodule/mymodule.go
```

3. Load the DLL module at runtime:

```go
loader := loader.NewDllLoader()
module, err := loader.Load("mymodule", "/path/to/mymodule.dll")
```

### RPC Modules

RPC modules communicate with the client via RPC. They can be written in any language that supports the RPC protocol.

To create an RPC module:

1. Implement the RPC server that exposes the module functionality
2. Start the RPC server
3. Connect to the RPC server from the client:

```go
loader := loader.NewRpcLoader()
module, err := loader.Load("mymodule", "localhost:8080")
```

## Module Isolation

DinoC2 provides several isolation mechanisms to protect the client from malicious or buggy modules:

### Process Isolation

Process isolation runs modules in separate processes:

```go
isolator := isolation.NewProcessIsolator()
isolatedModule := isolator.Isolate(module)
```

### Resource Limitation

Resource limitation restricts the resources available to modules:

```go
limits := isolation.ResourceLimits{
    Memory:    1024 * 1024 * 100, // 100 MB
    CPU:       1,                 // 1 CPU core
    Timeout:   60 * time.Second,  // 60 second timeout
}
isolator := isolation.NewResourceLimitIsolator(limits)
isolatedModule := isolator.Isolate(module)
```

### Sandbox Isolation

Sandbox isolation runs modules in a sandboxed environment:

```go
isolator := isolation.NewSandboxIsolator()
isolatedModule := isolator.Isolate(module)
```

## Module Security

DinoC2 provides several security features for modules:

### Module Signing

Modules can be signed to ensure their authenticity:

```bash
# Generate a signing key
openssl genrsa -out module_signing.key 2048

# Sign a module
dinoc2-builder sign -key module_signing.key -module mymodule.so
```

### Module Verification

Signed modules can be verified before loading:

```go
verifier := security.NewSignatureVerifier(security.DefaultSignatureOptions())
valid, err := verifier.VerifyModule(moduleData, signature, certificate)
if !valid {
    return fmt.Errorf("module verification failed: %w", err)
}
```

### Memory Protection

Sensitive data in modules can be protected:

```go
memProtect := security.NewMemoryProtection(security.DefaultMemoryProtectionOptions())
err := memProtect.Protect("credentials", []byte("sensitive_data"))
```

## Testing Modules

### Unit Testing

Create unit tests for your module:

```go
package mymodule

import (
    "testing"
)

func TestMyModule(t *testing.T) {
    module := NewMyModule()
    
    // Test initialization
    err := module.Init(map[string]interface{}{})
    if err != nil {
        t.Fatalf("Failed to initialize module: %v", err)
    }
    
    // Test execution
    result, err := module.Exec("hello", "Test")
    if err != nil {
        t.Fatalf("Failed to execute module: %v", err)
    }
    
    expected := "Hello, Test!"
    if result != expected {
        t.Fatalf("Unexpected result: got %v, want %v", result, expected)
    }
    
    // Test shutdown
    err = module.Shutdown()
    if err != nil {
        t.Fatalf("Failed to shutdown module: %v", err)
    }
}
```

### Integration Testing

Create integration tests for your module:

```go
package test

import (
    "dinoc2/pkg/module"
    "dinoc2/pkg/module/mymodule"
    "testing"
)

func TestMyModuleIntegration(t *testing.T) {
    // Register module
    module.RegisterModule("mymodule", mymodule.NewMyModule)
    
    // Create module manager
    manager, err := module.NewModuleManager()
    if err != nil {
        t.Fatalf("Failed to create module manager: %v", err)
    }
    
    // Load module
    err = manager.LoadModule("mymodule")
    if err != nil {
        t.Fatalf("Failed to load module: %v", err)
    }
    
    // Initialize module
    err = manager.InitModule("mymodule", map[string]interface{}{})
    if err != nil {
        t.Fatalf("Failed to initialize module: %v", err)
    }
    
    // Execute module
    result, err := manager.ExecModule("mymodule", "hello", "Test")
    if err != nil {
        t.Fatalf("Failed to execute module: %v", err)
    }
    
    expected := "Hello, Test!"
    if result != expected {
        t.Fatalf("Unexpected result: got %v, want %v", result, expected)
    }
    
    // Shutdown module
    err = manager.ShutdownModule("mymodule")
    if err != nil {
        t.Fatalf("Failed to shutdown module: %v", err)
    }
}
```

### Running Tests

Run tests for your module:

```bash
go test -v dinoc2/pkg/module/mymodule
```

## Module Distribution

### Building Modules

Build your module for distribution:

```bash
# Build native module
go build -o mymodule.so -buildmode=plugin pkg/module/mymodule/mymodule.go

# Build WebAssembly module
GOOS=js GOARCH=wasm go build -o mymodule.wasm pkg/module/mymodule/mymodule.go

# Build DLL module
go build -buildmode=c-shared -o mymodule.dll pkg/module/mymodule/mymodule.go
```

### Packaging Modules

Package your module for distribution:

```bash
# Create a module package
dinoc2-builder package -module mymodule.so -output mymodule.zip -type plugin
```

### Publishing Modules

Publish your module to a module repository:

```bash
# Publish a module
dinoc2-builder publish -package mymodule.zip -repo https://example.com/modules
```

## Best Practices

### Module Design

1. **Keep modules focused**: Each module should have a single responsibility
2. **Provide clear documentation**: Document your module's functionality, parameters, and return values
3. **Handle errors gracefully**: Properly handle and report errors
4. **Clean up resources**: Always clean up resources in the `Shutdown()` method
5. **Respect resource limits**: Be mindful of memory and CPU usage

### Security Considerations

1. **Validate input**: Always validate input parameters
2. **Protect sensitive data**: Use memory protection for sensitive data
3. **Limit permissions**: Request only the permissions your module needs
4. **Sign your modules**: Always sign your modules for distribution
5. **Avoid unsafe operations**: Be cautious with unsafe operations like direct memory access

### Performance Optimization

1. **Minimize dependencies**: Keep dependencies to a minimum
2. **Optimize resource usage**: Be efficient with memory and CPU usage
3. **Use appropriate data structures**: Choose the right data structures for your use case
4. **Implement caching**: Cache results when appropriate
5. **Profile your code**: Use profiling tools to identify bottlenecks

## Conclusion

This guide covers the basics of developing modules for the DinoC2 system. For more detailed information, refer to the following documentation:

- [Architecture](ARCHITECTURE.md): Overview of the system architecture
- [Security](SECURITY.md): Details on security features
- [User Guide](USER_GUIDE.md): Instructions for using the DinoC2 system
