package transport

import "net"

// TransportMode 定义传输模式
type TransportMode string

const (
	ModeAuto TransportMode = "auto" // 默认：优先 QUIC，失败后回退至 TCP
	ModeQUIC TransportMode = "quic" // 强制模式：仅使用 QUIC (UDP)
	ModeTCP  TransportMode = "tcp"  // 强制模式：仅使用 TCP + Yamux
)

// MuxSession 定义多路复用会话接口
// 所有传输层实现（QUIC、TCP+Yamux）都必须实现此接口
type MuxSession interface {
	net.Listener // 实现 Accept() 以接收来自对端的虚拟流 (Stream)

	// OpenStream 发起一个新的虚拟流 (Stream)
	OpenStream() (net.Conn, error)

	// Protocol 返回实际使用的协议 ("QUIC" 或 "TCP")
	Protocol() string

	// Close 关闭会话
	Close() error
}

// Config 定义拨号器配置
type Config struct {
	// Mode 传输模式：auto/quic/tcp
	Mode TransportMode

	// TLSConfig TLS 配置，为 nil 时会自动生成自签名证书
	TLSConfig interface{} // 实际使用时转换为 *tls.Config
}

// Dialer 定义拨号器接口
type Dialer interface {
	// Dial 根据配置建立多路复用会话
	Dial(target string) (MuxSession, error)
}
