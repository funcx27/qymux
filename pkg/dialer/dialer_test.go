package dialer

import (
	"testing"

	"github.com/funcx27/qymux/pkg/transport"
)

func TestNewDialer(t *testing.T) {
	config := &transport.Config{
		Mode: transport.ModeAuto,
	}

	d := NewDialer(config)
	if d == nil {
		t.Error("NewDialer() returned nil")
	}
}

func TestNewDialerWithNilConfig(t *testing.T) {
	d := NewDialer(nil)
	if d == nil {
		t.Error("NewDialer(nil) returned nil")
	}
}

func TestDialerDial(t *testing.T) {
	config := &transport.Config{
		Mode: transport.ModeTCP, // 使用 TCP 模式避免证书问题
	}

	d := NewDialer(config)

	// 连接到不存在的服务器应该失败
	session, err := d.Dial("localhost:9999")
	if err == nil {
		if session != nil {
			session.Close()
		}
		t.Error("Dial() to non-existent server should fail")
	}
}

func TestDialerDialInvalidAddr(t *testing.T) {
	config := &transport.Config{
		Mode: transport.ModeTCP,
	}

	d := NewDialer(config)

	// 无效地址
	_, err := d.Dial("invalid-address")
	if err == nil {
		t.Error("Dial() with invalid address should fail")
	}
}

func TestDialerDialEmptyAddr(t *testing.T) {
	config := &transport.Config{
		Mode: transport.ModeAuto,
	}

	d := NewDialer(config)

	// 空地址
	_, err := d.Dial("")
	if err == nil {
		t.Error("Dial() with empty address should fail")
	}
}

func TestDialerWithQUICMode(t *testing.T) {
	config := &transport.Config{
		Mode: transport.ModeQUIC,
	}

	d := NewDialer(config)
	if d == nil {
		t.Error("NewDialer() with QUIC mode returned nil")
	}

	// QUIC 连接可能因为证书失败，但不应该 panic
	_, err := d.Dial("localhost:9999")
	// 这里我们只验证不会 panic，错误是预期的
	_ = err
}
