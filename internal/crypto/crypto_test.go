package crypto

import (
	"testing"
)

var (
	hexSeed = "434670347c6bb7c791e3629fc79c38307315d625fc5b448a601abda6ba54f7efd0cfe70bf769f7e3545c970851f6fe9132ad658101ed1ff9cb2edfeb5dd2d19f"
)

func TestGetXRPLAddressFromKeyPair(t *testing.T) {
	tests := []struct {
		name     string
		hexSeed  string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid seed",
			hexSeed:  hexSeed,
			expected: "rUWaveCdPhssfFE3SiFV811w5vvaFxy1W1",
			wantErr:  false,
		},
		// Можно добавить дополнительные кейсы
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := GetKeyPairFromHexSeed(tt.hexSeed)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("failed to get key pair: %v", err)
				}
				return
			}
			address, err := GetXRPLAddressFromKeyPair(key)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetXRPLAddressFromKeyPair() error = %v, wantErr %v", err, tt.wantErr)
			}
			if address != tt.expected {
				t.Errorf("unexpected address: got %s, want %s", address, tt.expected)
			}
		})
	}
}

func TestGetXRPLSecretFromKeyPair(t *testing.T) {
	tests := []struct {
		name     string
		hexSeed  string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid seed",
			hexSeed:  hexSeed,
			expected: "3A5qELU2vgJ6sQBcyegzammQMFuHKDgGWPqk4mo1KJ8MvGh5ZAv",
			wantErr:  false,
		},
		// Можно добавить дополнительные кейсы
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := GetKeyPairFromHexSeed(tt.hexSeed)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("failed to get key pair: %v", err)
				}
				return
			}
			secret, err := GetXRPLSecretFromKeyPair(key)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetXRPLSecretFromKeyPair() error = %v, wantErr %v", err, tt.wantErr)
			}
			if secret != tt.expected {
				t.Errorf("unexpected secret: got %s, want %s", secret, tt.expected)
			}
		})
	}
}
