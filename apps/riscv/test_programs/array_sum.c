// 数组求和程序 - CKB-VM 测试用例
// 计算数组中所有元素的和

int array_sum(int* arr, int length) {
    int sum = 0;
    for (int i = 0; i < length; i++) {
        sum += arr[i];
    }
    return sum;
}

int main() {
    int arr[] = {1, 2, 3, 4, 5};
    int length = 5;
    return array_sum(arr, length);  // 应该返回 15
}
