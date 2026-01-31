package qymux

import (
	"testing"

	"github.com/funcx27/qymux/pkg/transport"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{"nil config", nil},
		{"auto mode", &Config{Mode: transport.ModeAuto}},
		{"quic mode", &Config{Mode: transport.ModeQUIC}},
		{"tcp mode", &Config{Mode: transport.ModeTCP}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := New(tt.config)
			if q == nil {
				t.Error("New() returned nil")
			}
		})
	}
}
