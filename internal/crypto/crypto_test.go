package crypto

import (
	"encoding/hex"
	"fmt"
	"testing"

	ac "github.com/CreatureDev/xrpl-go/address-codec"
	"github.com/CreatureDev/xrpl-go/keypairs"
	"github.com/decen-one/go-bip39"
	"github.com/stretchr/testify/assert"
)

var (
	hexSeed        = "434670347c6bb7c791e3629fc79c38307315d625fc5b448a601abda6ba54f7efd0cfe70bf769f7e3545c970851f6fe9132ad658101ed1ff9cb2edfeb5dd2d19f"
	derivationPath = "m/44'/144'/0'/0/0"
	address        = "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC"
)

func TestMnemonic(t *testing.T) {
	words, err := bip39.NewRandMnemonic("english", 12)
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
	generatedAddress := ac.Encode(accountID, []byte{ac.AccountAddressPrefix}, ac.AccountAddressLength)

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
