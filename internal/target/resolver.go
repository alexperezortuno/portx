package target

import (
	"fmt"
	"strconv"

	"github.com/alexperezortuno/portx/internal/config"
)

type Resolver struct {
	cfg *config.Config
}

func NewResolver(cfg *config.Config) *Resolver {
	return &Resolver{cfg: cfg}
}

func (r *Resolver) Resolve(input string) (*Resolved, error) {
	if input == "" {
		return nil, fmt.Errorf("target cannot be empty")
	}

	if resolved := r.tryPort(input); resolved != nil {
		return resolved, nil
	}

	if resolved := r.tryService(input); resolved != nil {
		return resolved, nil
	}

	return nil, fmt.Errorf("unknown target: %q", input)
}

func (r *Resolver) tryPort(input string) *Resolved {
	port, err := strconv.Atoi(input)
	if err != nil {
		return nil
	}
	if port < 1 || port > 65535 {
		return nil
	}
	return &Resolved{
		Type:      TypePort,
		Port:      port,
		LocalAddr: fmt.Sprintf("localhost:%d", port),
	}
}

func (r *Resolver) tryService(input string) *Resolved {
	if r.cfg == nil || r.cfg.Services == nil {
		return nil
	}

	svc, ok := r.cfg.Services[input]
	if !ok {
		return nil
	}

	return &Resolved{
		Type:      TypeService,
		Name:      input,
		Service:   &svc,
		Port:      svc.Port,
		LocalAddr: svc.LocalAddr(),
		Protocol:  svc.Protocol,
	}
}

type Resolved struct {
	Type      Type
	Name      string
	Service   *config.ServiceConfig
	Port      int
	LocalAddr string
	Protocol  string
}

type Type string

const (
	TypePort    Type = "port"
	TypeService Type = "service"
	TypeDocker  Type = "docker"
	TypeCompose Type = "compose"
	TypeK8s     Type = "kubernetes"
)

func (t Type) String() string {
	return string(t)
}
