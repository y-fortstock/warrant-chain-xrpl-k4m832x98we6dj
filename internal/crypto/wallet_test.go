package crypto

import (
	"testing"

	"github.com/CreatureDev/xrpl-go/model/transactions/types"
	"github.com/stretchr/testify/assert"
)

var (
	testHexSeed        = "434670347c6bb7c791e3629fc79c38307315d625fc5b448a601abda6ba54f7efd0cfe70bf769f7e3545c970851f6fe9132ad658101ed1ff9cb2edfeb5dd2d19f"
	testDerivationPath = "m/44'/144'/0'/0/0"
	testAddress        = "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC"
)

func TestNewWallet(t *testing.T) {
	tests := []struct {
		name          string
		address       string
		public        string
		private       string
		expectError   bool
		expectedError string
	}{
		{
			name:        "valid wallet data",
			address:     "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC",
			public:      "02A8A44DB3D4C73EEEE11DFE98DEDC90892FD38FC65E71E8D4D7F5F224A8B3323F",
			private:     "00A8A44DB3D4C73EEEE11DFE98DEDC90892FD38FC65E71E8D4D7F5F224A8B3323F",
			expectError: false,
		},
		{
			name:          "empty wallet data",
			address:       "",
			public:        "",
			private:       "",
			expectError:   true,
			expectedError: "wallet address cannot be empty",
		},
		{
			name:          "partial wallet data - empty address",
			address:       "",
			public:        "02A8A44DB3D4C73EEEE11DFE98DEDC90892FD38FC65E71E8D4D7F5F224A8B3323F",
			private:       "00A8A44DB3D4C73EEEE11DFE98DEDC90892FD38FC65E71E8D4D7F5F224A8B3323F",
			expectError:   true,
			expectedError: "wallet address cannot be empty",
		},
		{
			name:          "partial wallet data - empty public key",
			address:       "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC",
			public:        "",
			private:       "00A8A44DB3D4C73EEEE11DFE98DEDC90892FD38FC65E71E8D4D7F5F224A8B3323F",
			expectError:   true,
			expectedError: "wallet public key cannot be empty",
		},
		{
			name:          "partial wallet data - empty private key",
			address:       "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC",
			public:        "02A8A44DB3D4C73EEEE11DFE98DEDC90892FD38FC65E71E8D4D7F5F224A8B3323F",
			private:       "",
			expectError:   true,
			expectedError: "wallet private key cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet, err := NewWallet(types.Address(tt.address), tt.public, tt.private)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, wallet)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, wallet)
				assert.Equal(t, types.Address(tt.address), wallet.Address)
				assert.Equal(t, tt.public, wallet.PublicKey)
				assert.Equal(t, tt.private, wallet.PrivateKey)
			}
		})
	}
}

func TestNewWalletFromExtendedKey(t *testing.T) {
	t.Run("valid extended key", func(t *testing.T) {
		// Create a valid extended key first
		key, err := GetExtendedKeyFromHexSeedWithPath(testHexSeed, testDerivationPath)
		assert.NoError(t, err)
		assert.NotNil(t, key)

		// Test creating wallet from extended key
		wallet, err := NewWalletFromExtendedKey(key)
		assert.NoError(t, err)
		assert.NotNil(t, wallet)

		// Verify wallet fields are populated
		assert.NotEmpty(t, wallet.Address)
		assert.NotEmpty(t, wallet.PublicKey)
		assert.NotEmpty(t, wallet.PrivateKey)

		// Verify address format (XRPL addresses start with 'r')
		assert.Equal(t, uint8('r'), wallet.Address[0])

		// Verify public key is hex string
		assert.Greater(t, len(wallet.PublicKey), 0)
		assert.Greater(t, len(wallet.PrivateKey), 0)
	})

	t.Run("nil extended key", func(t *testing.T) {
		wallet, err := NewWalletFromExtendedKey(nil)
		assert.Error(t, err)
		assert.Nil(t, wallet)
	})
}

func TestNewWalletFromHexSeed(t *testing.T) {
	t.Run("valid hex seed and path", func(t *testing.T) {
		wallet, err := NewWalletFromHexSeed(testHexSeed, testDerivationPath)
		assert.NoError(t, err)
		assert.NotNil(t, wallet)

		// Verify wallet fields are populated
		assert.NotEmpty(t, wallet.Address)
		assert.NotEmpty(t, wallet.PublicKey)
		assert.NotEmpty(t, wallet.PrivateKey)

		// Verify address format
		assert.Equal(t, uint8('r'), wallet.Address[0])

		// Verify the address matches expected
		assert.Equal(t, types.Address(testAddress), wallet.Address)
	})

	t.Run("invalid hex seed", func(t *testing.T) {
		wallet, err := NewWalletFromHexSeed("invalid_hex", testDerivationPath)
		assert.Error(t, err)
		assert.Nil(t, wallet)
	})

	t.Run("empty hex seed", func(t *testing.T) {
		wallet, err := NewWalletFromHexSeed("", testDerivationPath)
		assert.Error(t, err)
		assert.Nil(t, wallet)
	})

	t.Run("invalid derivation path", func(t *testing.T) {
		wallet, err := NewWalletFromHexSeed(testHexSeed, "invalid/path")
		assert.Error(t, err)
		assert.Nil(t, wallet)
	})

	t.Run("empty derivation path", func(t *testing.T) {
		wallet, err := NewWalletFromHexSeed(testHexSeed, "")
		assert.Error(t, err)
		assert.Nil(t, wallet)
	})
}

