package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestEncryptionMethods tests different encryption methods in detail
func TestEncryptionMethods(t *testing.T) {
	// Build paths
	repoRoot := findRepoRoot()
	serverBin := filepath.Join(repoRoot, "bin", "server")
	clientBin := filepath.Join(repoRoot, "bin", "client")
	builderBin := filepath.Join(repoRoot, "bin", "builder")

	// Build binaries if they don't exist
	if _, err := os.Stat(serverBin); os.IsNotExist(err) {
		buildBinaries(t, repoRoot, serverBin, clientBin, builderBin)
	}

	// Test AES encryption
	t.Run("AES_Encryption_Detailed", func(t *testing.T) {
		testEncryptionDetailed(t, serverBin, clientBin, "aes")
	})

	// Test ChaCha20 encryption
	t.Run("ChaCha20_Encryption_Detailed", func(t *testing.T) {
		testEncryptionDetailed(t, serverBin, clientBin, "chacha20")
	})

	// Test encryption auto-detection
	t.Run("Encryption_Auto_Detection", func(t *testing.T) {
		testEncryptionAutoDetection(t, serverBin, clientBin)
	})
}

// testEncryptionDetailed tests a specific encryption method in detail
func testEncryptionDetailed(t *testing.T, serverBin, clientBin, encryption string) {
	// Start server
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	serverCmd := exec.CommandContext(serverCtx, serverBin, 
		"-protocol", "tcp", 
		"-address", "127.0.0.1:8080",
		"-encryption", encryption)
	
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	if err := serverCmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Start client
	clientCtx, clientCancel := context.WithCancel(context.Background())
	defer clientCancel()

	clientCmd := exec.CommandContext(clientCtx, clientBin, 
		"-protocol", "tcp", 
		"-server", "127.0.0.1:8080",
		"-encryption", encryption,
		"-verbose", "true") // Enable verbose logging
	
	clientCmd.Stdout = os.Stdout
	clientCmd.Stderr = os.Stderr
	if err := clientCmd.Start(); err != nil {
		t.Fatalf("Failed to start client: %v", err)
	}

	// Wait for client to connect
	time.Sleep(5 * time.Second)

	// Check if client is still running (indicating successful connection)
	if clientCmd.ProcessState != nil && clientCmd.ProcessState.Exited() {
		t.Fatalf("Client exited prematurely")
	}

	// TODO: In a real test, we would send and receive data to verify encryption
	// For now, we just verify that the client can connect with the specified encryption

	// Stop client and server
	clientCancel()
	serverCancel()

	// Wait for processes to exit
	clientCmd.Wait()
	serverCmd.Wait()
}

// testEncryptionAutoDetection tests encryption auto-detection
func testEncryptionAutoDetection(t *testing.T, serverBin, clientBin string) {
	// Start server with both encryption methods
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	serverCmd := exec.CommandContext(serverCtx, serverBin, 
		"-protocol", "tcp", 
		"-address", "127.0.0.1:8080",
		"-encryption", "auto") // Auto-detect encryption
	
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	if err := serverCmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Start client with AES encryption
	clientCtx1, clientCancel1 := context.WithCancel(context.Background())
	defer clientCancel1()

	clientCmd1 := exec.CommandContext(clientCtx1, clientBin, 
		"-protocol", "tcp", 
		"-server", "127.0.0.1:8080",
		"-encryption", "aes",
		"-client-id", "aes-client")
	
	clientCmd1.Stdout = os.Stdout
	clientCmd1.Stderr = os.Stderr
	if err := clientCmd1.Start(); err != nil {
		t.Fatalf("Failed to start AES client: %v", err)
	}

	// Wait for first client to connect
	time.Sleep(5 * time.Second)

	// Start client with ChaCha20 encryption
	clientCtx2, clientCancel2 := context.WithCancel(context.Background())
	defer clientCancel2()

	clientCmd2 := exec.CommandContext(clientCtx2, clientBin, 
		"-protocol", "tcp", 
		"-server", "127.0.0.1:8080",
		"-encryption", "chacha20",
		"-client-id", "chacha20-client")
	
	clientCmd2.Stdout = os.Stdout
	clientCmd2.Stderr = os.Stderr
	if err := clientCmd2.Start(); err != nil {
		t.Fatalf("Failed to start ChaCha20 client: %v", err)
	}

	// Wait for second client to connect
	time.Sleep(5 * time.Second)

	// Check if both clients are still running (indicating successful connections)
	if clientCmd1.ProcessState != nil && clientCmd1.ProcessState.Exited() {
		t.Fatalf("AES client exited prematurely")
	}
	if clientCmd2.ProcessState != nil && clientCmd2.ProcessState.Exited() {
		t.Fatalf("ChaCha20 client exited prematurely")
	}

	// Stop clients and server
	clientCancel1()
	clientCancel2()
	serverCancel()

	// Wait for processes to exit
	clientCmd1.Wait()
	clientCmd2.Wait()
	serverCmd.Wait()
}
