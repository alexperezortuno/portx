package provider

import (
	"context"
	"fmt"
)

type Provider interface {
	Name() string
	Start(ctx context.Context, cfg TunnelConfig) error
	Stop(ctx context.Context) error
	Status(ctx context.Context) (Status, error)
}

type TunnelConfig struct {
	LocalAddr  string
	RemoteAddr string
}

type Status string

const (
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
	StatusError   Status = "error"
)

func (s Status) String() string {
	return string(s)
}

type ConfiguredProvider struct {
	Provider Provider
	Config   TunnelConfig
}

func (c *ConfiguredProvider) Start(ctx context.Context) error {
	return c.Provider.Start(ctx, c.Config)
}

func (c *ConfiguredProvider) Stop(ctx context.Context) error {
	return c.Provider.Stop(ctx)
}

func (c *ConfiguredProvider) Status(ctx context.Context) (Status, error) {
	return c.Provider.Status(ctx)
}

var (
	ErrProviderNotFound          = fmt.Errorf("provider not found")
	ErrProviderAlreadyRegistered = fmt.Errorf("provider already registered")
	ErrProviderNotRunning        = fmt.Errorf("provider not running")
	ErrInvalidConfig             = fmt.Errorf("invalid tunnel config")
)
