package riscv

/*
#cgo CFLAGS: -I./ckb-vm-wrapper/src
#cgo LDFLAGS: -L./ckb-vm-wrapper/target/release -lckb_vm_wrapper -lpthread -ldl -lm
#include <stdlib.h>
#include <stdint.h>

// CKB-VM FFI types and functions
typedef struct {
    uint8_t* data;
    size_t length;
} ckb_vm_cell_data_t;

typedef struct {
    int exit_code;
    const char* error_message;
} ckb_vm_result_t;

// CKB-VM FFI function declarations
extern int ckb_vm_init(ckb_vm_cell_data_t* cell_data, size_t cell_data_length);
extern int ckb_vm_load_program(const uint8_t* program, size_t program_length);
extern int ckb_vm_run(void);
extern ckb_vm_result_t ckb_vm_get_result(void);
extern void ckb_vm_cleanup(void);
*/
import "C"

import (
	"errors"
	"fmt"

	"github.com/yu-org/yu/core/tripod"
)

// VMRunner represents a RISC-V VM runner using CKB-VM as a Bronze service
type VMRunner struct {
	*tripod.Bronze
	initialized bool
}

// NewVMRunner creates a new VM runner instance as a Bronze service
func NewVMRunner() *VMRunner {
	return &VMRunner{
		Bronze:      tripod.NewBronzeWithName("risc-v"),
		initialized: false,
	}
}

// Init initializes the CKB-VM
func (vm *VMRunner) Init() error {
	if vm.initialized {
		return errors.New("VM already initialized")
	}

	// Initialize CKB-VM with no cell data
	result := C.ckb_vm_init(nil, 0)
	if result != 0 {
		return fmt.Errorf("failed to initialize CKB-VM: %d", result)
	}

	vm.initialized = true
	return nil
}

// LoadProgram loads a RISC-V program into the VM
func (vm *VMRunner) LoadProgram(program []byte) error {
	if !vm.initialized {
		return errors.New("VM not initialized")
	}

	if len(program) == 0 {
		return errors.New("program cannot be empty")
	}

	// Allocate C memory for the program
	cProgram := C.CBytes(program)
	defer C.free(cProgram)

	cProgramLength := C.size_t(len(program))

	// Load program into VM
	result := C.ckb_vm_load_program((*C.uint8_t)(cProgram), cProgramLength)
	if result != 0 {
		return fmt.Errorf("failed to load program: %d", result)
	}

	return nil
}

// Run executes the loaded program in the VM
func (vm *VMRunner) Run() error {
	if !vm.initialized {
		return errors.New("VM not initialized")
	}

	// Run the program
	result := C.ckb_vm_run()
	if result != 0 {
		// Get execution result
		cResult := C.ckb_vm_get_result()
		if cResult.error_message != nil {
			return fmt.Errorf("VM execution failed: %s", C.GoString(cResult.error_message))
		}
		return fmt.Errorf("VM execution failed with exit code: %d", cResult.exit_code)
	}

	return nil
}

// GetExitCode returns the exit code of the last executed program
func (vm *VMRunner) GetExitCode() int {
	if !vm.initialized {
		return -1
	}

	cResult := C.ckb_vm_get_result()
	return int(cResult.exit_code)
}

// Cleanup cleans up the VM resources
func (vm *VMRunner) Cleanup() {
	if vm.initialized {
		C.ckb_vm_cleanup()
		vm.initialized = false
	}
}

// ExecuteProgram is a convenience method that initializes, loads, and runs a program
func (vm *VMRunner) ExecuteProgram(program []byte) error {
	if err := vm.Init(); err != nil {
		return fmt.Errorf("failed to initialize VM: %w", err)
	}
	defer vm.Cleanup()

	if err := vm.LoadProgram(program); err != nil {
		return fmt.Errorf("failed to load program: %w", err)
	}

	if err := vm.Run(); err != nil {
		return fmt.Errorf("failed to run program: %w", err)
	}

	return nil
}
