#!/bin/bash

# 后台运行 eth，并记录 PID
./yu_eth > yu_eth.log 2>&1 &
YUETH_PID=$!

# 检查 eth 是否启动成功
sleep 2
if ! ps -p $YUETH_PID > /dev/null; then
    echo "Error: Failed to start ./yu_eth"
    cat yu_eth.log
    exit 1
fi

echo "eth started with PID $YUETH_PID"

# 运行 transfer_test
echo "Running transfer_test..."
./bin/transfer_test --as-client true
TEST_EXIT_CODE=$?

# 停止 eth
echo "Stopping eth (PID $YUETH_PID)..."
kill $YUETH_PID
wait $YUETH_PID 2>/dev/null  # 等待进程结束，抑制 "No such process" 错误

# 根据测试结果退出
if [ $TEST_EXIT_CODE -ne 0 ]; then
    echo "Test failed with exit code $TEST_EXIT_CODE"
    exit $TEST_EXIT_CODE
else
    echo "Test completed successfully"
    exit 0
fi 