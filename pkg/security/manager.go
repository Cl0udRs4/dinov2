package security

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"
	"time"
)

// SecurityManagerOptions configures the security manager
type SecurityManagerOptions struct {
	EnableAntiDebug      bool
	EnableAntiSandbox    bool
	EnableMemoryProtection bool
	EnableAuthentication bool
	EnableSignatureVerification bool
	EnableIntegrityChecking bool
	EnableTrafficObfuscation bool
	SecurityViolationCallback func(violation string)
}

// DefaultSecurityManagerOptions returns default security manager options
func DefaultSecurityManagerOptions() SecurityManagerOptions {
	return SecurityManagerOptions{
		EnableAntiDebug:      true,
		EnableAntiSandbox:    true,
		EnableMemoryProtection: true,
		EnableAuthentication: true,
		EnableSignatureVerification: true,
		EnableIntegrityChecking: true,
		EnableTrafficObfuscation: true,
		SecurityViolationCallback: nil,
	}
}

// SecurityManager manages security components
type SecurityManager struct {
	options         SecurityManagerOptions
	mutex           sync.RWMutex
	antiDebugger    *AntiDebugger
	antiSandbox     *AntiSandbox
	memoryProtection *MemoryProtection
	authenticator   *Authenticator
	signatureVerifier *SignatureVerifier
	integrityChecker *IntegrityChecker
	trafficObfuscator *TrafficObfuscator
	violations      []string
}

// NewSecurityManager creates a new security manager with the specified options
func NewSecurityManager(options SecurityManagerOptions) *SecurityManager {
	return &SecurityManager{
		options:    options,
		violations: []string{},
	}
}

// Initialize initializes the security manager
func (s *SecurityManager) Initialize() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Initialize anti-debugger
	if s.options.EnableAntiDebug {
		s.antiDebugger = NewAntiDebugger(DefaultAntiDebugOptions())
	}

	// Initialize anti-sandbox
	if s.options.EnableAntiSandbox {
		s.antiSandbox = NewAntiSandbox(DefaultAntiSandboxOptions())
	}

	// Initialize memory protection
	if s.options.EnableMemoryProtection {
		s.memoryProtection = NewMemoryProtection(DefaultMemoryProtectionOptions())
	}

	// Initialize authenticator
	if s.options.EnableAuthentication {
		s.authenticator = NewAuthenticator(DefaultAuthenticationOptions())
		err := s.authenticator.InitCA()
		if err != nil {
			return fmt.Errorf("failed to initialize CA: %w", err)
		}
	}

	// Initialize signature verifier
	if s.options.EnableSignatureVerification {
		s.signatureVerifier = NewSignatureVerifier(DefaultSignatureOptions())
	}

	// Initialize integrity checker
	if s.options.EnableIntegrityChecking {
		s.integrityChecker = NewIntegrityChecker(DefaultIntegrityOptions())
		s.integrityChecker.Start()
	}

	// Initialize traffic obfuscator
	if s.options.EnableTrafficObfuscation {
		s.trafficObfuscator = NewTrafficObfuscator()
		s.trafficObfuscator.RegisterProfile(CreateHTTPProfile())
		s.trafficObfuscator.RegisterProfile(CreateDNSProfile())
		s.trafficObfuscator.RegisterProfile(CreateTLSProfile())
	}

	return nil
}

// RunSecurityChecks runs all security checks
func (s *SecurityManager) RunSecurityChecks() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Clear previous violations
	s.violations = []string{}

	// Run anti-debugging checks
	if s.options.EnableAntiDebug && s.antiDebugger != nil {
		if s.antiDebugger.RunChecks() {
			s.violations = append(s.violations, "debugger detected")
		}
	}

	// Run anti-sandbox checks
	if s.options.EnableAntiSandbox && s.antiSandbox != nil {
		if s.antiSandbox.RunChecks() {
			s.violations = append(s.violations, "sandbox detected")
		}
	}

	// Run integrity checks
	if s.options.EnableIntegrityChecking && s.integrityChecker != nil {
		if !s.integrityChecker.CheckIntegrity() {
			s.violations = append(s.violations, "integrity check failed")
		}
	}

	// Call violation callback if there are violations
	if len(s.violations) > 0 && s.options.SecurityViolationCallback != nil {
		for _, violation := range s.violations {
			s.options.SecurityViolationCallback(violation)
		}
	}

	return len(s.violations) == 0
}

// GetViolations returns a list of security violations
func (s *SecurityManager) GetViolations() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Create a copy of the violations slice
	violations := make([]string, len(s.violations))
	copy(violations, s.violations)

	return violations
}

// GetAntiDebugger returns the anti-debugger
func (s *SecurityManager) GetAntiDebugger() *AntiDebugger {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.antiDebugger
}

// GetAntiSandbox returns the anti-sandbox
func (s *SecurityManager) GetAntiSandbox() *AntiSandbox {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.antiSandbox
}

