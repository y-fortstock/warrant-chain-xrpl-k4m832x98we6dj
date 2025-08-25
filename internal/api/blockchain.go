package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	xrplClient   *client.XRPLClient
	systemWallet *crypto.Wallet
}

func NewBlockchain(cfg config.NetworkConfig) (*Blockchain, error) {
	rpcCfg, err := client.NewJsonRpcConfig(cfg.URL, client.WithHttpClient(&http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to create JSON-RPC config for %s: %w", cfg.URL, err)
	}
	client := jsonrpcclient.NewClient(rpcCfg)

	return &Blockchain{
		xrplClient:   client,
		systemWallet: crypto.NewWallet(types.Address(cfg.System.Account), cfg.System.Public, cfg.System.Secret),
	}, nil
}

func (b *Blockchain) GetBaseFeeAndReserve() (fee float32, reserve float32, err error) {
	resp, _, err := b.xrplClient.Server.ServerInfo(&server.ServerInfoRequest{})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get base fee and reserve: %w", err)
	}

	return resp.Info.ValidatedLedger.BaseFeeXRP, resp.Info.ValidatedLedger.ReserveBaseXRP, nil
}

func (b *Blockchain) SubmitTx(w *crypto.Wallet, tx transactions.Tx) (
	resp *clienttransactions.SubmitResponse, xrplResp client.XRPLResponse, err error) {
	if err := b.xrplClient.AutofillTx(w.Address, tx); err != nil {
		return nil, nil, fmt.Errorf("failed to autofill tx: %w", err)
	}

	baseTx := transactions.BaseTxForTransaction(tx)
	if baseTx == nil {
		return nil, nil, fmt.Errorf("failed to get base transaction")
	}
	baseTx.Fee = types.XRPCurrencyAmount(120)

	encodedForSigning, err := binarycodec.EncodeForSigning(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode for signing: %w", err)
	}
	signature, err := keypairs.Sign(encodedForSigning, w.PrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign: %w", err)
	}

	baseTx.Account = w.Address
	baseTx.SigningPubKey = w.PublicKey
	baseTx.TxnSignature = signature

	j, err := json.Marshal(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal: %w", err)
	}
	fmt.Println()
	fmt.Println("tx: ", string(j))

	txBlob, err := binarycodec.Encode(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode: %w", err)
	}
	fmt.Println()
	fmt.Println("txBlob: ", txBlob)

	submitReq := &clienttransactions.SubmitRequest{
		TxBlob: txBlob,
	}
	return b.xrplClient.Transaction.Submit(submitReq)
}

func (b *Blockchain) GetAccountInfo(address string) (*account.AccountInfoResponse, error) {
	accountInfoReq := &account.AccountInfoRequest{
		Account:     types.Address(address),
		LedgerIndex: clientcommon.VALIDATED,
	}
	accountInfo, _, err := b.xrplClient.Account.AccountInfo(accountInfoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}
	return accountInfo, nil
}

func (b *Blockchain) PaymentFromSystemAccount(to string, amount uint64) (hash string, err error) {
	return b.Payment(b.systemWallet, types.Address(to), amount)
}

func (b *Blockchain) PaymentToSystemAccount(from *crypto.Wallet, amount uint64) (hash string, err error) {
	return b.Payment(from, b.systemWallet.Address, amount)
}

func (b *Blockchain) Payment(from *crypto.Wallet, to types.Address, amount uint64) (hash string, err error) {
	payment := &transactions.Payment{
		Amount:      types.XRPCurrencyAmount(amount),
		Destination: to,
	}

	resp, _, err := b.SubmitTx(from, payment)
	if err != nil {
		return "", fmt.Errorf("failed to submit: %w", err)
	}

	if !strings.Contains(resp.EngineResult, "SUCCESS") {
		return "", fmt.Errorf("failed to submit: %s, %d, %s",
			resp.EngineResult,
			resp.EngineResultCode,
			resp.EngineResultMessage)
	}

	baseTx := transactions.BaseTxForTransaction(resp.Tx)
	if baseTx == nil {
		return "", fmt.Errorf("failed to get base transaction")
	}

	if baseTx.Hash == "" {
		return "", fmt.Errorf("transaction hash not available")
	}

	return string(baseTx.Hash), nil
}
