# mcp-bridge-go 🚀

High-performance, lightweight self-hosted Model Context Protocol (MCP) SSE bridge server written in Go.  
Replaces `mcp-bridge-cf` (Cloudflare Workers) to eliminate tier limits and cloud dependency.

---

## ✨ Features

- **Protocol Compatible**: 100% compatible with Model Context Protocol (MCP) JSON-RPC 2.0 and Server-Sent Events (SSE) specifications.
- **Zero Cloud Cost**: Run on any VPS, Raspberry Pi, homelab, or local server.
- **Ultra Lightweight & Fast**: Built with Go standard library and Gorilla WebSocket, with minimal memory footprint (< 15MB RAM).
- **Multi-room Support**: Isolated rooms via `/new` or URL query `?room=<roomId>`.
- **Automatic Multi-Platform Releases**: GitHub Actions CI builds binaries for Linux, macOS, and Windows (amd64 and arm64).

---

## 📡 Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `GET` / `POST` | `/new` or `/mcp/new` | Generate a new isolated room ID with connection URLs |
| `GET` | `/health?room=<roomId>` | Health status and connected tools count |
| `GET` | `/ws/extension?room=<roomId>` | WebSocket connection for Chrome Extension |
| `GET` | `/mcp?room=<roomId>` | Server-Sent Events (SSE) stream for MCP clients |
| `POST` | `/mcp?room=<roomId>` | JSON-RPC 2.0 endpoint (`initialize`, `tools/list`, `tools/call`, `ping`) |

---

## 🛠️ Quick Start

### 1. Download Pre-built Binary
Download the binary for your OS and architecture from the [GitHub Releases](https://github.com/kelvinzer0/mcp-bridge-go/releases) page.

```bash
# Example for Linux amd64
tar -xvf mcp-bridge-go-linux-amd64.tar.gz
./mcp-bridge-go -port 8080
```

### 2. Build from Source
```bash
git clone https://github.com/kelvinzer0/mcp-bridge-go.git
cd mcp-bridge-go
go build -o mcp-bridge-go .
./mcp-bridge-go -port 8080
```

### 3. Run with Docker
```bash
docker build -t mcp-bridge-go .
docker run -d -p 8080:8080 --name mcp-bridge mcp-bridge-go
```

---

## ⚙️ Configuration & Flags

| Flag | Env Var | Default | Description |
|---|---|---|---|
| `-port` | `PORT` | `8080` | Port to listen on |
| `-host` | - | `0.0.0.0` | Host IP to bind to |

---

## 📦 CI / CD Automated Releases

Releases are automatically triggered when a Git tag starting with `v*` (e.g., `v1.0.0`) is pushed to GitHub:

```bash
git tag v1.0.0
git push origin v1.0.0
```
GitHub Actions will automatically cross-compile and attach release archives + SHA256 checksums to the release.
