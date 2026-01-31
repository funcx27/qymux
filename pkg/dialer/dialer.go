package dialer

import (
	"crypto/tls"
	"log"
	"net"
	"sync"

	"github.com/funcx27/qymux/pkg/quic"
	"github.com/funcx27/qymux/pkg/tcp"
	"github.com/funcx27/qymux/pkg/transport"
	"github.com/funcx27/qymux/pkg/utils"
)

// Dialer 实现支持多种传输模式的拨号器
type Dialer struct {
	config     *transport.Config
	quicDialer *quic.Dialer
	tcpDialer  *tcp.Dialer
}

// NewDialer 创建新的拨号器
func NewDialer(config *transport.Config) *Dialer {
	if config == nil {
		config = &transport.Config{
			Mode:      transport.ModeAuto,
			TLSConfig: nil,
		}
	}

	d := &Dialer{
		config: config,
	}

	tlsConfig := extractTLSConfig(config.TLSConfig)
	d.quicDialer = quic.NewDialer(tlsConfig)
	d.tcpDialer = tcp.NewDialer(tlsConfig)

	return d
}

// extractTLSConfig 从配置中提取 TLS 配置
func extractTLSConfig(config interface{}) *tls.Config {
	if config == nil {
		return nil
	}
	return config.(*tls.Config)
}

// Dial 根据配置建立多路复用会话
func (d *Dialer) Dial(target string) (transport.MuxSession, error) {
	switch d.config.Mode {
	case transport.ModeQUIC:
		return d.dialQUIC(target)
	case transport.ModeTCP:
		return d.dialTCP(target)
	case transport.ModeAuto:
		return d.dialAuto(target)
	default:
		return d.dialAuto(target)
	}
}

// dialQUIC 仅使用 QUIC 拨号
func (d *Dialer) dialQUIC(target string) (transport.MuxSession, error) {
	return d.quicDialer.Dial(target)
}

// dialTCP 仅使用 TCP+Yamux 拨号
func (d *Dialer) dialTCP(target string) (transport.MuxSession, error) {
	return d.tcpDialer.Dial(target)
}

// dialAuto 优先 QUIC，失败后回退 TCP
func (d *Dialer) dialAuto(target string) (transport.MuxSession, error) {
	// 首先尝试 QUIC
	log.Printf("[Qymux] 尝试 QUIC 连接到 %s", target)
	session, err := d.quicDialer.Dial(target)
	if err == nil {
		log.Printf("[Qymux] QUIC 连接成功")
		return session, nil
	}

	// QUIC 失败，记录日志并回退到 TCP
	log.Printf("[Qymux] QUIC 连接失败: %v，回退到 TCP", err)
	log.Printf("[Qymux] 尝试 TCP 连接到 %s", target)

	return d.tcpDialer.Dial(target)
}

// Listener 支持多种传输模式的监听器
type Listener struct {
	config       *transport.Config
	quicListener *quic.Listener
	tcpListener  *tcp.Listener
	quicSession  chan transport.MuxSession
	quicErr      chan error
	tcpSession   chan transport.MuxSession
	tcpErr       chan error
	closed       bool
	wg           sync.WaitGroup // 等待 goroutine 退出
}

// NewListener 创建新的监听器
func NewListener(addr string, config *transport.Config) (*Listener, error) {
	if config == nil {
		config = &transport.Config{
			Mode:      transport.ModeAuto,
			TLSConfig: nil,
		}
	}

	l := &Listener{
		config:      config,
		quicSession: make(chan transport.MuxSession, 10),
		quicErr:     make(chan error, 1),
		tcpSession:  make(chan transport.MuxSession, 10),
		tcpErr:      make(chan error, 1),
	}

	tlsConfig := extractTLSConfig(config.TLSConfig)

	// 根据配置创建对应的监听器
	switch config.Mode {
	case transport.ModeQUIC:
		ln, err := quic.Listen(addr, tlsConfig)
		if err != nil {
			return nil, err
		}
		l.quicListener = ln

	case transport.ModeTCP:
		ln, err := tcp.Listen(addr, tlsConfig)
		if err != nil {
			return nil, err
		}
		l.tcpListener = ln

	default: // ModeAuto 或其他情况，同时支持两种模式
		// 对于监听器，需要同时监听 UDP (QUIC) 和 TCP
		// 这里我们创建两个监听器
		quicLn, err := quic.Listen(addr, tlsConfig)
		if err != nil {
			log.Printf("[Qymux] QUIC 监听失败: %v，仅使用 TCP", err)
		} else {
			l.quicListener = quicLn
		}

		tcpLn, err := tcp.Listen(addr, tlsConfig)
		if err != nil {
			if l.quicListener != nil {
				l.quicListener.Close()
			}
			return nil, err
		}
		l.tcpListener = tcpLn
	}

	// 启动监听 goroutine
	if l.quicListener != nil {
		l.wg.Add(1)
		go l.serveQuic()
	}
	if l.tcpListener != nil {
		l.wg.Add(1)
		go l.serveTCP()
	}

	return l, nil
}

// Accept 接受新连接
func (l *Listener) Accept() (transport.MuxSession, error) {
	if l.quicListener != nil && l.tcpListener != nil {
		return l.acceptDualMode()
	}
	if l.quicListener != nil {
		return l.quicListener.Accept()
	}
	return l.tcpListener.Accept()
}

// acceptDualMode 处理双模式监听（QUIC + TCP）
func (l *Listener) acceptDualMode() (transport.MuxSession, error) {
	for {
		select {
		case session, ok := <-l.quicSession:
			if !ok {
				// QUIC channel 已关闭，回退到 TCP
				return l.tcpListener.Accept()
			}
			return session, nil
		case session, ok := <-l.tcpSession:
			if !ok {
				// TCP channel 已关闭，回退到 QUIC
				return l.quicListener.Accept()
			}
			return session, nil
		case err, ok := <-l.quicErr:
			if ok && err != nil {
				// QUIC 监听失败，回退到 TCP
				log.Printf("[Qymux] QUIC 接受连接失败: %v，回退到 TCP", err)
				return l.tcpListener.Accept()
			}
		case err, ok := <-l.tcpErr:
			if ok && err != nil {
				// TCP 监听失败，回退到 QUIC
				log.Printf("[Qymux] TCP 接受连接失败: %v，回退到 QUIC", err)
				return l.quicListener.Accept()
			}
		}
	}
}

// serveQuic 持续监听 QUIC 连接
func (l *Listener) serveQuic() {
	defer l.wg.Done()
	for !l.closed {
		session, err := l.quicListener.Accept()
		if err != nil {
			l.quicErr <- err
			return
		}
		l.quicSession <- session
	}
}

// serveTCP 持续监听 TCP 连接
func (l *Listener) serveTCP() {
	defer l.wg.Done()
	for !l.closed {
		session, err := l.tcpListener.Accept()
		if err != nil {
			l.tcpErr <- err
			return
		}
		l.tcpSession <- session
	}
}

// Close 关闭监听器
func (l *Listener) Close() error {
	l.closed = true

	// 关闭底层监听器，这将导致 Accept() 返回错误
	err := utils.CloseAll(l.quicListener, l.tcpListener)

	// 等待 goroutine 退出
	l.wg.Wait()

	// 安全地关闭通道
	close(l.quicSession)
	close(l.quicErr)
	close(l.tcpSession)
	close(l.tcpErr)
	return err
}

// Addr 返回监听地址
func (l *Listener) Addr() net.Addr {
	if l.quicListener != nil {
		return l.quicListener.Addr()
	}
	return l.tcpListener.Addr()
}
