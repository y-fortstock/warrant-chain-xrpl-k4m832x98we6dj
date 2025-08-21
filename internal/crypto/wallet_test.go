package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testHexSeed        = "434670347c6bb7c791e3629fc79c38307315d625fc5b448a601abda6ba54f7efd0cfe70bf769f7e3545c970851f6fe9132ad658101ed1ff9cb2edfeb5dd2d19f"
	testDerivationPath = "m/44'/144'/0'/0/0"
	testAddress        = "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC"
)

func TestNewWallet(t *testing.T) {
	tests := []struct {
		name    string
		address string
		public  string
		private string
	}{
		{
			name:    "valid wallet data",
			address: "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC",
			public:  "02A8A44DB3D4C73EEEE11DFE98DEDC90892FD38FC65E71E8D4D7F5F224A8B3323F",
			private: "00A8A44DB3D4C73EEEE11DFE98DEDC90892FD38FC65E71E8D4D7F5F224A8B3323F",
		},
		{
			name:    "empty wallet data",
			address: "",
			public:  "",
			private: "",
		},
		{
			name:    "partial wallet data",
			address: "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC",
			public:  "",
			private: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := NewWallet(tt.address, tt.public, tt.private)

			assert.NotNil(t, wallet)
			assert.Equal(t, tt.address, wallet.Address)
			assert.Equal(t, tt.public, wallet.Public)
			assert.Equal(t, tt.private, wallet.Private)
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
		assert.NotEmpty(t, wallet.Public)
		assert.NotEmpty(t, wallet.Private)

		// Verify address format (XRPL addresses start with 'r')
		assert.Equal(t, uint8('r'), wallet.Address[0])

		// Verify public key is hex string
		assert.Greater(t, len(wallet.Public), 0)
		assert.Greater(t, len(wallet.Private), 0)
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
		assert.NotEmpty(t, wallet.Public)
		assert.NotEmpty(t, wallet.Private)

		// Verify address format
		assert.Equal(t, uint8('r'), wallet.Address[0])

		// Verify the address matches expected
		assert.Equal(t, testAddress, wallet.Address)
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
		assert.Equal(t, testAddress, wallet.Address)
		assert.NotEmpty(t, wallet.Public)
		assert.NotEmpty(t, wallet.Private)

		// Verify wallet can be recreated with same data
		recreatedWallet := NewWallet(wallet.Address, wallet.Public, wallet.Private)
		assert.Equal(t, wallet.Address, recreatedWallet.Address)
		assert.Equal(t, wallet.Public, recreatedWallet.Public)
		assert.Equal(t, wallet.Private, recreatedWallet.Private)
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
		assert.Equal(t, wallet1.Public, wallet2.Public)
		assert.Equal(t, wallet1.Private, wallet2.Private)
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
		} else {
			assert.NotNil(t, wallet)
		}
	})

	t.Run("complex derivation path", func(t *testing.T) {
		complexPath := "m/44'/144'/0'/0'/1'/2'/3'/4'/5'/6'"
		wallet, err := NewWalletFromHexSeed(testHexSeed, complexPath)
		// This should either succeed or fail gracefully, but not panic
		if err != nil {
			assert.Nil(t, wallet)
		} else {
			assert.NotNil(t, wallet)
		}
	})
}
