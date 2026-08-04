package frp

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type ProxyType string

const (
	ProxyTypeTCP   ProxyType = "tcp"
	ProxyTypeHTTP  ProxyType = "http"
	ProxyTypeHTTPS ProxyType = "https"
)

type Config struct {
	ServerAddr   string
	ServerPort   int
	Token        string
	ProxyType    ProxyType
	Subdomain    string
	CustomDomain string
	RemotePort   int
	User         string
	LocalPort    int
	TLSEnable    bool
}

func ParseConfig(serverAddr, token string, proxyType ProxyType, subdomain, customDomain string, remotePort int, user string, localPort int, tlsEnable bool) (*Config, error) {
	if serverAddr == "" {
		return nil, fmt.Errorf("server address is required")
	}

	host, portStr, err := net.SplitHostPort(serverAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid server address %q: %w", serverAddr, err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid server port %q: %w", portStr, err)
	}

	if token == "" {
		return nil, fmt.Errorf("authentication token is required")
	}

	switch proxyType {
	case ProxyTypeTCP, ProxyTypeHTTP, ProxyTypeHTTPS:
	default:
		return nil, fmt.Errorf("unsupported proxy type %q (supported: tcp, http, https)", proxyType)
	}

	if (proxyType == ProxyTypeHTTP || proxyType == ProxyTypeHTTPS) && subdomain == "" && customDomain == "" {
		return nil, fmt.Errorf("subdomain or custom-domain is required for http/https proxy types")
	}

	if proxyType == ProxyTypeTCP && remotePort < 0 {
		return nil, fmt.Errorf("remote-port must be >= 0 for tcp proxy type")
	}

	if localPort < 1 || localPort > 65535 {
		return nil, fmt.Errorf("local port %d out of range", localPort)
	}

	cfg := &Config{
		ServerAddr:   host,
		ServerPort:   port,
		Token:        token,
		ProxyType:    proxyType,
		Subdomain:    strings.TrimSpace(subdomain),
		CustomDomain: strings.TrimSpace(customDomain),
		RemotePort:   remotePort,
		User:         strings.TrimSpace(user),
		LocalPort:    localPort,
		TLSEnable:    tlsEnable,
	}

	return cfg, nil
}

func (c *Config) ProxyName() string {
	base := fmt.Sprintf("%s-%d", c.ProxyType, c.LocalPort)
	if c.User != "" {
		return c.User + "." + base
	}
	return base
}

func (c *Config) PublicURL() string {
	if c.ProxyType == ProxyTypeHTTP || c.ProxyType == ProxyTypeHTTPS {
		scheme := "http"
		if c.ProxyType == ProxyTypeHTTPS {
			scheme = "https"
		}
		host := c.CustomDomain
		if host == "" {
			host = c.Subdomain + "." + c.ServerAddr
		}
		return fmt.Sprintf("%s://%s", scheme, host)
	}
	return fmt.Sprintf("tcp://%s:%d", c.ServerAddr, c.RemotePort)
}
