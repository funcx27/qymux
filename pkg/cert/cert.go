package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"
)

const (
	// CertValidityDays 证书有效期（天）
	CertValidityDays = 3650 // 10年

	// ALPN qymux 协议标识
	ALPN = "qymux"
)

// GenerateSelfSignedConfig 生成自签名 TLS 配置
func GenerateSelfSignedConfig() (*tls.Config, error) {
	// 生成 RSA 2048 密钥对
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// 准备证书模板
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Qymux"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(CertValidityDays * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost"},
	}

	// 自签名证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	// 创建 TLS 配置
	cert := tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  priv,
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{ALPN},
	}, nil
}

// GenerateClientTLSConfig 生成客户端 TLS 配置（自动回退模式使用）
func GenerateClientTLSConfig() (*tls.Config, error) {
	// 客户端也生成自签名证书（为了双向认证能力）
	config, err := GenerateSelfSignedConfig()
	if err != nil {
		return nil, err
	}

	// 自动回退模式下，客户端跳过证书验证
	config.InsecureSkipVerify = true

	return config, nil
}

// CertPair 表示证书对
type CertPair struct {
	CertPEM []byte
	KeyPEM  []byte
}

// GenerateCertPair 生成并导出证书对为 PEM 格式
func GenerateCertPair() (*CertPair, error) {
	// 生成 RSA 2048 密钥对
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// 准备证书模板
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Qymux"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(CertValidityDays * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// 自签名证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	// 编码为 PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	return &CertPair{
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
	}, nil
}

// LoadCertPair 从 PEM 字节数据加载证书对
func LoadCertPair(certPEM, keyPEM []byte) (tls.Certificate, error) {
	return tls.X509KeyPair(certPEM, keyPEM)
}

// LoadCertConfig 从 PEM 字节数据加载 TLS 配置
func LoadCertConfig(certPEM, keyPEM []byte, isClient bool) (*tls.Config, error) {
	cert, err := LoadCertPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{ALPN},
	}

	if isClient {
		config.InsecureSkipVerify = true
	}

	return config, nil
}
