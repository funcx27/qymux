# Qymux

Qymux (QUIC-based Yamux-compatible Muxer) 是一个高性能的传输层库，提供基于 QUIC 和 TCP 的双模态反向隧道功能。

## 特性

- **双模态传输**：支持 QUIC (UDP) 和 TCP 两种传输层
- **自动回退**：QUIC 失败时自动回退到 TCP
- **统一接口**：向应用层暴露统一的 `net.Listener` 接口
- **gRPC 就绪**：专为 gRPC 设计的反向隧道支持
- **自签名证书**：内置证书生成工具，简化开发测试

## 架构

Qymux 是纯传输层库，类似于 `net/http` 包，专注于底层连接管理：

```
应用层 (gRPC 服务, 自定义协议)
        ↓
    Qymux (传输层)
        ↓
  QUIC / TCP (网络层)
```

**应用场景**：Qymux 适用于需要穿透 NAT 或防火墙的场景，例如：
- 反向代理
- 远程管理
- 内网穿透
- IoT 设备连接

对于 Agent 管理功能（注册、心跳、Exec、PTY 等），请使用 [Kcross](https://github.com/funcx27/kcross)。

## 安装

```bash
go get github.com/funcx27/qymux
```

## 快速开始

### Server 端

```go
package main

import (
    "github.com/funcx27/qymux/pkg/qymux"
    "github.com/funcx27/qymux/pkg/transport"
)

func main() {
    // 创建 Qymux 实例
    q := qymux.New(&qymux.Config{
        Mode: transport.ModeAuto, // 自动选择 QUIC 或 TCP
    })

    // 监听连接
    listener, err := q.Listen(":9090")
    if err != nil {
        panic(err)
    }
    defer listener.Close()

    // 接受连接
    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }

        // 在 conn 上运行你的协议 (gRPC, HTTP, 自定义协议等)
        go handleConnection(conn)
    }
}
```

### Client 端

```go
package main

import (
    "github.com/funcx27/qymux/pkg/qymux"
    "github.com/funcx27/qymux/pkg/transport"
)

func main() {
    // 创建 Qymux 实例
    q := qymux.New(&qymux.Config{
        Mode:         transport.ModeAuto,
        ServerAddr:   "example.com:9090",
    })

    // 连接到服务器
    session, err := q.Dial()
    if err != nil {
        panic(err)
    }
    defer session.Close()

    // 使用 session 创建你的协议连接
    // 例如：gRPC client 连接
}
```

## 配置

### 传输模式

```go
q := qymux.New(&qymux.Config{
    Mode: transport.ModeQUIC,  // 强制使用 QUIC
    // Mode: transport.ModeTCP,  // 强制使用 TCP
    // Mode: transport.ModeAuto, // 自动选择（推荐）
})
```

### TLS 配置

```go
q := qymux.New(&qymux.Config{
    TLSConfig: &tls.Config{
        // 自定义 TLS 配置
    },
})
```

### 连接选项

```go
q := qymux.New(&qymux.Config{
    KeepAlive:        30 * time.Second,  // 保活间隔
    ConnectTimeout:   10 * time.Second,  // 连接超时
    HandshakeTimeout: 5  * time.Second,  // 握手超时
})
```

## 工具

### 生成测试证书

```bash
make cert
```

证书将生成在 `pkg/cert/` 目录。

### 构建和测试

```bash
make build    # 构建所有包
make test     # 运行测试
make coverage # 生成覆盖率报告
```

## 项目结构

```
qymux/
├── pkg/
│   ├── qymux/       # 主要 API
│   ├── transport/   # 传输层抽象
│   ├── quic/        # QUIC 实现
│   ├── tcp/         # TCP + Yamux 实现
│   ├── dialer/      # 连接管理
│   ├── cert/        # 证书工具
│   ├── tls/         # TLS 配置
│   ├── errors/      # 错误定义
│   └── utils/       # 工具函数
└── go.mod
```

## 依赖

- `github.com/quic-go/quic-go` - QUIC 协议实现
- `github.com/hashicorp/yamux` - 多路复用
- `google.golang.org/grpc` - gRPC 支持

## 与 Kcross 的关系

| 项目 | 职责 |
|------|------|
| **Qymux** | 传输层库（类似 net/http） |
| **Kcross** | Agent 管理框架（基于 Qymux） |

- 使用 **Qymux** 如果只需要底层隧道功能
- 使用 **Kcross** 如果需要完整的 Agent 管理功能

## License

MIT
