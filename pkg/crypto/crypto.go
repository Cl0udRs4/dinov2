package crypto

import (
	"errors"
	"time"
)

// Algorithm represents supported encryption algorithms
type Algorithm string

const (
	AlgorithmAES      Algorithm = "aes"
	AlgorithmChacha20 Algorithm = "chacha20"
)

// Encryptor interface defines methods that all encryption implementations must support
type Encryptor interface {
	// Encrypt encrypts plaintext data
	Encrypt(plain []byte) ([]byte, error)
	
	// Decrypt decrypts ciphertext data
	Decrypt(cipher []byte) ([]byte, error)
	
	// Algorithm returns the encryption algorithm identifier
	Algorithm() Algorithm
	
	// ExchangeKey performs key exchange using the provided public key
	ExchangeKey(publicKey []byte) ([]byte, error)
	
	// RotateKey rotates the encryption key
	RotateKey() error
	
	// GetKeyFingerprint returns a fingerprint of the current key
	GetKeyFingerprint() []byte
	
	// GetLastRotation returns the time of the last key rotation
	GetLastRotation() time.Time
}

// Factory creates encryptors based on algorithm type
func Factory(algorithm Algorithm) (Encryptor, error) {
	switch algorithm {
	case AlgorithmAES:
		return NewAESEncryptor()
	case AlgorithmChacha20:
		return NewChacha20Encryptor()
	default:
		return nil, errors.New("unsupported encryption algorithm")
	}
}
