package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
)

// ECDHEKeyExchange implements Elliptic Curve Diffie-Hellman Ephemeral key exchange
type ECDHEKeyExchange struct {
	privateKey *ecdsa.PrivateKey
	curve      elliptic.Curve
}

// NewECDHEKeyExchange creates a new ECDHE key exchange with P-256 curve
func NewECDHEKeyExchange() (*ECDHEKeyExchange, error) {
	// Use P-256 curve for ECDHE
	curve := elliptic.P256()
	
	// Generate a new private key
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ECDHE private key: %w", err)
	}
	
	return &ECDHEKeyExchange{
		privateKey: privateKey,
		curve:      curve,
	}, nil
}

// GetPublicKey returns the public key in PEM format
func (e *ECDHEKeyExchange) GetPublicKey() ([]byte, error) {
	// Marshal the public key to DER format
	derBytes, err := x509.MarshalPKIXPublicKey(&e.privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	
	// Encode to PEM format
	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derBytes,
	}
	
	return pem.EncodeToMemory(pemBlock), nil
}

// DeriveSharedSecret derives a shared secret from a peer's public key
func (e *ECDHEKeyExchange) DeriveSharedSecret(peerPublicKeyPEM []byte) ([]byte, error) {
	// Decode PEM format
	block, _ := pem.Decode(peerPublicKeyPEM)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}
	
	// Parse the public key
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	
	// Ensure it's an ECDSA public key
	peerPublicKey, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("not an ECDSA public key")
	}
	
	// Compute the shared secret
	x, _ := e.curve.ScalarMult(peerPublicKey.X, peerPublicKey.Y, e.privateKey.D.Bytes())
	if x == nil {
		return nil, errors.New("failed to compute shared secret")
	}
	
	// Convert the shared point to bytes
	sharedSecret := x.Bytes()
	
	// Hash the shared secret to derive a symmetric key
	hash := sha256.Sum256(sharedSecret)
	return hash[:], nil
}

// RegenerateKeyPair regenerates the ECDHE key pair
func (e *ECDHEKeyExchange) RegenerateKeyPair() error {
	// Generate a new private key
	privateKey, err := ecdsa.GenerateKey(e.curve, rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to regenerate ECDHE private key: %w", err)
	}
	
	e.privateKey = privateKey
	return nil
}

// GenerateRandomBytes generates random bytes of the specified length
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}
