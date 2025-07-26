package api

import (
	"fmt"

	"github.com/CreatureDev/xrpl-go/client"
	jsonrpcclient "github.com/CreatureDev/xrpl-go/client/jsonrpc"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/config"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/crypto"
)

type Blockchain struct {
	xrplClient *client.XRPLClient
}

func NewBlockchain(cfg config.NetworkConfig) (*Blockchain, error) {
	rpcCfg, err := client.NewJsonRpcConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to create JSON-RPC config for %s: %w", cfg.URL, err)
	}
	client := jsonrpcclient.NewClient(rpcCfg)
	return &Blockchain{xrplClient: client}, nil
}

func (b *Blockchain) GetXRPLAddress(hexSeed string) (string, error) {
	keyPair, err := crypto.GetKeyPairFromHexSeed(hexSeed)
	if err != nil {
		return "", fmt.Errorf("failed to get key pair from hex seed: %w", err)
	}
	return crypto.GetXRPLAddressFromKeyPair(keyPair), nil
}

func (b *Blockchain) GetXRPLSecret(hexSeed string) (string, error) {
	keyPair, err := crypto.GetKeyPairFromHexSeed(hexSeed)
	if err != nil {
		return "", fmt.Errorf("failed to get key pair from hex seed: %w", err)
	}
	return crypto.GetXRPLSecretFromKeyPair(keyPair), nil
}
