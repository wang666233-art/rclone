#!/bin/bash
# 真正的流式播放脚本

FILENAME="$1"

if [ -z "$FILENAME" ]; then
    echo "用法: $0 <文件名>"
    echo "示例: $0 goldfallen.mp3"
    echo ""
    echo "或者启动服务器后手动播放:"
    echo "  ./stream_server.sh  (在另一个终端)"
    echo "  mpv http://localhost:8080/文件名.mp3"
    exit 1
fi

# 检查服务器是否已运行
if ! curl -s http://localhost:8080/ >/dev/null 2>&1; then
    echo "启动 rclone HTTP 服务器..."
    ./rclone serve http mybaidupan: --addr localhost:8080 --read-only >/dev/null 2>&1 &
    SERVER_PID=$!
    echo "服务器已启动 (PID: $SERVER_PID)"
    sleep 2
    STARTED_SERVER=1
fi

echo "正在流式播放: $FILENAME"
echo "可以快进/快退，不需要等待完整下载！"
mpv "http://localhost:8080/$FILENAME"

# 如果是我们启动的服务器，询问是否关闭
if [ "$STARTED_SERVER" = "1" ]; then
    read -p "是否关闭服务器? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kill $SERVER_PID 2>/dev/null
        echo "服务器已关闭"
    else
        echo "服务器继续运行 (PID: $SERVER_PID)"
        echo "要停止服务器: kill $SERVER_PID"
    fi
fi
