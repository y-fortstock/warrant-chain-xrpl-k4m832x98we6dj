package api

import (
	"fmt"
	"testing"

	"github.com/CreatureDev/xrpl-go/keypairs"
	"github.com/stretchr/testify/assert"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/config"
)

var (
	validHexSeed   = "434670347c6bb7c791e3629fc79c38307315d625fc5b448a601abda6ba54f7efd0cfe70bf769f7e3545c970851f6fe9132ad658101ed1ff9cb2edfeb5dd2d19f"
	invalidHexSeed = "invalid_hex_seed"
)

func TestBlockchain_signer(t *testing.T) {
	seed := "pNURfEJaBcFR15a1X4Zb6sJKuezyuVHZF5XVhTM9uFSCsyUw8WkRu"

	priv, pub, err := keypairs.DeriveKeypair(seed, false)
	assert.NoError(t, err)
	fmt.Println("priv: ", priv)
	fmt.Println("pub: ", pub)

	actual, err := keypairs.Sign([]byte("test"), priv)
	assert.NoError(t, err)
	fmt.Println("actual: ", actual)
}

func TestNewBlockchain(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.NetworkConfig
		wantErr bool
	}{
		{
			name: "valid network config",
			cfg: config.NetworkConfig{
				URL: "wss://s.altnet.rippletest.net:51233",
			},
			wantErr: false,
		},
		{
			name: "invalid URL",
			cfg: config.NetworkConfig{
				URL: "invalid://url",
			},
			wantErr: false,
		},
		{
			name: "empty URL",
			cfg: config.NetworkConfig{
				URL: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockchain, err := NewBlockchain(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBlockchain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && blockchain == nil {
				t.Error("NewBlockchain() returned nil blockchain when no error expected")
			}
			if !tt.wantErr && blockchain.xrplClient == nil {
				t.Error("NewBlockchain() returned blockchain with nil xrplClient")
			}
		})
	}
}

func TestBlockchain_GetXRPLAddress(t *testing.T) {
	// Создаем валидный blockchain для тестов
	cfg := config.NetworkConfig{
		URL: "wss://s.altnet.rippletest.net:51233",
	}
	blockchain, err := NewBlockchain(cfg)
	if err != nil {
		t.Fatalf("Failed to create blockchain for testing: %v", err)
	}

	tests := []struct {
		name     string
		hexSeed  string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid hex seed",
			hexSeed:  validHexSeed,
			expected: "rUWaveCdPhssfFE3SiFV811w5vvaFxy1W1",
			wantErr:  false,
		},
		{
			name:     "invalid hex seed",
			hexSeed:  invalidHexSeed,
			expected: "",
			wantErr:  true,
		},
		{
			name:     "empty hex seed",
			hexSeed:  "",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address, err := blockchain.GetXRPLAddress(tt.hexSeed)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetXRPLAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && address != tt.expected {
				t.Errorf("GetXRPLAddress() = %v, want %v", address, tt.expected)
			}
		})
	}
}

func TestBlockchain_GetXRPLSecret(t *testing.T) {
	// Создаем валидный blockchain для тестов
	cfg := config.NetworkConfig{
		URL: "wss://s.altnet.rippletest.net:51233",
	}
	blockchain, err := NewBlockchain(cfg)
	if err != nil {
		t.Fatalf("Failed to create blockchain for testing: %v", err)
	}

	tests := []struct {
		name     string
		hexSeed  string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid hex seed",
			hexSeed:  validHexSeed,
			expected: "3A5qELU2vgJ6sQBcyegzammQMFuHKDgGWPqk4mo1KJ8MvGh5ZAv",
			wantErr:  false,
		},
		{
			name:     "invalid hex seed",
			hexSeed:  invalidHexSeed,
			expected: "",
			wantErr:  true,
		},
		{
			name:     "empty hex seed",
			hexSeed:  "",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret, err := blockchain.GetXRPLSecret(tt.hexSeed)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetXRPLSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && secret != tt.expected {
				t.Errorf("GetXRPLSecret() = %v, want %v", secret, tt.expected)
			}
		})
	}
}

// Benchmark тесты для производительности
func BenchmarkBlockchain_GetXRPLAddress(b *testing.B) {
	cfg := config.NetworkConfig{
		URL: "wss://s.altnet.rippletest.net:51233",
	}
	blockchain, err := NewBlockchain(cfg)
	if err != nil {
		b.Fatalf("Failed to create blockchain for benchmarking: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := blockchain.GetXRPLAddress(validHexSeed)
		if err != nil {
			b.Fatalf("GetXRPLAddress failed: %v", err)
		}
	}
}

func BenchmarkBlockchain_GetXRPLSecret(b *testing.B) {
	cfg := config.NetworkConfig{
		URL: "wss://s.altnet.rippletest.net:51233",
	}
	blockchain, err := NewBlockchain(cfg)
	if err != nil {
		b.Fatalf("Failed to create blockchain for benchmarking: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := blockchain.GetXRPLSecret(validHexSeed)
		if err != nil {
			b.Fatalf("GetXRPLSecret failed: %v", err)
		}
	}
}
