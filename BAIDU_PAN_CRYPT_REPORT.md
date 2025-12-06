# 百度网盘 + rclone crypt 集成报告

本文档总结了在 `rclone` 中接入百度网盘（Baidu Pan）并叠加 `crypt` 加密层的完整流程，并给出在 Node.js 项目中直接调用 `crypt:` 远程的实践方案，避免通过 WebDAV 等额外协议中转。

---

## 1. 架构概览

- `baidupan` 后端：通过百度开放平台 API（OAuth 2.0 授权码 / 设备码）提供对象存储访问。
- `crypt` 后端：在任意远程之上透明加/解密文件名与文件内容。
- 调用路径：`Node.js 应用 → （嵌入式 librclone FFI / 自建 Go 服务） → crypt: → mybaidupan: → 百度网盘`，无需通过独立的 rclone 命令或 WebDAV。

扩展要点：

- 所有 API 请求需要显式传递 `access_token`（已在后端实现）。
- 百度网盘下载接口要求 `User-Agent: pan.baidu.com`，并必须允许重定向。
- 百度网盘强制分片大小 4 MiB，rclone 自动调整并给出警告提示属于正常。

---

## 2. 环境准备

1. **编译/安装 rclone**
   ```bash
   cd /home/wangkun/rclone
   go build -v
   ```
   > 如需代理：`go env -w GOPROXY=https://goproxy.cn,direct`

2. **确认二进制**
   ```bash
   ./rclone version
   ```

3. **确保 OAuth 凭据可用**
   - `client_id`, `client_secret`
   - 已在百度开放平台设置回调地址 `http://127.0.0.1:53682/`

---

## 3. 创建远程配置

### 3.1 Baidu Netdisk 远程

```bash
./rclone config
# name: mybaidupan
# type: baidupan
# auth_flow: 1 (授权码) 或 2 (设备码)
```

常见问题：
- **WSL 无法自动开浏览器**：选择 `n`，手动复制终端给出的 URL。
- **Invalid Bduss**：确保回调完成后，所有 API 请求 URL 都附带 `access_token`（该后端已修复）。

### 3.2 crypt 远程

假设加密内容存放在 `mybaidupan:/加密`：

```bash
./rclone config
# name: crypt
# type: crypt
# remote: mybaidupan:/加密
# password/password2: 建议使用 rclone 自动加密（输入明文后确认）
```

可选项：

- `Filename encryption`: `standard`（默认，推荐）或 `off/obfuscate`
- `Directory name encryption`: `true`（默认）
- `Password2`: 盐值，加密目录名时建议设置
- `--crypt-strict-names=false`：忽略无法解密的对象，避免遍历被中断

---

## 4. 基础操作速查

### 4.1 列表 & JSON API

```bash
# 人类可读
./rclone lsd crypt:

# 机器友好 (Node.js 建议解析该 JSON)
./rclone lsjson crypt: --max-depth 1
```

### 4.2 单文件读取

```bash
# 输出到标准输出，可被 Node.js 进程读取
./rclone cat "crypt:歌词/track01.flac"
```

### 4.3 文件传输

```bash
# 下载到本地
./rclone copy "crypt:lyrics/track01.flac" ./downloads

# 上传
./rclone copy ./local-folder crypt:lyrics
```

所有命令都可辅以 `-P` 查看进度，`--transfers` 调整并发。

---

## 5. Node.js 无缝流式播放方案（唯一推荐）

> 目标重申：在 Node.js 项目中，以最高性能流式播放（含快进/快退） `crypt:` 上的百度网盘音频及其他大文件，不依赖 WebDAV 或命令行工具链。

整个方案由两部分组成：

1. **Go Streaming Gateway（核心）**：直接复用 rclone 的 `crypt + baidupan` 实现，提供精简 HTTP API。
2. **Node.js 客户端**：调用 Gateway，获得目录、元数据与 Range 流，实现播放器或业务逻辑。

此架构兼顾了集成度、性能与可维护性，其余方案仅作为参考，不再展开。

### 5.1 Gateway 设计概览

