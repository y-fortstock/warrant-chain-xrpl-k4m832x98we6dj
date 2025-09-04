package crypto

import (
	"encoding/hex"
	"fmt"
	"testing"

	ac "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/keypairs"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/decen-one/go-bip39"
	"github.com/stretchr/testify/assert"
)

var (
	hexSeed        = "434670347c6bb7c791e3629fc79c38307315d625fc5b448a601abda6ba54f7efd0cfe70bf769f7e3545c970851f6fe9132ad658101ed1ff9cb2edfeb5dd2d19f"
	derivationPath = "m/44'/144'/0'/0/0"
	address        = "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC"
)

func TestMnemonic(t *testing.T) {
	ent, err := bip39.NewEntropy(128)
	assert.NoError(t, err)
	words, err := bip39.NewMnemonic("english", ent)
	assert.NoError(t, err)
	fmt.Println(words)

	seed, err := bip39.NewSeedWithErrorChecking("english", words, "password")
	assert.NoError(t, err)
	fmt.Println(hex.EncodeToString(seed))
}

// TestGetExtendedKeyFromHexSeedWithPath тестирует получение расширенного ключа из hex seed
func TestGetExtendedKeyFromHexSeedWithPath(t *testing.T) {
	key, err := GetExtendedKeyFromHexSeedWithPath(hexSeed, derivationPath)
	assert.NoError(t, err)
	assert.NotNil(t, key)

	// Проверяем, что ключ можно сериализовать
	serialized := key.String()
	assert.NotEmpty(t, serialized)
}

// TestGetExtendedKeyFromSeedWithPath тестирует получение расширенного ключа из seed байтов
func TestGetExtendedKeyFromSeedWithPath(t *testing.T) {
	seed, err := hex.DecodeString(hexSeed)
	assert.NoError(t, err)

	key, err := GetExtendedKeyFromSeedWithPath(seed, derivationPath)
	assert.NoError(t, err)
	assert.NotNil(t, key)
}

