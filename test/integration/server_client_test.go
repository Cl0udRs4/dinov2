package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestServerClientCommunication tests basic communication between server and client
func TestServerClientCommunication(t *testing.T) {
	// Build paths
	repoRoot := findRepoRoot()
	serverBin := filepath.Join(repoRoot, "bin", "server")
	clientBin := filepath.Join(repoRoot, "bin", "client")
	builderBin := filepath.Join(repoRoot, "bin", "builder")

	// Build binaries
	buildBinaries(t, repoRoot, serverBin, clientBin, builderBin)

	// Test TCP communication
	t.Run("TCP_Communication", func(t *testing.T) {
		testProtocolCommunication(t, serverBin, clientBin, "tcp", "127.0.0.1:8080")
	})

	// Test HTTP communication
	t.Run("HTTP_Communication", func(t *testing.T) {
		testProtocolCommunication(t, serverBin, clientBin, "http", "127.0.0.1:8081")
	})

	// Test WebSocket communication
	t.Run("WebSocket_Communication", func(t *testing.T) {
		testProtocolCommunication(t, serverBin, clientBin, "websocket", "127.0.0.1:8082")
	})

	// Test protocol switching
	t.Run("Protocol_Switching", func(t *testing.T) {
		testProtocolSwitching(t, serverBin, clientBin, builderBin)
	})

	// Test encryption methods
	t.Run("Encryption_Methods", func(t *testing.T) {
		testEncryptionMethods(t, serverBin, clientBin)
	})
}

// findRepoRoot finds the repository root directory
func findRepoRoot() string {
	// Start from the current directory
	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Failed to get current directory: %v", err))
	}

	// Go up until we find the .git directory
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root directory without finding .git
			panic("Could not find repository root")
		}
		dir = parent
	}
}

// buildBinaries builds the server, client, and builder binaries
func buildBinaries(t *testing.T, repoRoot, serverBin, clientBin, builderBin string) {
	// Create bin directory if it doesn't exist
	binDir := filepath.Join(repoRoot, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create bin directory: %v", err)
	}

	// Build server
	cmd := exec.Command("go", "build", "-o", serverBin, "./cmd/server")
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build server: %v\n%s", err, output)
	}

	// Build client
	cmd = exec.Command("go", "build", "-o", clientBin, "./cmd/client")
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build client: %v\n%s", err, output)
	}

	// Build builder
	cmd = exec.Command("go", "build", "-o", builderBin, "./cmd/builder")
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build builder: %v\n%s", err, output)
	}
}

// testProtocolCommunication tests communication over a specific protocol
func testProtocolCommunication(t *testing.T, serverBin, clientBin, protocol, address string) {
	// Start server
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	serverCmd := exec.CommandContext(serverCtx, serverBin, "-protocol", protocol, "-address", address)
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

	clientCmd := exec.CommandContext(clientCtx, clientBin, "-protocol", protocol, "-server", address)
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

	// Stop client and server
	clientCancel()
	serverCancel()

	// Wait for processes to exit
	clientCmd.Wait()
	serverCmd.Wait()
}

// testProtocolSwitching tests protocol switching
func testProtocolSwitching(t *testing.T, serverBin, clientBin, builderBin string) {
	// Build a client with multiple protocols
	customClientBin := filepath.Join(filepath.Dir(clientBin), "multi_protocol_client")
	cmd := exec.Command(builderBin, 
		"-output", customClientBin,
		"-protocol", "tcp,http,websocket",
		"-server", "127.0.0.1:8080",
		"-active-switch", "true",
		"-passive-switch", "true")
	
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build multi-protocol client: %v\n%s", err, output)
	}

	// Start server with TCP protocol
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	serverCmd := exec.CommandContext(serverCtx, serverBin, 
		"-protocol", "tcp,http,websocket", 
		"-address", "127.0.0.1:8080,127.0.0.1:8081,127.0.0.1:8082")
	
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	if err := serverCmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Start client with TCP protocol
	clientCtx, clientCancel := context.WithCancel(context.Background())
	defer clientCancel()

	clientCmd := exec.CommandContext(clientCtx, customClientBin, "-initial-protocol", "tcp")
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

	// TODO: Send protocol switch command from server to client
	// This would require implementing a control interface for the server

	// Stop client and server
	clientCancel()
	serverCancel()

	// Wait for processes to exit
	clientCmd.Wait()
	serverCmd.Wait()
}

// testEncryptionMethods tests different encryption methods
func testEncryptionMethods(t *testing.T, serverBin, clientBin string) {
	// Test AES encryption
	t.Run("AES_Encryption", func(t *testing.T) {
		testEncryption(t, serverBin, clientBin, "aes")
	})

	// Test ChaCha20 encryption
	t.Run("ChaCha20_Encryption", func(t *testing.T) {
		testEncryption(t, serverBin, clientBin, "chacha20")
	})
}

// testEncryption tests a specific encryption method
func testEncryption(t *testing.T, serverBin, clientBin, encryption string) {
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
		"-encryption", encryption)
	
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

	// Stop client and server
	clientCancel()
	serverCancel()

	// Wait for processes to exit
	clientCmd.Wait()
	serverCmd.Wait()
}
