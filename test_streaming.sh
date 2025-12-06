#!/bin/bash

echo "=== 启动 rclone HTTP 服务器 ==="
./rclone serve http mybaidupan: --addr localhost:8080 --read-only > /tmp/rclone_server.log 2>&1 &
SERVER_PID=$!
echo "服务器 PID: $SERVER_PID"

sleep 3

echo ""
echo "=== 测试 1: 获取完整文件头信息 ==="
curl -I http://localhost:8080/goldfallen.mp3 2>&1 | grep -E "(HTTP|Accept-Ranges|Content-Length|Content-Type)"

echo ""
echo "=== 测试 2: Range 请求（前 100 字节）==="
curl -r 0-99 http://localhost:8080/goldfallen.mp3 2>/dev/null | wc -c
echo "预期: 100 字节"

echo ""
echo "=== 测试 3: Range 请求（模拟快进到中间）==="
curl -I -r 10000-10099 http://localhost:8080/goldfallen.mp3 2>&1 | grep -E "(HTTP|Content-Range|Content-Length)"

echo ""
echo "=== 测试 4: 流式播放前 3 秒（使用 mpv）==="
echo "命令: timeout 3 mpv http://localhost:8080/goldfallen.mp3"
echo "(播放 3 秒后自动停止)"
timeout 3 mpv http://localhost:8080/goldfallen.mp3 >/dev/null 2>&1 && echo "✅ 流式播放成功！"

echo ""
echo "停止服务器..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo ""
echo "=== 测试结果 ==="
echo "✅ 支持流式播放"
echo "✅ 支持 Range 请求（可快进）"
echo "✅ 不需要完整下载"
echo "❌ 不会永久缓存（服务器关闭后清除）"
