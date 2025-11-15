#!/bin/bash

# 百度网盘原生Range请求测试脚本
# 基于API分析，测试不使用缓存的纯Range请求方案

echo "=== 百度网盘原生Range请求测试 ==="
echo "基于API文档分析，测试最小化缓存方案"
echo

# 配置参数
# 支持命令行参数：./test_range_native.sh [remote_path] [port]
REMOTE=${1:-"baidupan:"}
PORT=${2:-9999}

# 如果用户传递了crypt:，需要指定实际的路径
if [[ "$REMOTE" == "crypt:" ]]; then
    MUSIC_DIR="/"
    echo "检测到crypt存储，使用根路径"
else
    MUSIC_DIR="/音乐"
fi

LOG_FILE="/tmp/rclone_range_test.log"

# 最小化缓存配置（仅用于连接复用，不用于文件缓存）
CACHE_DIR="/tmp/rclone_min_cache"
mkdir -p "$CACHE_DIR"

echo "测试配置："
echo "- 远程路径: $REMOTE$MUSIC_DIR"
echo "- 本地端口: $PORT"
echo "- 缓存目录: $CACHE_DIR (最小化)"
echo "- 日志文件: $LOG_FILE"
echo

# 停止现有服务
echo "停止现有rclone服务..."
pkill -f "rclone serve http" || true
sleep 2

# 清理最小缓存
echo "清理最小缓存..."
rm -rf "$CACHE_DIR"/*
mkdir -p "$CACHE_DIR"

echo "启动最小化缓存测试服务..."
echo "这个配置主要测试："
echo "1. 原生Range请求支持"
echo "2. dlink生命周期管理"
echo "3. 连接复用优化"
echo

# 启动最小化缓存服务
nohup rclone serve http "$REMOTE$MUSIC_DIR" \
    --addr "0.0.0.0:$PORT" \
    --log-file "$LOG_FILE" \
    --log-level INFO \
    --buffer-size 16M \
    --dir-cache-time 5m \
    --poll-interval 1m \
    --vfs-cache-mode minimal \
    --vfs-cache-max-size 1G \
    --vfs-cache-max-age 5m \
    --vfs-read-ahead 0 \
    --vfs-read-chunk-size 4M \
    --vfs-read-chunk-offset 0 \
    --allow-non-empty \
    --no-checksum \
    --no-modtime \
    --timeout 30s \
    --contimeout 10s \
    --low-level-retries 3 \
    --retries 3 \
    --tpslimit 5 \
    --tpslimit-burst 10 \
    --use-server-modtime \
    --ignore-checksum \
    --ignore-size \
    --max-read-ahead 0 \
    --dir-cache-duration 5m \
    --chunker-chunk-size 4M \
    --chunker-hash-type md5 \
    --chunker-hash-type md5 \
    --chunker-hash-type md5 > /dev/null 2>&1 &

RCLONE_PID=$!
echo "服务已启动，PID: $RCLONE_PID"

# 等待服务启动
echo "等待服务启动..."
sleep 5

# 检查服务状态
if curl -s "http://localhost:$PORT/" > /dev/null; then
    echo "✅ 服务启动成功！"
    echo
    echo "测试访问地址："
    echo "http://localhost:$PORT/"
    echo
    echo "这个最小化缓存方案的特点："
    echo "🔹 vfs-cache-mode: minimal (仅缓存元数据，不缓存文件内容)"
    echo "🔹 vfs-read-ahead: 0 (关闭预读，强制使用Range请求)"
    echo "🔹 vfs-read-chunk-size: 4M (匹配百度网盘API限制)"
    echo "🔹 buffer-size: 16M (适中的缓冲区)"
    echo "🔹 禁用校验和和时间检查，减少API调用"
    echo
    echo "🎵 现在可以测试音频播放和seek功能！"
    echo "📝 监控日志: tail -f $LOG_FILE"
    echo "🛑 停止服务: kill $RCLONE_PID"
    echo
    echo "如果这个方案仍有seek问题，说明问题在于："
    echo "1. dlink生命周期管理"
    echo "2. 百度网盘服务器的Range请求限制"
    echo "3. 302重定向后的连接处理"
    echo
    echo "在这种情况下，缓存方案（stable/hybrid）是必要的workaround。"
else
    echo "❌ 服务启动失败，请检查配置"
    kill $RCLONE_PID 2>/dev/null
    exit 1
fi