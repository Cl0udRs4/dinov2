package crypto

import (
	"bytes"
	"testing"
	"time"
)

func TestAESEncryption(t *testing.T) {
	// Create a new AES encryptor
	encryptor, err := NewAESEncryptor()
	if err != nil {
		t.Fatalf("Failed to create AES encryptor: %v", err)
	}
	
	// Test encryption and decryption
	plaintext := []byte("This is a test message for AES encryption")
	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}
	
	// Ensure ciphertext is different from plaintext
	if bytes.Equal(plaintext, ciphertext) {
		t.Fatal("Ciphertext should be different from plaintext")
	}
	
	// Decrypt the ciphertext
	decrypted, err := encryptor.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}
	
	// Verify the decrypted text matches the original
	if !bytes.Equal(plaintext, decrypted) {
		t.Fatal("Decrypted text does not match original plaintext")
	}
}

func TestChacha20Encryption(t *testing.T) {
	// Create a new ChaCha20 encryptor
	encryptor, err := NewChacha20Encryptor()
	if err != nil {
		t.Fatalf("Failed to create ChaCha20 encryptor: %v", err)
	}
	
	// Test encryption and decryption
	plaintext := []byte("This is a test message for ChaCha20 encryption")
	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}
	
	// Ensure ciphertext is different from plaintext
	if bytes.Equal(plaintext, ciphertext) {
		t.Fatal("Ciphertext should be different from plaintext")
	}
	
	// Decrypt the ciphertext
	decrypted, err := encryptor.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}
	
	// Verify the decrypted text matches the original
	if !bytes.Equal(plaintext, decrypted) {
		t.Fatal("Decrypted text does not match original plaintext")
	}
}

func TestECDHEKeyExchange(t *testing.T) {
	// Create two ECDHE key exchanges
	exchange1, err := NewECDHEKeyExchange()
	if err != nil {
		t.Fatalf("Failed to create first ECDHE key exchange: %v", err)
	}
	
	exchange2, err := NewECDHEKeyExchange()
	if err != nil {
		t.Fatalf("Failed to create second ECDHE key exchange: %v", err)
	}
	
	// Get public keys
	pubKey1, err := exchange1.GetPublicKey()
	if err != nil {
		t.Fatalf("Failed to get first public key: %v", err)
	}
	
	pubKey2, err := exchange2.GetPublicKey()
	if err != nil {
		t.Fatalf("Failed to get second public key: %v", err)
	}
	
	// Derive shared secrets
	secret1, err := exchange1.DeriveSharedSecret(pubKey2)
	if err != nil {
		t.Fatalf("Failed to derive first shared secret: %v", err)
	}
	
	secret2, err := exchange2.DeriveSharedSecret(pubKey1)
	if err != nil {
		t.Fatalf("Failed to derive second shared secret: %v", err)
	}
	
	// Verify both sides derived the same secret
	if !bytes.Equal(secret1, secret2) {
		t.Fatal("Shared secrets do not match")
	}
}

func TestEncryptorKeyExchange(t *testing.T) {
	// Create two encryptors
	aes1, err := NewAESEncryptor()
	if err != nil {
		t.Fatalf("Failed to create first AES encryptor: %v", err)
	}
	
	aes2, err := NewAESEncryptor()
	if err != nil {
		t.Fatalf("Failed to create second AES encryptor: %v", err)
	}
	
	// Get public key from first encryptor
	pubKey1, err := aes1.keyExchange.GetPublicKey()
	if err != nil {
		t.Fatalf("Failed to get first public key: %v", err)
	}
	
	// Exchange keys
	pubKey2, err := aes2.ExchangeKey(pubKey1)
	if err != nil {
		t.Fatalf("Failed second key exchange: %v", err)
	}
	
	_, err = aes1.ExchangeKey(pubKey2)
	if err != nil {
		t.Fatalf("Failed first key exchange: %v", err)
	}
	
	// Test encryption with the exchanged keys
	plaintext := []byte("Testing encryption after key exchange")
	ciphertext, err := aes1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt with first encryptor: %v", err)
	}
	
	// Decrypt with the second encryptor
	decrypted, err := aes2.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Failed to decrypt with second encryptor: %v", err)
	}
	
	// Verify the decrypted text matches the original
	if !bytes.Equal(plaintext, decrypted) {
		t.Fatal("Decrypted text does not match original plaintext after key exchange")
	}
}

func TestSessionManager(t *testing.T) {
	// Create a session manager
	manager := NewSessionManager()
	defer manager.Shutdown()
	
	// Create a session
	session1, err := manager.CreateSession("client1", AlgorithmAES)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	
	// Verify session was created
	if session1 == nil {
		t.Fatal("Session should not be nil")
	}
	
	// Get the session
	session2, err := manager.GetSession("client1")
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}
	
	// Verify it's the same session
	if session1 != session2 {
		t.Fatal("Sessions should be the same")
	}
	
	// Create another session with a different algorithm
	session3, err := manager.CreateSession("client2", AlgorithmChacha20)
	if err != nil {
		t.Fatalf("Failed to create second session: %v", err)
	}
	
	// Verify the algorithm
	if session3.Encryptor.Algorithm() != AlgorithmChacha20 {
		t.Fatalf("Expected ChaCha20 algorithm, got %s", session3.Encryptor.Algorithm())
	}
	
	// Test session count
	if count := manager.GetSessionCount(); count != 2 {
		t.Fatalf("Expected 2 sessions, got %d", count)
	}
	
	// Test key rotation
	if err := manager.RotateSessionKey("client1"); err != nil {
		t.Fatalf("Failed to rotate key: %v", err)
	}
	
	// Test session removal
	if err := manager.RemoveSession("client1"); err != nil {
		t.Fatalf("Failed to remove session: %v", err)
	}
	
	// Verify session was removed
	if _, err := manager.GetSession("client1"); err == nil {
		t.Fatal("Session should have been removed")
	}
	
	// Test session count after removal
	if count := manager.GetSessionCount(); count != 1 {
		t.Fatalf("Expected 1 session after removal, got %d", count)
	}
}
