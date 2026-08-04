package frp

import (
	"context"
	"testing"

	"github.com/alexperezortuno/portx/internal/provider"
)

func TestProvider_Name(t *testing.T) {
	p := New("frp-test")
	if p.Name() != "frp-test" {
		t.Errorf("expected name 'frp-test', got %q", p.Name())
	}
}

func TestProvider_Status_Stopped(t *testing.T) {
	p := New("frp-test")
	ctx := context.Background()

	status, err := p.Status(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != provider.StatusStopped {
		t.Errorf("expected status stopped, got %s", status)
	}
}

func TestProvider_URL_Empty(t *testing.T) {
	p := New("frp-test")
	if p.URL() != "" {
		t.Errorf("expected empty URL, got %q", p.URL())
	}
}

func TestParsePort_Valid(t *testing.T) {
	tests := []struct {
		addr   string
		expect int
	}{
		{"localhost:8080", 8080},
		{"127.0.0.1:3000", 3000},
		{"0.0.0.0:80", 80},
	}

	for _, tt := range tests {
		port, err := parsePort(tt.addr)
		if err != nil {
			t.Errorf("parsePort(%q) unexpected error: %v", tt.addr, err)
			continue
		}
		if port != tt.expect {
			t.Errorf("parsePort(%q) = %d, want %d", tt.addr, port, tt.expect)
		}
	}
}

func TestParsePort_Invalid(t *testing.T) {
	tests := []string{
		"localhost",
		"invalid",
		":8080",
		"localhost:abc",
		"localhost:0",
		"localhost:65536",
		"::1:9000",
	}

	for _, addr := range tests {
		_, err := parsePort(addr)
		if err == nil {
			t.Errorf("parsePort(%q) expected error, got nil", addr)
		}
	}
}

func TestConfig_ProxyName(t *testing.T) {
	cfg := &Config{
		ProxyType: ProxyTypeTCP,
		LocalPort: 8080,
		User:      "testuser",
	}
	if cfg.ProxyName() != "testuser.tcp-8080" {
		t.Errorf("expected 'testuser.tcp-8080', got %q", cfg.ProxyName())
	}
}

func TestConfig_PublicURL_TCP(t *testing.T) {
	cfg := &Config{
		ProxyType:  ProxyTypeTCP,
		ServerAddr: "frp.example.com",
		ServerPort: 7000,
		RemotePort: 12345,
	}
	url := cfg.PublicURL()
	if url != "tcp://frp.example.com:12345" {
		t.Errorf("expected 'tcp://frp.example.com:12345', got %q", url)
	}
}

func TestConfig_PublicURL_HTTP(t *testing.T) {
	cfg := &Config{
		ProxyType:  ProxyTypeHTTP,
		ServerAddr: "frp.example.com",
		ServerPort: 7000,
		Subdomain:  "myapp",
	}
	url := cfg.PublicURL()
	if url != "http://myapp.frp.example.com" {
		t.Errorf("expected 'http://myapp.frp.example.com', got %q", url)
	}
}

func TestConfig_PublicURL_HTTPS_CustomDomain(t *testing.T) {
	cfg := &Config{
		ProxyType:    ProxyTypeHTTPS,
		ServerAddr:   "frp.example.com",
		ServerPort:   7000,
		CustomDomain: "api.example.com",
	}
	url := cfg.PublicURL()
	if url != "https://api.example.com" {
		t.Errorf("expected 'https://api.example.com', got %q", url)
	}
}

func TestParseConfig_Valid(t *testing.T) {
	cfg, err := ParseConfig(
		"frp.example.com:7000",
		"mytoken",
		ProxyTypeHTTP,
		"myapp",
		"",
		0,
		"testuser",
		8080,
		false,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ServerAddr != "frp.example.com" {
		t.Errorf("expected server addr 'frp.example.com', got %q", cfg.ServerAddr)
	}
	if cfg.ServerPort != 7000 {
		t.Errorf("expected server port 7000, got %d", cfg.ServerPort)
	}
	if cfg.Token != "mytoken" {
		t.Errorf("expected token 'mytoken', got %q", cfg.Token)
	}
	if cfg.ProxyType != ProxyTypeHTTP {
		t.Errorf("expected proxy type http, got %s", cfg.ProxyType)
	}
	if cfg.Subdomain != "myapp" {
		t.Errorf("expected subdomain 'myapp', got %q", cfg.Subdomain)
	}
	if cfg.User != "testuser" {
		t.Errorf("expected user 'testuser', got %q", cfg.User)
	}
	if cfg.LocalPort != 8080 {
		t.Errorf("expected local port 8080, got %d", cfg.LocalPort)
	}
}

func TestParseConfig_MissingServerAddr(t *testing.T) {
	_, err := ParseConfig("", "token", ProxyTypeTCP, "", "", 0, "", 8080, false)
	if err == nil {
		t.Error("expected error for missing server address")
	}
}

func TestParseConfig_MissingToken(t *testing.T) {
	_, err := ParseConfig("frp.example.com:7000", "", ProxyTypeTCP, "", "", 0, "", 8080, false)
	if err == nil {
		t.Error("expected error for missing token")
	}
}

func TestParseConfig_InvalidProxyType(t *testing.T) {
	_, err := ParseConfig("frp.example.com:7000", "token", "invalid", "", "", 0, "", 8080, false)
	if err == nil {
		t.Error("expected error for invalid proxy type")
	}
}

func TestParseConfig_HTTP_MissingSubdomain(t *testing.T) {
	_, err := ParseConfig("frp.example.com:7000", "token", ProxyTypeHTTP, "", "", 0, "", 8080, false)
	if err == nil {
		t.Error("expected error for missing subdomain/custom-domain for http")
	}
}
