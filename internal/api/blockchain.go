// Package api provides the gRPC API implementations for the XRPL blockchain service.
// It includes implementations for account management, token operations, and blockchain interactions.
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
	// xrpToDrops represents the conversion factor from XRP to drops.
	// 1 XRP = 1,000,000 drops in the XRPL network.
	xrpToDrops = 1000000
)

// Blockchain represents the main interface to the XRPL blockchain.
// It provides methods for interacting with the XRPL network, including
// account operations, transaction submission, and token management.
type Blockchain struct {
	xrplClient   *client.XRPLClient
	SystemWallet *crypto.Wallet
}

// NewBlockchain creates and returns a new Blockchain instance.
// It initializes the XRPL client connection and system wallet using the provided configuration.
//
// Parameters:
// - cfg: Network configuration containing RPC URL, timeout, and system account details
//
// Returns a configured Blockchain instance or an error if initialization fails.
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

// GetBaseFeeAndReserve retrieves the current base fee and reserve requirements from the XRPL network.
// This information is used to calculate transaction costs and minimum account balances.
//
// Returns server ledger information including base fee and reserve amounts, or an error if the request fails.
func (b *Blockchain) GetBaseFeeAndReserve() (info *server.ServerLedgerInfo, err error) {
	resp, _, err := b.xrplClient.Server.ServerInfo(&server.ServerInfoRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	return resp.Info.ValidatedLedger, nil
}

// SubmitTx submits a transaction to the XRPL network using the provided wallet.
// The function handles transaction signing, encoding, and submission to the network.
//
// Parameters:
// - w: The wallet used to sign the transaction
// - tx: The transaction to submit
//
// Returns the submit response, XRPL response, and any error that occurred during submission.
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

// GetAccountInfo retrieves detailed information about an XRPL account.
// This includes the account's balance, sequence number, and other account-specific data.
//
// Parameters:
// - address: The XRPL account address to query
//
// Returns account information or an error if the request fails.
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

// GetTransactionInfo retrieves detailed information about a specific transaction.
// This includes transaction metadata, base transaction details, and validation status.
//
// Parameters:
// - hash: The transaction hash to query
//
// Returns transaction response, metadata, base transaction, and any error that occurred.
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

// PaymentFromSystemAccount transfers XRP from the system account to the specified destination.
// This is typically used for funding new accounts or providing liquidity.
//
// Parameters:
// - to: The destination account address
// - amount: The amount to transfer in drops
//
// Returns the transaction hash if successful, or an error if the transfer fails.
func (b *Blockchain) PaymentFromSystemAccount(to string, amount uint64) (hash string, err error) {
	return b.Payment(b.SystemWallet, types.Address(to), amount)
}

// PaymentToSystemAccount transfers XRP from the specified source wallet to the system account.
// This is typically used for reclaiming funds or collecting fees.
//
// Parameters:
// - from: The source wallet
// - amount: The amount to transfer in drops
//
// Returns the transaction hash if successful, or an error if the transfer fails.
func (b *Blockchain) PaymentToSystemAccount(from *crypto.Wallet, amount uint64) (hash string, err error) {
	return b.Payment(from, b.SystemWallet.Address, amount)
}

// Payment executes a payment transaction between two accounts.
// The function creates, signs, and submits a payment transaction to the XRPL network.
//
// Parameters:
// - from: The source wallet
// - to: The destination account address
// - amount: The amount to transfer in drops
//
// Returns the transaction hash if successful, or an error if the payment fails.
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

	baseTx, err := b.getBaseTx(resp)
	if err != nil {
		return "", fmt.Errorf("failed to get base transaction")
	}

	return string(baseTx.Hash), nil
}

