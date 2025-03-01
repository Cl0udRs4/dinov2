package security

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// AntiSandboxOptions configures anti-sandbox behavior
type AntiSandboxOptions struct {
	EnableHardwareChecks   bool
	EnableVirtualizationChecks bool
	EnableUserChecks       bool
	EnableNetworkChecks    bool
	EnableTimeChecks       bool
	EnableProcessChecks    bool
	EnableFileChecks       bool
	DelayExecution         bool
	DelayDuration          time.Duration
}

// DefaultAntiSandboxOptions returns default anti-sandbox options
func DefaultAntiSandboxOptions() AntiSandboxOptions {
	return AntiSandboxOptions{
		EnableHardwareChecks:   true,
		EnableVirtualizationChecks: true,
		EnableUserChecks:       true,
		EnableNetworkChecks:    true,
		EnableTimeChecks:       true,
		EnableProcessChecks:    true,
		EnableFileChecks:       true,
		DelayExecution:         true,
		DelayDuration:          30 * time.Second,
	}
}

// AntiSandbox implements anti-sandbox techniques
type AntiSandbox struct {
	options     AntiSandboxOptions
	isSandboxed bool
	detections  map[string]int
}

// NewAntiSandbox creates a new anti-sandbox with the specified options
func NewAntiSandbox(options AntiSandboxOptions) *AntiSandbox {
	return &AntiSandbox{
		options:     options,
		isSandboxed: false,
		detections:  make(map[string]int),
	}
}

// RunChecks performs anti-sandbox checks
func (a *AntiSandbox) RunChecks() bool {
	// Initialize detection result
	detected := false

	// Delay execution if enabled
	if a.options.DelayExecution {
		time.Sleep(a.options.DelayDuration)
	}

	// Run platform-specific checks
	switch runtime.GOOS {
	case "windows":
		detected = a.runWindowsChecks() || detected
	case "linux":
		detected = a.runLinuxChecks() || detected
	case "darwin":
		detected = a.runDarwinChecks() || detected
	}

	// Run common checks
	if a.options.EnableTimeChecks {
		if a.checkTimeAcceleration() {
			a.recordDetection("time")
			detected = true
		}
	}

	if a.options.EnableUserChecks {
		if a.checkUserActivity() {
			a.recordDetection("user")
			detected = true
		}
	}

	// Update sandboxed state
	a.isSandboxed = detected

	return detected
}

// IsSandboxed returns true if a sandbox has been detected
func (a *AntiSandbox) IsSandboxed() bool {
	return a.isSandboxed
}

// GetDetections returns a map of detection types and counts
func (a *AntiSandbox) GetDetections() map[string]int {
	return a.detections
}

// recordDetection records a detection of the specified type
func (a *AntiSandbox) recordDetection(detectionType string) {
	a.detections[detectionType]++
}

// checkTimeAcceleration checks for time acceleration often used in sandboxes
func (a *AntiSandbox) checkTimeAcceleration() bool {
	// Measure actual elapsed time vs. expected time
	start := time.Now()
	time.Sleep(100 * time.Millisecond)
	elapsed := time.Since(start)

	// If elapsed time is significantly less than expected, time might be accelerated
	if elapsed < 50*time.Millisecond {
		return true
	}

	return false
}

// checkUserActivity checks for signs of user activity
func (a *AntiSandbox) checkUserActivity() bool {
	// In a real implementation, this would check for:
	// - Mouse movement
	// - Keyboard input
	// - Window focus changes
	// - Process creation/termination patterns
	
	// Placeholder for user activity checks
	return false
}

// runWindowsChecks runs Windows-specific anti-sandbox checks
func (a *AntiSandbox) runWindowsChecks() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	detected := false

	// Check for virtualization artifacts
	if a.options.EnableVirtualizationChecks {
		// In a real implementation, this would check for:
		// - VMware registry keys and devices
		// - VirtualBox registry keys and devices
		// - Hyper-V specific artifacts
		// - QEMU/KVM specific artifacts
		// - Xen specific artifacts
	}

	// Check for sandbox-specific processes
	if a.options.EnableProcessChecks {
		// In a real implementation, this would check for:
		// - Sandbox analysis tools
		// - Monitoring processes
		// - VM tools processes
	}

	// Check for sandbox-specific files
	if a.options.EnableFileChecks {
		// In a real implementation, this would check for:
		// - VM tools installation files
		// - Sandbox-specific files
		// - Analysis tool files
	}

	return detected
}

