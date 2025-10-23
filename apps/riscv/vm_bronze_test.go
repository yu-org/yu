package riscv

import (
	"testing"
)

func TestVMRunnerAsBronze(t *testing.T) {
	vm := NewVMRunner()
	defer vm.Cleanup()

	// Test initialization
	err := vm.Init()
	if err != nil {
		t.Fatalf("Failed to initialize VM: %v", err)
	}

	if !vm.initialized {
		t.Error("VM should be initialized")
	}

	// Test with empty program
	program := []byte{}
	err = vm.LoadProgram(program)
	if err != nil {
		t.Logf("Expected error when loading empty program: %v", err)
	}

	// Test GetExitCode without running
	exitCode := vm.GetExitCode()
	t.Logf("Exit code: %d", exitCode)

	// Test cleanup
	vm.Cleanup()
	if vm.initialized {
		t.Error("VM should not be initialized after cleanup")
	}
}

func TestExecuteProgram(t *testing.T) {
	vm := NewVMRunner()

	// Test with empty program
	program := []byte{}

	err := vm.ExecuteProgram(program)
	if err != nil {
		t.Logf("Expected error when executing empty program: %v", err)
	}
}

func TestBronzeName(t *testing.T) {
	vm := NewVMRunner()

	// Test that it's properly configured as a Bronze
	if vm.Bronze == nil {
		t.Fatal("Bronze should not be nil")
	}

	name := vm.Bronze.Name()
	if name != "vm" {
		t.Errorf("Expected Bronze name to be 'vm', got '%s'", name)
	}

	t.Logf("Bronze name: %s", name)
}
