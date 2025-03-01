package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuthenticationOptions configures authentication behavior
type AuthenticationOptions struct {
	EnableCertAuth      bool
	EnablePSKAuth       bool
	RequireMutualAuth   bool
	CertValidityDays    int
	KeySize             int
	CertificateAuthority *x509.Certificate
	CAPrivateKey        *ecdsa.PrivateKey
	TrustedCerts        []*x509.Certificate
	PreSharedKeys       map[string]string
}

// DefaultAuthenticationOptions returns default authentication options
func DefaultAuthenticationOptions() AuthenticationOptions {
	return AuthenticationOptions{
		EnableCertAuth:    true,
		EnablePSKAuth:     true,
		RequireMutualAuth: true,
		CertValidityDays:  365,
		KeySize:           256,
		PreSharedKeys:     make(map[string]string),
	}
}

// Authenticator implements authentication mechanisms
type Authenticator struct {
	options     AuthenticationOptions
	mutex       sync.RWMutex
	tlsConfig   *tls.Config
	certificates map[string]*tls.Certificate
}

// NewAuthenticator creates a new authenticator with the specified options
func NewAuthenticator(options AuthenticationOptions) *Authenticator {
	return &Authenticator{
		options:     options,
		certificates: make(map[string]*tls.Certificate),
	}
}

// InitCA initializes a certificate authority
func (a *Authenticator) InitCA() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Generate CA private key
	caPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate CA private key: %w", err)
	}

	// Prepare CA certificate template
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{"DinoC2 CA"},
			CommonName:   "DinoC2 Root CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years validity
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	// Create CA certificate
	caBytes, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// Parse the CA certificate
	caCert, err := x509.ParseCertificate(caBytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Store CA certificate and private key
	a.options.CertificateAuthority = caCert
	a.options.CAPrivateKey = caPrivKey

	return nil
}

// GenerateCertificate generates a certificate signed by the CA
func (a *Authenticator) GenerateCertificate(name, commonName string, isServer bool, ipAddresses []net.IP, dnsNames []string) (*tls.Certificate, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Check if CA is initialized
	if a.options.CertificateAuthority == nil || a.options.CAPrivateKey == nil {
		return nil, fmt.Errorf("CA not initialized")
	}

	// Generate private key
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Prepare certificate template
	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{"DinoC2"},
			CommonName:   commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, a.options.CertValidityDays),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	// Set appropriate key usage based on certificate type
	if isServer {
		certTemplate.ExtKeyUsage = append(certTemplate.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
	} else {
		certTemplate.ExtKeyUsage = append(certTemplate.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	}

	// Add IP addresses and DNS names
	certTemplate.IPAddresses = ipAddresses
	certTemplate.DNSNames = dnsNames

	// Create certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, a.options.CertificateAuthority, &privKey.PublicKey, a.options.CAPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Create tls.Certificate
	cert := &tls.Certificate{
		Certificate: [][]byte{certBytes, a.options.CertificateAuthority.Raw},
		PrivateKey:  privKey,
		Leaf:        certTemplate,
	}

	// Store certificate
	a.certificates[name] = cert

	return cert, nil
}

// SaveCertificateToFile saves a certificate to a file
func (a *Authenticator) SaveCertificateToFile(name, certFile, keyFile string) error {
	a.mutex.RLock()
	cert, exists := a.certificates[name]
	a.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("certificate not found: %s", name)
	}

	// Create directory if it doesn't exist
	certDir := filepath.Dir(certFile)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	keyDir := filepath.Dir(keyFile)
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Save certificate
	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certOut.Close()

	// Write certificate PEM
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Save private key
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyOut.Close()

	// Get private key bytes
	privKey, ok := cert.PrivateKey.(*ecdsa.PrivateKey)
	if !ok {
		return fmt.Errorf("private key is not ECDSA")
	}

	privKeyBytes, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	// Write private key PEM
	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privKeyBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	return nil
}

