package riscv

import (
	"io"
	"os"
	"testing"
)

func TestSimpleReturnProgram(t *testing.T) {
	// Read compiled RISC-V ELF file
	file, err := os.Open("test_programs/compiled/simple_return.elf")
	if err != nil {
		t.Fatalf("Failed to open simple_return.elf: %v", err)
	}
	defer file.Close()

	program, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read simple_return.elf: %v", err)
	}

	vm := NewVMRunner()
	defer vm.Cleanup()

	t.Logf("Executing simple_return program (%d bytes)", len(program))

	// Execute program
	err = vm.ExecuteProgram(program)
	if err != nil {
		t.Logf("Program execution result: %v", err)
		exitCode := vm.GetExitCode()
		t.Logf("Exit code: %d", exitCode)
		t.Errorf("Program execution failed with exit code: %d, expected: 0", exitCode)
	} else {
		t.Log("Program executed successfully")
		exitCode := vm.GetExitCode()
		t.Logf("Exit code: %d", exitCode)
		if exitCode != 0 {
			t.Errorf("Program execution failed with exit code: %d, expected: 0", exitCode)
		}
	}
}

func TestSimpleAddProgram(t *testing.T) {
	// Read compiled RISC-V ELF file
	file, err := os.Open("test_programs/compiled/simple_add.elf")
	if err != nil {
		t.Fatalf("Failed to open simple_add.elf: %v", err)
	}
	defer file.Close()

	program, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read simple_add.elf: %v", err)
	}

	vm := NewVMRunner()
	defer vm.Cleanup()

	t.Logf("Executing simple_add program (%d bytes)", len(program))

	// Execute program
	err = vm.ExecuteProgram(program)
	if err != nil {
		t.Logf("Program execution result: %v", err)
		exitCode := vm.GetExitCode()
		t.Logf("Exit code: %d", exitCode)
		t.Errorf("Program execution failed with exit code: %d, expected: 0", exitCode)
	} else {
		t.Log("Program executed successfully")
		exitCode := vm.GetExitCode()
		t.Logf("Exit code: %d", exitCode)
		if exitCode != 0 {
			t.Errorf("Program execution failed with exit code: %d, expected: 0", exitCode)
		}
	}
}

func TestFibonacciProgram(t *testing.T) {
	// Read compiled RISC-V ELF file
	file, err := os.Open("test_programs/compiled/fibonacci.elf")
	if err != nil {
		t.Fatalf("Failed to open fibonacci.elf: %v", err)
	}
	defer file.Close()

	program, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read fibonacci.elf: %v", err)
	}

	vm := NewVMRunner()
	defer vm.Cleanup()

	t.Logf("Executing fibonacci program (%d bytes)", len(program))

	// Execute program
	err = vm.ExecuteProgram(program)
	if err != nil {
		t.Logf("Program execution result: %v", err)
		exitCode := vm.GetExitCode()
		t.Logf("Exit code: %d", exitCode)
		t.Errorf("Program execution failed with exit code: %d, expected: 0", exitCode)
	} else {
		t.Log("Program executed successfully")
		exitCode := vm.GetExitCode()
		t.Logf("Exit code: %d", exitCode)
		if exitCode != 0 {
			t.Errorf("Program execution failed with exit code: %d, expected: 0", exitCode)
		}
	}
}
