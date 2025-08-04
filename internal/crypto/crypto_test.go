package crypto

import (
	"encoding/hex"
	"fmt"
	"testing"

	ac "github.com/CreatureDev/xrpl-go/address-codec"
	"github.com/CreatureDev/xrpl-go/keypairs"
	"github.com/stretchr/testify/assert"
)

var (
	hexSeed        = "434670347c6bb7c791e3629fc79c38307315d625fc5b448a601abda6ba54f7efd0cfe70bf769f7e3545c970851f6fe9132ad658101ed1ff9cb2edfeb5dd2d19f"
	derivationPath = "m/44'/144'/0'/0/0"
	address        = "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC"
)

const (
	SECP256K1_PREFIX = 0x21 // 33
	ED25519_PREFIX   = 0xED // 237
)

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

	walletAddress, privateKey, err := GetXRPLWallet(key)
	assert.NoError(t, err)
	assert.NotEmpty(t, walletAddress)
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
	walletAddress, privateKey, err := GetXRPLWallet(key)
	assert.NoError(t, err)
	assert.NotEmpty(t, walletAddress)
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

// func TestGetKeyPairFromSeed_1(t *testing.T) {
// 	familySeed := "pNURfEJaBcFR15a1X4Zb6sJKuezyuVHZF5XVhTM9uFSCsyUw8WkRu"

// 	priv, pub, err := keypairs.DeriveKeypair(familySeed, false)
// 	assert.NoError(t, err)
// 	fmt.Println("priv: ", priv)
// 	fmt.Println("pub: ", pub)

// 	// Получаем адрес из публичного ключа
// 	pubKeyBytes, err := hex.DecodeString(pub)
// 	assert.NoError(t, err)

// 	// Используем правильный способ генерации адреса XRPL
// 	accountID := ac.Sha256RipeMD160(pubKeyBytes)
// 	accountAddress := ac.Encode(accountID, []byte{ac.AccountAddressPrefix}, ac.AccountAddressLength)
// 	fmt.Println("accountAddress: ", accountAddress)

// 	// Получаем текущий sequence number для аккаунта
// 	rpcCfg, err := client.NewJsonRpcConfig("https://s.altnet.rippletest.net:51234", client.WithHttpClient(&http.Client{
// 		Timeout: time.Duration(30) * time.Second,
// 	}))
// 	assert.NoError(t, err)

// 	cli := jsonrpcclient.NewClient(rpcCfg)

// 	// Получаем информацию об аккаунте
// 	accountInfoReq := &clientaccount.AccountInfoRequest{
// 		Account:     types.Address(accountAddress),
// 		LedgerIndex: clientcommon.VALIDATED,
// 	}
// 	accountInfo, _, err := cli.Account.AccountInfo(accountInfoReq)
// 	var sequence uint32 = 1
// 	if err != nil {
// 		fmt.Printf("Warning: Could not get account info for %s: %v\n", accountAddress, err)
// 		fmt.Println("This might be a new account that needs funding")
// 		// Для нового аккаунта используем sequence = 1
// 	} else {
// 		sequence = accountInfo.AccountData.Sequence
// 	}

// 	// Получаем текущий ledger
// 	ledgerReq := &clientledger.LedgerRequest{
// 		LedgerIndex: clientcommon.VALIDATED,
// 	}
// 	ledgerResp, _, err := cli.Ledger.Ledger(ledgerReq)
// 	assert.NoError(t, err)

// 	// Конвертируем LedgerIndex в uint32
// 	ledgerIndex := uint32(ledgerResp.LedgerIndex) + 20

// 	payment := &transactions.Payment{
// 		BaseTx: transactions.BaseTx{
// 			Account:            types.Address(accountAddress),
// 			TransactionType:    transactions.PaymentTx,
// 			Fee:                types.XRPCurrencyAmount(12000), // Увеличиваем fee
// 			Sequence:           sequence,
// 			LastLedgerSequence: ledgerIndex, // Добавляем LastLedgerSequence
// 			SigningPubKey:      pub,         // Добавляем публичный ключ для подписи
// 		},
// 		Amount:      types.XRPCurrencyAmount(1000000),
// 		Destination: types.Address("ra5nK24KXen9AHvsdFTKHSANinZseWnPcX"),
// 	}
// 	encodedForSigning, err := binarycodec.EncodeForSigning(payment)
// 	assert.NoError(t, err)
// 	fmt.Println("encodedForSigning: ", encodedForSigning)

// 	signature, err := keypairs.Sign(encodedForSigning, priv)
// 	assert.NoError(t, err)
// 	fmt.Println("signature: ", signature)

// 	payment.TxnSignature = signature

// 	txBlob, err := binarycodec.Encode(payment)
// 	assert.NoError(t, err)
// 	fmt.Println("txBlob: ", txBlob)

// 	submitReq := &clienttransactions.SubmitRequest{
// 		TxBlob: txBlob,
// 	}

// 	resp, xrplResp, err := cli.Transaction.Submit(submitReq)
// 	if err != nil {
// 		fmt.Printf("Submit error: %v\n", err)
// 		if xrplResp != nil {
// 			fmt.Printf("XRPL Response: %+v\n", xrplResp)
// 		}
// 		// Не делаем assert.NoError здесь, так как аккаунт может не иметь средств
// 		return
// 	}
// 	fmt.Println("resp: ", resp)
// 	fmt.Println("xrplResp: ", xrplResp)
// }
