# RISC-V VM Integration

This module provides RISC-V virtual machine integration for the Yu blockchain framework using CKB-VM through CGO.

## Overview

The RISC-V VM module allows you to execute RISC-V programs on the Yu blockchain, providing a flexible and secure way to run custom logic. It integrates with CKB-VM, a RISC-V virtual machine implementation used in the CKB blockchain.

## Features

- **RISC-V Program Execution**: Execute RISC-V binaries using CKB-VM v0.24.0
- **CGO Integration**: Seamless integration with CKB-VM via Rust FFI wrapper
- **Yu Tripod Integration**: Full integration with Yu framework as a Tripod
- **Transaction Support**: Execute scripts through blockchain transactions
- **State Management**: Store and query script execution results

## Prerequisites

### CKB-VM Integration

This module uses a Rust FFI wrapper to integrate with CKB-VM. The wrapper is automatically built as part of the build process.

### Required Dependencies

- **Rust**: For building the CKB-VM wrapper (latest stable version)
- **CGO**: Enabled in your Go environment
- **C Compiler**: GCC or Clang
- **Git**: For downloading CKB-VM dependencies

## Installation

1. **Build the module**:
   ```bash
   cd apps/riscv
   make all
   ```

   This will automatically:
   - Build the Rust FFI wrapper
   - Build the Go module with CGO integration

2. **Alternative manual build**:
   ```bash
   # Build the Rust wrapper first
   make build-wrapper
   
   # Then build the Go module
   go build
   ```

## Usage

### Basic VM Usage

```go
package main

import (
    "fmt"
    "github.com/yu-org/yu/apps/riscv"
)

func main() {
    vm := riscv.NewVMRunner()
    defer vm.Cleanup()

    // Initialize VM with cell data
    cellData := []byte{0x01, 0x02, 0x03}
    err := vm.Init(cellData)
    if err != nil {
        panic(err)
    }

    // Load RISC-V program
    program := []byte{0x13, 0x05, 0x00, 0x00} // Example RISC-V instructions
    err = vm.LoadProgram(program)
    if err != nil {
        panic(err)
    }

    // Execute program
    err = vm.Run()
    if err != nil {
        fmt.Printf("Execution failed: %v\n", err)
        return
    }

    // Get exit code
    exitCode := vm.GetExitCode()
    fmt.Printf("Program exited with code: %d\n", exitCode)
}
```

### Yu Tripod Integration

```go
package main

import (
    "github.com/yu-org/yu/apps/riscv"
    "github.com/yu-org/yu/core/startup"
)

func main() {
    // Create RISC-V Tripod
    riscvTripod := riscv.NewRiscvTripod()

    // Start Yu with RISC-V Tripod
    startup.DefaultStartup(riscvTripod)
}
```

### Transaction-based Script Execution

```go
// Execute a script through a transaction
txData := map[string]interface{}{
    "script_data": []byte{0x13, 0x05, 0x00, 0x00}, // RISC-V program
    "cell_data":   []byte{0x01, 0x02, 0x03},        // Input data
}

// Submit transaction to blockchain
// The script will be executed when the transaction is processed
```

## API Reference

### VMRunner

#### `NewVMRunner() *VMRunner`
Creates a new VM runner instance.

#### `Init(cellData []byte) error`
Initializes the VM with cell data.

#### `LoadProgram(program []byte) error`
Loads a RISC-V program into the VM.

#### `Run() error`
Executes the loaded program.

#### `GetExitCode() int`
Returns the exit code of the last executed program.

#### `Cleanup()`
Cleans up VM resources.

#### `ExecuteProgram(cellData, program []byte) error`
Convenience method that initializes, loads, and runs a program.

### RiscvTripod

#### `NewRiscvTripod() *RiscvTripod`
Creates a new RISC-V Tripod.

#### `ExecuteScript(ctx *context.WriteContext) error`
Executes a RISC-V script through a blockchain transaction.

#### `QueryScriptResult(ctx *context.ReadContext)`
Queries the result of a script execution.

#### `ExecuteRiscvScript(scriptData, cellData []byte) (int, error)`
Helper method to execute RISC-V scripts directly.

## Configuration

### CGO Flags

The module uses the following CGO flags:

```go
#cgo CFLAGS: -I./ckb-vm-wrapper/src
#cgo LDFLAGS: -L./ckb-vm-wrapper/target/release -lckb_vm_wrapper -lpthread -ldl -lm
```

The wrapper is compiled as a static library (`libckb_vm_wrapper.a`, ~19MB) for easier deployment and distribution. It integrates CKB-VM v0.24.0 through a Rust FFI wrapper.

### Build Tags

You can use build tags to conditionally compile the CGO code:

```bash
go build -tags cgo
```

## Testing

Run the tests:

```bash
go test -v
```

## Troubleshooting

### Common Issues

1. **CGO compilation errors**:
   - Ensure CKB-VM is properly built and installed
   - Check that library paths are correctly set
   - Verify C compiler is available

2. **Library not found errors**:
   - Ensure the Rust wrapper is built: `make build-wrapper`
   - Check that `libckb_vm_wrapper` is in the correct path
   - Verify CGO can find the library

3. **Rust build errors**:
   - Ensure Rust is properly installed
   - Check that CKB-VM can be downloaded from GitHub
   - Verify Rust dependencies are resolved

### Debug Mode

Enable debug logging:

```go
logrus.SetLevel(logrus.DebugLevel)
```

## Performance Considerations

- **VM Initialization**: VM initialization has overhead, consider reusing instances
- **Memory Usage**: RISC-V programs consume memory, monitor usage in production
- **Execution Time**: Set appropriate gas limits for script execution
- **Concurrency**: VM instances are not thread-safe, use one per goroutine

## Security Considerations

- **Script Validation**: Validate RISC-V programs before execution
- **Resource Limits**: Set appropriate limits on execution time and memory
- **Sandboxing**: Consider additional sandboxing for untrusted scripts
- **Input Validation**: Validate all inputs to script execution functions

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This module is part of the Yu project and follows the same license terms.
