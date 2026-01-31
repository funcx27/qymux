package quic

import (
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

// Conn 将 quic.Stream 封装为 net.Conn
type Conn struct {
	stream     *quic.Stream
	localAddr  net.Addr
	remoteAddr net.Addr
}

// NewConn 创建新的 QUIC 连接封装
func NewConn(stream *quic.Stream, localAddr, remoteAddr net.Addr) *Conn {
	return &Conn{
		stream:     stream,
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
	}
}

// Read 从流中读取数据
func (c *Conn) Read(b []byte) (n int, err error) {
	return c.stream.Read(b)
}

// Write 向流中写入数据
func (c *Conn) Write(b []byte) (n int, err error) {
	return c.stream.Write(b)
}

// Close 关闭流
func (c *Conn) Close() error {
	return c.stream.Close()
}

// LocalAddr 返回本地地址
func (c *Conn) LocalAddr() net.Addr {
	return c.localAddr
}

// RemoteAddr 返回远程地址
func (c *Conn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

// SetDeadline 设置读写截止时间
func (c *Conn) SetDeadline(t time.Time) error {
	return c.stream.SetDeadline(t)
}

// SetReadDeadline 设置读截止时间
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.stream.SetReadDeadline(t)
}

// SetWriteDeadline 设置写截止时间
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.stream.SetWriteDeadline(t)
}

// StreamID 返回流 ID（QUIC 特有）
func (c *Conn) StreamID() quic.StreamID {
	return c.stream.StreamID()
}

// CancelRead 取消读取
func (c *Conn) CancelRead(errorCode quic.StreamErrorCode) {
	c.stream.CancelRead(errorCode)
}

// CancelWrite 取消写入
func (c *Conn) CancelWrite(errorCode quic.StreamErrorCode) {
	c.stream.CancelWrite(errorCode)
}
