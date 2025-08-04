package api

import (
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
	clientledger "github.com/CreatureDev/xrpl-go/model/client/ledger"
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
	systemPublic  string
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
		xrplClient:    client,
		systemAccount: cfg.System.Account,
		systemSecret:  cfg.System.Secret,
		systemPublic:  cfg.System.Public,
	}, nil
}

func (b *Blockchain) GetXRPLWallet(hexSeed string, path string) (address string, public string, private string, err error) {
	extendedKey, err := crypto.GetExtendedKeyFromHexSeedWithPath(hexSeed, path)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get extended key from hex seed: %w", err)
	}
	address, public, private, err = crypto.GetXRPLWallet(extendedKey)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get xrpl wallet: %w", err)
	}
	return address, public, private, nil
}

func (b *Blockchain) GetBaseFeeAndReserve() (float32, float32, error) {
	serverInfoReq := &server.ServerInfoRequest{}
	resp, _, err := b.xrplClient.Server.ServerInfo(serverInfoReq)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get base fee: %w", err)
	}

	fmt.Println("resp.Info.ValidatedLedger.BaseFeeXRP: ", resp.Info.ValidatedLedger.BaseFeeXRP)
	fmt.Println("resp.Info.ValidatedLedger.ReserveBaseXRP: ", resp.Info.ValidatedLedger.ReserveBaseXRP)
	return resp.Info.ValidatedLedger.BaseFeeXRP,
		resp.Info.ValidatedLedger.ReserveBaseXRP, nil
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

func (b *Blockchain) GetLedger() (*clientledger.LedgerResponse, error) {
	ledgerReq := &clientledger.LedgerRequest{
		LedgerIndex: clientcommon.VALIDATED,
	}
	ledgerResp, _, err := b.xrplClient.Ledger.Ledger(ledgerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledger: %w", err)
	}
	return ledgerResp, nil
}

func (b *Blockchain) PaymentFromSystemAccount(to string, fee, amount uint64, sequence uint32) (string, error) {
	return b.Payment(b.systemAccount, b.systemPublic, b.systemSecret, to, fee, amount, sequence)
}

func (b *Blockchain) PaymentToSystemAccount(from, public, private string,
	fee, amount uint64,
	sequence uint32) (string, error) {
	return b.Payment(from, public, private, b.systemAccount, fee, amount, sequence)
}

func (b *Blockchain) Payment(from, public, private, to string,
	fee, amount uint64,
	sequence uint32) (string, error) {
	ledgerResp, err := b.GetLedger()
	if err != nil {
		return "", fmt.Errorf("failed to get ledger: %w", err)
	}
	ledgerIndex := uint32(ledgerResp.LedgerIndex) + 20

	payment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account:            types.Address(from),
			TransactionType:    transactions.PaymentTx,
			Fee:                types.XRPCurrencyAmount(fee),
			Sequence:           sequence,
			LastLedgerSequence: ledgerIndex,
			SigningPubKey:      public,
		},
		Amount:      types.XRPCurrencyAmount(amount),
		Destination: types.Address(to),
	}
	encodedForSigning, err := binarycodec.EncodeForSigning(payment)
	if err != nil {
		return "", fmt.Errorf("failed to encode for signing: %w", err)
	}

	payment.TxnSignature, err = keypairs.Sign(encodedForSigning, private)
	if err != nil {
		return "", fmt.Errorf("failed to sign: %w", err)
	}

	txBlob, err := binarycodec.Encode(payment)
	if err != nil {
		return "", fmt.Errorf("failed to encode: %w", err)
	}

	submitReq := &clienttransactions.SubmitRequest{
		TxBlob: txBlob,
	}

	resp, _, err := b.xrplClient.Transaction.Submit(submitReq)
	if err != nil {
		return "", fmt.Errorf("failed to submit: %w", err)
	}
	if !strings.Contains(resp.EngineResult, "SUCCESS") {
		return "", fmt.Errorf("failed to submit: %s, %s, %s",
			resp.EngineResult,
			resp.EngineResultCode,
			resp.EngineResultMessage)
	}

	decodedTx, err := binarycodec.Decode(resp.TxBlob)
	if err != nil {
		return "", fmt.Errorf("failed to decode: %w", err)
	}
	fmt.Println("resp: ", resp)
	fmt.Println("decodedTx: ", decodedTx)

	if txnSignature, ok := decodedTx["TxnSignature"].(string); ok {
		return txnSignature, nil
	}
	return "", fmt.Errorf("failed to get txn signature")
}
