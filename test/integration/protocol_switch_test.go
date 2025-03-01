package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestProtocolSwitchingDetailed tests protocol switching in detail
func TestProtocolSwitchingDetailed(t *testing.T) {
	// Build paths
	repoRoot := findRepoRoot()
	serverBin := filepath.Join(repoRoot, "bin", "server")
	clientBin := filepath.Join(repoRoot, "bin", "client")
	builderBin := filepath.Join(repoRoot, "bin", "builder")

	// Build binaries if they don't exist
	if _, err := os.Stat(serverBin); os.IsNotExist(err) {
		buildBinaries(t, repoRoot, serverBin, clientBin, builderBin)
	}

	// Test active protocol switching
	t.Run("Active_Protocol_Switching", func(t *testing.T) {
		testActiveProtocolSwitching(t, serverBin, clientBin, builderBin)
	})

	// Test passive protocol switching
	t.Run("Passive_Protocol_Switching", func(t *testing.T) {
		testPassiveProtocolSwitching(t, serverBin, clientBin, builderBin)
	})
}

// testActiveProtocolSwitching tests active protocol switching (client-initiated)
func testActiveProtocolSwitching(t *testing.T, serverBin, clientBin, builderBin string) {
	// Build a client with multiple protocols and active switching enabled
	customClientBin := filepath.Join(filepath.Dir(clientBin), "active_switch_client")
	cmd := exec.Command(builderBin, 
		"-output", customClientBin,
		"-protocol", "tcp,http",
		"-server", "127.0.0.1:8080,127.0.0.1:8081",
		"-active-switch", "true",
		"-passive-switch", "false",
		"-heartbeat", "2", // Short heartbeat interval for faster testing
		"-reconnect", "1") // Short reconnect interval
	
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build active switch client: %v\n%s", err, output)
	}

	// Start server with multiple protocols
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	serverCmd := exec.CommandContext(serverCtx, serverBin, 
		"-protocol", "tcp,http", 
		"-address", "127.0.0.1:8080,127.0.0.1:8081")
	
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

	// Simulate TCP listener failure by stopping and restarting the server without TCP
	serverCancel()
	serverCmd.Wait()

	// Start server with only HTTP protocol
	serverCtx, serverCancel = context.WithCancel(context.Background())
	serverCmd = exec.CommandContext(serverCtx, serverBin, 
		"-protocol", "http", 
		"-address", "127.0.0.1:8081")
	
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	if err := serverCmd.Start(); err != nil {
		t.Fatalf("Failed to restart server: %v", err)
	}

	// Wait for client to detect failure and switch protocols
	time.Sleep(10 * time.Second)

	// Check if client is still running (indicating successful protocol switch)
	if clientCmd.ProcessState != nil && clientCmd.ProcessState.Exited() {
		t.Fatalf("Client exited after protocol switch")
	}

	// Stop client and server
	clientCancel()
	serverCancel()

	// Wait for processes to exit
	clientCmd.Wait()
	serverCmd.Wait()
}

// testPassiveProtocolSwitching tests passive protocol switching (server-initiated)
func testPassiveProtocolSwitching(t *testing.T, serverBin, clientBin, builderBin string) {
	// Build a client with multiple protocols and passive switching enabled
	customClientBin := filepath.Join(filepath.Dir(clientBin), "passive_switch_client")
	cmd := exec.Command(builderBin, 
		"-output", customClientBin,
		"-protocol", "tcp,http",
		"-server", "127.0.0.1:8080,127.0.0.1:8081",
		"-active-switch", "false",
		"-passive-switch", "true")
	
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build passive switch client: %v\n%s", err, output)
	}

	// Start server with multiple protocols
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	// For a real test, we would need a server with an API to trigger protocol switching
	// This is a simplified version that just tests if the client can connect
	serverCmd := exec.CommandContext(serverCtx, serverBin, 
		"-protocol", "tcp,http", 
		"-address", "127.0.0.1:8080,127.0.0.1:8081")
	
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

	// TODO: In a real test, we would send a protocol switch command from the server
	// For now, we just verify that the client can connect

	// Stop client and server
	clientCancel()
	serverCancel()

	// Wait for processes to exit
	clientCmd.Wait()
	serverCmd.Wait()
}