// TestParseDerivationPath тестирует парсинг пути деривации
func TestParseDerivationPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []uint32
		hasError bool
	}{
		{
			name:     "valid path with m/ prefix",
			path:     "m/44'/144'/0'/0/0",
			expected: []uint32{2147483692, 2147483792, 2147483648, 0, 0},
			hasError: false,
		},
		{
			name:     "valid path without m/ prefix",
			path:     "44'/144'/0'/0/0",
			expected: []uint32{2147483692, 2147483792, 2147483648, 0, 0},
			hasError: false,
		},
		{
			name:     "path with mixed hardened and normal",
			path:     "m/44'/144'/0'/0/1",
			expected: []uint32{2147483692, 2147483792, 2147483648, 0, 1},
			hasError: false,
		},
		{
			name:     "empty path",
			path:     "",
			expected: nil,
			hasError: true,
		},
		{
			name:     "invalid path component",
			path:     "m/44'/abc'/0'/0/0",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDerivationPath(tt.path)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestGetXRPLWallet тестирует получение XRPL кошелька из расширенного ключа
func TestGetXRPLWallet(t *testing.T) {
	key, err := GetExtendedKeyFromHexSeedWithPath(hexSeed, derivationPath)
	assert.NoError(t, err)

	walletAddress, publicKey, privateKey, err := GetXRPLWallet(key)
	// fmt.Println("walletAddress: ", walletAddress)
	// fmt.Println("publicKey: ", publicKey)
	// fmt.Println("privateKey: ", privateKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, walletAddress)
	assert.NotEmpty(t, publicKey)
	assert.NotEmpty(t, privateKey)

	// Проверяем, что адрес начинается с 'r' (XRPL адрес)
	assert.Equal(t, uint8('r'), walletAddress[0])

	// Проверяем, что приватный ключ можно использовать для получения публичного ключа
	// Используем секрет вместо приватного ключа напрямую
	secret, err := getXRPLSecret(key)
	assert.NoError(t, err)
	_, pubKey, err := keypairs.DeriveKeypair(secret, false)
	assert.NoError(t, err)
	assert.NotEmpty(t, pubKey)
}

// TestGetXRPLSecret тестирует получение XRPL секрета из расширенного ключа
func TestGetXRPLSecret(t *testing.T) {
	key, err := GetExtendedKeyFromHexSeedWithPath(hexSeed, derivationPath)
	assert.NoError(t, err)

	secret, err := getXRPLSecret(key)
	assert.NoError(t, err)
	assert.NotEmpty(t, secret)

	// Проверяем, что секрет можно использовать для получения ключевой пары
	privKey, pubKey, err := keypairs.DeriveKeypair(secret, false)
	assert.NoError(t, err)
	assert.NotEmpty(t, privKey)
	assert.NotEmpty(t, pubKey)
}

// TestFullDerivationFlow тестирует полный процесс деривации адреса из hexSeed
func TestFullDerivationFlow(t *testing.T) {
	// Получаем расширенный ключ из hex seed
	key, err := GetExtendedKeyFromHexSeedWithPath(hexSeed, derivationPath)
	assert.NoError(t, err)
	assert.NotNil(t, key)

	// Получаем XRPL кошелек
	walletAddress, publicKey, privateKey, err := GetXRPLWallet(key)
	assert.NoError(t, err)
	assert.NotEmpty(t, walletAddress)
	assert.NotEmpty(t, publicKey)
	assert.NotEmpty(t, privateKey)

	// Проверяем, что адрес соответствует ожидаемому
	fmt.Printf("Derived address: %s\n", walletAddress)
	fmt.Printf("Expected address: %s\n", address)

	// Проверяем, что адрес начинается с 'r' (XRPL адрес)
	assert.Equal(t, uint8('r'), walletAddress[0])

	// Проверяем, что приватный ключ работает через секрет
	secret, err := getXRPLSecret(key)
	assert.NoError(t, err)
	_, pubKey, err := keypairs.DeriveKeypair(secret, false)
	assert.NoError(t, err)

	// Проверяем, что публичный ключ можно декодировать
	pubKeyBytes, err := hex.DecodeString(pubKey)
	assert.NoError(t, err)

	// Генерируем адрес из публичного ключа
	accountID := ac.Sha256RipeMD160(pubKeyBytes)
	generatedAddress, err := ac.Encode(accountID, []byte{ac.AccountAddressPrefix}, ac.AccountAddressLength)
	if err != nil {
		t.Fatalf("Failed to encode address: %v", err)
	}

	// Адрес должен совпадать с полученным из кошелька
	assert.Equal(t, walletAddress, generatedAddress)

	// Проверяем, что полученный адрес совпадает с ожидаемым
	assert.Equal(t, address, walletAddress)
}

// TestInvalidInputs тестирует обработку некорректных входных данных
func TestInvalidInputs(t *testing.T) {
	// Тест с некорректным hex seed
	_, err := GetExtendedKeyFromHexSeedWithPath("invalid_hex", derivationPath)
	assert.Error(t, err)

	// Тест с некорректным путем деривации
	_, err = GetExtendedKeyFromHexSeedWithPath(hexSeed, "invalid/path")
	assert.Error(t, err)

	// Тест с пустым hex seed
	_, err = GetExtendedKeyFromHexSeedWithPath("", derivationPath)
	assert.Error(t, err)
}

func TestNewWalletFromExtendedKey(t *testing.T) {
	t.Run("valid extended key", func(t *testing.T) {
		// Create a valid extended key first
		key, err := GetExtendedKeyFromHexSeedWithPath(hexSeed, derivationPath)
		assert.NoError(t, err)
		assert.NotNil(t, key)

		// Test creating wallet from extended key
		wallet, err := NewWalletFromExtendedKey(key)
		assert.NoError(t, err)
		assert.NotNil(t, wallet)

		// Verify wallet fields are populated
		assert.NotEmpty(t, wallet.ClassicAddress)
		assert.NotEmpty(t, wallet.PublicKey)
		assert.NotEmpty(t, wallet.PrivateKey)

		// Verify address format (XRPL addresses start with 'r')
		assert.Equal(t, uint8('r'), wallet.ClassicAddress[0])

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
		wallet, err := NewWalletFromHexSeed(hexSeed, derivationPath)
		assert.NoError(t, err)
		assert.NotNil(t, wallet)

		// Verify wallet fields are populated
		assert.NotEmpty(t, wallet.ClassicAddress)
		assert.NotEmpty(t, wallet.PublicKey)
		assert.NotEmpty(t, wallet.PrivateKey)

		// Verify address format
		assert.Equal(t, uint8('r'), wallet.ClassicAddress[0])

		// Verify the address matches expected
		assert.Equal(t, types.Address(address), wallet.ClassicAddress)
	})

	t.Run("invalid hex seed", func(t *testing.T) {
		wallet, err := NewWalletFromHexSeed("invalid_hex", derivationPath)
		assert.Error(t, err)
		assert.Nil(t, wallet)
	})

	t.Run("empty hex seed", func(t *testing.T) {
		wallet, err := NewWalletFromHexSeed("", derivationPath)
		assert.Error(t, err)
		assert.Nil(t, wallet)
	})

	t.Run("invalid derivation path", func(t *testing.T) {
		wallet, err := NewWalletFromHexSeed(hexSeed, "invalid/path")
		assert.Error(t, err)
		assert.Nil(t, wallet)
	})

	t.Run("empty derivation path", func(t *testing.T) {
		wallet, err := NewWalletFromHexSeed(hexSeed, "")
		assert.Error(t, err)
		assert.Nil(t, wallet)
	})
}

func TestWalletIntegration(t *testing.T) {
	t.Run("full wallet creation flow", func(t *testing.T) {
		// Test the complete flow from hex seed to wallet
		wallet, err := NewWalletFromHexSeed(hexSeed, derivationPath)
		assert.NoError(t, err)
		assert.NotNil(t, wallet)

		// Verify all wallet components
		assert.Equal(t, types.Address(address), wallet.ClassicAddress)
		assert.NotEmpty(t, wallet.PublicKey)
		assert.NotEmpty(t, wallet.PrivateKey)

		// Verify wallet can be recreated with same data
		recreatedWallet, err := NewWallet(wallet.ClassicAddress, wallet.PublicKey, wallet.PrivateKey)
		assert.NoError(t, err)
		assert.Equal(t, wallet.ClassicAddress, recreatedWallet.ClassicAddress)
		assert.Equal(t, wallet.PublicKey, recreatedWallet.PublicKey)
		assert.Equal(t, wallet.PrivateKey, recreatedWallet.PrivateKey)
	})

	t.Run("wallet consistency", func(t *testing.T) {
		// Create wallet using hex seed method
		wallet1, err := NewWalletFromHexSeed(hexSeed, derivationPath)
		assert.NoError(t, err)

		// Create wallet using extended key method
		key, err := GetExtendedKeyFromHexSeedWithPath(hexSeed, derivationPath)
		assert.NoError(t, err)
		wallet2, err := NewWalletFromExtendedKey(key)
		assert.NoError(t, err)

		// Both wallets should be identical
		assert.Equal(t, wallet1.ClassicAddress, wallet2.ClassicAddress)
		assert.Equal(t, wallet1.PublicKey, wallet2.PublicKey)
		assert.Equal(t, wallet1.PrivateKey, wallet2.PrivateKey)
	})
}

func TestWalletEdgeCases(t *testing.T) {
	t.Run("very long hex seed", func(t *testing.T) {
		longSeed := "434670347c6bb7c791e3629fc79c38307315d625fc5b448a601abda6ba54f7efd0cfe70bf769f7e3545c970851f6fe9132ad658101ed1ff9cb2edfeb5dd2d19f" +
			"434670347c6bb7c791e3629fc79c38307315d625fc5b448a601abda6ba54f7efd0cfe70bf769f7e3545c970851f6fe9132ad658101ed1ff9cb2edfeb5dd2d19f"

		wallet, err := NewWalletFromHexSeed(longSeed, derivationPath)
		// This should either succeed or fail gracefully, but not panic
		if err != nil {
			assert.Nil(t, wallet)
			// Log the error for debugging but don't fail the test
			t.Logf("Expected error for long seed: %v", err)
		} else {
			assert.NotNil(t, wallet)
			// Verify the wallet has valid data
			assert.NotEmpty(t, wallet.ClassicAddress)
			assert.NotEmpty(t, wallet.PublicKey)
			assert.NotEmpty(t, wallet.PrivateKey)
		}
	})

	t.Run("complex derivation path", func(t *testing.T) {
		complexPath := "m/44'/144'/0'/0'/1'/2'/3'/4'/5'/6'"
		wallet, err := NewWalletFromHexSeed(hexSeed, complexPath)
		// This should either succeed or fail gracefully, but not panic
		if err != nil {
			assert.Nil(t, wallet)
			// Log the error for debugging but don't fail the test
			t.Logf("Expected error for complex path: %v", err)
		} else {
			assert.NotNil(t, wallet)
			// Verify the wallet has valid data
			assert.NotEmpty(t, wallet.ClassicAddress)
			assert.NotEmpty(t, wallet.PublicKey)
			assert.NotEmpty(t, wallet.PrivateKey)
		}
	})

	t.Run("malformed hex seed", func(t *testing.T) {
		malformedSeed := "not_a_hex_string"
		wallet, err := NewWalletFromHexSeed(malformedSeed, derivationPath)
		assert.Error(t, err)
		assert.Nil(t, wallet)
	})

	t.Run("invalid derivation path format", func(t *testing.T) {
		invalidPath := "invalid/path/format"
		wallet, err := NewWalletFromHexSeed(hexSeed, invalidPath)
		assert.Error(t, err)
		assert.Nil(t, wallet)
	})
}
