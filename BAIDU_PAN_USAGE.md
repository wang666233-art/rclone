# 百度网盘 rclone 使用指南

## 🎵 音频播放方式对比

### 方式 1: 流式播放（推荐）✨

**特点:**
- ✅ **真正的流式播放** - 边下边播
- ✅ **支持快进/快退** - HTTP Range 请求
- ✅ **不占用本地存储** - 仅缓存正在播放的部分
- ✅ **即点即播** - 无需等待下载

**使用方法:**

```bash
# 方法 A: 一键播放（自动启动服务器）
./play_streaming.sh goldfallen.mp3

# 方法 B: 手动启动服务器（可播放多个文件）
./stream_server.sh  # 在一个终端启动
mpv http://localhost:8080/goldfallen.mp3  # 在另一个终端播放
mpv http://localhost:8080/REC001.mp3     # 播放其他文件
```

**工作原理:**
1. rclone 启动 HTTP 服务器
2. mpv 通过 HTTP 访问文件
3. 支持 Range 请求，可以快进
4. 仅下载播放需要的部分

---

### 方式 2: 下载后播放

**特点:**
- ❌ **不是流式** - 需要先下载完整文件
- ✅ **完全缓存** - 文件保存在 /tmp
- ✅ **可快进** - 本地文件
- ❌ **占用存储** - 需要完整文件大小的空间

**使用方法:**

```bash
./play_audio.sh goldfallen.mp3
```

**适用场景:**
- 网络不稳定时
- 想要离线保存文件
- 需要反复播放同一文件

---

## 📋 其他常用命令

### 列出文件
```bash
# 列出根目录所有文件
./rclone ls mybaidupan: --max-depth 1

# 只列出 MP3
./rclone ls mybaidupan: --include "*.mp3"

# 列出某个文件夹
./rclone ls mybaidupan:音乐
```

### 下载文件
```bash
# 下载单个文件
./rclone copy mybaidupan:song.mp3 /local/folder

# 下载整个文件夹
./rclone copy mybaidupan:音乐 /local/music

# 只下载 MP3
./rclone copy mybaidupan: /local/music --include "*.mp3"
```

### 上传文件
```bash
# 上传文件到根目录
./rclone copy /local/file.mp3 mybaidupan:

# 上传到指定文件夹
./rclone copy /local/music mybaidupan:我的音乐
```

### 查看空间
```bash
./rclone about mybaidupan:
```

---

## 🎯 推荐工作流

### 方案 1: 常驻 HTTP 服务器（最佳体验）

```bash
# 终端 1: 启动服务器
./stream_server.sh

# 终端 2: 随时播放任何文件
mpv http://localhost:8080/文件名.mp3

# 浏览器访问: http://localhost:8080/
# 可以看到所有文件列表！
```

### 方案 2: 使用 rclone mount（类似本地磁盘）

```bash
# 挂载网盘到本地目录
mkdir -p ~/baidupan
./rclone mount mybaidupan: ~/baidupan --daemon

# 直接访问文件（像本地文件一样）
mpv ~/baidupan/goldfallen.mp3
ls ~/baidupan/

# 卸载
fusermount -u ~/baidupan
```

---

## 🔧 技术细节

### HTTP 服务器端口
- 默认: `localhost:8080`
- 修改: `./rclone serve http mybaidupan: --addr localhost:端口号 --read-only`

### 缓存策略
- 流式播放: rclone 会缓存正在读取的块，播放器关闭后自动清理
- 下载播放: 文件保存在 /tmp，系统重启后清理

### Range 请求支持
- ✅ 支持 HTTP Range 请求
- ✅ 可以快进到任意位置
- ✅ 播放器可以显示总时长
- ✅ 不需要下载完整文件

---

## 📊 性能对比

| 特性 | 流式播放 | 下载后播放 |
|------|---------|-----------|
| 启动速度 | ⚡ 即时 | 🐌 需等待下载 |
| 存储占用 | ✅ 最小 | ❌ 完整文件 |
| 网络使用 | ✅ 按需 | ❌ 完整下载 |
| 快进支持 | ✅ Range | ✅ 本地文件 |
| 离线播放 | ❌ 需网络 | ✅ 可离线 |
| 重复播放 | 🔄 重新请求 | ✅ 无需重下 |

**推荐:** 日常使用流式播放，重要文件下载后保存。
