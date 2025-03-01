package security

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// SignatureOptions configures signature behavior
type SignatureOptions struct {
	EnableSignatureVerification bool
	RequireSignature           bool
	TrustedSigners             []*x509.Certificate
	SigningKey                 *ecdsa.PrivateKey
}

// DefaultSignatureOptions returns default signature options
func DefaultSignatureOptions() SignatureOptions {
	return SignatureOptions{
		EnableSignatureVerification: true,
		RequireSignature:           true,
		TrustedSigners:             []*x509.Certificate{},
	}
}

// SignatureVerifier implements signature verification
type SignatureVerifier struct {
	options     SignatureOptions
	mutex       sync.RWMutex
}

// NewSignatureVerifier creates a new signature verifier with the specified options
func NewSignatureVerifier(options SignatureOptions) *SignatureVerifier {
	return &SignatureVerifier{
		options: options,
	}
}

// Sign signs data with the signing key
func (s *SignatureVerifier) Sign(data []byte) ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Check if signing key is available
	if s.options.SigningKey == nil {
		return nil, fmt.Errorf("signing key not available")
	}

	// Calculate hash
	hash := sha256.Sum256(data)

	// Sign hash
	signature, err := s.options.SigningKey.Sign(rand.Reader, hash[:], crypto.SHA256)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	return signature, nil
}

// Verify verifies a signature
func (s *SignatureVerifier) Verify(data, signature []byte, cert *x509.Certificate) (bool, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Check if signature verification is enabled
	if !s.options.EnableSignatureVerification {
		return true, nil
	}

	// Check if certificate is trusted
	if !s.isTrustedCertificate(cert) {
		return false, fmt.Errorf("certificate not trusted")
	}

	// Get public key
	pubKey, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("certificate public key is not ECDSA")
	}

	// Calculate hash
	hash := sha256.Sum256(data)

	// Verify signature
	valid := ecdsa.VerifyASN1(pubKey, hash[:], signature)
	if !valid {
		return false, fmt.Errorf("signature verification failed")
	}

	return true, nil
}

// isTrustedCertificate checks if a certificate is trusted
func (s *SignatureVerifier) isTrustedCertificate(cert *x509.Certificate) bool {
	// If no trusted signers are specified, trust all certificates
	if len(s.options.TrustedSigners) == 0 {
		return true
	}

	// Check if certificate is in trusted signers
	for _, trustedCert := range s.options.TrustedSigners {
		if cert.Equal(trustedCert) {
			return true
		}
	}

	return false
}

// AddTrustedSigner adds a trusted signer certificate
func (s *SignatureVerifier) AddTrustedSigner(cert *x509.Certificate) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.options.TrustedSigners = append(s.options.TrustedSigners, cert)
}

// RemoveTrustedSigner removes a trusted signer certificate
func (s *SignatureVerifier) RemoveTrustedSigner(cert *x509.Certificate) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var newTrustedSigners []*x509.Certificate
	for _, trustedCert := range s.options.TrustedSigners {
		if !cert.Equal(trustedCert) {
			newTrustedSigners = append(newTrustedSigners, trustedCert)
		}
	}

	s.options.TrustedSigners = newTrustedSigners
}

// LoadSigningKeyFromFile loads a signing key from a file
func (s *SignatureVerifier) LoadSigningKeyFromFile(keyFile string) error {
	// Load private key
	keyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	// Parse private key
	block, _ := pem.Decode(keyPEM)
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return fmt.Errorf("failed to decode private key PEM")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	// Store signing key
	s.mutex.Lock()
	s.options.SigningKey = key
	s.mutex.Unlock()

	return nil
}

// LoadTrustedSignersFromFile loads trusted signer certificates from a file
func (s *SignatureVerifier) LoadTrustedSignersFromFile(certFile string) error {
	// Load certificates
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Parse certificates
	var certs []*x509.Certificate
	for {
		block, rest := pem.Decode(certPEM)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			certPEM = rest
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse certificate: %w", err)
		}

		certs = append(certs, cert)
		certPEM = rest
	}

	// Store trusted signers
	s.mutex.Lock()
	s.options.TrustedSigners = append(s.options.TrustedSigners, certs...)
	s.mutex.Unlock()

	return nil
}

// SignModule signs a module
func (s *SignatureVerifier) SignModule(moduleData []byte) ([]byte, error) {
	return s.Sign(moduleData)
}

// VerifyModule verifies a module signature
func (s *SignatureVerifier) VerifyModule(moduleData, signature []byte, cert *x509.Certificate) (bool, error) {
	return s.Verify(moduleData, signature, cert)
}

// GenerateSigningKey generates a new signing key
func (s *SignatureVerifier) GenerateSigningKey() error {
	// Generate private key
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate signing key: %w", err)
	}

	// Store signing key
	s.mutex.Lock()
	s.options.SigningKey = key
	s.mutex.Unlock()

	return nil
}

// SaveSigningKeyToFile saves the signing key to a file
func (s *SignatureVerifier) SaveSigningKeyToFile(keyFile string) error {
	s.mutex.RLock()
	key := s.options.SigningKey
	s.mutex.RUnlock()

	if key == nil {
		return fmt.Errorf("signing key not available")
	}

	// Create directory if it doesn't exist
	keyDir := filepath.Dir(keyFile)
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Marshal private key
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	// Create key file
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyOut.Close()

	// Write private key PEM
	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	return nil
}
