package crypto

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

// Chacha20Encryptor implements the Encryptor interface using ChaCha20-Poly1305
type Chacha20Encryptor struct {
	key []byte
}

// NewChacha20Encryptor creates a new ChaCha20 encryptor with a random key
func NewChacha20Encryptor() (*Chacha20Encryptor, error) {
	// Generate a random 32-byte key for ChaCha20
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	
	return &Chacha20Encryptor{key: key}, nil
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
	// TODO: Implement ECDHE key exchange
	return nil, errors.New("ECDHE key exchange not yet implemented")
}

// RotateKey implements the Encryptor interface
func (e *Chacha20Encryptor) RotateKey() error {
	// Generate a new random key
	newKey := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(rand.Reader, newKey); err != nil {
		return err
	}
	
	e.key = newKey
	return nil
}
