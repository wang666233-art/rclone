# Baidu Pan (百度网盘) Backend for Rclone

This backend provides rclone support for Baidu Pan (Baidu Netdisk).

## Features

### Core Features
- ✅ OAuth 2.0 authentication (Authorization Code flow and Device Code flow)
- ✅ List files and directories
- ✅ Upload files (with chunked upload for large files)
- ✅ Download files (supports Range requests for streaming)
- ✅ Delete files and directories
- ✅ Create directories
- ✅ Get file metadata (size, modification time, MD5 hash)

### Advanced Features
- ✅ Server-side Move operations
- ✅ Server-side Copy operations
- ✅ Get quota information (About)
- ✅ Generate public share links
- ✅ Rapid upload (秒传) - if file exists on Baidu Pan, skip upload
- ✅ MD5 hash support

### Media File Support
- ✅ **Audio file streaming** - Open() method supports HTTP Range requests for audio playback
- ✅ **Video file streaming** - Same Range support for video streaming
- ✅ **Direct file access** - Can be mounted with `rclone mount` for direct media playback

## Configuration

### Prerequisites
1. Register at [Baidu Pan Open Platform](https://pan.baidu.com/union/doc/nksg0sbfs)
2. Create an application to get your AppKey (Client ID) and AppSecret (Client Secret)
3. Configure OAuth redirect URL in your app settings

### Setup

Run `rclone config` and choose:
- Type: `baidupan`
- client_id: Your Baidu Pan AppKey
- client_secret: Your Baidu Pan AppSecret
- auth_flow: Choose between:
  - `code` (recommended) - Authorization code flow, requires browser
  - `device` - Device code flow for headless environments

Example configuration:
```
[mybaidupan]
type = baidupan
client_id = YOUR_APP_KEY
client_secret = YOUR_APP_SECRET
auth_flow = code
```

## Usage Examples

### List files
```bash
rclone ls mybaidupan:
rclone ls mybaidupan:path/to/folder
```

### Upload files
```bash
rclone copy local/file.txt mybaidupan:remote/path/
```

### Download files
```bash
rclone copy mybaidupan:remote/file.txt local/path/
```

### Mount for media playback
```bash
# Mount Baidu Pan
rclone mount mybaidupan: /mnt/baidupan

# Now you can play audio/video files directly
vlc /mnt/baidupan/music/song.mp3
mpv /mnt/baidupan/videos/movie.mp4
```

### Sync directories
```bash
rclone sync local/folder mybaidupan:backup/
```

### Get quota information
```bash
rclone about mybaidupan:
```

### Generate share link
```bash
rclone link mybaidupan:path/to/file.txt
```

## Audio/Video Streaming Support

The backend fully supports streaming audio and video files:

1. **Range Requests**: The `Open()` method properly handles HTTP Range headers, allowing:
   - Audio players to seek to any position in the track
   - Video players to skip forward/backward
   - Progressive download while playing

2. **Mount and Play**: After mounting with `rclone mount`, any media player can directly access files:
   ```bash
   rclone mount mybaidupan: /mnt/baidupan
   
   # Play audio
   vlc /mnt/baidupan/Music/song.mp3
   mpv /mnt/baidupan/Music/album/
   
   # Play video
   vlc /mnt/baidupan/Videos/movie.mp4
   ```

3. **Supported Formats**: All formats supported by Baidu Pan:
   - Audio: MP3, FLAC, WAV, AAC, APE, etc.
   - Video: MP4, AVI, MKV, MOV, etc.

## Technical Details

### Upload Process
1. Calculate file MD5 and block MD5s (4MB chunks)
2. Call precreate API (enables rapid upload if file exists)
3. Upload only required blocks (if not rapid upload)
4. Call create API to finalize the file

### Chunk Size
- Fixed at 4MB (百度网盘要求)
- Cannot be changed as it's required by Baidu Pan API

### Authentication
- Uses OAuth 2.0 with automatic token refresh
- Supports both authorization code flow and device code flow
- Tokens stored securely in rclone config

### API Endpoints
- Main API: `https://pan.baidu.com`
- Upload API: `https://d.pcs.baidu.com`
- OAuth: `https://openapi.baidu.com`

## Limitations

1. **Modification Time**: Cannot set custom modification times (Baidu Pan limitation)
2. **Chunk Size**: Fixed at 4MB, cannot be changed
3. **Rate Limiting**: Subject to Baidu Pan API rate limits

## Error Handling

The backend handles common errors:
- Automatic retry on network errors
- Token refresh on expiration
- Rate limit handling (429 errors)
- Proper error mapping to rclone error types

## Development

### File Structure
```
backend/baidupan/
├── api/
│   └── types.go       # API data structures
├── baidupan.go        # Main implementation (Fs and Object interfaces)
├── upload.go          # Upload logic with chunking
├── baidupan_test.go   # Tests
└── README.md          # This file
```

### Building
```bash
cd /path/to/rclone
go build
```

### Testing
```bash
# Configure a test remote first
rclone config

# Run tests
go test ./backend/baidupan/...
```

## Links

- [Baidu Pan Open Platform](https://pan.baidu.com/union/doc/nksg0sbfs)
- [Rclone Documentation](https://rclone.org/docs/)
- [OAuth 2.0 Specification](https://oauth.net/2/)

## License

Same as rclone (MIT License)









