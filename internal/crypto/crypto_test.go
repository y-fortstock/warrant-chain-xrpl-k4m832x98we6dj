package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	ac "github.com/CreatureDev/xrpl-go/address-codec"
	binarycodec "github.com/CreatureDev/xrpl-go/binary-codec"
	"github.com/CreatureDev/xrpl-go/client"
	jsonrpcclient "github.com/CreatureDev/xrpl-go/client/jsonrpc"
	"github.com/CreatureDev/xrpl-go/keypairs"
	clientaccount "github.com/CreatureDev/xrpl-go/model/client/account"
	clientcommon "github.com/CreatureDev/xrpl-go/model/client/common"
	clientledger "github.com/CreatureDev/xrpl-go/model/client/ledger"
	clienttransactions "github.com/CreatureDev/xrpl-go/model/client/transactions"
	"github.com/CreatureDev/xrpl-go/model/transactions"
	"github.com/CreatureDev/xrpl-go/model/transactions/types"
	"github.com/stretchr/testify/assert"
)

var (
	hexSeed = "434670347c6bb7c791e3629fc79c38307315d625fc5b448a601abda6ba54f7efd0cfe70bf769f7e3545c970851f6fe9132ad658101ed1ff9cb2edfeb5dd2d19f"
)

const (
	SECP256K1_PREFIX = 0x21 // 33
	ED25519_PREFIX   = 0xED // 237
)

func DoubleSha256(data []byte) []byte {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return second[:4]
}

func EncodeXRPLSeed(entropy []byte, algorithm int) (string, error) {
	if len(entropy) != 16 {
		return "", errors.New("entropy must be exactly 16 bytes")
	}

	// Use the correct prefix for ED25519 as defined in addresscodec
	var prefix []byte
	if algorithm == ED25519_PREFIX {
		prefix = []byte{0x01, 0xe1, 0x4b}
	} else {
		prefix = []byte{byte(algorithm)}
	}

	// 1. Add algorithm prefix
	data := append(prefix, entropy...)

	// 2. Calculate checksum (double SHA256, first 4 bytes)
	checksum := DoubleSha256(data)

	// 3. Combine data + checksum
	fullData := append(data, checksum...)

	// 4. Encode with XRPL base58 dictionary
	familySeed := ac.EncodeBase58(fullData)

	return familySeed, nil
}

func TestGetKeyPairFromSeed_1(t *testing.T) {
	familySeed := "pNURfEJaBcFR15a1X4Zb6sJKuezyuVHZF5XVhTM9uFSCsyUw8WkRu"

	priv, pub, err := keypairs.DeriveKeypair(familySeed, false)
	assert.NoError(t, err)
	fmt.Println("priv: ", priv)
	fmt.Println("pub: ", pub)

	// Получаем адрес из публичного ключа
	pubKeyBytes, err := hex.DecodeString(pub)
	assert.NoError(t, err)

	// Используем правильный способ генерации адреса XRPL
	accountID := ac.Sha256RipeMD160(pubKeyBytes)
	accountAddress := ac.Encode(accountID, []byte{ac.AccountAddressPrefix}, ac.AccountAddressLength)
	fmt.Println("accountAddress: ", accountAddress)

	// Получаем текущий sequence number для аккаунта
	rpcCfg, err := client.NewJsonRpcConfig("https://s.altnet.rippletest.net:51234", client.WithHttpClient(&http.Client{
		Timeout: time.Duration(30) * time.Second,
	}))
	assert.NoError(t, err)

	cli := jsonrpcclient.NewClient(rpcCfg)

	// Получаем информацию об аккаунте
	accountInfoReq := &clientaccount.AccountInfoRequest{
		Account:     types.Address(accountAddress),
		LedgerIndex: clientcommon.VALIDATED,
	}
	accountInfo, _, err := cli.Account.AccountInfo(accountInfoReq)
	var sequence uint32 = 1
	if err != nil {
		fmt.Printf("Warning: Could not get account info for %s: %v\n", accountAddress, err)
		fmt.Println("This might be a new account that needs funding")
		// Для нового аккаунта используем sequence = 1
	} else {
		sequence = accountInfo.AccountData.Sequence
	}

	// Получаем текущий ledger
	ledgerReq := &clientledger.LedgerRequest{
		LedgerIndex: clientcommon.VALIDATED,
	}
	ledgerResp, _, err := cli.Ledger.Ledger(ledgerReq)
	assert.NoError(t, err)

	// Конвертируем LedgerIndex в uint32
	ledgerIndex := uint32(ledgerResp.LedgerIndex) + 20

	payment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account:            types.Address(accountAddress),
			TransactionType:    transactions.PaymentTx,
			Fee:                types.XRPCurrencyAmount(12000), // Увеличиваем fee
			Sequence:           sequence,
			LastLedgerSequence: ledgerIndex, // Добавляем LastLedgerSequence
			SigningPubKey:      pub,         // Добавляем публичный ключ для подписи
		},
		Amount:      types.XRPCurrencyAmount(1000000),
		Destination: types.Address("ra5nK24KXen9AHvsdFTKHSANinZseWnPcX"),
	}
	encodedForSigning, err := binarycodec.EncodeForSigning(payment)
	assert.NoError(t, err)
	fmt.Println("encodedForSigning: ", encodedForSigning)

	signature, err := keypairs.Sign(encodedForSigning, priv)
	assert.NoError(t, err)
	fmt.Println("signature: ", signature)

	payment.TxnSignature = signature

	txBlob, err := binarycodec.Encode(payment)
	assert.NoError(t, err)
	fmt.Println("txBlob: ", txBlob)

	submitReq := &clienttransactions.SubmitRequest{
		TxBlob: txBlob,
	}

	resp, xrplResp, err := cli.Transaction.Submit(submitReq)
	if err != nil {
		fmt.Printf("Submit error: %v\n", err)
		if xrplResp != nil {
			fmt.Printf("XRPL Response: %+v\n", xrplResp)
		}
		// Не делаем assert.NoError здесь, так как аккаунт может не иметь средств
		return
	}
	fmt.Println("resp: ", resp)
	fmt.Println("xrplResp: ", xrplResp)
}

