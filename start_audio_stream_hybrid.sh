#!/bin/bash

# 混合缓存策略 - 平衡性能和空间占用
# 使用方法: ./start_audio_stream_hybrid.sh [remote:path] [port]

REMOTE_PATH=${1:-"./audio"}
PORT=${2:-8080}
SERVER_ADDR="0.0.0.0:$PORT"

echo "启动rclone音频流媒体服务 (混合缓存策略)..."
echo "远程路径: $REMOTE_PATH"
echo "监听地址: $SERVER_ADDR"
echo "策略: 智能缓存 + 优化seek，减少磁盘占用"
echo ""

# 创建音频目录（如果不存在）
if [ "$REMOTE_PATH" = "./audio" ] && [ ! -d "./audio" ]; then
    mkdir -p ./audio
    echo "已创建本地音频目录: ./audio"
fi

# 混合策略参数:
# --vfs-cache-mode writes: 只缓存写入，读取直接从远程
# --buffer-size 8M: 适中的内存缓冲
# --vfs-read-ahead 8M: 较小的预读，减少浪费
# --vfs-read-chunk-size 4M: 适配百度云限制
# --vfs-read-chunk-streams 2: 双流，提高seek响应但不过度
# --vfs-cache-max-size 5G: 较小的缓存限制
# --vfs-cache-max-age 6h: 较短的缓存时间
# --vfs-read-wait 2s: 增加seek等待时间
./rclone serve http "$REMOTE_PATH" \
    --addr "$SERVER_ADDR" \
    --vfs-cache-mode writes \
    --buffer-size 8M \
    --vfs-read-ahead 8M \
    --vfs-read-chunk-size 4M \
    --vfs-read-chunk-streams 2 \
    --dir-cache-time 1h \
    --no-modtime \
    --allow-origin "*" \
    --server-read-timeout 6h \
    --server-write-timeout 6h \
    --max-header-bytes 32768 \
    --vfs-cache-max-size 5G \
    --vfs-cache-max-age 6h \
    --vfs-write-back 60s \
    --vfs-read-wait 2s \
    --vfs-write-wait 5s \
    --vfs-fast-fingerprint \
    --read-only \
    -v

echo ""
echo "rclone音频流媒体服务已停止"