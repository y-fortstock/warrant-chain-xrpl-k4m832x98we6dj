package api

import (
	"fmt"
	"net/http"
	"time"

	binarycodec "github.com/CreatureDev/xrpl-go/binary-codec"
	"github.com/CreatureDev/xrpl-go/client"
	jsonrpcclient "github.com/CreatureDev/xrpl-go/client/jsonrpc"
	"github.com/CreatureDev/xrpl-go/keypairs"
	"github.com/CreatureDev/xrpl-go/model/client/account"
	clientcommon "github.com/CreatureDev/xrpl-go/model/client/common"
	"github.com/CreatureDev/xrpl-go/model/client/server"
	clienttransactions "github.com/CreatureDev/xrpl-go/model/client/transactions"
	"github.com/CreatureDev/xrpl-go/model/transactions"
	"github.com/CreatureDev/xrpl-go/model/transactions/types"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/config"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/crypto"
)

const (
	xrpToDrops = 1000000
)

type Blockchain struct {
	xrplClient    *client.XRPLClient
	systemAccount string
	systemSecret  string
}

func NewBlockchain(cfg config.NetworkConfig) (*Blockchain, error) {
	rpcCfg, err := client.NewJsonRpcConfig(cfg.URL, client.WithHttpClient(&http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to create JSON-RPC config for %s: %w", cfg.URL, err)
	}
	client := jsonrpcclient.NewClient(rpcCfg)

	systemAccount := cfg.System.Account
	systemSecret := cfg.System.Secret
	return &Blockchain{xrplClient: client, systemAccount: systemAccount, systemSecret: systemSecret}, nil
}

func (b *Blockchain) GetXRPLAddress(hexSeed string) (string, error) {
	keyPair, err := crypto.GetKeyPairFromHexSeed(hexSeed)
	if err != nil {
		return "", fmt.Errorf("failed to get key pair from hex seed: %w", err)
	}
	return crypto.GetXRPLAddressFromKeyPair(keyPair)
}

func (b *Blockchain) GetXRPLSecret(hexSeed string) (string, error) {
	keyPair, err := crypto.GetKeyPairFromHexSeed(hexSeed)
	if err != nil {
		return "", fmt.Errorf("failed to get key pair from hex seed: %w", err)
	}
	return crypto.GetXRPLSecretFromKeyPair(keyPair)
}

// GetAccountBalance получает баланс аккаунта XRPL
func (b *Blockchain) GetAccountBalance(address string) (uint64, error) {
	xrplReq := &account.AccountInfoRequest{
		Account:     types.Address(address),
		LedgerIndex: clientcommon.VALIDATED,
	}
	resp, xrplRes, err := b.xrplClient.Account.AccountInfo(xrplReq)
	if err != nil {
		return 0, fmt.Errorf("failed to get account info for %s: %w (xrplRes: %v)", address, err, xrplRes)
	}

	return uint64(resp.AccountData.Balance), nil
}

func (b *Blockchain) GetBaseFee() (uint64, error) {
	xrplReq := &server.FeeRequest{}
	resp, xrplRes, err := b.xrplClient.Server.Fee(xrplReq)
	if err != nil {
		return 0, fmt.Errorf("failed to get base fee: %w (xrplRes: %v)", err, xrplRes)
	}
	return uint64(resp.Drops.BaseFee), nil
}

func (b *Blockchain) PaymentFromSystemAccount(to string, fee, amount uint64) (string, error) {
	payment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account:         types.Address(b.systemAccount),
			TransactionType: transactions.PaymentTx,
			Fee:             types.XRPCurrencyAmount(fee),
		},
		Amount:      types.XRPCurrencyAmount(amount),
		Destination: types.Address(to),
	}

	fmt.Println("payment", payment)

	encodedForSigning, err := binarycodec.EncodeForSigning(payment)
	if err != nil {
		return "", fmt.Errorf("failed to encode for signing: %w", err)
	}

	priv, _, err := keypairs.DeriveKeypair(b.systemSecret, false)
	if err != nil {
		return "", fmt.Errorf("failed to derive key pair: %w", err)
	}

	signature, err := keypairs.Sign(encodedForSigning, priv)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Add signature to the transaction
	payment.TxnSignature = signature

	// Encode the complete signed transaction
	txBlob, err := binarycodec.Encode(payment)
	if err != nil {
		return "", fmt.Errorf("failed to encode signed transaction: %w", err)
	}

	submitReq := &clienttransactions.SubmitRequest{
		TxBlob: txBlob,
	}

	_, xrplResp, err := b.xrplClient.Transaction.Submit(submitReq)
	if err != nil {
		return "", fmt.Errorf("failed to submit transaction: %w", err)
	}

	// Get hash from XRPLResponse by parsing the raw JSON
	var rawResponse map[string]interface{}
	if err := xrplResp.GetResult(&rawResponse); err != nil {
		return "", fmt.Errorf("failed to get result from response: %w", err)
	}

	fmt.Println(rawResponse)

	// Try to get hash from tx_json.hash
	if txJSON, ok := rawResponse["tx_json"].(map[string]interface{}); ok {
		if hash, ok := txJSON["hash"].(string); ok {
			return hash, nil
		}
	}

	return "", fmt.Errorf("hash not found in response")
}
