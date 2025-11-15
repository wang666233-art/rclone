# rclone音频流媒体服务策略对比

## 策略总结

| 策略 | 磁盘占用 | 内存占用 | Seek性能 | 首次加载 | 适用场景 |
|------|----------|----------|-----------|----------|----------|
| **稳定版** (full) | 高 (20GB) | 低 (4MB) | ⭐⭐⭐⭐⭐ | 慢 | 频繁seek，磁盘充足 |
| **混合版** (writes) | 中 (5GB) | 低 (8MB) | ⭐⭐⭐⭐ | 中等 | 平衡性能和空间 |
| **内存版** (minimal) | 低 (1GB) | 中 (16MB) | ⭐⭐⭐ | 快 | 内存充足，磁盘紧张 |
| **无缓存版** (off) | 无 | 高 (32MB) | ⭐⭐ | 最快 | 磁盘极小，只顺序播放 |

## 详细参数对比

### 稳定版 (start_audio_stream_stable.sh)
```bash
--vfs-cache-mode full          # 完整缓存
--vfs-cache-max-size 20G      # 大缓存
--vfs-cache-max-age 168h      # 7天缓存
--buffer-size 4M             # 小内存缓冲
--vfs-read-wait 1s           # 长等待时间
```
**优点**: Seek性能最佳，连接稳定
**缺点**: 占用大量磁盘空间

### 混合版 (start_audio_stream_hybrid.sh)
```bash
--vfs-cache-mode writes       # 只缓存写入
--vfs-cache-max-size 5G      # 中等缓存
--vfs-cache-max-age 6h       # 6小时缓存
--buffer-size 8M             # 中等内存缓冲
--vfs-read-wait 2s           # 中等等待时间
```
**优点**: 平衡了性能和空间占用
**缺点**: Seek性能略低于完整缓存

### 内存版 (start_audio_stream_memory.sh)
```bash
--vfs-cache-mode minimal     # 最小缓存
--vfs-cache-max-size 1G      # 小缓存
--vfs-cache-max-age 1h       # 1小时缓存
--buffer-size 16M            # 大内存缓冲
--vfs-read-wait 500ms        # 短等待时间
```
**优点**: 磁盘占用小，响应较快
**缺点**: 内存占用较高，Seek性能一般

### 无缓存版 (start_audio_stream_nocache.sh)
```bash
--vfs-cache-mode off         # 无缓存
--buffer-size 32M            # 最大内存缓冲
--vfs-read-chunk-streams 6   # 多流并发
--vfs-read-wait 100ms        # 最短等待时间
```
**优点**: 零磁盘占用，启动最快
**缺点**: Seek性能差，适合顺序播放

## 推荐使用场景

### 🎯 如果你：
- 有充足磁盘空间 (>20GB)
- 经常快进/快退音频
- 希望最稳定的播放体验
**→ 使用稳定版**

### ⚖️ 如果你：
- 磁盘空间有限 (5-10GB)
- 偶尔需要seek操作
- 希望平衡性能和空间
**→ 使用混合版**

### 💾 如果你：
- 磁盘空间很小 (<2GB)
- 内存充足 (>4GB)
- 主要顺序播放
**→ 使用内存版**

### 🚀 如果你：
- 磁盘空间极小
- 只顺序播放音频
- 可以接受偶尔的卡顿
**→ 使用无缓存版**

## 使用方法

```bash
# 停止当前服务
pkill -f "rclone serve http"

# 启动不同策略的服务
./start_audio_stream_stable.sh crypt: 9999    # 稳定版
./start_audio_stream_hybrid.sh crypt: 9999    # 混合版  
./start_audio_stream_memory.sh crypt: 9999    # 内存版
./start_audio_stream_nocache.sh crypt: 9999    # 无缓存版
```

## 监控缓存使用

查看当前缓存使用情况：
```bash
# 查看rclone进程的缓存文件
du -sh ~/.cache/rclone/vfs/

# 或者查看服务日志中的缓存信息
# 日志会显示: "total size 1.452Gi"
```