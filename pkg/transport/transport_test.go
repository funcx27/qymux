package transport

import "testing"

func TestTransportModeValues(t *testing.T) {
	tests := []struct {
		name string
		mode TransportMode
	}{
		{"auto mode", ModeAuto},
		{"quic mode", ModeQUIC},
		{"tcp mode", ModeTCP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mode == "" {
				t.Error("TransportMode should not be empty")
			}
		})
	}
}
