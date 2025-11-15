#!/bin/bash

# 百度网盘原生Range测试 - 简化版
# 基于成功启动的配置

echo "=== 百度网盘原生Range测试 ==="
echo

# 参数处理
REMOTE=${1:-"crypt:"}
PORT=${2:-9999}

echo "配置：$REMOTE → 端口 $PORT"
echo

# 停止现有服务
pkill -f "./rclone serve http" || true
sleep 2

echo "启动原生Range服务..."
echo "特点：最小缓存 + 4MB块 + 无预读"
echo

# 启动服务（基于成功启动的配置）
nohup ./rclone serve http "$REMOTE" \
    --addr "0.0.0.0:$PORT" \
    --log-file /tmp/rclone_pure_range.log \
    --log-level INFO \
    --vfs-cache-mode minimal \
    --vfs-cache-max-size 512M \
    --vfs-cache-max-age 2m \
    --vfs-read-ahead 0 \
    --vfs-read-chunk-size 4M \
    --vfs-read-chunk-offset 0 \
    --buffer-size 16M \
    --dir-cache-time 2m \
    --allow-non-empty \
    --timeout 15s \
    --contimeout 10s \
    --low-level-retries 2 \
    --retries 2 \
    --tpslimit 3 > /dev/null 2>&1 &

RCLONE_PID=$!
echo "服务已启动，PID: $RCLONE_PID"

# 等待启动
sleep 5

# 检查状态
if curl -s --max-time 3 "http://localhost:$PORT/" > /dev/null; then
    echo "✅ 服务启动成功！"
    echo
    echo "🌐 http://localhost:$PORT/"
    echo "📝 tail -f /tmp/rclone_pure_range.log"
    echo "🛑 kill $RCLONE_PID"
    echo
    echo "💡 现在测试音频seek功能！"
else
    echo "❌ 启动失败，查看日志："
    tail -10 /tmp/rclone_pure_range.log 2>/dev/null
    kill $RCLONE_PID 2>/dev/null
fi