// LoadCertificateFromFile loads a certificate from a file
func (a *Authenticator) LoadCertificateFromFile(name, certFile, keyFile string) error {
	// Load certificate
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("failed to load certificate: %w", err)
	}

	// Parse certificate
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Set leaf certificate
	cert.Leaf = leaf

	// Store certificate
	a.mutex.Lock()
	a.certificates[name] = &cert
	a.mutex.Unlock()

	return nil
}

// GetTLSConfig returns a TLS configuration for the server or client
func (a *Authenticator) GetTLSConfig(isServer bool) *tls.Config {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	// Create certificate pool for CA
	certPool := x509.NewCertPool()
	if a.options.CertificateAuthority != nil {
		certPool.AddCert(a.options.CertificateAuthority)
	}

	// Add trusted certificates
	for _, cert := range a.options.TrustedCerts {
		certPool.AddCert(cert)
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		RootCAs:      certPool,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		},
	}

	// Set client or server specific options
	if isServer {
		// Server configuration
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		if !a.options.RequireMutualAuth {
			tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven
		}
	} else {
		// Client configuration
		tlsConfig.InsecureSkipVerify = false
	}

	// Add certificates
	var certs []tls.Certificate
	for _, cert := range a.certificates {
		certs = append(certs, *cert)
	}
	
	if len(certs) > 0 {
		tlsConfig.Certificates = certs
	}

	return tlsConfig
}

// AddPreSharedKey adds a pre-shared key
func (a *Authenticator) AddPreSharedKey(id, key string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.options.PreSharedKeys[id] = key
}

// VerifyPreSharedKey verifies a pre-shared key
func (a *Authenticator) VerifyPreSharedKey(id, key string) bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	storedKey, exists := a.options.PreSharedKeys[id]
	if !exists {
		return false
	}

	return storedKey == key
}

// SaveCAToFile saves the CA certificate to a file
func (a *Authenticator) SaveCAToFile(certFile, keyFile string) error {
	a.mutex.RLock()
	ca := a.options.CertificateAuthority
	caKey := a.options.CAPrivateKey
	a.mutex.RUnlock()

	if ca == nil || caKey == nil {
		return fmt.Errorf("CA not initialized")
	}

	// Create directory if it doesn't exist
	certDir := filepath.Dir(certFile)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	keyDir := filepath.Dir(keyFile)
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Save CA certificate
	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate file: %w", err)
	}
	defer certOut.Close()

	// Write CA certificate PEM
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: ca.Raw}); err != nil {
		return fmt.Errorf("failed to write CA certificate: %w", err)
	}

	// Save CA private key
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return fmt.Errorf("failed to create CA key file: %w", err)
	}
	defer keyOut.Close()

	// Get CA private key bytes
	caKeyBytes, err := x509.MarshalECPrivateKey(caKey)
	if err != nil {
		return fmt.Errorf("failed to marshal CA private key: %w", err)
	}

	// Write CA private key PEM
	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: caKeyBytes}); err != nil {
		return fmt.Errorf("failed to write CA private key: %w", err)
	}

	return nil
}

// LoadCAFromFile loads the CA certificate from a file
func (a *Authenticator) LoadCAFromFile(certFile, keyFile string) error {
	// Load CA certificate
	caCertPEM, err := os.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate file: %w", err)
	}

	// Parse CA certificate
	block, _ := pem.Decode(caCertPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return fmt.Errorf("failed to decode CA certificate PEM")
	}

	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Load CA private key
	caKeyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("failed to read CA key file: %w", err)
	}

	// Parse CA private key
	block, _ = pem.Decode(caKeyPEM)
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return fmt.Errorf("failed to decode CA private key PEM")
	}

	caKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA private key: %w", err)
	}

	// Store CA certificate and private key
	a.mutex.Lock()
	a.options.CertificateAuthority = caCert
	a.options.CAPrivateKey = caKey
	a.mutex.Unlock()

	return nil
}
