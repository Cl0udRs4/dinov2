package server

import (
	"crypto/tls"
	"dinoc2/pkg/security"
	"fmt"
	"time"
)

// SecurityOptions configures server security behavior
type SecurityOptions struct {
	EnableAntiDebug      bool
	EnableMemoryProtection bool
	EnableAuthentication bool
	EnableSignatureVerification bool
	EnableIntegrityChecking bool
	EnableTrafficObfuscation bool
	SecurityViolationCallback func(violation string)
}

// DefaultSecurityOptions returns default security options
func DefaultSecurityOptions() SecurityOptions {
	return SecurityOptions{
		EnableAntiDebug:      true,
		EnableMemoryProtection: true,
		EnableAuthentication: true,
		EnableSignatureVerification: true,
		EnableIntegrityChecking: true,
		EnableTrafficObfuscation: true,
		SecurityViolationCallback: nil,
	}
}

// Server represents the C2 server (forward declaration)
type Server struct{}

// SecurityIntegration integrates security features with the server
type SecurityIntegration struct {
	options         SecurityOptions
	securityManager *security.SecurityManager
	server          *Server
}

// NewSecurityIntegration creates a new security integration
func NewSecurityIntegration(server *Server, options SecurityOptions) (*SecurityIntegration, error) {
	// Create security manager options
	managerOptions := security.SecurityManagerOptions{
		EnableAntiDebug:      options.EnableAntiDebug,
		EnableAntiSandbox:    false, // Server doesn't need anti-sandbox
		EnableMemoryProtection: options.EnableMemoryProtection,
		EnableAuthentication: options.EnableAuthentication,
		EnableSignatureVerification: options.EnableSignatureVerification,
		EnableIntegrityChecking: options.EnableIntegrityChecking,
		EnableTrafficObfuscation: options.EnableTrafficObfuscation,
		SecurityViolationCallback: options.SecurityViolationCallback,
	}

	// Create security manager
	manager := security.NewSecurityManager(managerOptions)
	err := manager.Initialize()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize security manager: %w", err)
	}

	return &SecurityIntegration{
		options:         options,
		securityManager: manager,
		server:          server,
	}, nil
}

// Initialize initializes the security integration
func (s *SecurityIntegration) Initialize() error {
	// Run security checks
	if !s.securityManager.RunSecurityChecks() {
		violations := s.securityManager.GetViolations()
		return fmt.Errorf("security checks failed: %v", violations)
	}

	// Start security monitoring
	s.securityManager.StartSecurityMonitoring(30 * time.Second)

	return nil
}

// ProtectData protects sensitive data in memory
func (s *SecurityIntegration) ProtectData(id string, data []byte) error {
	return s.securityManager.ProtectData(id, data)
}

// AccessProtectedData accesses protected data in memory
func (s *SecurityIntegration) AccessProtectedData(id string) ([]byte, error) {
	return s.securityManager.AccessProtectedData(id)
}

// ObfuscateTraffic obfuscates outgoing traffic
func (s *SecurityIntegration) ObfuscateTraffic(data []byte) ([]byte, error) {
	return s.securityManager.ObfuscateTraffic(data)
}

// DeobfuscateTraffic deobfuscates incoming traffic
func (s *SecurityIntegration) DeobfuscateTraffic(data []byte) ([]byte, error) {
	return s.securityManager.DeobfuscateTraffic(data)
}

// SignModule signs a module
func (s *SecurityIntegration) SignModule(moduleData []byte) ([]byte, error) {
	return s.securityManager.SignModule(moduleData)
}

// GetTLSConfig returns a TLS configuration for the server
func (s *SecurityIntegration) GetTLSConfig() (*tls.Config, error) {
	return s.securityManager.GetTLSConfig(true)
}

// Shutdown shuts down the security integration
func (s *SecurityIntegration) Shutdown() {
	s.securityManager.Shutdown()
}