// GetMemoryProtection returns the memory protection
func (s *SecurityManager) GetMemoryProtection() *MemoryProtection {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.memoryProtection
}

// GetAuthenticator returns the authenticator
func (s *SecurityManager) GetAuthenticator() *Authenticator {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.authenticator
}

// GetSignatureVerifier returns the signature verifier
func (s *SecurityManager) GetSignatureVerifier() *SignatureVerifier {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.signatureVerifier
}

// GetIntegrityChecker returns the integrity checker
func (s *SecurityManager) GetIntegrityChecker() *IntegrityChecker {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.integrityChecker
}

// GetTrafficObfuscator returns the traffic obfuscator
func (s *SecurityManager) GetTrafficObfuscator() *TrafficObfuscator {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.trafficObfuscator
}

// ProtectData protects sensitive data in memory
func (s *SecurityManager) ProtectData(id string, data []byte) error {
	if !s.options.EnableMemoryProtection || s.memoryProtection == nil {
		return fmt.Errorf("memory protection not enabled")
	}

	return s.memoryProtection.Protect(id, data)
}

// AccessProtectedData accesses protected data in memory
func (s *SecurityManager) AccessProtectedData(id string) ([]byte, error) {
	if !s.options.EnableMemoryProtection || s.memoryProtection == nil {
		return nil, fmt.Errorf("memory protection not enabled")
	}

	return s.memoryProtection.Access(id)
}

// RemoveProtectedData removes protected data from memory
func (s *SecurityManager) RemoveProtectedData(id string) error {
	if !s.options.EnableMemoryProtection || s.memoryProtection == nil {
		return fmt.Errorf("memory protection not enabled")
	}

	return s.memoryProtection.Remove(id)
}

// ObfuscateTraffic obfuscates outgoing traffic
func (s *SecurityManager) ObfuscateTraffic(data []byte) ([]byte, error) {
	if !s.options.EnableTrafficObfuscation || s.trafficObfuscator == nil {
		return data, nil
	}

	return s.trafficObfuscator.ObfuscateOutgoing(data)
}

// DeobfuscateTraffic deobfuscates incoming traffic
func (s *SecurityManager) DeobfuscateTraffic(data []byte) ([]byte, error) {
	if !s.options.EnableTrafficObfuscation || s.trafficObfuscator == nil {
		return data, nil
	}

	return s.trafficObfuscator.DeobfuscateIncoming(data)
}

// VerifyModuleSignature verifies a module signature
func (s *SecurityManager) VerifyModuleSignature(moduleData, signature []byte, cert *x509.Certificate) (bool, error) {
	if !s.options.EnableSignatureVerification || s.signatureVerifier == nil {
		return true, nil
	}

	return s.signatureVerifier.VerifyModule(moduleData, signature, cert)
}

// SignModule signs a module
func (s *SecurityManager) SignModule(moduleData []byte) ([]byte, error) {
	if !s.options.EnableSignatureVerification || s.signatureVerifier == nil {
		return nil, fmt.Errorf("signature verification not enabled")
	}

	return s.signatureVerifier.SignModule(moduleData)
}

// GetTLSConfig returns a TLS configuration for the server or client
func (s *SecurityManager) GetTLSConfig(isServer bool) (*tls.Config, error) {
	if !s.options.EnableAuthentication || s.authenticator == nil {
		return nil, fmt.Errorf("authentication not enabled")
	}

	return s.authenticator.GetTLSConfig(isServer), nil
}

// StartSecurityMonitoring starts continuous security monitoring
func (s *SecurityManager) StartSecurityMonitoring(interval time.Duration) {
	// Start anti-debugging monitoring
	if s.options.EnableAntiDebug && s.antiDebugger != nil {
		s.antiDebugger.StartMonitoring(interval, func(detected bool) {
			if detected && s.options.SecurityViolationCallback != nil {
				s.options.SecurityViolationCallback("debugger detected")
			}
		})
	}

	// Start anti-sandbox monitoring
	if s.options.EnableAntiSandbox && s.antiSandbox != nil {
		s.antiSandbox.StartMonitoring(interval, func(detected bool) {
			if detected && s.options.SecurityViolationCallback != nil {
				s.options.SecurityViolationCallback("sandbox detected")
			}
		})
	}

	// Start integrity monitoring
	if s.options.EnableIntegrityChecking && s.integrityChecker != nil {
		s.integrityChecker.Start()
	}
}

// Shutdown shuts down the security manager
func (s *SecurityManager) Shutdown() {
	// Stop memory protection
	if s.options.EnableMemoryProtection && s.memoryProtection != nil {
		s.memoryProtection.Stop()
	}

	// Stop integrity checker
	if s.options.EnableIntegrityChecking && s.integrityChecker != nil {
		s.integrityChecker.Stop()
	}
}
