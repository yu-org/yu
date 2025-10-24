#!/bin/bash
# 编译测试程序为 RISC-V 字节码

echo "🔧 编译 RISC-V 测试程序..."

# 检查 RISC-V 工具链
if ! command -v riscv64-elf-gcc &> /dev/null; then
    echo "❌ 未找到 riscv64-elf-gcc，请先安装 RISC-V 工具链"
    exit 1
fi

# 创建输出目录
mkdir -p compiled

# 编译 simple_add.c
echo "📦 编译 simple_add.c..."
riscv64-elf-gcc -nostdlib -static -o compiled/simple_add.elf simple_add.c
if [ $? -eq 0 ]; then
    echo "✅ simple_add.elf 编译成功"
    # 提取 .text 段作为字节码
    riscv64-elf-objcopy -O binary --only-section=.text compiled/simple_add.elf compiled/simple_add.bin
    echo "✅ simple_add.bin 提取成功"
else
    echo "❌ simple_add.c 编译失败"
fi

# 编译 fibonacci.c
echo "📦 编译 fibonacci.c..."
riscv64-elf-gcc -nostdlib -static -o compiled/fibonacci.elf fibonacci.c
if [ $? -eq 0 ]; then
    echo "✅ fibonacci.elf 编译成功"
    riscv64-elf-objcopy -O binary --only-section=.text compiled/fibonacci.elf compiled/fibonacci.bin
    echo "✅ fibonacci.bin 提取成功"
else
    echo "❌ fibonacci.c 编译失败"
fi

# 编译 array_sum.c
echo "📦 编译 array_sum.c..."
riscv64-elf-gcc -nostdlib -static -o compiled/array_sum.elf array_sum.c
if [ $? -eq 0 ]; then
    echo "✅ array_sum.elf 编译成功"
    riscv64-elf-objcopy -O binary --only-section=.text compiled/array_sum.elf compiled/array_sum.bin
    echo "✅ array_sum.bin 提取成功"
else
    echo "❌ array_sum.c 编译失败"
fi

echo "🎉 所有测试程序编译完成！"
echo "📁 编译结果在 compiled/ 目录中"
