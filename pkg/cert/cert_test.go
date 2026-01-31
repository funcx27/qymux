package cert

import "testing"

func TestGenerateSelfSignedConfig(t *testing.T) {
	config, err := GenerateSelfSignedConfig()
	if err != nil {
		t.Fatalf("GenerateSelfSignedConfig() error = %v", err)
	}

	if config == nil {
		t.Error("GenerateSelfSignedConfig() returned nil config")
	}

	if len(config.Certificates) == 0 {
		t.Error("Config should have certificates")
	}
}

func TestGenerateClientTLSConfig(t *testing.T) {
	config, err := GenerateClientTLSConfig()
	if err != nil {
		t.Fatalf("GenerateClientTLSConfig() error = %v", err)
	}

	if config == nil {
		t.Error("GenerateClientTLSConfig() returned nil config")
	}

	// 客户端应该跳过验证
	if !config.InsecureSkipVerify {
		t.Error("Client config should have InsecureSkipVerify=true")
	}
}

func TestALPNConstant(t *testing.T) {
	if ALPN == "" {
		t.Error("ALPN constant should not be empty")
	}
}