- **职责**：封装 rclone 逻辑，暴露 JSON + Range Streaming 接口；负责 OAuth、刷新令牌、错误重试、缓存。
- **核心依赖**：`github.com/rclone/rclone/fs` 及 `backend/baidupan`、`backend/crypt`。
- **运行方式**：独立进程/容器，或与 Node.js 共用宿主机；监听本地端口（如 `localhost:9000`）。
- **API 约定**：
  - `GET /dir?path=<path>` → 返回目录与文件（类似 `lsjson`）
  - `GET /meta?path=<path>` → 返回单个文件的尺寸、修改时间、哈希
  - `GET /stream?path=<path>` → 支持 `Range` 的 206 响应，直接推送加解密后的原始字节
  - `POST /upload`、`DELETE /file` 等按需扩展

只要确保这些接口覆盖原有 rclone 功能点（列目录、读取、写入、删除等），Node.js 端即可无缝替换。

### 5.2 Go 实现要点

```go
package gateway

import (
  "context"
  "encoding/json"
  "net/http"

  "github.com/rclone/rclone/fs"
  "github.com/rclone/rclone/fs/operations"
  _ "github.com/rclone/rclone/backend/baidupan"
  _ "github.com/rclone/rclone/backend/crypt"
)

type Server struct {
  crypt fs.Fs
}

func NewServer() *Server {
  ctx := context.Background()
  f, err := fs.NewFs(ctx, "crypt:") // 重用 rclone 配置
  if err != nil {
    panic(err)
  }
  return &Server{crypt: f}
}

// 目录列表
func (s *Server) Dir(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  remote := r.URL.Query().Get("path")

  var items []*operations.ListJSONItem
  err := operations.ListJSON(ctx, s.crypt, remote, &operations.ListJSONOpt{}, func(item *operations.ListJSONItem) error {
    items = append(items, item)
    return nil
  })
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(items)
}
```

#### Range 流式读取

```go
func (s *Server) Stream(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  remote := r.URL.Query().Get("path")

  obj, err := s.crypt.NewObject(ctx, remote)
  if err != nil {
    http.Error(w, err.Error(), http.StatusNotFound)
    return
  }

  // 解析 Range 头
  start, end, err := parseRange(r.Header.Get("Range"), obj.Size())
  if err != nil {
    http.Error(w, err.Error(), http.StatusRequestedRangeNotSatisfiable)
    return
  }

  opts := []fs.OpenOption{}
  if start != 0 || end != obj.Size()-1 {
    opts = append(opts, &fs.RangeOption{Start: start, End: end})
  }

  rc, err := obj.Open(ctx, opts...)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  defer rc.Close()

  // 设置响应头
  statusCode := http.StatusOK
  contentLength := obj.Size()
  if len(opts) > 0 {
    statusCode = http.StatusPartialContent
    contentLength = end - start + 1
    w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, obj.Size()))
  }
  w.Header().Set("Accept-Ranges", "bytes")
  w.Header().Set("Content-Type", detectMime(remote))
  w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
  w.WriteHeader(statusCode)

  // 直接把解密后的数据写给客户端
  if _, err := io.Copy(w, rc); err != nil {
    fs.Errorf(obj, "stream aborted: %v", err)
  }
}
```

`parseRange`、`detectMime` 可使用标准库或现有实现（例如参考 `net/http` 和 `mime` 包）。

#### 上传 / 秒传

可复用 `operations.Copy`、`operations.Move` 等 API。对于大文件上传，按照 rclone 的分片逻辑执行：

```go
func (s *Server) Upload(w http.ResponseWriter, r *http.Request) {
  // 将请求体流式写入临时对象，然后调用 operations.Copy 或直接使用 Put/Update
}
```

### 5.3 Node.js 客户端示例

