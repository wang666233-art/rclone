#!/bin/bash
# 启动 rclone HTTP 服务器用于流式播放
# 用法: ./stream_server.sh

echo "Starting rclone HTTP server on http://localhost:8080"
echo "You can access files at: http://localhost:8080/"
echo "Example: mpv http://localhost:8080/goldfallen.mp3"
echo ""
echo "Press Ctrl+C to stop"

./rclone serve http crypt: --addr localhost:8080 --read-only
