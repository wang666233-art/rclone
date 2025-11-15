#!/bin/bash

# 内存优先策略 - 最小磁盘占用
# 使用方法: ./start_audio_stream_memory.sh [remote:path] [port]

REMOTE_PATH=${1:-"./audio"}
PORT=${2:-8080}
SERVER_ADDR="0.0.0.0:$PORT"

echo "启动rclone音频流媒体服务 (内存优先策略)..."
echo "远程路径: $REMOTE_PATH"
echo "监听地址: $SERVER_ADDR"
echo "策略: 内存缓存 + 优化网络，最小磁盘占用"
echo ""

# 创建音频目录（如果不存在）
if [ "$REMOTE_PATH" = "./audio" ] && [ ! -d "./audio" ]; then
    mkdir -p ./audio
    echo "已创建本地音频目录: ./audio"
fi

# 内存优先策略:
# --vfs-cache-mode minimal: 最小缓存模式
# --buffer-size 16M: 大内存缓冲区
# --vfs-read-ahead 4M: 最小预读
# --vfs-read-chunk-size 4M: 适配百度云
# --vfs-read-chunk-streams 4: 多流提高性能
# --vfs-cache-max-size 1G: 极小缓存
# --vfs-cache-max-age 1h: 短缓存时间
./rclone serve http "$REMOTE_PATH" \
    --addr "$SERVER_ADDR" \
    --vfs-cache-mode minimal \
    --buffer-size 16M \
    --vfs-read-ahead 4M \
    --vfs-read-chunk-size 4M \
    --vfs-read-chunk-streams 4 \
    --dir-cache-time 30m \
    --no-modtime \
    --allow-origin "*" \
    --server-read-timeout 4h \
    --server-write-timeout 4h \
    --max-header-bytes 16384 \
    --vfs-cache-max-size 1G \
    --vfs-cache-max-age 1h \
    --vfs-write-back 30s \
    --vfs-read-wait 500ms \
    --vfs-write-wait 2s \
    --vfs-fast-fingerprint \
    --read-only \
    -v

echo ""
echo "rclone音频流媒体服务已停止"