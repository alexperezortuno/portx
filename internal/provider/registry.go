package provider

import "fmt"

type Registry interface {
	Register(p Provider) error
	Unregister(name string) error
	Get(name string) (Provider, bool)
	List() []string
}

type ProviderRegistry interface {
	Registry
}

func NewRegistry() ProviderRegistry {
	return &inMemoryRegistry{
		providers: make(map[string]Provider),
	}
}

type inMemoryRegistry struct {
	providers map[string]Provider
}

func (r *inMemoryRegistry) Register(p Provider) error {
	if p == nil {
		return ErrProviderNotFound
	}
	name := p.Name()
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("%w: %s", ErrProviderAlreadyRegistered, name)
	}
	r.providers[name] = p
	return nil
}

func (r *inMemoryRegistry) Unregister(name string) error {
	if _, exists := r.providers[name]; !exists {
		return fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}
	delete(r.providers, name)
	return nil
}

func (r *inMemoryRegistry) Get(name string) (Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

func (r *inMemoryRegistry) List() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}
