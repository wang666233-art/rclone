#!/bin/bash

# 百度网盘原生Range测试 - 最终版
# 基于验证成功的配置

echo "=== 百度网盘原生Range测试 - 最终版 ==="
echo

# 参数处理
REMOTE=${1:-"crypt:"}
PORT=${2:-9997}

echo "配置：$REMOTE → 端口 $PORT"
echo

# 停止现有服务
echo "清理现有服务..."
pkill -f "./rclone serve http" || true
sleep 3

echo "启动原生Range服务..."
echo "🔹 最小缓存模式 (minimal)"
echo "🔹 4MB读取块 (匹配百度API)"  
echo "🔹 关闭预读 (强制Range请求)"
echo "🔹 16MB缓冲区"
echo

# 启动服务
./rclone serve http "$REMOTE" \
    --addr "0.0.0.0:$PORT" \
    --log-level INFO \
    --vfs-cache-mode minimal \
    --vfs-cache-max-size 512M \
    --vfs-cache-max-age 2m \
    --vfs-read-ahead 0 \
    --vfs-read-chunk-size 4M \
    --buffer-size 16M \
    --dir-cache-time 2m \
    --allow-non-empty \
    --timeout 15s \
    --contimeout 10s \
    --low-level-retries 2 \
    --retries 2 \
    --tpslimit 3 &

RCLONE_PID=$!
echo "服务已启动，PID: $RCLONE_PID"

# 等待启动
echo "等待服务启动..."
sleep 8

# 检查服务
if curl -s --max-time 3 "http://localhost:$PORT/" > /dev/null 2>&1; then
    echo "✅ 服务启动成功！"
    echo
    echo "🌐 访问地址："
    echo "   http://localhost:$PORT/"
    echo
    echo "🎵 现在可以测试音频播放和seek功能！"
    echo
    echo "💡 这个配置测试百度网盘原生Range请求支持："
    echo "   - 如果seek正常，说明原生Range支持良好"
    echo "   - 如果仍有问题，说明需要缓存方案"
    echo
    echo "🛑 停止服务：kill $RCLONE_PID"
    echo
    echo "📊 实时监控："
    echo "   watch -n 1 'ps aux | grep rclone'"
    echo
else
    echo "❌ 服务启动失败"
    kill $RCLONE_PID 2>/dev/null
    exit 1
fi