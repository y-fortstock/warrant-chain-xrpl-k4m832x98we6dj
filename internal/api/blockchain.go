package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/CreatureDev/xrpl-go/client"
	jsonrpcclient "github.com/CreatureDev/xrpl-go/client/jsonrpc"
	"github.com/CreatureDev/xrpl-go/model/client/account"
	"github.com/CreatureDev/xrpl-go/model/transactions/types"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/config"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/crypto"
)

type Blockchain struct {
	xrplClient *client.XRPLClient
}

func NewBlockchain(cfg config.NetworkConfig) (*Blockchain, error) {
	rpcCfg, err := client.NewJsonRpcConfig(cfg.URL, client.WithHttpClient(&http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}))
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

// GetAccountBalance получает баланс аккаунта XRPL
func (b *Blockchain) GetAccountBalance(address string) (uint64, error) {
	xrplReq := &account.AccountInfoRequest{
		Account: types.Address(address),
	}
	resp, xrplRes, err := b.xrplClient.Account.AccountInfo(xrplReq)
	if err != nil {
		return 0, fmt.Errorf("failed to get account info for %s: %w (xrplRes: %v)", address, err, xrplRes)
	}

	return uint64(resp.AccountData.Balance), nil
}
