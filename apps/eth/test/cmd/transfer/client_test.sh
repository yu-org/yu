#!/bin/bash

# 后台运行 reddio，并记录 PID
./reddio > reddio.log 2>&1 &
REDDIO_PID=$!

# 检查 reddio 是否启动成功
sleep 2
if ! ps -p $REDDIO_PID > /dev/null; then
    echo "Error: Failed to start ./reddio"
    cat reddio.log
    exit 1
fi

echo "reddio started with PID $REDDIO_PID"

# 运行 transfer_test
echo "Running transfer_test..."
./bin/transfer_test --as-client true
TEST_EXIT_CODE=$?

# 停止 reddio
echo "Stopping reddio (PID $REDDIO_PID)..."
kill $REDDIO_PID
wait $REDDIO_PID 2>/dev/null  # 等待进程结束，抑制 "No such process" 错误

# 根据测试结果退出
if [ $TEST_EXIT_CODE -ne 0 ]; then
    echo "Test failed with exit code $TEST_EXIT_CODE"
    exit $TEST_EXIT_CODE
else
    echo "Test completed successfully"
    exit 0
fi 