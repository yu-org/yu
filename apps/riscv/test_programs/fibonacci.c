// 斐波那契数列程序 - CKB-VM 测试用例
// 计算第 n 个斐波那契数

int fibonacci(int n) {
    if (n <= 1) {
        return n;
    }
    return fibonacci(n - 1) + fibonacci(n - 2);
}

int main() {
    // 计算第 10 个斐波那契数
    return fibonacci(10);  // 应该返回 55
}
