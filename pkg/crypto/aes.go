package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// AESEncryptor implements the Encryptor interface using AES-GCM
type AESEncryptor struct {
	key []byte
}

// NewAESEncryptor creates a new AES encryptor with a random key
func NewAESEncryptor() (*AESEncryptor, error) {
	// Generate a random 32-byte key for AES-256
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	
	return &AESEncryptor{key: key}, nil
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
	// TODO: Implement ECDHE key exchange
	return nil, errors.New("ECDHE key exchange not yet implemented")
}

// RotateKey implements the Encryptor interface
func (e *AESEncryptor) RotateKey() error {
	// Generate a new random key
	newKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, newKey); err != nil {
		return err
	}
	
	e.key = newKey
	return nil
}
