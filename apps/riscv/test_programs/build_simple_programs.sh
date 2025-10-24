#!/bin/bash
# 编译简单的 RISC-V 程序，直接使用 ELF 文件

echo "🔧 编译简单的 RISC-V 程序..."

# 检查 RISC-V 工具链
if ! command -v riscv64-elf-gcc &> /dev/null; then
    echo "❌ 未找到 riscv64-elf-gcc，请先安装 RISC-V 工具链"
    exit 1
fi

# 创建输出目录
mkdir -p compiled

# 编译一个非常简单的程序
echo "📦 编译 simple_return.c..."
cat > compiled/simple_return.c << 'EOF'
// 最简单的 RISC-V 程序 - 直接返回 42
int main() {
    return 42;
}
EOF

riscv64-elf-gcc -nostdlib -static -o compiled/simple_return.elf compiled/simple_return.c
if [ $? -eq 0 ]; then
    echo "✅ simple_return.elf 编译成功"
    # 直接使用整个 ELF 文件
    cp compiled/simple_return.elf compiled/simple_return.bin
    echo "✅ simple_return.bin 复制成功"
else
    echo "❌ simple_return.c 编译失败"
fi

# 编译一个简单的加法程序
echo "📦 编译 simple_add.c..."
cat > compiled/simple_add.c << 'EOF'
// 简单的加法程序
int add(int a, int b) {
    return a + b;
}

int main() {
    return add(3, 4);  // 返回 7
}
EOF

riscv64-elf-gcc -nostdlib -static -o compiled/simple_add.elf compiled/simple_add.c
if [ $? -eq 0 ]; then
    echo "✅ simple_add.elf 编译成功"
    cp compiled/simple_add.elf compiled/simple_add.bin
    echo "✅ simple_add.bin 复制成功"
else
    echo "❌ simple_add.c 编译失败"
fi

echo "🎉 简单程序编译完成！"
echo "📁 编译结果在 compiled/ 目录中"
