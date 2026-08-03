package sshutil

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type Client = ssh.Client

type Config struct {
	User       string
	Host       string
	Port       int
	Password   string
	PrivateKey string
	RemoteAddr string
	UseAgent   bool
}

func (c *Config) ClientConfig() (*ssh.ClientConfig, error) {
	var auth []ssh.AuthMethod

	if c.Password != "" {
		auth = append(auth, ssh.Password(c.Password))
	}

	if c.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(c.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("parsing private key: %w", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}

	if c.UseAgent {
		agentAuth, err := c.agentAuth()
		if err == nil {
			auth = append(auth, agentAuth)
		}
	}

	if len(auth) == 0 {
		return nil, fmt.Errorf("no SSH authentication method provided; set --ssh-password, --ssh-private-key, or --ssh-use-agent")
	}

	return &ssh.ClientConfig{
		User:            c.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}

func (c *Config) agentAuth() (ssh.AuthMethod, error) {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return nil, fmt.Errorf("SSH_AUTH_SOCK not set")
	}

	conn, err := net.Dial("unix", sock)
	if err != nil {
		return nil, fmt.Errorf("dialing SSH_AUTH_SOCK: %w", err)
	}

	ag := agent.NewClient(conn)
	signers, err := ag.Signers()
	if err != nil {
		return nil, fmt.Errorf("getting signers from agent: %w", err)
	}

	if len(signers) == 0 {
		return nil, fmt.Errorf("no keys in SSH agent")
	}

	return ssh.PublicKeys(signers...), nil
}

type Dialer struct {
	Config *Config
}

func NewDialer(cfg *Config) *Dialer {
	return &Dialer{Config: cfg}
}

func (d *Dialer) Dial() (*ssh.Client, error) {
	sshConfig, err := d.Config.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("building SSH config: %w", err)
	}

	addr := net.JoinHostPort(d.Config.Host, fmt.Sprintf("%d", d.Config.Port))
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("SSH dial %s: %w", addr, err)
	}

	return client, nil
}

type Forward struct {
	Client     *ssh.Client
	RemoteAddr string
	LocalAddr  string
	listener   net.Listener
}

func (f *Forward) Start(ctx context.Context) error {
	network, addr, err := ParseHostAddr(f.RemoteAddr)
	if err != nil {
		return err
	}

	listener, err := f.Client.Listen(network, addr)
	if err != nil {
		return fmt.Errorf("SSH listen %s: %w", addr, err)
	}
	f.listener = listener

	go f.serve(ctx)
	return nil
}

func (f *Forward) serve(ctx context.Context) {
	for {
		conn, err := f.listener.Accept()
		if err != nil {
			return
		}
		go f.handleConn(ctx, conn)
	}
}

func (f *Forward) handleConn(ctx context.Context, remote net.Conn) {
	defer remote.Close()

	local, err := net.Dial("tcp", f.LocalAddr)
	if err != nil {
		return
	}
	defer local.Close()

	go io.Copy(local, remote)
	go io.Copy(remote, local)
}

func (f *Forward) Stop() {
	if f.listener != nil {
		f.listener.Close()
	}
}

func ParseHostAddr(addr string) (string, string, error) {
	host, strPort, err := net.SplitHostPort(addr)
	if err != nil {
		return "", "", fmt.Errorf("split host port: %w", err)
	}
	return "tcp", net.JoinHostPort(host, strPort), nil
}

func ParsePrivateKey(pem string) (ssh.Signer, error) {
	pem = strings.TrimSpace(pem)
	signer, err := ssh.ParsePrivateKey([]byte(pem))
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}
	return signer, nil
}
