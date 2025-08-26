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

func NewWallet(address types.Address, publicKey, privateKey string) (*Wallet, error) {
	w := &Wallet{Address: address, PublicKey: publicKey, PrivateKey: privateKey}
	if err := w.Validate(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w Wallet) String() string {
	return string(w.Address)
}

func (w Wallet) Validate() error {
	if w.Address == "" {
		return fmt.Errorf("wallet address cannot be empty")
	}
	if w.PublicKey == "" {
		return fmt.Errorf("wallet public key cannot be empty")
	}
	if w.PrivateKey == "" {
		return fmt.Errorf("wallet private key cannot be empty")
	}
	return w.Address.Validate()
}

func NewWalletFromExtendedKey(key *hdkeychain.ExtendedKey) (*Wallet, error) {
	if key == nil {
		return nil, fmt.Errorf("extended key cannot be nil")
	}
	address, public, private, err := GetXRPLWallet(key)
	if err != nil {
		return nil, err
	}
	return NewWallet(types.Address(address), public, private)
}

func NewWalletFromHexSeed(hexSeed string, path string) (*Wallet, error) {
	key, err := GetExtendedKeyFromHexSeedWithPath(hexSeed, path)
	if err != nil {
		return nil, err
	}
	return NewWalletFromExtendedKey(key)
}
