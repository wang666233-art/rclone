#!/bin/bash
# 从百度网盘流式播放音频
# 用法: ./play_audio.sh <filename>

FILENAME="$1"
if [ -z "$FILENAME" ]; then
    echo "Usage: $0 <filename>"
    echo "Example: $0 goldfallen.mp3"
    exit 1
fi

# 通过 rclone copy 快速下载并用管道传给 mpv
./rclone copy "mybaidupan:" /tmp/rclone_stream --include "$FILENAME" --max-depth 1 --transfers 1 2>/dev/null && mpv "/tmp/rclone_stream/$FILENAME"

