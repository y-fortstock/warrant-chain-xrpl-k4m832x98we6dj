package crypto

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
)

type Wallet struct {
	Address string
	Public  string
	Private string
}

func NewWallet(address, public, private string) *Wallet {
	return &Wallet{Address: address, Public: public, Private: private}
}

func NewWalletFromExtendedKey(key *hdkeychain.ExtendedKey) (*Wallet, error) {
	if key == nil {
		return nil, fmt.Errorf("extended key cannot be nil")
	}
	address, public, private, err := GetXRPLWallet(key)
	if err != nil {
		return nil, err
	}
	return NewWallet(address, public, private), nil
}

func NewWalletFromHexSeed(hexSeed string, path string) (*Wallet, error) {
	key, err := GetExtendedKeyFromHexSeedWithPath(hexSeed, path)
	if err != nil {
		return nil, err
	}
	return NewWalletFromExtendedKey(key)
}
