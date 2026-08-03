package target

import (
	"testing"

	"github.com/alexperezortuno/portx/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestResolver_TryPort(t *testing.T) {
	cfg := &config.Config{}
	r := NewResolver(cfg)

	tests := []struct {
		input    string
		wantPort int
		wantOk   bool
	}{
		{"3000", 3000, true},
		{"8080", 8080, true},
		{"80", 80, true},
		{"65535", 65535, true},
		{"0", 0, false},
		{"65536", 0, false},
		{"-1", 0, false},
		{"abc", 0, false},
		{"", 0, false},
		{"8080abc", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			resolved := r.tryPort(tt.input)
			if tt.wantOk {
				assert.NotNil(t, resolved)
				assert.Equal(t, tt.wantPort, resolved.Port)
				assert.Equal(t, TypePort, resolved.Type)
			} else {
				assert.Nil(t, resolved)
			}
		})
	}
}

func TestResolver_TryService(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]config.ServiceConfig{
			"frontend": {Name: "frontend", Port: 3000, Protocol: "http"},
			"api":      {Name: "api", Port: 8080, Protocol: "http"},
		},
	}
	r := NewResolver(cfg)

	tests := []struct {
		input    string
		wantOk   bool
		wantPort int
	}{
		{"frontend", true, 3000},
		{"api", true, 8080},
		{"nonexistent", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			resolved := r.tryService(tt.input)
			if tt.wantOk {
				assert.NotNil(t, resolved)
				assert.Equal(t, TypeService, resolved.Type)
				assert.Equal(t, tt.wantPort, resolved.Port)
			} else {
				assert.Nil(t, resolved)
			}
		})
	}
}

func TestResolver_Resolve(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]config.ServiceConfig{
			"frontend": {Name: "frontend", Port: 3000, Protocol: "http"},
		},
	}
	r := NewResolver(cfg)

	tests := []struct {
		input        string
		expectedType Type
		wantOk       bool
	}{
		{"3000", TypePort, true},
		{"frontend", TypeService, true},
		{"nonexistent", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			resolved, err := r.Resolve(tt.input)
			if tt.wantOk {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedType, resolved.Type)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestResolver_ResolveEmpty(t *testing.T) {
	r := NewResolver(nil)
	_, err := r.Resolve("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}