```javascript
// gateway-client.js
import fetch from "node-fetch";

const BASE = "http://localhost:9000";

export async function listDir(path = "") {
  const res = await fetch(`${BASE}/dir?path=${encodeURIComponent(path)}`);
  if (!res.ok) throw new Error(`dir failed: ${res.status}`);
  return res.json();
}

export async function fetchMeta(path) {
  const res = await fetch(`${BASE}/meta?path=${encodeURIComponent(path)}`);
  if (!res.ok) throw new Error(`meta failed: ${res.status}`);
  return res.json();
}

export async function streamFile(path, { start, end } = {}) {
  const headers = {};
  if (start !== undefined || end !== undefined) {
    headers.Range = `bytes=${start ?? ""}-${end ?? ""}`;
  }
  const res = await fetch(`${BASE}/stream?path=${encodeURIComponent(path)}`, { headers });
  if (!(res.ok || res.status === 206)) throw new Error(`stream failed: ${res.status}`);
  return res.body; // Node.js ReadableStream，可直接 pipe 给播放器
}
```

- **播放音频**：`streamFile("music/track.flac", { start: position })` → pipe 到 `mpv`, `ffmpeg`, `HLS` 处理器或 Browser `ReadableStream`.
- **快进/快退**：播放器根据时长换算字节位置，发起新的 Range 请求即可。
- **缓存策略**：Node.js 可自行处理，仅当前端 or 服务端缓存分块。

### 5.4 性能与上线建议

- **多线程下载**：rclone 支持多线程 `Open`，在 Gateway 中可按文件大小启用 `fs.MultiThreadOption` 优化高码率视频。
- **连接池**：重用 Gateway → 百度网盘 HTTP 客户端，开启 Keep-Alive。
- **日志与监控**：使用 `fs.Debugf`、`fs.Logf` 并输出到结构化日志；对接 Prometheus/StatsD。
- **容错**：Baidu API 可能返回 302/403；Gateway 应复用 rclone 的 `pacer` 与 `shouldRetry` 逻辑。
- **安全**：对 Gateway 的 API 添加鉴权（JWT、API Key），防止未授权访问你的网盘。
- **部署**：推荐容器化（Docker），通过 `ENV RCLONE_CONFIG=/data/rclone.conf` 挂载配置；也可 systemd 服务。

> 如未来确需单进程嵌入，可参考 `librclone` FFI；此处不再展开，以避免分散重点。

---

## 6. 性能与限制

- **块大小**：百度网盘要求 4 MiB 分片；警告信息可忽略。
- **速率限制**：后端已实现节流与重试 (`pacer`)；大量并发时建议调低 `--transfers`。
- **范围读取**：`rclone cat` 支持 `--offset` 和 `--count`，可自行在 Node.js 中构建 Range 逻辑。
- **文件名兼容**：如遇 `illegal base32 data`，请确认 crypt 密钥一致或关闭严格校验。

---

## 7. 排错辅助

| 症状 | 可能原因 | 处理建议 |
|------|----------|----------|
| `Invalid Bduss` | access_token 缺失/过期 | 重新 `rclone config reconnect mybaidupan:` |
| `Skipping undecryptable dir name` | 密码/盐不匹配或非 crypt 数据 | 校验配置或使用 `--crypt-strict-names=false` |
| `HTTP error 302` | 下载未跟随重定向 | 确保后端使用默认 `http.Client` 跟随重定向 |
| `Failed to recognize file format` | 播放器直接消费 `cat` 输出 | 使用 `rclone serve http` 或 `mpv <(rclone cat ...)` |

---

## 8. 推荐工作流总结

1. `./rclone config` 完成 `mybaidupan` 与 `crypt` 配置；
2. Node.js 以“方案 A”或“方案 B”调用 `crypt:`，读取/写入加密文件；
3. 如需流式在线播放，仍可临时使用 `./stream_server.sh`（HTTP）；
4. 定期运行 `./rclone about mybaidupan:` 获取容量，`./rclone cryptcheck` 校验加密数据完整性。

---

## 9. 参考资料

- `docs/content/rc.md`：RC API 文档
- `docs/content/commands/rclone_lsjson.md`
- `docs/content/commands/rclone_cat.md`
- 百度开放平台文档：<https://pan.baidu.com/union/doc/nksg0sbfs>
- `BAIDU_PAN_USAGE.md`：本仓库内基础用法指南

如需进一步自动化脚本，可在此基础上扩展任务队列、缓存和错误重试逻辑。欢迎根据业务需求调整上述示例。


