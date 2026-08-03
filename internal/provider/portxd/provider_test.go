package portxd

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/alexperezortuno/portx/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cfg := ProviderConfig{
		LocalAddr:  "localhost:8080",
		ServerPort: 7222,
	}

	p := New("portxd-test", cfg)
	assert.Equal(t, "portxd-test", p.name)
	assert.Equal(t, provider.StatusStopped, p.status)
}

func TestProvider_Name(t *testing.T) {
	p := New("my-portxd", ProviderConfig{})
	assert.Equal(t, "my-portxd", p.Name())
}

func TestProvider_Status(t *testing.T) {
	p := New("test", ProviderConfig{})

	status, err := p.Status(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, provider.StatusStopped, status)
}

func TestProvider_StartStop(t *testing.T) {
	ln, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	localPort := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	p := New("test", ProviderConfig{
		LocalAddr:  fmt.Sprintf("localhost:%d", localPort),
		ServerPort: 0,
	})

	err = p.Start(context.Background(), provider.TunnelConfig{
		LocalAddr: fmt.Sprintf("localhost:%d", localPort),
	})
	require.NoError(t, err)

	status, err := p.Status(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, provider.StatusRunning, status)

	err = p.Stop(context.Background())
	assert.NoError(t, err)

	status, err = p.Status(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, provider.StatusStopped, status)
}

func TestServer_StartStop(t *testing.T) {
	srv := &Server{
		localAddr:  "localhost:8080",
		serverPort: 0,
	}

	ln, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	srv.serverPort = ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	err = srv.Start(context.Background())
	require.NoError(t, err)

	srv.Stop()
}

func TestServer_StopWithoutStart(t *testing.T) {
	srv := &Server{}
	srv.Stop()
}
