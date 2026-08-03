package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `
version: 1
log_level: info
provider:
  default: portxd
services:
  frontend:
    port: 3000
    protocol: http
  api:
    port: 8080
`
	err := os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err)

	cfg, err := Load(configPath)
	require.NoError(t, err)

	assert.Equal(t, 1, cfg.Version)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "portxd", cfg.Provider.Default)
	require.Len(t, cfg.Services, 2)
	assert.Equal(t, 3000, cfg.Services["frontend"].Port)
	assert.Equal(t, "http", cfg.Services["frontend"].Protocol)
}

func TestLoadNotFound(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid",
			cfg: Config{
				LogLevel: "info",
			},
			wantErr: false,
		},
		{
			name: "invalid_log_level",
			cfg: Config{
				LogLevel: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_DefaultProvider(t *testing.T) {
	cfg := &Config{
		Provider: ProviderConfig{Default: "ssh"},
	}
	assert.Equal(t, "ssh", cfg.DefaultProvider())

	cfg2 := &Config{}
	assert.Equal(t, "portxd", cfg2.DefaultProvider())
}

func TestConfig_GetService(t *testing.T) {
	cfg := &Config{
		Services: map[string]ServiceConfig{
			"frontend": {Port: 3000, Protocol: "http"},
		},
	}

	svc, ok := cfg.GetService("frontend")
	assert.True(t, ok)
	assert.Equal(t, 3000, svc.Port)

	_, ok = cfg.GetService("nonexistent")
	assert.False(t, ok)
}

func TestServiceConfig_LocalAddr(t *testing.T) {
	svc := &ServiceConfig{Host: "localhost", Port: 3000}
	assert.Equal(t, "localhost:3000", svc.LocalAddr())

	svc2 := &ServiceConfig{Host: "", Port: 8080}
	assert.Equal(t, "localhost:8080", svc2.LocalAddr())
}

func TestTunnelConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		t       TunnelConfig
		wantErr bool
	}{
		{
			name: "valid",
			t: TunnelConfig{
				LocalAddr: "localhost:8080",
			},
			wantErr: false,
		},
		{
			name: "missing_local_addr",
			t: TunnelConfig{
				LocalAddr: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.t.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
