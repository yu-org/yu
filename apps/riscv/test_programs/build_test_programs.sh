#!/bin/bash
# Compile RISC-V programs to ELF format for CKB-VM

echo "🔧 Compiling RISC-V programs to ELF format..."

# Check RISC-V toolchain
if ! command -v riscv64-elf-gcc &> /dev/null; then
    echo "❌ riscv64-elf-gcc not found, please install RISC-V toolchain first"
    exit 1
fi

# Create output directory
mkdir -p compiled

# Compile simple add program
echo "📦 Compiling simple_add.c..."
riscv64-elf-gcc -nostdlib -static -o compiled/simple_add.elf simple_add.c
if [ $? -eq 0 ]; then
    echo "✅ simple_add.elf compiled successfully"
else
    echo "❌ simple_add.c compilation failed"
fi

# Compile fibonacci program
echo "📦 Compiling fibonacci.c..."
riscv64-elf-gcc -nostdlib -static -o compiled/fibonacci.elf fibonacci.c
if [ $? -eq 0 ]; then
    echo "✅ fibonacci.elf compiled successfully"
else
    echo "❌ fibonacci.c compilation failed"
fi

# Compile array sum program
echo "📦 Compiling array_sum.c..."
riscv64-elf-gcc -nostdlib -static -o compiled/array_sum.elf array_sum.c
if [ $? -eq 0 ]; then
    echo "✅ array_sum.elf compiled successfully"
else
    echo "❌ array_sum.c compilation failed"
fi

echo "🎉 RISC-V programs compiled to ELF format!"
echo "📁 Compiled files are in compiled/ directory"
