package test

import (
	"dinoc2/pkg/module"
	"dinoc2/pkg/module/isolation"
	"dinoc2/pkg/module/loader"
	"dinoc2/pkg/module/manager"
	"dinoc2/pkg/module/shell"
	"fmt"
	"testing"
	"time"
)

func TestModuleSystem(t *testing.T) {
	// Create module manager
	mgr, err := manager.NewModuleManager()
	if err != nil {
		t.Fatalf("Failed to create module manager: %v", err)
	}

	// Register shell module
	module.RegisterModule("shell", shell.NewShellModule)

	// Load shell module
	mod, err := mgr.LoadModule("shell", "shell", loader.LoaderTypeNative)
	if err != nil {
		t.Fatalf("Failed to load shell module: %v", err)
	}

	// Initialize shell module
	err = mgr.InitModule("shell", nil)
	if err != nil {
		t.Fatalf("Failed to initialize shell module: %v", err)
	}

	// Execute command on shell module
	result, err := mgr.ExecModule("shell", "echo Hello, World!")
	if err != nil {
		t.Fatalf("Failed to execute command on shell module: %v", err)
	}

	// Verify result
	fmt.Printf("Shell module result: %v\n", result)

	// Test module isolation
	isolatedMod := isolation.NewIsolatedModule(mod, "shell", 5*time.Second)

	// Execute command on isolated module
	result, err = isolatedMod.Exec("echo Isolated module test")
	if err != nil {
		t.Fatalf("Failed to execute command on isolated module: %v", err)
	}

	// Verify result
	fmt.Printf("Isolated module result: %v\n", result)

	// Shutdown module
	err = mgr.ShutdownModule("shell")
	if err != nil {
		t.Fatalf("Failed to shutdown shell module: %v", err)
	}

	// Verify module is shutdown
	status, err := mgr.GetModuleInfo("shell")
	if err != nil {
		t.Fatalf("Failed to get module info: %v", err)
	}

	if status.Status.Running {
		t.Fatalf("Module should be shutdown")
	}

	fmt.Println("Module system test completed successfully")
}
