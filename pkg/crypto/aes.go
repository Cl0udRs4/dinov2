package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"time"
)

// AESEncryptor implements the Encryptor interface using AES-GCM
type AESEncryptor struct {
	key       []byte
	keyExchange *ECDHEKeyExchange
}

// NewAESEncryptor creates a new AES encryptor with a random key
func NewAESEncryptor() (*AESEncryptor, error) {
	// Generate a random 32-byte key for AES-256
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	
	// Create ECDHE key exchange
	keyExchange, err := NewECDHEKeyExchange()
	if err != nil {
		return nil, err
	}
	
	return &AESEncryptor{
		key:       key,
		keyExchange: keyExchange,
	}, nil
}

// Encrypt implements the Encryptor interface
func (e *AESEncryptor) Encrypt(plain []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}
	
	// Create a GCM cipher mode
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	// Create a nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	
	// Encrypt and seal the data
	ciphertext := aesGCM.Seal(nonce, nonce, plain, nil)
	return ciphertext, nil
}

// Decrypt implements the Encryptor interface
func (e *AESEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}
	
	// Create a GCM cipher mode
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	// Verify the ciphertext is long enough
	if len(ciphertext) < aesGCM.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}
	
	// Extract the nonce and ciphertext
	nonce, ciphertext := ciphertext[:aesGCM.NonceSize()], ciphertext[aesGCM.NonceSize():]
	
	// Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	
	return plaintext, nil
}

// Algorithm implements the Encryptor interface
func (e *AESEncryptor) Algorithm() Algorithm {
	return AlgorithmAES
}

// ExchangeKey implements the Encryptor interface
func (e *AESEncryptor) ExchangeKey(publicKey []byte) ([]byte, error) {
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
func (e *AESEncryptor) RotateKey() error {
	// Regenerate the ECDHE key pair
	if err := e.keyExchange.RegenerateKeyPair(); err != nil {
		return fmt.Errorf("failed to regenerate key pair: %w", err)
	}
	
	// Generate a new random key
	newKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, newKey); err != nil {
		return err
	}
	
	e.key = newKey
	return nil
}

// GetKeyFingerprint returns a fingerprint of the current key
func (e *AESEncryptor) GetKeyFingerprint() []byte {
	// Create a simple fingerprint by hashing the key
	hash := sha256.Sum256(e.key)
	return hash[:8] // Return first 8 bytes as fingerprint
}

// GetLastRotation returns the time of the last key rotation
func (e *AESEncryptor) GetLastRotation() time.Time {
	// This would normally track the last rotation time
	// For now, just return the current time
	return time.Now()
}
