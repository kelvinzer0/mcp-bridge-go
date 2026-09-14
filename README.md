# mcp-bridge-go

Self-hosted Model Context Protocol (MCP) bridge daemon written in Go.
Provides Server-Sent Events (SSE) and JSON-RPC 2.0 transport for tool execution between local environments and browser extensions.

## Architecture & Security

- Strictly binds to `127.0.0.1` by default. Binding to global addresses (`0.0.0.0`) is prohibited to prevent exposure without a secured reverse proxy.
- Implements MCP specification `protocolVersion: 2025-03-26`.
- Thread-safe in-memory room management for session isolation.

## Endpoints

- `GET /new`: Allocates an isolated room ID and generates WebSocket/SSE endpoints.
- `GET /health?room=<room_id>`: Returns room status and count of active tools.
- `GET /ws/extension?room=<room_id>`: WebSocket endpoint for extension communication.
- `GET /mcp?room=<room_id>`: Server-Sent Events stream delivering endpoint metadata and keepalive heartbeats.
- `POST /mcp?room=<room_id>`: JSON-RPC 2.0 endpoint (`initialize`, `tools/list`, `tools/call`, `ping`).

## Installation on Linux

### Quick Install via Release Tarball

Download and extract the pre-built Linux package for your architecture:

```bash
tar -xvf mcp-bridge-linux-amd64.tar.gz
cd mcp-bridge-linux-amd64
sudo ./install.sh
```

The installer performs the following actions:
1. Installs binary to `/usr/local/bin/mcp-bridge`.
2. Creates default configuration at `/etc/mcp-bridge/mcp-bridge.env`.
3. Registers and starts the service under `systemd` (or `init.d` fallback).

### Service Management

Manage the service using standard Linux commands:

```bash
service mcp-bridge start
service mcp-bridge stop
service mcp-bridge restart
service mcp-bridge status
```

Or via `systemctl`:

```bash
systemctl start mcp-bridge
systemctl status mcp-bridge
```

### Configuration

Edit `/etc/mcp-bridge/mcp-bridge.env`:

```ini
HOST=127.0.0.1
PORT=8080
```

Restart the service after editing:

```bash
service mcp-bridge restart
```

### Uninstallation

```bash
sudo ./uninstall.sh
```

## Building from Source

Requirements: Go 1.22+

```bash
git clone https://github.com/kelvinzer0/mcp-bridge-go.git
cd mcp-bridge-go
go build -trimpath -ldflags="-s -w" -o bin/mcp-bridge ./cmd/mcp-bridge
./bin/mcp-bridge -host 127.0.0.1 -port 8080
```
