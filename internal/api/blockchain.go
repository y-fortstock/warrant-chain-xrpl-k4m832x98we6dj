package api

import "gitlab.com/warrant1/warrant/chain-xrpl/internal/crypto"

type Blockchain struct {
}

func NewBlockchain() *Blockchain {
	return &Blockchain{}
}

func (b *Blockchain) GetXRPLAddress(hexSeed string) (string, error) {
	keyPair, err := crypto.GetKeyPairFromHexSeed(hexSeed)
	if err != nil {
		return "", err
	}
	return crypto.GetXRPLAddressFromKeyPair(keyPair)
}

func (b *Blockchain) GetXRPLSecret(hexSeed string) (string, error) {
	keyPair, err := crypto.GetKeyPairFromHexSeed(hexSeed)
	if err != nil {
		return "", err
	}
	return crypto.GetXRPLSecretFromKeyPair(keyPair)
}
