package tls

import (
	"crypto/tls"
	"testing"
)

func TestEnsureServerTLSConfig(t *testing.T) {
	// 测试 nil 输入
	config, err := EnsureServerTLSConfig(nil)
	if err != nil {
		t.Fatalf("EnsureServerTLSConfig(nil) error = %v", err)
	}
	if config == nil {
		t.Error("EnsureServerTLSConfig(nil) should return config")
	}

	// 验证配置有效
	if len(config.Certificates) == 0 {
		t.Error("Server config should have certificates")
	}
}

func TestEnsureClientTLSConfig(t *testing.T) {
	// 测试 nil 输入
	config := EnsureClientTLSConfig(nil)
	if config == nil {
		t.Error("EnsureClientTLSConfig(nil) should return config")
	}

	// 客户端应该跳过验证
	if !config.InsecureSkipVerify {
		t.Error("Client config should have InsecureSkipVerify=true")
	}
}

func TestEnsureServerTLSConfigWithExisting(t *testing.T) {
	existing := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	config, err := EnsureServerTLSConfig(existing)
	if err != nil {
		t.Fatalf("EnsureServerTLSConfig() error = %v", err)
	}
	if config == nil {
		t.Error("EnsureServerTLSConfig() should return config")
	}

	// 应该保留原有配置
	if config.MinVersion != tls.VersionTLS12 {
		t.Error("Existing config should be preserved")
	}
}

func TestEnsureClientTLSConfigWithExisting(t *testing.T) {
	existing := &tls.Config{
		ServerName: "example.com",
	}

	config := EnsureClientTLSConfig(existing)
	if config == nil {
		t.Error("EnsureClientTLSConfig() should return config")
	}

	// 应该保留原有配置
	if config.ServerName != "example.com" {
		t.Error("Existing config should be preserved")
	}
}
