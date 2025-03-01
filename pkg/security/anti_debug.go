package security

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

// AntiDebugOptions configures anti-debugging behavior
type AntiDebugOptions struct {
	EnableTimingChecks    bool
	EnableTraceChecks     bool
	EnableParentChecks    bool
	EnableProcessChecks   bool
	EnableMemoryChecks    bool
	EnableEnvironmentChecks bool
	SelfDestruct          bool
	ObfuscateDetection    bool
}

// DefaultAntiDebugOptions returns default anti-debugging options
func DefaultAntiDebugOptions() AntiDebugOptions {
	return AntiDebugOptions{
		EnableTimingChecks:    true,
		EnableTraceChecks:     true,
		EnableParentChecks:    true,
		EnableProcessChecks:   true,
		EnableMemoryChecks:    true,
		EnableEnvironmentChecks: true,
		SelfDestruct:          false,
		ObfuscateDetection:    true,
	}
}

// AntiDebugger implements anti-debugging techniques
type AntiDebugger struct {
	options     AntiDebugOptions
	isDebugged  bool
	lastCheck   time.Time
	checkCount  int
	detections  map[string]int
}

// NewAntiDebugger creates a new anti-debugger with the specified options
func NewAntiDebugger(options AntiDebugOptions) *AntiDebugger {
	return &AntiDebugger{
		options:    options,
		isDebugged: false,
		lastCheck:  time.Now(),
		checkCount: 0,
		detections: make(map[string]int),
	}
}

// RunChecks performs anti-debugging checks
func (a *AntiDebugger) RunChecks() bool {
	// Record check time for timing analysis
	now := time.Now()
	elapsed := now.Sub(a.lastCheck)
	a.lastCheck = now
	a.checkCount++

	// Initialize detection result
	detected := false

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
	if a.options.EnableTimingChecks {
		if a.checkTimingAnomalies(elapsed) {
			a.recordDetection("timing")
			detected = true
		}
	}

	if a.options.EnableEnvironmentChecks {
		if a.checkEnvironmentVariables() {
			a.recordDetection("environment")
			detected = true
		}
	}

	// Update debugged state
	a.isDebugged = detected

	// Take action if debugger detected
	if detected && a.options.SelfDestruct {
		a.selfDestruct()
	}

	return detected
}

// IsDebugged returns true if a debugger has been detected
func (a *AntiDebugger) IsDebugged() bool {
	return a.isDebugged
}

// GetDetections returns a map of detection types and counts
func (a *AntiDebugger) GetDetections() map[string]int {
	return a.detections
}

// recordDetection records a detection of the specified type
func (a *AntiDebugger) recordDetection(detectionType string) {
	a.detections[detectionType]++
}

// checkTimingAnomalies checks for timing anomalies that might indicate debugging
func (a *AntiDebugger) checkTimingAnomalies(elapsed time.Duration) bool {
	// Skip the first few checks to establish a baseline
	if a.checkCount < 5 {
		return false
	}

	// Check for suspicious timing (e.g., unusually long pauses between checks)
	// This could indicate breakpoints or single-stepping
	if elapsed > 500*time.Millisecond {
		return true
	}

	return false
}

// checkEnvironmentVariables checks for environment variables that might indicate debugging
func (a *AntiDebugger) checkEnvironmentVariables() bool {
	// Check for common debugging environment variables
	debugEnvVars := []string{
		"TERM_PROGRAM", // Look for debugging terminals
		"DEBUG",
		"_DEBUG",
		"DEBUGGER",
		"LD_PRELOAD", // Common for hooking libraries
		"LD_AUDIT",
	}

	for _, envVar := range debugEnvVars {
		if val, exists := os.LookupEnv(envVar); exists {
			// For some variables, just existing is suspicious
			if envVar == "DEBUG" || envVar == "_DEBUG" || envVar == "DEBUGGER" {
				return true
			}

			// For others, check the content
			if envVar == "LD_PRELOAD" || envVar == "LD_AUDIT" {
				if val != "" {
					return true
				}
			}

			// Check terminal programs for debugging tools
			if envVar == "TERM_PROGRAM" {
				debugTerms := []string{"gdb", "lldb", "debug", "ida", "ghidra"}
				for _, term := range debugTerms {
					if strings.Contains(strings.ToLower(val), term) {
						return true
					}
				}
			}
		}
	}

	return false
}

// runWindowsChecks runs Windows-specific anti-debugging checks
func (a *AntiDebugger) runWindowsChecks() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	// In a real implementation, this would include:
	// - IsDebuggerPresent API call
	// - CheckRemoteDebuggerPresent API call
	// - NtQueryInformationProcess with ProcessDebugPort
	// - Hardware breakpoint detection
	// - Timing checks specific to Windows debugging

	// Placeholder for Windows-specific checks
	return false
}

// runLinuxChecks runs Linux-specific anti-debugging checks
func (a *AntiDebugger) runLinuxChecks() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	detected := false

	// Check for tracers using ptrace
	if a.options.EnableTraceChecks {
		// In a real implementation, this would use ptrace(PTRACE_TRACEME, 0, 0, 0)
		// If it returns an error, we're being traced
		// For now, we'll check /proc/self/status for TracerPid
		
		data, err := os.ReadFile("/proc/self/status")
		if err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "TracerPid:") {
					fields := strings.Fields(line)
					if len(fields) >= 2 && fields[1] != "0" {
						a.recordDetection("tracer")
						detected = true
					}
				}
			}
		}
	}

	// Check parent process
	if a.options.EnableParentChecks {
		// Check if parent process is a known debugger
		ppid := os.Getppid()
		procPath := fmt.Sprintf("/proc/%d/comm", ppid)
		data, err := os.ReadFile(procPath)
		if err == nil {
			parentName := strings.TrimSpace(string(data))
			debuggers := []string{"gdb", "lldb", "strace", "ltrace", "valgrind", "ida", "ghidra"}
			for _, debugger := range debuggers {
				if strings.Contains(strings.ToLower(parentName), debugger) {
					a.recordDetection("parent")
					detected = true
				}
			}
		}
	}

	return detected
}

// runDarwinChecks runs macOS-specific anti-debugging checks
func (a *AntiDebugger) runDarwinChecks() bool {
	if runtime.GOOS != "darwin" {
		return false
	}

	// In a real implementation, this would include:
	// - sysctl call to check if process is being traced
	// - P_TRACED flag check
	// - Exception port checks
	// - Hardware breakpoint detection

	// Placeholder for macOS-specific checks
	return false
}

// selfDestruct performs actions to protect the binary when debugging is detected
func (a *AntiDebugger) selfDestruct() {
	// In a real implementation, this might:
	// - Corrupt memory
	// - Delete files
	// - Encrypt sensitive data
	// - Exit the process

	// For safety, we'll just exit
	fmt.Println("Debugger detected, exiting...")
	os.Exit(1)
}

// ObfuscatedCheck performs anti-debugging checks with obfuscation
func (a *AntiDebugger) ObfuscatedCheck() bool {
	// This function would be heavily obfuscated in a real implementation
	// to make static analysis more difficult
	
	// Use runtime.GC() and debug.FreeOSMemory() to make timing analysis harder
	runtime.GC()
	debug.FreeOSMemory()
	
	// Use indirect function calls and pointer manipulation to obfuscate logic
	return a.RunChecks()
}

// StartMonitoring starts continuous monitoring for debuggers
func (a *AntiDebugger) StartMonitoring(interval time.Duration, callback func(bool)) {
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
