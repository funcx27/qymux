package qymux

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/funcx27/qymux/pkg/dialer"
	"github.com/funcx27/qymux/pkg/transport"
)

// Qymux 提供方便第三方调用的高层 API
type Qymux struct {
	config *Config
	dialer *dialer.Dialer
}

// Config 定义 Qymux 配置
type Config struct {
	// Mode 传输模式：auto/quic/tcp
	Mode transport.TransportMode

	// TLSConfig TLS 配置，为 nil 时会自动生成自签名证书
	TLSConfig *tls.Config

	// ServerAddr 服务器地址
	ServerAddr string

	// ListenAddr 监听地址（用于 Server 模式）
	ListenAddr string
}

// New 创建新的 Qymux 实例
func New(config *Config) *Qymux {
	if config == nil {
		config = &Config{
			Mode: transport.ModeAuto,
		}
	}

	transportConfig := &transport.Config{
		Mode:      config.Mode,
		TLSConfig: config.TLSConfig,
	}

	return &Qymux{
		config: config,
		dialer: dialer.NewDialer(transportConfig),
	}
}

// Dial 连接到服务器并建立隧道
func (q *Qymux) Dial() (transport.MuxSession, error) {
	return q.dialer.Dial(q.config.ServerAddr)
}

// Listen 启动服务器监听
func (q *Qymux) Listen() (*dialer.Listener, error) {
	return dialer.NewListener(q.config.ListenAddr, &transport.Config{
		Mode:      q.config.Mode,
		TLSConfig: q.config.TLSConfig,
	})
}

// DialAgent 在 Server 端连接 Agent（gRPC 隧道）
// 这个函数实现了 gRPC 隧道适配器，将 MuxSession 转换为 gRPC 客户端连接
func DialAgent(sess transport.MuxSession, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// 配置 gRPC Keepalive，快速检测连接断开
	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second, // 每10秒发送一次 ping（避免 too_many_pings）
		Timeout:             2 * time.Second,  // ping 超时2秒认为连接断开
		PermitWithoutStream: true,            // 没有活跃 stream 也发送 ping
	}

	defaultOpts := []grpc.DialOption{
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			// 每次发起请求时，在多路复用 Session 上打开一个新流
			return sess.OpenStream()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 隧道已加密，gRPC 层使用 Insecure
		grpc.WithKeepaliveParams(kacp),                          // 启用 keepalive
	}

	opts = append(defaultOpts, opts...)

	return grpc.Dial("", opts...)
}

// StartGRPCServer 在 Session 上启动 gRPC 服务器
// 这个函数用于 Agent 端，在 MuxSession 上启动 gRPC 服务
// 返回一个 channel，当 server.Serve() 退出时会关闭该 channel
func StartGRPCServer(sess transport.MuxSession, server *grpc.Server) (<-chan struct{}, error) {
	log.Printf("[Qymux] 在 %s 协议上启动 gRPC 服务器", sess.Protocol())

	// 创建退出通知 channel
	stopped := make(chan struct{})

	// MuxSession 已经实现了 net.Listener 接口
	// 直接在 MuxSession 上启动 gRPC 服务器
	go func() {
		server.Serve(sess) // 当连接断开时，Serve() 会返回
		close(stopped)     // 通知主循环 server 已停止
	}()

	return stopped, nil
}

