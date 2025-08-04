package crypto

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	ac "github.com/CreatureDev/xrpl-go/address-codec"
	"github.com/CreatureDev/xrpl-go/keypairs"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
)

func GetExtendedKeyFromHexSeedWithPath(hexSeed string, path string) (*hdkeychain.ExtendedKey, error) {
	seed, err := hex.DecodeString(hexSeed)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex seed: %w", err)
	}
	return GetExtendedKeyFromSeedWithPath(seed, path)
}

func GetExtendedKeyFromSeedWithPath(seed []byte, path string) (*hdkeychain.ExtendedKey, error) {
	// Создаем master key с параметрами MainNet
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create master key: %w", err)
	}

	// Парсим путь BIP-44
	derivationPath, err := parseDerivationPath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse derivation path: %w", err)
	}

	// Проходим по всем уровням пути
	currentKey := masterKey
	for i, childIndex := range derivationPath {
		currentKey, err = currentKey.Derive(childIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to derive key at level %d (index %d): %w", i, childIndex, err)
		}
	}

	return currentKey, nil
}

// parseDerivationPath парсит строку пути BIP-44 в массив индексов
func parseDerivationPath(path string) ([]uint32, error) {
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}

	// Убираем префикс "m/" если он есть
	if len(path) >= 2 && path[:2] == "m/" {
		path = path[2:]
	}

	// Разбиваем путь по "/"
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid path format")
	}

	derivationPath := make([]uint32, len(parts))
	for i, part := range parts {
		// Проверяем на hardened derivation (с апострофом)
		hardened := false
		if strings.HasSuffix(part, "'") {
			hardened = true
			part = part[:len(part)-1]
		}

		// Парсим число
		index, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid path component %s: %w", part, err)
		}

		if hardened {
			derivationPath[i] = hdkeychain.HardenedKeyStart + uint32(index)
		} else {
			derivationPath[i] = uint32(index)
		}
	}

	return derivationPath, nil
}

func GetXRPLWallet(key *hdkeychain.ExtendedKey) (address string, private string, err error) {
	secret, err := getXRPLSecret(key)
	if err != nil {
		return "", "", fmt.Errorf("failed to get secret from key: %w", err)
	}

	privKey, pubKeyHex, err := keypairs.DeriveKeypair(secret, false)
	if err != nil {
		return "", "", fmt.Errorf("failed to derive keypair: %w", err)
	}

	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode public key: %w", err)
	}

	accountID := ac.Sha256RipeMD160(pubKeyBytes)
	address = ac.Encode(accountID, []byte{ac.AccountAddressPrefix}, ac.AccountAddressLength)
	return address, privKey, nil
}

func getXRPLSecret(key *hdkeychain.ExtendedKey) (string, error) {
	privKey, err := key.ECPrivKey()
	if err != nil {
		return "", fmt.Errorf("failed to get private key: %w", err)
	}

	privKeyBytes := privKey.Serialize()

	secret := ac.Encode(privKeyBytes,
		[]byte{0x01, 0xe1, 0x4b},
		32,
	)
	return secret, nil
}
