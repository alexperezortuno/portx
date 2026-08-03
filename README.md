# PortX

Provider-agnostic tunneling platform written in Go.

## Features

-   Multi-provider (SSH, Cloudflare, FRP, Tailscale, PortXD)
-   Single binary
-   Open Source
-   Cross-platform (Linux, macOS, Windows)
-   Extensible provider architecture
-   Structured logging with slog
-   Configuration via file or environment variables

## Providers

| Provider | Status |
|----------|--------|
| SSH | Implemented |
| PortXD | Implemented |
| Cloudflare | Planned |
| FRP | Planned |
| Tailscale | Planned |

## Install

```bash
go install github.com/alexperezortuno/portx@latest
```

Or build from source:

```bash
go build -o portx ./cmd/portx
```

## Quick Start

### Expose a Port

```bash
portx expose 3000
```

This exposes local port 3000 using the default provider (portxd).

### Expose with SSH

```bash
portx expose 3000 --provider ssh --ssh-host example.com --ssh-use-agent
```

### Expose a Named Service

Define services in `~/.portx/config.yaml`:

```bash
portx expose frontend
```

## Commands

| Command | Description |
|---------|-------------|
| `portx expose [target]` | Expose a local service through a tunnel |
| `portx list` | List active tunnels |
| `portx stop` | Stop a tunnel or all tunnels |
| `portx config` | Manage configuration |
| `portx doctor` | Check system requirements |
| `portx version` | Print version information |

### Expose

Target can be:
- **Port number**: `portx expose 3000`
- **Named service**: `portx expose frontend`

### Expose Flags

| Flag | Description | Default |
|------|-------------|---------|
| `target` | Port number or service name | Required |
| `--provider` | Tunnel provider (ssh, portxd, cloudflare) | portxd |
| `--local-addr` | Local service address (host:port) | localhost:{target} |
| `--hostname` | Hostname for the tunnel | - |
| `--ssh-host` | SSH server host | - |
| `--ssh-user` | SSH username | - |
| `--ssh-port` | SSH server port | 22 |
| `--ssh-password` | SSH password | - |
| `--ssh-private-key` | SSH private key content (PEM format) | - |
| `--ssh-use-agent` | Use SSH agent for authentication | false |
| `--portxd-port` | PortXD server port | 7222 |
| `--port` | Local port (deprecated: use positional argument) | - |

### List Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--output` | Output format (table, json) | table |

### Stop Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--all` | Stop all tunnels | false |

### Config Subcommands

| Command | Description |
|---------|-------------|
| `portx config view` | View resolved configuration |
| `portx config validate [path]` | Validate configuration file |

### Config View Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--output` | Output format (yaml, json) | yaml |

## Configuration

PortX loads configuration from `~/.portx/config.yaml` or the current directory.

Example `~/.portx/config.yaml`:

```yaml
version: 1

provider:
  default: portxd

storage:
  path: ~/.portx

services:
  frontend:
    port: 3000
    protocol: http
  api:
    port: 8080
    protocol: http
  postgres:
    port: 5432
    protocol: tcp

providers:
  cloudflare:
    enabled: true
  portxd:
    enabled: true
  ssh:
    enabled: true
```

**Security Note:** For SSH tunnels, prefer `--ssh-use-agent` or `--ssh-private-key` over `--ssh-password` when possible.

**Security Note:** For SSH tunnels, prefer `--ssh-private-key` over `--ssh-password` when possible.

## Architecture

```
cmd/portx
  └── internal/
      ├── cli/           # Cobra commands
      ├── config/        # Configuration (Viper)
      ├── logger/        # slog wrapper
      ├── provider/      # Provider interface + registry
      │   ├── ssh/       # SSH provider
      │   └── portxd/   # PortXD provider
      ├── runtime/       # Composition root
      ├── sshutil/       # Shared SSH utilities
      ├── target/        # Target resolution (port, service, docker, k8s)
      └── tunnel/        # Tunnel lifecycle
```

### Provider Interface

All providers implement the `Provider` interface:

```go
type Provider interface {
    Name() string
    Start(ctx context.Context, cfg TunnelConfig) error
    Stop(ctx context.Context) error
    Status(ctx context.Context) (Status, error)
}
```

Providers are registered in a central registry. The CLI never depends on concrete provider implementations.

### Target Resolution

PortX resolves targets in this order:

1. **Port**: If input is a number, treat as a port number
2. **Service**: If input matches a service name in config
3. **Docker**: If input matches a Docker container name
4. **Compose**: If input matches a Docker Compose service name
5. **Kubernetes**: If input matches a Kubernetes service name
6. **Error**: Unknown target

## Requirements

- Go 1.24+

## Contributing

See CONTRIBUTING.md
