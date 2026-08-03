package ssh

import (
	"context"
	"testing"

	"github.com/alexperezortuno/portx/internal/provider"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cfg := SSHConfig{
		User: "test",
		Host: "example.com",
		Port: 22,
	}

	p := New("ssh-test", cfg)
	assert.Equal(t, "ssh-test", p.name)
	assert.Equal(t, provider.StatusStopped, p.status)
}

func TestSSHProvider_Name(t *testing.T) {
	p := New("my-ssh", SSHConfig{})
	assert.Equal(t, "my-ssh", p.Name())
}

func TestSSHProvider_Status(t *testing.T) {
	p := New("test", SSHConfig{})

	status, err := p.Status(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, provider.StatusStopped, status)
}

func TestSSHProvider_StartWithoutAuth(t *testing.T) {
	p := New("test", SSHConfig{
		User: "test",
		Host: "example.com",
		Port: 22,
	})

	err := p.Start(context.Background(), provider.TunnelConfig{
		LocalAddr:  "localhost:8080",
		RemoteAddr: "0.0.0.0:10000",
	})
	assert.Error(t, err)
}

func TestParseHostAddr(t *testing.T) {
	tests := []struct {
		addr     string
		wantNet  string
		wantAddr string
		wantErr  bool
	}{
		{"0.0.0.0:10000", "tcp", "0.0.0.0:10000", false},
		{"localhost:8080", "tcp", "localhost:8080", false},
		{"192.168.1.1:22", "tcp", "192.168.1.1:22", false},
		{"[::1]:10000", "tcp", "[::1]:10000", false},
	}

	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			net, addr, err := parseHostAddr(tt.addr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantNet, net)
				assert.Equal(t, tt.wantAddr, addr)
			}
		})
	}
}

func TestReverseForward_Stop(t *testing.T) {
	f := &reverseForward{}
	f.stop()
}
