package qymux

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/funcx27/qymux/pkg/transport"
)

// DialHTTP 创建 HTTP 客户端，通过 MuxSession 发送 HTTP 请求
// 返回的 http.Client 会将所有请求通过 Session 上的虚拟流发送
func DialHTTP(sess transport.MuxSession, targetURL string) (*http.Client, error) {
	// 解析目标 URL
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("parse target URL failed: %w", err)
	}

	// 创建自定义 Transport
	transport := &sessionTransport{
		session: sess,
		target:  target,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}, nil
}

// sessionTransport 实现 http.RoundTripper 接口
// 它将 HTTP 请求通过 Session 的虚拟流发送
type sessionTransport struct {
	session transport.MuxSession
	target  *url.URL
	mu      sync.Mutex
}

// RoundTrip 执行 HTTP 请求
func (t *sessionTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 1. 修改请求的 URL，指向目标服务器
	req.URL.Scheme = t.target.Scheme
	req.URL.Host = t.target.Host

	// 2. 在 Session 上打开一个新流
	stream, err := t.session.OpenStream()
	if err != nil {
		return nil, fmt.Errorf("open stream failed: %w", err)
	}
	defer stream.Close()

	// 设置流的截止时间（超时）
	stream.SetDeadline(time.Now().Add(30 * time.Second))

	// 3. 写入 HTTP 请求到流
	// 格式：METHOD PATH HTTP/1.1\r\nHeaders\r\n\r\nBody
	if err := req.Write(stream); err != nil {
		return nil, fmt.Errorf("write request to stream failed: %w", err)
	}

	// 4. 从流读取 HTTP 响应
	resp, err := http.ReadResponse(bufio.NewReader(stream), req)
	if err != nil {
		return nil, fmt.Errorf("read response from stream failed: %w", err)
	}

	return resp, nil
}

// StartHTTPServer 在 Session 上启动 HTTP 代理服务器
// 接收来自对端的 HTTP 请求流，转发到 targetURL，然后返回响应
//
// targetURL: 例如 "http://localhost:8001" 或 "https://kubernetes.default.svc"
func StartHTTPServer(sess transport.MuxSession, targetURL string) error {
	target, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("parse target URL failed: %w", err)
	}

	log.Printf("[Qymux-HTTP] 启动 HTTP 代理服务器，转发到: %s (%s)", targetURL, sess.Protocol())

	go func() {
		for {
			// 1. 接受新的虚拟流（每个流对应一个 HTTP 请求）
			stream, err := sess.Accept()
			if err != nil {
				// Session 关闭或其他错误
				log.Printf("[Qymux-HTTP] Accept stream failed: %v", err)
				return
			}

			// 2. 在 goroutine 中处理这个 HTTP 请求
			go handleStream(stream, target)
		}
	}()

	return nil
}

// handleStream 处理单个 HTTP 请求流
func handleStream(stream net.Conn, target *url.URL) {
	defer stream.Close()

	// 设置超时
	stream.SetDeadline(time.Now().Add(30 * time.Second))

	// 1. 从流读取 HTTP 请求
	req, err := http.ReadRequest(bufio.NewReader(stream))
	if err != nil {
		log.Printf("[Qymux-HTTP] Read request failed: %v", err)
		return
	}

	// 2. 修改请求的 URL，指向目标服务器
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.Host = target.Host

	// 3. 转发请求到目标服务器
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		// 返回错误响应
		log.Printf("[Qymux-HTTP] Forward request failed: %v", err)
		errorResp := &http.Response{
			StatusCode: http.StatusBadGateway,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("Bad Gateway: " + err.Error())),
		}
		errorResp.Write(stream)
		return
	}
	defer resp.Body.Close()

	// 4. 将响应写回流
	// 移除 Connection 头（HTTP/1.1）
	resp.Header.Del("Connection")
	resp.Header.Del("Transfer-Encoding")
	resp.Close = true

	if err := resp.Write(stream); err != nil {
		log.Printf("[Qymux-HTTP] Write response failed: %v", err)
		return
	}

	log.Printf("[Qymux-HTTP] %s %s -> %d", req.Method, req.URL.Path, resp.StatusCode)
}

// GetHTTPDialer 返回一个 http.RoundTripper，可以用于创建自定义的 http.Client
// 这个函数提供了更灵活的配置方式
func GetHTTPDialer(sess transport.MuxSession, targetURL string) (http.RoundTripper, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("parse target URL failed: %w", err)
	}

	return &sessionTransport{
		session: sess,
		target:  target,
	}, nil
}
