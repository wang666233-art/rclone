#!/bin/bash

# 无缓存策略 - 零磁盘占用
# 使用方法: ./start_audio_stream_nocache.sh [remote:path] [port]

REMOTE_PATH=${1:-"./audio"}
PORT=${2:-8080}
SERVER_ADDR="0.0.0.0:$PORT"

echo "启动rclone音频流媒体服务 (无缓存策略)..."
echo "远程路径: $REMOTE_PATH"
echo "监听地址: $SERVER_ADDR"
echo "策略: 直接流式传输，零磁盘占用"
echo "注意: seek性能可能较差，适合顺序播放"
echo ""

# 创建音频目录（如果不存在）
if [ "$REMOTE_PATH" = "./audio" ] && [ ! -d "./audio" ]; then
    mkdir -p ./audio
    echo "已创建本地音频目录: ./audio"
fi

# 无缓存策略:
# --vfs-cache-mode off: 关闭VFS缓存
# --buffer-size 32M: 大内存缓冲区
# --vfs-read-chunk-size 4M: 适配百度云
# --vfs-read-chunk-streams 6: 多流提高并发
# 无缓存相关参数
./rclone serve http "$REMOTE_PATH" \
    --addr "$SERVER_ADDR" \
    --vfs-cache-mode off \
    --buffer-size 32M \
    --vfs-read-chunk-size 4M \
    --vfs-read-chunk-streams 6 \
    --dir-cache-time 15m \
    --no-modtime \
    --allow-origin "*" \
    --server-read-timeout 2h \
    --server-write-timeout 2h \
    --max-header-bytes 8192 \
    --vfs-read-wait 100ms \
    --vfs-write-wait 1s \
    --read-only \
    -v

echo ""
echo "rclone音频流媒体服务已停止"