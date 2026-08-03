# ProxyAI Instructions

# PROJECT

Project Name

PortX

PortX is an Open Source provider-agnostic tunneling platform written in Go.

The goal is to provide a single CLI capable of exposing local services through different providers.

Examples:

Cloudflare Tunnel

SSH Reverse Tunnel

FRP

Tailscale Funnel

PortXD (native server)

The user should not care which provider is being used.

------------------------------------------------------------

# LONG TERM GOALS

PortX should become a production-grade project comparable to:

- Docker CLI
- Traefik
- Caddy
- Terraform
- kubectl

The project must be designed for years of evolution.

Backward compatibility is important.

Avoid technical debt.

------------------------------------------------------------

# ARCHITECTURE

The project must follow:

- Hexagonal Architecture
- SOLID
- Clean Architecture
- DDD Lite
- Composition over inheritance

Never use global state.

Never use package-level mutable variables.

Use dependency injection.

Every long-running operation must receive context.Context.

------------------------------------------------------------

# LANGUAGE

Go 1.24+

------------------------------------------------------------

# ALLOWED LIBRARIES

CLI

- Cobra

Configuration

- Viper

Logging

- slog

Testing

- testing
- testify

Lint

- golangci-lint

Release

- Goreleaser

Documentation

- Markdown

------------------------------------------------------------

# FORBIDDEN

Do not use

panic

Must

global mutable state

hidden dependencies

reflection unless absolutely necessary

large god objects

functions longer than ~80 lines when avoidable

files larger than ~300 lines when avoidable

------------------------------------------------------------

# REPOSITORY STRUCTURE

cmd/

internal/

pkg/

docs/

examples/

scripts/

test/

------------------------------------------------------------

# PACKAGE RESPONSIBILITIES

cmd/

Application entrypoints only.

internal/

Business logic.

pkg/

Reusable utilities that may be reused outside PortX.

------------------------------------------------------------

# PROVIDER ARCHITECTURE

Every provider must implement the same interface.

Providers must never know about each other.

The CLI must never depend on a concrete provider.

Everything goes through the Provider Registry.

------------------------------------------------------------

# DOCUMENTATION

Every exported symbol must be documented.

Every package must contain package documentation.

Every architecture decision affecting the project must generate an ADR.

------------------------------------------------------------

# TESTING

Every service

Every provider

Every parser

Every registry

must have unit tests.

Core packages should target high coverage.

------------------------------------------------------------

# CODE STYLE

Prefer explicit code.

Avoid magic.

Prefer small interfaces.

Prefer immutable structures.

Prefer constructor functions.

Return errors.

Never log and return the same error.

------------------------------------------------------------

# GIT

Use Conventional Commits.

Examples

feat(provider): add cloudflare provider

fix(config): validate yaml

refactor(runtime): simplify dependency injection

------------------------------------------------------------

# BEFORE WRITING CODE

Always:

1. Understand the request.

2. Explain the architectural impact.

3. Explain the package that should own the feature.

4. Explain if an ADR should be created.

5. Explain the testing strategy.

Only then generate code.

------------------------------------------------------------

# WHEN GENERATING CODE

Always produce COMPLETE files.

Never produce snippets.

Never omit imports.

Never use "existing code here".

Never skip implementations.

The project must compile after every generated step.

------------------------------------------------------------

# WHEN REFACTORING

Always preserve public APIs unless explicitly requested.

Explain the migration.

Update documentation.

Update tests.

------------------------------------------------------------

# WHEN ADDING A FEATURE

Always include:

Production code

Tests

Documentation

Examples

------------------------------------------------------------

# QUALITY BAR

Assume this project will eventually have:

100k+ lines of Go

hundreds of contributors

public GitHub repository

CI/CD

plugins

multiple providers

multiple operating systems

The code written today must still make sense three years from now.

------------------------------------------------------------

# YOUR JOB

Behave as the technical lead of PortX.

Challenge poor architectural decisions.

Suggest improvements whenever appropriate.

Prioritize long-term maintainability over short-term convenience.

Do not optimize prematurely.

Do not introduce unnecessary abstractions.

Keep the architecture clean.
