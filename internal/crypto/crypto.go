package crypto

import (
	"encoding/hex"
	"fmt"

	ac "github.com/CreatureDev/xrpl-go/address-codec"
	bip32 "github.com/tyler-smith/go-bip32"
)

func GetKeyPairFromHexSeed(hexSeed string) (*bip32.Key, error) {
	if hexSeed == "" {
		return nil, fmt.Errorf("hex seed is empty")
	}
	seed, err := hex.DecodeString(hexSeed)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex seed: %w", err)
	}
	return GetKeyPairFromSeed(seed)
}

func GetKeyPairFromSeed(seed []byte) (*bip32.Key, error) {
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to create master key: %w", err)
	}

	purpose, err := masterKey.NewChildKey(bip32.FirstHardenedChild + 44)
	if err != nil {
		return nil, fmt.Errorf("failed to create purpose: %w", err)
	}

	// 144' = hardened 144 (XRP coin type)
	coinType, err := purpose.NewChildKey(bip32.FirstHardenedChild + 144)
	if err != nil {
		return nil, fmt.Errorf("failed to create coin type: %w", err)
	}

	// 0' = hardened 0
	account, err := coinType.NewChildKey(bip32.FirstHardenedChild + 0)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	// 0 = change (external)
	change, err := account.NewChildKey(0)
	if err != nil {
		return nil, fmt.Errorf("failed to create change: %w", err)
	}

	// 0 = address index
	addressKey, err := change.NewChildKey(0)
	if err != nil {
		return nil, fmt.Errorf("failed to create address key: %w", err)
	}

	// return addressKey, nil
	return addressKey, nil
}

func GetXRPLAddressFromKeyPair(key *bip32.Key) string {
	ripemd160 := ac.Sha256RipeMD160(key.PublicKey().Key)
	address := ac.Encode(ripemd160,
		[]byte{ac.AccountAddressPrefix},
		ac.AccountAddressLength,
	)
	return address
}

func GetXRPLSecretFromKeyPair(key *bip32.Key) string {
	secret := ac.Encode(key.Key,
		[]byte{ac.ED25519Prefix},
		32,
	)
	return secret
}