func TestWalletIntegration(t *testing.T) {
	t.Run("full wallet creation flow", func(t *testing.T) {
		// Test the complete flow from hex seed to wallet
		wallet, err := NewWalletFromHexSeed(testHexSeed, testDerivationPath)
		assert.NoError(t, err)
		assert.NotNil(t, wallet)

		// Verify all wallet components
		assert.Equal(t, types.Address(testAddress), wallet.Address)
		assert.NotEmpty(t, wallet.PublicKey)
		assert.NotEmpty(t, wallet.PrivateKey)

		// Verify wallet can be recreated with same data
		recreatedWallet, err := NewWallet(wallet.Address, wallet.PublicKey, wallet.PrivateKey)
		assert.NoError(t, err)
		assert.Equal(t, wallet.Address, recreatedWallet.Address)
		assert.Equal(t, wallet.PublicKey, recreatedWallet.PublicKey)
		assert.Equal(t, wallet.PrivateKey, recreatedWallet.PrivateKey)
	})

	t.Run("wallet consistency", func(t *testing.T) {
		// Create wallet using hex seed method
		wallet1, err := NewWalletFromHexSeed(testHexSeed, testDerivationPath)
		assert.NoError(t, err)

		// Create wallet using extended key method
		key, err := GetExtendedKeyFromHexSeedWithPath(testHexSeed, testDerivationPath)
		assert.NoError(t, err)
		wallet2, err := NewWalletFromExtendedKey(key)
		assert.NoError(t, err)

		// Both wallets should be identical
		assert.Equal(t, wallet1.Address, wallet2.Address)
		assert.Equal(t, wallet1.PublicKey, wallet2.PublicKey)
		assert.Equal(t, wallet1.PrivateKey, wallet2.PrivateKey)
	})
}

func TestWalletEdgeCases(t *testing.T) {
	t.Run("very long hex seed", func(t *testing.T) {
		longSeed := "434670347c6bb7c791e3629fc79c38307315d625fc5b448a601abda6ba54f7efd0cfe70bf769f7e3545c970851f6fe9132ad658101ed1ff9cb2edfeb5dd2d19f" +
			"434670347c6bb7c791e3629fc79c38307315d625fc5b448a601abda6ba54f7efd0cfe70bf769f7e3545c970851f6fe9132ad658101ed1ff9cb2edfeb5dd2d19f"

		wallet, err := NewWalletFromHexSeed(longSeed, testDerivationPath)
		// This should either succeed or fail gracefully, but not panic
		if err != nil {
			assert.Nil(t, wallet)
			// Log the error for debugging but don't fail the test
			t.Logf("Expected error for long seed: %v", err)
		} else {
			assert.NotNil(t, wallet)
			// Verify the wallet has valid data
			assert.NotEmpty(t, wallet.Address)
			assert.NotEmpty(t, wallet.PublicKey)
			assert.NotEmpty(t, wallet.PrivateKey)
		}
	})

	t.Run("complex derivation path", func(t *testing.T) {
		complexPath := "m/44'/144'/0'/0'/1'/2'/3'/4'/5'/6'"
		wallet, err := NewWalletFromHexSeed(testHexSeed, complexPath)
		// This should either succeed or fail gracefully, but not panic
		if err != nil {
			assert.Nil(t, wallet)
			// Log the error for debugging but don't fail the test
			t.Logf("Expected error for complex path: %v", err)
		} else {
			assert.NotNil(t, wallet)
			// Verify the wallet has valid data
			assert.NotEmpty(t, wallet.Address)
			assert.NotEmpty(t, wallet.PublicKey)
			assert.NotEmpty(t, wallet.PrivateKey)
		}
	})

	t.Run("malformed hex seed", func(t *testing.T) {
		malformedSeed := "not_a_hex_string"
		wallet, err := NewWalletFromHexSeed(malformedSeed, testDerivationPath)
		assert.Error(t, err)
		assert.Nil(t, wallet)
	})

	t.Run("invalid derivation path format", func(t *testing.T) {
		invalidPath := "invalid/path/format"
		wallet, err := NewWalletFromHexSeed(testHexSeed, invalidPath)
		assert.Error(t, err)
		assert.Nil(t, wallet)
	})
}