// runLinuxChecks runs Linux-specific anti-sandbox checks
func (a *AntiSandbox) runLinuxChecks() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	detected := false

	// Check for virtualization artifacts
	if a.options.EnableVirtualizationChecks {
		// Check /proc/cpuinfo for virtualization flags
		data, err := os.ReadFile("/proc/cpuinfo")
		if err == nil {
			cpuinfo := string(data)
			
			// Check for hypervisor flag
			if strings.Contains(cpuinfo, "hypervisor") {
				a.recordDetection("hypervisor")
				detected = true
			}
			
			// Check for common VM CPU models
			vmCpus := []string{"QEMU", "KVM", "VMware", "VirtualBox", "Xen"}
			for _, vmCpu := range vmCpus {
				if strings.Contains(cpuinfo, vmCpu) {
					a.recordDetection("vm_cpu")
					detected = true
				}
			}
		}
		
		// Check for VM-specific devices
		if _, err := os.Stat("/dev/vboxguest"); err == nil {
			a.recordDetection("vbox_device")
			detected = true
		}
		
		if _, err := os.Stat("/dev/vmware"); err == nil {
			a.recordDetection("vmware_device")
			detected = true
		}
	}

	// Check for sandbox-specific processes
	if a.options.EnableProcessChecks {
		// Check for VM tools processes
		vmProcesses := []string{"VBoxService", "VBoxClient", "vmtoolsd", "qemu"}
		for _, proc := range vmProcesses {
			// In a real implementation, this would use ps or /proc to check for these processes
			// For now, we'll just check if the process name exists in /proc
			matches, _ := filepath.Glob(fmt.Sprintf("/proc/*/comm"))
			for _, match := range matches {
				data, err := os.ReadFile(match)
				if err == nil && strings.Contains(string(data), proc) {
					a.recordDetection("vm_process")
					detected = true
				}
			}
		}
	}

	// Check for sandbox-specific files
	if a.options.EnableFileChecks {
		// Check for VM tools files
		vmFiles := []string{
			"/etc/vmware-tools",
			"/etc/virtualbox",
			"/etc/xen",
			"/var/lib/cuckoo",
		}
		
		for _, file := range vmFiles {
			if _, err := os.Stat(file); err == nil {
				a.recordDetection("vm_file")
				detected = true
			}
		}
	}

	// Check hardware characteristics
	if a.options.EnableHardwareChecks {
		// Check for low memory (sandboxes often have limited resources)
		var memInfo runtime.MemStats
		runtime.ReadMemStats(&memInfo)
		
		// If total memory is suspiciously low (less than 2GB)
		if memInfo.TotalAlloc < 2*1024*1024*1024 {
			a.recordDetection("low_memory")
			detected = true
		}
		
		// Check number of CPUs (sandboxes often have few CPUs)
		if runtime.NumCPU() < 2 {
			a.recordDetection("low_cpu")
			detected = true
		}
	}

	return detected
}

// runDarwinChecks runs macOS-specific anti-sandbox checks
func (a *AntiSandbox) runDarwinChecks() bool {
	if runtime.GOOS != "darwin" {
		return false
	}

	detected := false

	// Check for virtualization artifacts
	if a.options.EnableVirtualizationChecks {
		// In a real implementation, this would check for:
		// - VMware tools
		// - VirtualBox artifacts
		// - Parallels artifacts
	}

	// Check for sandbox-specific processes
	if a.options.EnableProcessChecks {
		// In a real implementation, this would check for:
		// - VM tools processes
		// - Analysis tool processes
	}

	return detected
}

// StartMonitoring starts continuous monitoring for sandboxes
func (a *AntiSandbox) StartMonitoring(interval time.Duration, callback func(bool)) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				detected := a.RunChecks()
				if callback != nil {
					callback(detected)
				}
			}
		}
	}()
}
