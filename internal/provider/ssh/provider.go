package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/alexperezortuno/portx/internal/provider"
	"golang.org/x/crypto/ssh"
)

type SSHProvider struct {
	name      string
	config    SSHConfig
	client    *ssh.Client
	status    provider.Status
	mu        sync.RWMutex
	forwarder *reverseForward
}

type SSHConfig struct {
	User       string
	Host       string
	Port       int
	Password   string
	PrivateKey string
	RemoteAddr string
	LocalAddr  string
}

func New(name string, cfg SSHConfig) *SSHProvider {
	return &SSHProvider{
		name:   name,
		config: cfg,
		status: provider.StatusStopped,
	}
}

func (p *SSHProvider) Name() string { return p.name }

func (p *SSHProvider) Start(ctx context.Context, cfg provider.TunnelConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status == provider.StatusRunning {
		return nil
	}

	sshConfig, err := p.buildSSHConfig()
	if err != nil {
		return fmt.Errorf("building SSH config: %w", err)
	}

	addr := net.JoinHostPort(p.config.Host, fmt.Sprintf("%d", p.config.Port))
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("SSH dial %s: %w", addr, err)
	}

	remoteAddr := cfg.RemoteAddr
	if remoteAddr == "" {
		remoteAddr = p.config.RemoteAddr
	}

	fwd := &reverseForward{
		client:     client,
		remoteAddr: remoteAddr,
		localAddr:  cfg.LocalAddr,
	}

	if err := fwd.start(ctx); err != nil {
		client.Close()
		return fmt.Errorf("starting reverse forward: %w", err)
	}

	p.client = client
	p.forwarder = fwd
	p.status = provider.StatusRunning

	return nil
}

func (p *SSHProvider) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status != provider.StatusRunning {
		return nil
	}

	if p.forwarder != nil {
		p.forwarder.stop()
		p.forwarder = nil
	}

	if p.client != nil {
		p.client.Close()
		p.client = nil
	}

	p.status = provider.StatusStopped
	return nil
}

func (p *SSHProvider) Status(ctx context.Context) (provider.Status, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status, nil
}

func (p *SSHProvider) buildSSHConfig() (*ssh.ClientConfig, error) {
	var auth []ssh.AuthMethod

	if p.config.Password != "" {
		auth = append(auth, ssh.Password(p.config.Password))
	}

	if p.config.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(p.config.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("parsing private key: %w", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}

	if len(auth) == 0 {
		return nil, fmt.Errorf("no SSH authentication method provided")
	}

	return &ssh.ClientConfig{
		User:            p.config.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}

type reverseForward struct {
	client     *ssh.Client
	remoteAddr string
	localAddr  string
	listener   net.Listener
}

func (f *reverseForward) start(ctx context.Context) error {
	network, addr, err := parseHostAddr(f.remoteAddr)
	if err != nil {
		return err
	}

	listener, err := f.client.Listen(network, addr)
	if err != nil {
		return fmt.Errorf("SSH listen %s: %w", addr, err)
	}
	f.listener = listener

	go f.serve(ctx)
	return nil
}

func (f *reverseForward) serve(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := f.listener.Accept()
			if err != nil {
				return
			}
			go f.handleConn(ctx, conn)
		}
	}
}

func (f *reverseForward) handleConn(ctx context.Context, remote net.Conn) {
	defer remote.Close()

	local, err := net.Dial("tcp", f.localAddr)
	if err != nil {
		return
	}
	defer local.Close()

	go io.Copy(local, remote)
	go io.Copy(remote, local)
}

func (f *reverseForward) stop() {
	if f.listener != nil {
		f.listener.Close()
	}
}

func parseHostAddr(addr string) (string, string, error) {
	host, strPort, err := net.SplitHostPort(addr)
	if err != nil {
		return "", "", fmt.Errorf("split host port: %w", err)
	}
	return "tcp", net.JoinHostPort(host, strPort), nil
}

var _ provider.Provider = (*SSHProvider)(nil)
