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
`
	err := os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err)

	cfg, err := Load(configPath)
	require.NoError(t, err)

	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "ssh", cfg.Provider)
	require.Len(t, cfg.Tunnels, 1)
	assert.Equal(t, "web", cfg.Tunnels[0].Name)
	assert.Equal(t, "localhost:8080", cfg.Tunnels[0].LocalAddr)
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
