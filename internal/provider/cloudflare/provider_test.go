package cloudflare

import (
	"context"
	"testing"

	"github.com/alexperezortuno/portx/internal/provider"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	p := New("cloudflare-test")
	assert.Equal(t, "cloudflare-test", p.name)
	assert.Equal(t, provider.StatusStopped, p.status)
}

func TestProvider_Name(t *testing.T) {
	p := New("my-cf")
	assert.Equal(t, "my-cf", p.Name())
}

func TestProvider_Status(t *testing.T) {
	p := New("test")

	status, err := p.Status(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, provider.StatusStopped, status)
}

func TestProvider_StopWithoutStart(t *testing.T) {
	p := New("test")
	err := p.Stop(context.Background())
	assert.NoError(t, err)

	status, err := p.Status(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, provider.StatusStopped, status)
}

func TestProvider_DoubleStart(t *testing.T) {
	p := New("test")

	err := p.Start(context.Background(), provider.TunnelConfig{
		LocalAddr: "localhost:99999",
	})
	assert.Error(t, err)

	status, err := p.Status(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, provider.StatusStopped, status)
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		addr    string
		want    int
		wantErr bool
	}{
		{"localhost:8080", 8080, false},
		{"127.0.0.1:3000", 3000, false},
		{"[::1]:9000", 9000, false},
		{"localhost:http", 0, true},
		{"localhost:0", 0, true},
		{"localhost:70000", 0, true},
		{"example.com:8080", 0, true},
		{"192.168.1.1:8080", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			port, err := parsePort(tt.addr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, port)
			}
		})
	}
}

func TestProvider_URL(t *testing.T) {
	p := New("test")
	assert.Equal(t, "", p.URL())
}
