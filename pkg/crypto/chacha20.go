package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
)

// Chacha20Encryptor implements the Encryptor interface using ChaCha20-Poly1305
type Chacha20Encryptor struct {
	key         []byte
	keyExchange *ECDHEKeyExchange
}

// NewChacha20Encryptor creates a new ChaCha20 encryptor with a random key
func NewChacha20Encryptor() (*Chacha20Encryptor, error) {
	// Generate a random 32-byte key for ChaCha20
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	
	// Create ECDHE key exchange
	keyExchange, err := NewECDHEKeyExchange()
	if err != nil {
		return nil, err
	}
	
	return &Chacha20Encryptor{
		key:         key,
		keyExchange: keyExchange,
	}, nil
}

// Encrypt implements the Encryptor interface
func (e *Chacha20Encryptor) Encrypt(plain []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(e.key)
	if err != nil {
		return nil, err
	}
	
	// Create a nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	
	// Encrypt and seal the data
	ciphertext := aead.Seal(nonce, nonce, plain, nil)
	return ciphertext, nil
}

// Decrypt implements the Encryptor interface
func (e *Chacha20Encryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(e.key)
	if err != nil {
		return nil, err
	}
	
	// Verify the ciphertext is long enough
	if len(ciphertext) < aead.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}
	
	// Extract the nonce and ciphertext
	nonce, ciphertext := ciphertext[:aead.NonceSize()], ciphertext[aead.NonceSize():]
	
	// Decrypt the data
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	
	return plaintext, nil
}

// Algorithm implements the Encryptor interface
func (e *Chacha20Encryptor) Algorithm() Algorithm {
	return AlgorithmChacha20
}

// ExchangeKey implements the Encryptor interface
func (e *Chacha20Encryptor) ExchangeKey(publicKey []byte) ([]byte, error) {
	// Get our public key to send to the peer
	ourPublicKey, err := e.keyExchange.GetPublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}
	
	// Derive shared secret using peer's public key
	sharedSecret, err := e.keyExchange.DeriveSharedSecret(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive shared secret: %w", err)
	}
	
	// Use the shared secret as our new key
	e.key = sharedSecret
	
	// Return our public key so the peer can derive the same shared secret
	return ourPublicKey, nil
}

// RotateKey implements the Encryptor interface
func (e *Chacha20Encryptor) RotateKey() error {
	// Regenerate the ECDHE key pair
	if err := e.keyExchange.RegenerateKeyPair(); err != nil {
		return fmt.Errorf("failed to regenerate key pair: %w", err)
	}
	
	// Generate a new random key
	newKey := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(rand.Reader, newKey); err != nil {
		return err
	}
	
	e.key = newKey
	return nil
}

// GetKeyFingerprint returns a fingerprint of the current key
func (e *Chacha20Encryptor) GetKeyFingerprint() []byte {
	// Create a simple fingerprint by hashing the key
	hash := sha256.Sum256(e.key)
	return hash[:8] // Return first 8 bytes as fingerprint
}

// GetLastRotation returns the time of the last key rotation
func (e *Chacha20Encryptor) GetLastRotation() time.Time {
	// This would normally track the last rotation time
	// For now, just return the current time
	return time.Now()
}
