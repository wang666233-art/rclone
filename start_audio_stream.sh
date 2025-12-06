#!/bin/bash

# 专为音频文件流式播放优化的rclone HTTP服务启动脚本
# 使用方法: ./start_audio_stream.sh [remote:path] [port]

# 默认参数
REMOTE_PATH=${1:-"./audio"}  # 默认使用本地audio目录
PORT=${2:-8080}              # 默认端口8080
SERVER_ADDR="0.0.0.0:$PORT"  # 监听所有网络接口

# 音频流媒体优化参数说明:
# --vfs-cache-mode full: 启用完整VFS缓存，支持随机访问和快速跳转
# --buffer-size 32M: 设置32MB内存缓冲区，提高seek操作响应速度
# --vfs-read-ahead 128M: 预读128MB数据，大幅减少网络延迟对播放的影响
# --vfs-read-chunk-size 4M: 4MB块大小，适配百度云限制并优化音频流读取
# --vfs-read-chunk-streams 8: 8个并行流，显著提高大文件读取和seek性能
# --dir-cache-time 1h: 目录缓存1小时，减少频繁的目录列表请求
# --no-modtime: 不读取修改时间，减少API调用提高性能
# --allow-origin "*": 允许跨域访问，支持Web播放器
# --server-read-timeout 4h: 服务器读取超时4小时，支持长音频播放
# --server-write-timeout 4h: 服务器写入超时4小时
# --max-header-bytes 16384: 增加HTTP头部大小限制到16KB
# --vfs-cache-max-size 5G: 最大缓存5GB，可缓存多个音频文件
# --vfs-cache-max-age 48h: 缓存文件48小时，减少重复下载
# --vfs-write-back 30s: 写回延迟30秒，减少频繁的磁盘写入
# --vfs-read-wait 200ms: 读取等待时间200ms，优化seek响应
# --vfs-write-wait 2s: 写入等待时间2秒，提高写入稳定性

echo "启动rclone音频流媒体服务..."
echo "远程路径: $REMOTE_PATH"
echo "监听地址: $SERVER_ADDR"
echo "优化参数: 音频流媒体优化"
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
    --buffer-size 8M \
    --vfs-read-ahead 32M \
    --vfs-read-chunk-size 4M \
    --vfs-read-chunk-streams 2 \
    --dir-cache-time 2h \
    --no-modtime \
    --allow-origin "*" \
    --server-read-timeout 6h \
    --server-write-timeout 6h \
    --max-header-bytes 32768 \
    --vfs-cache-max-size 10G \
    --vfs-cache-max-age 72h \
    --vfs-write-back 60s \
    --vfs-read-wait 500ms \
    --vfs-write-wait 5s \
    --read-only \
    -v

echo ""
echo "rclone音频流媒体服务已停止"