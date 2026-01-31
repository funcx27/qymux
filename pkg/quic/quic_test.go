package quic

import (
	"testing"
)

func TestNewSession(t *testing.T) {
	// NewSession 需要 *quic.Conn，这里测试 nil 不会 panic
	session := NewSession(nil, nil, nil)
	if session == nil {
		t.Error("NewSession() returned nil")
	}

	// 测试 Protocol 方法
	if session.Protocol() != "QUIC" {
		t.Errorf("Protocol() = %v, want %v", session.Protocol(), "QUIC")
	}
}

func TestSessionAddr(t *testing.T) {
	session := NewSession(nil, nil, nil)
	addr := session.Addr()
	if addr != nil {
		t.Error("Session with nil addresses should have nil Addr()")
	}
}

func TestSessionImplementsMuxSession(t *testing.T) {
	// 验证 Session 实现了接口
	session := NewSession(nil, nil, nil)

	// 这些方法应该存在而不 panic
	_ = session.Protocol()
	_ = session.Addr()
}
