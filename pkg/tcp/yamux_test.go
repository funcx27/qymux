package tcp

import (
	"crypto/tls"
	"testing"
)

func TestNewDialer(t *testing.T) {
	dialer := NewDialer(nil)
	if dialer == nil {
		t.Error("NewDialer() returned nil")
	}
}

func TestNewDialerWithConfig(t *testing.T) {
	config := &tls.Config{}
	dialer := NewDialer(config)
	if dialer == nil {
		t.Error("NewDialer() with config returned nil")
	}
}

func TestNewSession(t *testing.T) {
	// NewSession 需要 *yamux.Session, 这里只测试 nil 不会 panic
	session := NewSession(nil, nil, nil)
	if session == nil {
		t.Error("NewSession() returned nil")
	}

	// 测试 Protocol 方法
	if session.Protocol() != "TCP" {
		t.Errorf("Protocol() = %v, want %v", session.Protocol(), "TCP")
	}
}

func TestSessionAddr(t *testing.T) {
	session := NewSession(nil, nil, nil)
	addr := session.Addr()
	if addr != nil {
		t.Error("Session with nil localAddr should have nil Addr()")
	}
}
