#!/bin/bash

# 专为音频文件流式播放优化的rclone HTTP服务启动脚本 (稳定版)
# 使用方法: ./start_audio_stream_stable.sh [remote:path] [port]

unset http_proxy
unset https_proxy
unset all_proxy
unset HTTP_PROXY
unset HTTPS_PROXY
unset ALL_PROXY

# 默认参数
REMOTE_PATH=${1:-"./audio"}  # 默认使用本地audio目录
PORT=${2:-8080}              # 默认端口8080
SERVER_ADDR="0.0.0.0:$PORT"  # 监听所有网络接口

# 音频流媒体稳定优化参数说明:
# --vfs-cache-mode full: 启用完整VFS缓存，这是解决seek问题的关键
# --buffer-size 4M: 较小的4MB缓冲区，减少内存压力
# --vfs-read-ahead 16M: 适中的预读，避免过度预读导致的问题
# --vfs-read-chunk-size 4M: 4MB块大小，适配百度云限制
# --vfs-read-chunk-streams 1: 单线程流，避免并发导致的连接问题
# --dir-cache-time 2h: 目录缓存2小时
# --no-modtime: 不读取修改时间，减少API调用
# --allow-origin "*": 允许跨域访问
# --server-read-timeout 12h: 超长的读取超时，确保长音频播放
# --server-write-timeout 12h: 超长的写入超时
# --max-header-bytes 65536: 大幅增加HTTP头部限制
# --vfs-cache-max-size 20G: 大缓存，确保整个音频文件可以被缓存
# --vfs-cache-max-age 168h: 7天缓存，减少重复下载
# --vfs-write-back 120s: 延长写回时间
# --vfs-read-wait 1s: 增加读取等待时间，给seek操作更多时间
# --vfs-write-wait 10s: 增加写入等待时间
# --vfs-fast-fingerprint: 使用快速指纹，减少文件检查时间

echo "启动rclone音频流媒体服务 (稳定版)..."
echo "远程路径: $REMOTE_PATH"
echo "监听地址: $SERVER_ADDR"
echo "优化策略: 稳定性优先，解决seek操作问题"
echo ""

# 创建音频目录（如果不存在）
if [ "$REMOTE_PATH" = "./audio" ] && [ ! -d "./audio" ]; then
    mkdir -p ./audio
    echo "已创建本地音频目录: ./audio"
    echo "请将音频文件放入此目录中"
fi

# 启动rclone HTTP服务
./rclone serve http "$REMOTE_PATH" \
    --addr "$SERVER_ADDR" \
    --vfs-cache-mode full \
    --buffer-size 4M \
    --vfs-read-ahead 16M \
    --vfs-read-chunk-size 4M \
    --vfs-read-chunk-streams 1 \
    --dir-cache-time 2h \
    --no-modtime \
    --allow-origin "*" \
    --server-read-timeout 12h \
    --server-write-timeout 12h \
    --max-header-bytes 65536 \
    --vfs-cache-max-size 20G \
    --vfs-cache-max-age 168h \
    --vfs-write-back 120s \
    --vfs-read-wait 1s \
    --vfs-write-wait 10s \
    --vfs-fast-fingerprint \
    --read-only \
    -v

echo ""
echo "rclone音频流媒体服务已停止"
