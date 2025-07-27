package crypto

import (
	"encoding/hex"
	"fmt"

	ac "github.com/CreatureDev/xrpl-go/address-codec"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
)

func GetKeyPairFromHexSeed(hexSeed string) (*hdkeychain.ExtendedKey, error) {
	if hexSeed == "" {
		return nil, fmt.Errorf("hex seed is empty")
	}
	seed, err := hex.DecodeString(hexSeed)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex seed: %w", err)
	}
	return GetKeyPairFromSeed(seed)
}

func GetKeyPairFromSeed(seed []byte) (*hdkeychain.ExtendedKey, error) {
	// Создаем master key с параметрами MainNet
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create master key: %w", err)
	}

	// m/44' (purpose)
	purpose, err := masterKey.Derive(hdkeychain.HardenedKeyStart + 44)
	if err != nil {
		return nil, fmt.Errorf("failed to create purpose: %w", err)
	}

	// m/44'/144' (XRP coin type)
	coinType, err := purpose.Derive(hdkeychain.HardenedKeyStart + 144)
	if err != nil {
		return nil, fmt.Errorf("failed to create coin type: %w", err)
	}

	// m/44'/144'/0' (account)
	account, err := coinType.Derive(hdkeychain.HardenedKeyStart + 0)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	// m/44'/144'/0'/0 (change - external)
	change, err := account.Derive(0)
	if err != nil {
		return nil, fmt.Errorf("failed to create change: %w", err)
	}

	// m/44'/144'/0'/0/0 (address index)
	addressKey, err := change.Derive(0)
	if err != nil {
		return nil, fmt.Errorf("failed to create address key: %w", err)
	}

	return addressKey, nil
}

func GetXRPLAddressFromKeyPair(key *hdkeychain.ExtendedKey) (string, error) {
	// Получаем публичный ключ
	pubKey, err := key.ECPubKey()
	if err != nil {
		return "", fmt.Errorf("failed to get public key: %w", err)
	}

	// Сериализуем публичный ключ в сжатом формате
	pubKeyBytes := pubKey.SerializeCompressed()

	ripemd160 := ac.Sha256RipeMD160(pubKeyBytes)
	address := ac.Encode(ripemd160,
		[]byte{ac.AccountAddressPrefix},
		ac.AccountAddressLength,
	)
	return address, nil
}

func GetXRPLSecretFromKeyPair(key *hdkeychain.ExtendedKey) (string, error) {
	// Получаем приватный ключ
	privKey, err := key.ECPrivKey()
	if err != nil {
		return "", fmt.Errorf("failed to get private key: %w", err)
	}

	// Сериализуем приватный ключ в байты
	privKeyBytes := privKey.Serialize()

	secret := ac.Encode(privKeyBytes,
		[]byte{0x01, 0xe1, 0x4b},
		32,
	)
	return secret, nil
}
