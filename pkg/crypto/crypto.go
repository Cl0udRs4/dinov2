package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"time"
)

// Algorithm represents an encryption algorithm
type Algorithm string

// Encryption algorithms
const (
	AlgorithmAES      Algorithm = "aes"
	AlgorithmChacha20 Algorithm = "chacha20"
)

// SessionID represents a unique session identifier
type SessionID string

// Encryptor interface defines methods that all encryption implementations must support
type Encryptor interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
	Algorithm() Algorithm
}

// NewEncryptor creates a new encryptor based on the specified algorithm
func NewEncryptor(algorithm Algorithm) (Encryptor, error) {
	switch algorithm {
	case AlgorithmAES:
		return NewAESEncryptor()
	case AlgorithmChacha20:
		return NewChacha20Encryptor()
	default:
		return nil, errors.New("unsupported encryption algorithm")
	}
}

// GenerateSessionID generates a new random session ID
func GenerateSessionID() SessionID {
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		// If random fails, use a timestamp-based ID
		return SessionID(fmt.Sprintf("client-%d", time.Now().UnixNano()))
	}
	return SessionID(fmt.Sprintf("client-%x", b))
}
