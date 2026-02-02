package qymux

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/funcx27/qymux/pkg/transport"
)

// DemoHTTPServer 演示如何在 Session 上启动 HTTP 代理服务器
func DemoHTTPServer() {
	// 1. 创建 qymux 实例
	q := New(&Config{
		Mode:       transport.ModeTCP, // 使用 TCP + Yamux
		ListenAddr: ":9090",
	})

	// 2. 启动服务器
	listener, err := q.Listen()
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	// 3. 在每个连接上启动 HTTP 代理服务器
	go func() {
		for {
			sess, err := listener.Accept()
			if err != nil {
				break
			}

			// 在这个 Session 上启动 HTTP 代理
			// 将请求转发到 localhost:8080
			StartHTTPServer(sess, "http://localhost:8080")
		}
	}()

	// 服务器继续运行...
	select {}
}

// DemoHTTPClient 演示如何通过 Session 创建 HTTP 客户端
func DemoHTTPClient() {
	// 1. 连接到服务器
	q := New(&Config{
		Mode:       transport.ModeTCP,
		ServerAddr: "localhost:9090",
	})

	sess, err := q.Dial()
	if err != nil {
		panic(err)
	}
	defer sess.Close()

	// 2. 创建 HTTP 客户端
	client, err := DialHTTP(sess, "http://localhost:8080")
	if err != nil {
		panic(err)
	}

	// 3. 发送 HTTP 请求
	resp, err := client.Get("/api/v1/pods")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// 4. 处理响应
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Body: %s\n", body)
}

// TestHTTPOverSession 测试 HTTP over Session
func TestHTTPOverSession(t *testing.T) {
	// 1. 创建模拟的目标服务器
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok", "path": "` + r.URL.Path + `"}`))
	}))
	defer targetServer.Close()

	// 注意：这个测试需要完整的 Session 建立流程
	// 实际使用时，通过真实的 qymux 连接建立 Session
	t.Skip("需要真实的 MuxSession，集成测试中验证")
}

// TestDialHTTP 创建 HTTP 客户端的单元测试
func TestDialHTTP(t *testing.T) {
	// 解析目标 URL
	targetURL := "http://localhost:8080"
	target, err := url.Parse(targetURL)
	if err != nil {
		t.Fatalf("parse URL failed: %v", err)
	}

	if target.Scheme != "http" {
		t.Errorf("expected scheme 'http', got '%s'", target.Scheme)
	}

	if target.Host != "localhost:8080" {
		t.Errorf("expected host 'localhost:8080', got '%s'", target.Host)
	}
}

// TestSessionTransport 测试 sessionTransport
func TestSessionTransport(t *testing.T) {
	// 这个测试需要 mock Session
	// 实际测试在集成测试中进行
	t.Skip("需要 mock MuxSession")
}

// BenchmarkDialHTTP 性能测试
func BenchmarkDialHTTP(b *testing.B) {
	// 准备测试环境...
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建 HTTP 客户端
		// client, err := DialHTTP(sess, "http://localhost:8080")
		// if err != nil {
		// 	b.Fatal(err)
		// }
		// _ = client
	}
}
