#!/bin/bash

# 百度网盘原生Range能力测试脚本
# 最小化配置，纯粹测试API的Range支持

echo "=== 百度网盘原生Range能力测试 ==="
echo "最小化配置，测试纯Range请求支持"
echo

# 参数处理
REMOTE=${1:-"crypt:"}
PORT=${2:-9999}

echo "配置："
echo "- 远程存储: $REMOTE"
echo "- 服务端口: $PORT"
echo

# 停止现有服务
echo "停止现有服务..."
pkill -f "./rclone serve http" || true
sleep 3

# 清理环境
echo "清理缓存环境..."
rm -rf /tmp/rclone_range_*
mkdir -p /tmp/rclone_range_cache

echo "启动原生Range测试服务..."
echo "配置特点："
echo "🔹 最小缓存模式 (minimal)"
echo "🔹 关闭预读 (read-ahead=0)"
echo "🔹 4MB读取块 (匹配百度API限制)"
echo "🔹 16MB缓冲区"
echo

# 启动服务 - 纯粹测试原生Range能力
nohup ./rclone serve http "$REMOTE" \
    --addr "0.0.0.0:$PORT" \
    --log-file /tmp/rclone_range_test.log \
    --log-level INFO \
    --vfs-cache-mode minimal \
    --vfs-cache-max-size 512M \
    --vfs-cache-max-age 2m \
    --vfs-read-ahead 0 \
    --vfs-read-chunk-size 4M \
    --vfs-read-chunk-offset 0 \
    --buffer-size 16M \
    --dir-cache-time 2m \
    --poll-interval 30s \
    --allow-non-empty \
    --timeout 15s \
    --contimeout 10s \
    --low-level-retries 2 \
    --retries 2 \
    --tpslimit 3 \
    --use-server-modtime \
    --ignore-checksum \
    --ignore-size > /dev/null 2>&1 &

RCLONE_PID=$!
echo "服务已启动，PID: $RCLONE_PID"

# 等待服务启动
echo "等待服务启动..."
sleep 8

# 检查服务状态
if curl -s --max-time 5 "http://localhost:$PORT/" > /dev/null 2>&1; then
    echo "✅ 服务启动成功！"
    echo
    echo "🌐 访问地址："
    echo "http://localhost:$PORT/"
    echo
    echo "🎵 现在可以测试音频播放和seek功能！"
    echo
    echo "📊 监控日志："
    echo "tail -f /tmp/rclone_range_test.log"
    echo
    echo "🛑 停止服务："
    echo "kill $RCLONE_PID"
    echo
    echo "💡 如果这个配置仍有seek问题，"
    echo "   说明需要缓存方案来绕过dlink管理复杂性。"
else
    echo "❌ 服务启动失败！"
    echo "查看错误日志："
    tail -20 /tmp/rclone_range_test.log 2>/dev/null || echo "日志文件不存在"
    kill $RCLONE_PID 2>/dev/null
    exit 1
fi