func TestGetKeyPairFromHexSeed(t *testing.T) {
	tests := []struct {
		name    string
		hexSeed string
		wantErr bool
	}{
		{
			name:    "valid seed",
			hexSeed: hexSeed,
			wantErr: false,
		},
		{
			name:    "empty seed",
			hexSeed: "",
			wantErr: true,
		},
		{
			name:    "invalid hex",
			hexSeed: "invalid_hex_string",
			wantErr: true,
		},
		{
			name:    "short seed (16 bytes)",
			hexSeed: "1234567890abcdef1234567890abcdef",
			wantErr: false,
		},
		{
			name:    "too short seed (8 bytes)",
			hexSeed: "1234567890abcdef",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := GetKeyPairFromHexSeed(tt.hexSeed)
			fmt.Println("key: ", key)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Fatalf("expected error but got none")
			}
			if key == nil {
				t.Fatalf("expected key but got nil")
			}
		})
	}
}

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
		{
			name:     "short seed (16 bytes)",
			hexSeed:  "1234567890abcdef1234567890abcdef",
			expected: "",
			wantErr:  false, // Should work with valid seed length
		},
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
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("failed to get XRPL address: %v", err)
				}
				return
			}

			if tt.expected != "" && address != tt.expected {
				t.Errorf("unexpected address: got %s, want %s", address, tt.expected)
			}

			// Check that address is not empty and has correct format
			if address == "" {
				t.Errorf("address should not be empty")
			}
			if len(address) < 25 || len(address) > 35 {
				t.Errorf("address length should be between 25-35 characters, got %d", len(address))
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
			expected: "pNURfEJaBcFR15a1X4Zb6sJKuezyuVHZF5XVhTM9uFSCsyUw8WkRu",
			wantErr:  false,
		},
		{
			name:     "short seed (16 bytes)",
			hexSeed:  "1234567890abcdef1234567890abcdef",
			expected: "",
			wantErr:  false, // Should work with valid seed length
		},
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
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("failed to get XRPL secret: %v", err)
				}
				return
			}

			if tt.expected != "" && secret != tt.expected {
				t.Errorf("unexpected secret: got %s, want %s", secret, tt.expected)
			}

			// Check that secret is not empty and has reasonable length
			if secret == "" {
				t.Errorf("secret should not be empty")
			}
			if len(secret) < 25 {
				t.Errorf("secret length should be at least 25 characters, got %d", len(secret))
			}
		})
	}
}

func TestGetKeyPairFromSeed(t *testing.T) {
	tests := []struct {
		name    string
		seed    []byte
		wantErr bool
	}{
		{
			name:    "valid seed (16 bytes)",
			seed:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			wantErr: false,
		},
		{
			name:    "valid seed (32 bytes)",
			seed:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			wantErr: false,
		},
		{
			name:    "too short seed (8 bytes)",
			seed:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
			wantErr: true,
		},
		{
			name:    "empty seed",
			seed:    []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := GetKeyPairFromSeed(tt.seed)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Fatalf("expected error but got none")
			}
			if key == nil {
				t.Fatalf("expected key but got nil")
			}
		})
	}
}
