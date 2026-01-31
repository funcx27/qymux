// Package tls 提供 TLS 配置的辅助函数
package tls

import (
	"crypto/tls"

	"github.com/funcx27/qymux/pkg/cert"
)

// EnsureClientTLSConfig 确保客户端 TLS 配置存在
// 如果传入的 config 为 nil，则自动生成一个客户端 TLS 配置
// 如果生成失败，则创建一个基本的配置（跳过证书验证）
func EnsureClientTLSConfig(config *tls.Config) *tls.Config {
	if config != nil {
		return config
	}

	clientConfig, err := cert.GenerateClientTLSConfig()
	if err != nil {
		// 如果生成失败，创建一个基本的配置
		clientConfig = &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{cert.ALPN},
		}
	}

	return clientConfig
}

// EnsureServerTLSConfig 确保服务端 TLS 配置存在
// 如果传入的 config 为 nil，则自动生成一个自签名证书的服务端 TLS 配置
func EnsureServerTLSConfig(config *tls.Config) (*tls.Config, error) {
	if config != nil {
		return config, nil
	}

	return cert.GenerateSelfSignedConfig()
}