// MPTokenIssuanceCreate creates a new Multi-Purpose Token (MPT) on the XRPL network.
// This function handles the creation of token metadata and submission of the issuance transaction.
//
// Parameters:
// - w: The wallet that will own the token
// - mpt: The MPToken containing document hash and signature information
//
// Returns the transaction hash and issuance ID if successful, or an error if creation fails.
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
	if !strings.Contains(resp.EngineResult, "SUCCESS") {
		return "", "", fmt.Errorf("failed to submit: %s, %d, %s",
			resp.EngineResult,
			resp.EngineResultCode,
			resp.EngineResultMessage)
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

// AuthorizeMPToken authorizes an MPT for use by the specified wallet.
// This is required before the token can be transferred or used in transactions.
//
// Parameters:
// - w: The wallet to authorize the token for
// - issuanceId: The ID of the token issuance to authorize
//
// Returns the transaction hash if successful, or an error if authorization fails.
func (b *Blockchain) AuthorizeMPToken(w *crypto.Wallet, issuanceId string) (hash string, err error) {
	tx := &transactions.MPTokenAuthorize{
		MPTokenIssuanceID: types.Hash192(issuanceId),
	}

	resp, _, err := b.SubmitTx(w, tx)
	if err != nil {
		return "", fmt.Errorf("failed to submit tx: %w", err)
	}
	// TODO: remove this
	fmt.Println("resp", resp)

	if !strings.Contains(resp.EngineResult, "SUCCESS") {
		return "", fmt.Errorf("failed to submit: %s, %d, %s",
			resp.EngineResult,
			resp.EngineResultCode,
			resp.EngineResultMessage)
	}

	baseTx, err := b.getBaseTx(resp)
	if err != nil {
		return "", fmt.Errorf("failed to get base transaction: %w", err)
	}

	return string(baseTx.Hash), nil
}

// TransferMPToken transfers an MPT from one account to another.
// The sender must be authorized to use the token before the transfer can succeed.
//
// Parameters:
// - w: The sender's wallet
// - issuanceId: The ID of the token issuance to transfer
// - to: The destination account address
//
// Returns the transaction hash if successful, or an error if the transfer fails.
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
	if !strings.Contains(resp.EngineResult, "SUCCESS") {
		return "", fmt.Errorf("failed to submit: %s, %d, %s",
			resp.EngineResult,
			resp.EngineResultCode,
			resp.EngineResultMessage)
	}

	baseTx, err := b.getBaseTx(resp)
	if err != nil {
		return "", fmt.Errorf("failed to get base transaction: %w", err)
	}

	return string(baseTx.Hash), nil
}

// GetIssuerAddressFromIssuanceID extracts the issuer's address from a token issuance ID.
// This is useful for determining the original creator of a token.
//
// Parameters:
// - issuanceId: The token issuance ID to extract the issuer from
//
// Returns the issuer's address as a string, or an error if extraction fails.
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

// getBaseTx is a helper function that extracts base transaction information from a submit response.
// It validates the response and returns the base transaction if successful.
//
// Parameters:
// - resp: The submit response from the XRPL network
//
// Returns the base transaction or an error if extraction fails.
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

// MPToken represents a Multi-Purpose Token with associated metadata.
// It contains document hash and signature information for asset-backed tokens.
type MPToken struct {
	DocumentHash string
	Signature    string
}

// NewMPToken creates and returns a new MPToken instance.
// It requires a document hash and signature for token creation.
func NewMPToken(docHash, signature string) MPToken {
	return MPToken{
		DocumentHash: docHash,
		Signature:    signature,
	}
}

// CreateMetadata generates the metadata structure required for MPT creation.
// This includes token details, URLs, and additional information like document hash and signature.
//
// Returns the metadata structure or an error if creation fails.
func (m MPToken) CreateMetadata() (ledger.MPTokenMetadata, error) {
	addInfo, err := json.Marshal(map[string]string{
		"document_hash": m.DocumentHash,
		// "signature":     m.Signature,
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

// CreateIssuanceID generates a unique issuance ID for the token.
// This ID combines the issuer's account ID with the transaction sequence number.
//
// Parameters:
// - issuer: The issuer's account address
// - sequence: The transaction sequence number
//
// Returns the issuance ID as a string, or an error if generation fails.
func (m MPToken) CreateIssuanceID(issuer string, sequence uint32) (string, error) {
	_, accountID, err := addresscodec.DecodeClassicAddressToAccountID(issuer)
	if err != nil {
		return "", fmt.Errorf("failed to decode classic address to account id: %w", err)
	}
	accountIDHex := fmt.Sprintf("%X", accountID)
	return fmt.Sprintf("%08X%s", sequence, accountIDHex), nil
}
