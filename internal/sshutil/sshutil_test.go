package sshutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ClientConfig_NoAuth(t *testing.T) {
	cfg := &Config{
		User: "test",
		Host: "example.com",
		Port: 22,
	}

	_, err := cfg.ClientConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no SSH authentication method")
}

func TestConfig_ClientConfig_Password(t *testing.T) {
	cfg := &Config{
		User:     "test",
		Host:     "example.com",
		Port:     22,
		Password: "secret",
	}

	clientCfg, err := cfg.ClientConfig()
	assert.NoError(t, err)
	assert.NotNil(t, clientCfg)
	assert.Equal(t, "test", clientCfg.User)
}

func TestConfig_ClientConfig_InvalidKey(t *testing.T) {
	cfg := &Config{
		User:       "test",
		Host:       "example.com",
		Port:       22,
		PrivateKey: "invalid key content",
	}

	_, err := cfg.ClientConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parsing private key")
}

func TestNewDialer(t *testing.T) {
	cfg := &Config{
		User:     "test",
		Host:     "example.com",
		Port:     22,
		Password: "secret",
	}

	dialer := NewDialer(cfg)
	assert.NotNil(t, dialer)
	assert.Equal(t, cfg, dialer.Config)
}

func TestForward_Stop(t *testing.T) {
	f := &Forward{}
	f.Stop()
}

func TestParseHostAddr(t *testing.T) {
	tests := []struct {
		addr     string
		wantNet  string
		wantHost string
		wantErr  bool
	}{
		{"0.0.0.0:10000", "tcp", "0.0.0.0:10000", false},
		{"localhost:8080", "tcp", "localhost:8080", false},
		{"192.168.1.1:22", "tcp", "192.168.1.1:22", false},
		{"[::1]:10000", "tcp", "[::1]:10000", false},
		{"invalid", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			net, addr, err := ParseHostAddr(tt.addr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantNet, net)
				assert.Equal(t, tt.wantHost, addr)
			}
		})
	}
}
