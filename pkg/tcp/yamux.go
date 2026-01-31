package tcp

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/hashicorp/yamux"
	tlsconfig "github.com/funcx27/qymux/pkg/tls"
	"github.com/funcx27/qymux/pkg/transport"
)

// Session 实现 transport.MuxSession 接口
type Session struct {
	session   *yamux.Session
	tlsConn   *tls.Conn
	localAddr net.Addr
}

// NewSession 创建新的 TCP+Yamux 会话适配器
func NewSession(session *yamux.Session, tlsConn *tls.Conn, localAddr net.Addr) *Session {
	return &Session{
		session:   session,
		tlsConn:   tlsConn,
		localAddr: localAddr,
	}
}

// Accept 接受来自对端的虚拟流
func (s *Session) Accept() (net.Conn, error) {
	conn, err := s.session.Accept()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// OpenStream 发起一个新的虚拟流
func (s *Session) OpenStream() (net.Conn, error) {
	return s.session.OpenStream()
}

// Protocol 返回协议类型
func (s *Session) Protocol() string {
	return "TCP"
}

// Close 关闭会话
func (s *Session) Close() error {
	// 关闭 Yamux 会话
	if err := s.session.Close(); err != nil {
		return err
	}
	// 关闭 TLS 连接
	if s.tlsConn != nil {
		return s.tlsConn.Close()
	}
	return nil
}

// Addr 返回监听地址（为了实现 net.Listener）
func (s *Session) Addr() net.Addr {
	return s.localAddr
}

// Dialer 实现 TCP+Yamux 拨号器
type Dialer struct {
	tlsConfig *tls.Config
	timeout   time.Duration
}

// NewDialer 创建新的 TCP+Yamux 拨号器
func NewDialer(tlsConfig *tls.Config) *Dialer {
	tlsConfig = tlsconfig.EnsureClientTLSConfig(tlsConfig)

	return &Dialer{
		tlsConfig: tlsConfig,
		timeout:   10 * time.Second,
	}
}

// Dial 建立到目标地址的 TCP+Yamux 连接
func (d *Dialer) Dial(target string) (transport.MuxSession, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	// 建立 TCP 连接
	conn, err := net.DialTimeout("tcp", target, d.timeout)
	if err != nil {
		return nil, err
	}

	// 建立 TLS 连接
	tlsConn := tls.Client(conn, d.tlsConfig)
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		conn.Close()
		return nil, err
	}

	// 在 TLS 上创建 Yamux 会话
	session, err := yamux.Client(tlsConn, defaultYamuxConfig())
	if err != nil {
		tlsConn.Close()
		return nil, err
	}

	return NewSession(session, tlsConn, tlsConn.LocalAddr()), nil
}

// Listener 实现 TCP+Yamux 监听器
type Listener struct {
	ln        net.Listener
	tlsConfig *tls.Config
	localAddr net.Addr
}

// NewListener 创建新的 TCP+Yamux 监听器
func NewListener(ln net.Listener, tlsConfig *tls.Config, localAddr net.Addr) *Listener {
	return &Listener{
		ln:        ln,
		tlsConfig: tlsConfig,
		localAddr: localAddr,
	}
}

// Accept 接受新连接
func (l *Listener) Accept() (transport.MuxSession, error) {
	// 接受 TCP 连接
	conn, err := l.ln.Accept()
	if err != nil {
		return nil, err
	}

	// 建立 TLS 连接
	tlsConn := tls.Server(conn, l.tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		return nil, err
	}

	// 在 TLS 上创建 Yamux 会话
	session, err := yamux.Server(tlsConn, defaultYamuxConfig())
	if err != nil {
		tlsConn.Close()
		return nil, err
	}

	return NewSession(session, tlsConn, l.localAddr), nil
}

// Close 关闭监听器
func (l *Listener) Close() error {
	return l.ln.Close()
}

// Addr 返回监听地址
func (l *Listener) Addr() net.Addr {
	return l.localAddr
}

// Listen 创建 TCP+Yamux 监听器
func Listen(addr string, tlsConfig *tls.Config) (*Listener, error) {
	var err error
	tlsConfig, err = tlsconfig.EnsureServerTLSConfig(tlsConfig)
	if err != nil {
		return nil, err
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return NewListener(ln, tlsConfig, ln.Addr()), nil
}

// defaultYamuxConfig 返回默认的 Yamux 配置
func defaultYamuxConfig() *yamux.Config {
	config := yamux.DefaultConfig()
	config.Logger = nil // 禁用 yamux 的日志
	return config
}
