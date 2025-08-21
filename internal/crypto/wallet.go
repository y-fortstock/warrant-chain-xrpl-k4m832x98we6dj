package crypto

import (
	"fmt"

	"github.com/CreatureDev/xrpl-go/model/transactions/types"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
)

type Wallet struct {
	Address    types.Address
	PublicKey  string
	PrivateKey string
}

func NewWallet(address types.Address, publicKey, privateKey string) *Wallet {
	return &Wallet{Address: address, PublicKey: publicKey, PrivateKey: privateKey}
}

func NewWalletFromExtendedKey(key *hdkeychain.ExtendedKey) (*Wallet, error) {
	if key == nil {
		return nil, fmt.Errorf("extended key cannot be nil")
	}
	address, public, private, err := GetXRPLWallet(key)
	if err != nil {
		return nil, err
	}
	return NewWallet(types.Address(address), public, private), nil
}

func NewWalletFromHexSeed(hexSeed string, path string) (*Wallet, error) {
	key, err := GetExtendedKeyFromHexSeedWithPath(hexSeed, path)
	if err != nil {
		return nil, err
	}
	return NewWalletFromExtendedKey(key)
}
