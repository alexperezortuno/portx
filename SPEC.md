# SPEC.md

# PortX Functional Specification

## Vision

PortX provides a unified CLI for securely exposing local services using
multiple tunnel providers.

## Functional Requirements

1.  CLI commands
2.  Provider abstraction
3.  Provider registry
4.  Config management
5.  Logging
6.  Tunnel lifecycle
7.  Dashboard
8.  REST API
9.  Plugin system
10. Auto-update

## Providers

-   Cloudflare
-   SSH Reverse Tunnel
-   FRP
-   Tailscale Funnel
-   Native PortXD

## Commands

-   portx expose
-   portx stop
-   portx list
-   portx doctor
-   portx config
-   portx provider
-   portx dashboard
-   portx version

## Non Functional

-   Go 1.24+
-   Linux/macOS/Windows
-   ARM64/AMD64
-   Apache-2.0
-   High test coverage
-   Structured logging
-   Graceful shutdown
