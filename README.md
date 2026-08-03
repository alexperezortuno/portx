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

## Quick Start

### Install

```bash
go install github.com/alexperezortuno/portx@latest
```

Or build from source:

```bash
go build -o portx ./cmd/portx
```

### SSH Tunnel

```bash
portx expose --provider ssh \
  --ssh-user tunnel \
  --ssh-host example.com \
  --local-port 8080
```

### PortXD Local Tunnel

```bash
portx expose --provider portxd --local-port 8080
```

PortXD starts an embedded server on port 7222 by default. Connect clients to that port.

## Usage

```
portx expose --provider <name> [flags]
```

### Common Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--provider` | Tunnel provider (ssh, portxd) | Required |
| `--local-port` | Local service address (host:port) | Required |
| `--tunnel-port` | Tunnel address on remote (host:port) | - |
| `--ssh-user` | SSH username | - |
| `--ssh-host` | SSH server host | - |
| `--ssh-port` | SSH server port | 22 |
| `--ssh-password` | SSH password | - |
| `--ssh-private-key` | SSH private key content | - |
| `--portxd-port` | PortXD server port | 7222 |

### Commands

```
portx expose    Expose a local service through a tunnel
portx doctor    Check system requirements
portx version   Print version information
```

## Architecture

```
cmd/portx
  └── internal/
      ├── cli/           # Cobra commands
      ├── config/        # Viper configuration
      ├── logger/        # slog wrapper
      ├── provider/      # Provider interface + registry
      │   ├── ssh/       # SSH reverse tunnel
      │   └── portxd/   # PortXD embedded server
      ├── runtime/       # Composition root
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

## Configuration

PortX loads configuration from `~/.portx/config.yaml` or the path specified with `--config`.

Example `config.yaml`:

```yaml
log_level: info
provider: ssh
tunnels:
  - name: web
    provider: ssh
    local_addr: localhost:8080
    remote_addr: 0.0.0.0:10000
    ssh_user: tunnel
    ssh_host: example.com
    ssh_port: 22
```

## Requirements

- Go 1.24+

## Contributing

See CONTRIBUTING.md
