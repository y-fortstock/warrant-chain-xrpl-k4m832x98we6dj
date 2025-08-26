package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	addresscodec "github.com/CreatureDev/xrpl-go/address-codec"
	binarycodec "github.com/CreatureDev/xrpl-go/binary-codec"
	"github.com/CreatureDev/xrpl-go/client"
	jsonrpcclient "github.com/CreatureDev/xrpl-go/client/jsonrpc"
	"github.com/CreatureDev/xrpl-go/keypairs"
	"github.com/CreatureDev/xrpl-go/model/client/account"
	clientcommon "github.com/CreatureDev/xrpl-go/model/client/common"
	"github.com/CreatureDev/xrpl-go/model/client/server"
	clienttransactions "github.com/CreatureDev/xrpl-go/model/client/transactions"
	"github.com/CreatureDev/xrpl-go/model/ledger"
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
	SystemWallet *crypto.Wallet
}

func NewBlockchain(cfg config.NetworkConfig) (*Blockchain, error) {
	rpcCfg, err := client.NewJsonRpcConfig(cfg.URL, client.WithHttpClient(&http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to create JSON-RPC config for %s: %w", cfg.URL, err)
	}
	client := jsonrpcclient.NewClient(rpcCfg)

	systemWallet, err := crypto.NewWallet(types.Address(cfg.System.Account), cfg.System.Public, cfg.System.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create system wallet: %w", err)
	}
	return &Blockchain{
		xrplClient:   client,
		SystemWallet: systemWallet,
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
	if w == nil {
		return nil, nil, fmt.Errorf("wallet cannot be nil")
	}
	if tx == nil {
		return nil, nil, fmt.Errorf("transaction cannot be nil")
	}
	if err := w.Validate(); err != nil {
		return nil, nil, fmt.Errorf("wallet is invalid: %w", err)
	}

	if err := b.xrplClient.AutofillTx(w.Address, tx); err != nil {
		return nil, nil, fmt.Errorf("failed to autofill tx: %w", err)
	}

	baseTx := transactions.BaseTxForTransaction(tx)
	if baseTx == nil {
		return nil, nil, fmt.Errorf("failed to get base transaction")
	}
	baseTx.Account = w.Address
	baseTx.SigningPubKey = w.PublicKey

	encodedForSigning, err := binarycodec.EncodeForSigning(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode for signing: %w", err)
	}
	signature, err := keypairs.Sign(encodedForSigning, w.PrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign: %w", err)
	}
	baseTx.TxnSignature = signature

	txBlob, err := binarycodec.Encode(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode: %w", err)
	}

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

func (b *Blockchain) GetTransactionInfo(hash string) (
	resp *clienttransactions.TxResponse,
	meta transactions.TxObjMeta,
	baseTx *transactions.BaseTx,
	err error) {
	resp, _, err = b.xrplClient.Transaction.Tx(&clienttransactions.TxRequest{
		Transaction: hash,
	})
	if err != nil {
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to get transaction info: %w", err)
	}

	if resp.Meta == nil {
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("metadata is nil")
	}
	if resp.Tx == nil {
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("transaction is nil")
	}

	if objMeta, ok := resp.Meta.(transactions.TxObjMeta); ok {
		meta = objMeta
	}
	baseTx = transactions.BaseTxForTransaction(resp.Tx)
	if baseTx == nil {
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to extract base transaction from transaction")
	}

	return resp, meta, baseTx, nil
}

func (b *Blockchain) PaymentFromSystemAccount(to string, amount uint64) (hash string, err error) {
	return b.Payment(b.SystemWallet, types.Address(to), amount)
}

func (b *Blockchain) PaymentToSystemAccount(from *crypto.Wallet, amount uint64) (hash string, err error) {
	return b.Payment(from, b.SystemWallet.Address, amount)
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

	baseTx, err := b.getBaseTx(resp)
	if err != nil {
		return "", fmt.Errorf("failed to get base transaction")
	}

	return string(baseTx.Hash), nil
}

func (b *Blockchain) MPTokenIssuanceCreate(w *crypto.Wallet, mpt MPToken) (hash, issuanceID string, err error) {
	md, err := mpt.CreateMetadata()
	if err != nil {
		return "", "", fmt.Errorf("failed to create metadata: %w", err)
	}

	blob, err := md.GetBlob()
	if err != nil {
		return "", "", fmt.Errorf("failed to get blob: %w", err)
	}

	tx := &transactions.MPTokenIssuanceCreate{
		MPTokenMetadata: blob,
		MaximumAmount:   "1",
		TransferFee:     0,
		Flags:           types.NewFlag().SetFlag(types.TfMPTCanEscrow).SetFlag(types.TfMPTCanTrade).SetFlag(types.TfMPTCanTransfer),
	}

	resp, _, err := b.SubmitTx(w, tx)
	if err != nil {
		return "", "", fmt.Errorf("failed to submit tx: %w", err)
	}
	issuanceID, err = mpt.CreateIssuanceID(string(w.Address), tx.Sequence)
	if err != nil {
		return "", "", fmt.Errorf("failed to create issuance id: %w", err)
	}

	baseTx, err := b.getBaseTx(resp)
	if err != nil {
		return "", "", fmt.Errorf("failed to get base transaction: %w", err)
	}

	return string(baseTx.Hash), issuanceID, nil
}

func (b *Blockchain) AuthorizeMPToken(w *crypto.Wallet, issuanceId string) (hash string, err error) {
	tx := &transactions.MPTokenAuthorize{
		MPTokenIssuanceID: types.Hash192(issuanceId),
	}

	resp, _, err := b.SubmitTx(w, tx)
	if err != nil {
		return "", fmt.Errorf("failed to submit tx: %w", err)
	}

	baseTx, err := b.getBaseTx(resp)
	if err != nil {
		return "", fmt.Errorf("failed to get base transaction: %w", err)
	}

	return string(baseTx.Hash), nil
}

func (b *Blockchain) TransferMPToken(w *crypto.Wallet, issuanceId, to string) (hash string, err error) {
	tx := &transactions.Payment{
		Amount: types.MPTCurrencyAmount{
			Value:         "1",
			MPTIssuanceID: types.Hash192(issuanceId),
		},
		Destination: types.Address(to),
	}

	resp, _, err := b.SubmitTx(w, tx)
	if err != nil {
		return "", fmt.Errorf("failed to submit tx: %w", err)
	}

	baseTx, err := b.getBaseTx(resp)
	if err != nil {
		return "", fmt.Errorf("failed to get base transaction: %w", err)
	}

	return string(baseTx.Hash), nil
}

func (b *Blockchain) GetIssuerAddressFromIssuanceID(issuanceId string) (issuer string, err error) {
	amount := types.MPTCurrencyAmount{
		Value:         "1",
		MPTIssuanceID: types.Hash192(issuanceId),
	}
	issuerAccId, err := amount.IssuerAccountID()
	if err != nil {
		return "", fmt.Errorf("failed to get issuer account id: %w", err)
	}

	issuerAddr, err := addresscodec.EncodeAccountIDToClassicAddress(issuerAccId)
	if err != nil {
		return "", fmt.Errorf("failed to encode account id %s to classic address: %w",
			string(issuerAccId), err)
	}

	return string(issuerAddr), nil
}

func (b *Blockchain) getBaseTx(resp *clienttransactions.SubmitResponse) (baseTx *transactions.BaseTx, err error) {
	if !strings.Contains(resp.EngineResult, "SUCCESS") {
		return nil, fmt.Errorf("failed to submit: %s, %d, %s",
			resp.EngineResult,
			resp.EngineResultCode,
			resp.EngineResultMessage)
	}

	baseTx = transactions.BaseTxForTransaction(resp.Tx)
	if baseTx == nil {
		return nil, fmt.Errorf("failed to cast to base transaction")
	}

	if baseTx.Hash == "" {
		return nil, fmt.Errorf("transaction hash not available")
	}

	return baseTx, nil
}

type MPToken struct {
	DocumentHash string
	Signature    string
}

func NewMPToken(docHash, signature string) MPToken {
	return MPToken{
		DocumentHash: docHash,
		Signature:    signature,
	}
}

func (m MPToken) CreateMetadata() (ledger.MPTokenMetadata, error) {
	addInfo, err := json.Marshal(map[string]string{
		"document_hash": m.DocumentHash,
		"signature":     m.Signature,
	})
	if err != nil {
		return ledger.MPTokenMetadata{}, fmt.Errorf("failed to marshal additional info: %w", err)
	}

	return ledger.MPTokenMetadata{
		Ticker:        "FSWRNT",
		Name:          "FortStock Warrant",
		Desc:          "Digital representation of real-world asset-backed warrants",
		AssetClass:    "rwa",
		AssetSubclass: "other",
		Urls: []ledger.MPTokenMetadataUrl{
			{
				Url:   "https://fortstock.io",
				Type:  "website",
				Title: "Home",
			},
			{
				Url:   "https://fortstock.io/rulebook/",
				Type:  "document",
				Title: "Legal framework",
			},
		},
		AdditionalInfo: addInfo,
	}, nil
}

func (m MPToken) CreateIssuanceID(issuer string, sequence uint32) (string, error) {
	_, accountID, err := addresscodec.DecodeClassicAddressToAccountID(issuer)
	if err != nil {
		return "", fmt.Errorf("failed to decode classic address to account id: %w", err)
	}
	accountIDHex := fmt.Sprintf("%X", accountID)
	return fmt.Sprintf("%08X%s", sequence, accountIDHex), nil
}
