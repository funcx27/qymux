package quic

import (
	"context"
	"crypto/tls"
	"net"

	tlsconfig "github.com/funcx27/qymux/pkg/tls"
	"github.com/funcx27/qymux/pkg/transport"
	"github.com/quic-go/quic-go"
)

// Session 实现 transport.MuxSession 接口
type Session struct {
	conn       *quic.Conn
	localAddr  net.Addr
	remoteAddr net.Addr
}

// NewSession 创建新的 QUIC 会话适配器
func NewSession(conn *quic.Conn, localAddr, remoteAddr net.Addr) *Session {
	return &Session{
		conn:       conn,
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
	}
}

// Accept 接受来自对端的虚拟流
func (s *Session) Accept() (net.Conn, error) {
	stream, err := s.conn.AcceptStream(s.context())
	if err != nil {
		return nil, err
	}

	return &Conn{
		stream:    stream,
		localAddr:  s.localAddr,
		remoteAddr: s.remoteAddr,
	}, nil
}

// OpenStream 发起一个新的虚拟流
func (s *Session) OpenStream() (net.Conn, error) {
	stream, err := s.conn.OpenStreamSync(s.context())
	if err != nil {
		return nil, err
	}

	return &Conn{
		stream:    stream,
		localAddr:  s.localAddr,
		remoteAddr: s.remoteAddr,
	}, nil
}

// Protocol 返回协议类型
func (s *Session) Protocol() string {
	return "QUIC"
}

// Close 关闭会话
func (s *Session) Close() error {
	return s.conn.CloseWithError(0, "")
}

// Addr 返回监听地址（为了实现 net.Listener）
func (s *Session) Addr() net.Addr {
	return s.localAddr
}

// context 返回默认上下文
func (s *Session) context() context.Context {
	return context.Background()
}

// Dialer 实现 QUIC 拨号器
type Dialer struct {
	tlsConfig *tls.Config
	config    *quic.Config
}

// NewDialer 创建新的 QUIC 拨号器
func NewDialer(tlsConfig *tls.Config) *Dialer {
	tlsConfig = tlsconfig.EnsureClientTLSConfig(tlsConfig)

	return &Dialer{
		tlsConfig: tlsConfig,
		config:    &quic.Config{},
	}
}

// Dial 建立到目标地址的 QUIC 连接
func (d *Dialer) Dial(target string) (transport.MuxSession, error) {
	// 建立到目标地址的 QUIC 连接
	quicConn, err := quic.DialAddr(context.Background(), target, d.tlsConfig, d.config)
	if err != nil {
		return nil, err
	}

	// 返回适配的会话
	return NewSession(quicConn, quicConn.LocalAddr(), quicConn.RemoteAddr()), nil
}

// Listener 实现 QUIC 监听器
type Listener struct {
	ln         *quic.Listener
	localAddr  net.Addr
	acceptChan chan *Session
	errChan    chan error
}

// NewListener 创建新的 QUIC 监听器
func NewListener(ln *quic.Listener, localAddr net.Addr) *Listener {
	l := &Listener{
		ln:         ln,
		localAddr:  localAddr,
		acceptChan: make(chan *Session, 10),
		errChan:    make(chan error, 1),
	}

	// 启动接受 goroutine
	go l.acceptLoop()

	return l
}

// acceptLoop 持续接受新连接
func (l *Listener) acceptLoop() {
	for {
		conn, err := l.ln.Accept(context.Background())
		if err != nil {
			l.errChan <- err
			return
		}

		session := NewSession(conn, l.localAddr, conn.RemoteAddr())
		l.acceptChan <- session
	}
}

// Accept 接受新连接
func (l *Listener) Accept() (transport.MuxSession, error) {
	select {
	case session := <-l.acceptChan:
		return session, nil
	case err := <-l.errChan:
		return nil, err
	}
}

// Close 关闭监听器
func (l *Listener) Close() error {
	return l.ln.Close()
}

// Addr 返回监听地址
func (l *Listener) Addr() net.Addr {
	return l.localAddr
}

// Listen 创建 QUIC 监听器
func Listen(addr string, tlsConfig *tls.Config) (*Listener, error) {
	var err error
	tlsConfig, err = tlsconfig.EnsureServerTLSConfig(tlsConfig)
	if err != nil {
		return nil, err
	}

	ln, err := quic.ListenAddr(addr, tlsConfig, &quic.Config{})
	if err != nil {
		return nil, err
	}

	return NewListener(ln, ln.Addr()), nil
}
