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

### SSH Tunnel (with password)

```bash
portx expose --provider ssh \
  --ssh-user tunnel \
  --ssh-host example.com \
  --ssh-password "your-password" \
  --local-port 8080
```

### SSH Tunnel (with private key)

```bash
portx expose --provider ssh \
  --ssh-user tunnel \
  --ssh-host example.com \
  --ssh-private-key "$(cat ~/.ssh/id_ed25519)" \
  --local-port 8080
```

### SSH Tunnel (with SSH agent)

```bash
# Requires SSH_AUTH_SOCK to be set (e.g., from ssh-agent or keychain)
portx expose --provider ssh \
  --ssh-user tunnel \
  --ssh-host example.com \
  --ssh-use-agent \
  --local-port 8080
```

### PortXD Local Tunnel

```bash
portx expose --provider portxd --local-port 8080
```

PortXD starts an embedded server on port 7222 by default.

## Commands

| Command | Description |
|---------|-------------|
| `portx expose` | Expose a local service through a tunnel |
| `portx list` | List active tunnels |
| `portx stop` | Stop a tunnel or all tunnels |
| `portx config` | Manage configuration |
| `portx doctor` | Check system requirements |
| `portx version` | Print version information |

### Expose Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--provider` | Tunnel provider (ssh, portxd) | Required |
| `--local-port` | Local service address (host:port) | Required |
| `--tunnel-port` | Tunnel address on remote (host:port) | - |
| `--ssh-user` | SSH username | - |
| `--ssh-host` | SSH server host | - |
| `--ssh-port` | SSH server port | 22 |
| `--ssh-password` | SSH password | - |
| `--ssh-private-key` | SSH private key content (PEM format) | - |
| `--ssh-use-agent` | Use SSH agent for authentication | false |
| `--portxd-port` | PortXD server port | 7222 |

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
    ssh_private_key: |
      -----BEGIN OPENSSH PRIVATE KEY-----
      ...your private key content...
      -----END OPENSSH PRIVATE KEY-----
  - name: api
    provider: portxd
    local_addr: localhost:3000
    portxd_port: 7222
```

**Security Note:** For SSH tunnels, prefer `--ssh-private-key` over `--ssh-password` when possible.

## Architecture

```
cmd/portx
  └── internal/
      ├── cli/           # Cobra commands
      ├── config/        # Viper configuration
      ├── logger/        # slog wrapper
      ├── provider/      # Provider interface + registry
      │   ├── ssh/       # SSH provider
      │   └── portxd/   # PortXD provider
      ├── runtime/       # Composition root
      ├── ssh/           # Shared SSH utilities
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

## Requirements

- Go 1.24+

## Contributing

See CONTRIBUTING.md
