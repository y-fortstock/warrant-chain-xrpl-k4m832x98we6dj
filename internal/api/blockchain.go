// Package api provides the gRPC API implementations for the XRPL blockchain service.
// It includes implementations for account management, token operations, and blockchain interactions.
package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	account "github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	servertypes "github.com/Peersyst/xrpl-go/xrpl/queries/server/types"
	requests "github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	rpctypes "github.com/Peersyst/xrpl-go/xrpl/rpc/types"
	transactions "github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/config"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/crypto"
)

const (
	// xrpToDrops represents the conversion factor from XRP to drops.
	// 1 XRP = 1,000,000 drops in the XRPL network.
	xrpToDrops = 1000000
)

type SubmittableTransaction interface {
	TxType() transactions.TxType
	Flatten() transactions.FlatTransaction
}

// Blockchain represents the main interface to the XRPL blockchain.
// It provides methods for interacting with the XRPL network, including
// account operations, transaction submission, and token management.
type Blockchain struct {
	mu sync.RWMutex
	c  *rpc.Client
	w  *wallet.Wallet
}

// NewBlockchain creates and returns a new Blockchain instance.
// It initializes the XRPL client connection and system wallet using the provided configuration.
//
// Parameters:
// - cfg: Network configuration containing RPC URL, timeout, and system account details
//
// Returns a configured Blockchain instance or an error if initialization fails.
func NewBlockchain(cfg config.NetworkConfig) (*Blockchain, error) {
	rpcCfg, err := rpc.NewClientConfig(cfg.URL, rpc.WithHTTPClient(&http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to create JSON-RPC config for %s: %w", cfg.URL, err)
	}
	client := rpc.NewClient(rpcCfg)

	w, err := crypto.NewWallet(types.Address(cfg.System.Account), cfg.System.Public, cfg.System.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return &Blockchain{
		c: client,
		w: w,
	}, nil
}

// Lock acquires an exclusive lock on the blockchain instance.
// This method should be called before performing any operations that require
// exclusive access to the blockchain state.
func (b *Blockchain) Lock() {
	b.mu.Lock()
}

// Unlock releases the exclusive lock on the blockchain instance.
// This method should be called after completing operations that required
// exclusive access to the blockchain state.
func (b *Blockchain) Unlock() {
	b.mu.Unlock()
}

// GetBaseFeeAndReserve retrieves the current base fee and reserve requirements from the XRPL network.
// This information is used to calculate transaction costs and minimum account balances.
//
// Returns server ledger information including base fee and reserve amounts, or an error if the request fails.
func (b *Blockchain) GetBaseFeeAndReserve() (info servertypes.ClosedLedger, err error) {
	resp, err := b.c.GetServerInfo(&server.InfoRequest{})
	if err != nil {
		return servertypes.ClosedLedger{}, fmt.Errorf("failed to get server info: %w", err)
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
func (b *Blockchain) SubmitTx(w *wallet.Wallet, tx SubmittableTransaction) (
	hash string, err error) {
	if w == nil {
		return "", fmt.Errorf("wallet cannot be nil")
	}
	if tx == nil {
		return "", fmt.Errorf("transaction cannot be nil")
	}

	// Access BaseTx fields directly since all transaction types embed BaseTx
	flattenedTx := tx.Flatten()
	flattenedTx["Account"] = w.ClassicAddress.String()
	flattenedTx["SigningPubKey"] = w.PublicKey

	resp, err := b.c.SubmitTx(flattenedTx, &rpctypes.SubmitOptions{
		Autofill: true,
		FailHard: false,
		Wallet:   w,
	})
	if err != nil {
		return "", fmt.Errorf("failed to submit tx: %w", err)
	}

	if resp.EngineResult != string(transactions.TesSUCCESS) {
		return "", &rpc.ClientError{ErrorString: "transaction failed to submit with engine result: " + resp.EngineResult}
	}

	hash = resp.Tx["hash"].(string)
	if hash == "" {
		return "", fmt.Errorf("hash is empty")
	}

	return hash, nil
}

// SubmitTxWithSequence submits a transaction to the XRPL network and returns the hash and sequence.
func (b *Blockchain) SubmitTxWithSequence(w *wallet.Wallet, tx SubmittableTransaction) (
	hash string, sequence uint32, err error) {
	if w == nil {
		return "", 0, fmt.Errorf("wallet cannot be nil")
	}
	if tx == nil {
		return "", 0, fmt.Errorf("transaction cannot be nil")
	}

	// Access BaseTx fields directly since all transaction types embed BaseTx
	flattenedTx := tx.Flatten()
	flattenedTx["Account"] = w.ClassicAddress.String()
	flattenedTx["SigningPubKey"] = w.PublicKey

	resp, err := b.c.SubmitTx(flattenedTx, &rpctypes.SubmitOptions{
		Autofill: true,
		FailHard: false,
		Wallet:   w,
	})
	if err != nil {
		return "", 0, fmt.Errorf("failed to submit tx: %w", err)
	}

	if resp.EngineResult != string(transactions.TesSUCCESS) {
		return "", 0, &rpc.ClientError{ErrorString: "transaction failed to submit with engine result: " + resp.EngineResult}
	}

	hash = resp.Tx["hash"].(string)
	if hash == "" {
		return "", 0, fmt.Errorf("hash is empty")
	}

	// Get sequence from the response
	sequenceValue, ok := resp.Tx["Sequence"]
	if !ok {
		return "", 0, fmt.Errorf("sequence not found in response")
	}

	// Handle different numeric types that might be returned
	switch v := sequenceValue.(type) {
	case uint32:
		sequence = v
	case int:
		sequence = uint32(v)
	case float64:
		sequence = uint32(v)
	case int64:
		sequence = uint32(v)
	case json.Number:
		// Handle json.Number type
		intVal, err := v.Int64()
		if err != nil {
			return "", 0, fmt.Errorf("failed to convert sequence to int64: %w", err)
		}
		sequence = uint32(intVal)
	default:
		return "", 0, fmt.Errorf("sequence has unexpected type: %T", v)
	}

	return hash, sequence, nil
}

// GetAccountInfo retrieves detailed information about an XRPL account.
// This includes the account's balance, sequence number, and other account-specific data.
//
// Parameters:
// - address: The XRPL account address to query
//
// Returns account information or an error if the request fails.
func (b *Blockchain) GetAccountInfo(address string) (*account.InfoResponse, error) {
	accountInfoReq := &account.InfoRequest{
		Account:     types.Address(address),
		LedgerIndex: common.Validated,
	}
	accountInfo, err := b.c.GetAccountInfo(accountInfoReq)
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
	resp *requests.TxResponse,
	meta transactions.TxObjMeta,
	baseTx *transactions.BaseTx,
	err error) {
	res, err := b.c.Request(&requests.TxRequest{
		Transaction: hash,
	})
	if err != nil {
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to get transaction info: %w", err)
	}

	var txResp requests.TxResponse
	err = res.GetResult(&txResp)
	if err != nil {
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to parse transaction response: %w", err)
	}

	if txResp.Meta == nil {
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("metadata is nil")
	}
	if len(txResp.Tx) == 0 {
		// Check if this is a "not found" case by looking at the response
		if txResp.LedgerIndex == 0 && !txResp.Validated {
			return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("transaction not found or not yet confirmed")
		}
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("transaction is nil or empty (ledger_index: %v, validated: %v)", txResp.LedgerIndex, txResp.Validated)
	}

	if objMeta, ok := txResp.Meta.(transactions.TxObjMeta); ok {
		meta = objMeta
	} else {
		// Try to convert from map[string]interface{} to TxObjMeta using JSON marshaling/unmarshaling
		if metaMap, ok := txResp.Meta.(map[string]interface{}); ok {
			// Convert map to JSON and then unmarshal to TxObjMeta
			jsonData, err := json.Marshal(metaMap)
			if err != nil {
				return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to marshal metadata: %w", err)
			}
			err = json.Unmarshal(jsonData, &meta)
			if err != nil {
				return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to unmarshal metadata to TxObjMeta: %w", err)
			}
		} else {
			return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to cast metadata to TxObjMeta, got type: %T", txResp.Meta)
		}
	}

	// Safely extract fields from transaction with type assertions
	account, ok := txResp.Tx["Account"].(string)
	if !ok {
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to extract Account from transaction")
	}

	// Try different types for Fee
	var fee float64
	if feeFloat, ok := txResp.Tx["Fee"].(float64); ok {
		fee = feeFloat
	} else if feeString, ok := txResp.Tx["Fee"].(string); ok {
		// Try to parse string to float64
		if parsedFee, err := strconv.ParseFloat(feeString, 64); err == nil {
			fee = parsedFee
		} else {
			return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to parse Fee string '%s': %w", feeString, err)
		}
	} else if feeNumber, ok := txResp.Tx["Fee"].(json.Number); ok {
		// Try to parse json.Number to float64
		if parsedFee, err := feeNumber.Float64(); err == nil {
			fee = parsedFee
		} else {
			return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to parse Fee json.Number '%s': %w", feeNumber, err)
		}
	} else {
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to extract Fee from transaction, got type: %T", txResp.Tx["Fee"])
	}

	// Try different types for Flags
	var flags float64
	if txResp.Tx["Flags"] == nil {
		// Flags can be nil if not set
		flags = 0
	} else if flagsFloat, ok := txResp.Tx["Flags"].(float64); ok {
		flags = flagsFloat
	} else if flagsString, ok := txResp.Tx["Flags"].(string); ok {
		if parsedFlags, err := strconv.ParseFloat(flagsString, 64); err == nil {
			flags = parsedFlags
		} else {
			return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to parse Flags string '%s': %w", flagsString, err)
		}
	} else if flagsNumber, ok := txResp.Tx["Flags"].(json.Number); ok {
		if parsedFlags, err := flagsNumber.Float64(); err == nil {
			flags = parsedFlags
		} else {
			return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to parse Flags json.Number '%s': %w", flagsNumber, err)
		}
	} else {
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to extract Flags from transaction, got type: %T", txResp.Tx["Flags"])
	}

	// Try different types for LastLedgerSequence
	var lastLedgerSeq float64
	if txResp.Tx["LastLedgerSequence"] == nil {
		// LastLedgerSequence can be nil if not set
		lastLedgerSeq = 0
	} else if lastLedgerSeqFloat, ok := txResp.Tx["LastLedgerSequence"].(float64); ok {
		lastLedgerSeq = lastLedgerSeqFloat
	} else if lastLedgerSeqString, ok := txResp.Tx["LastLedgerSequence"].(string); ok {
		if parsedLastLedgerSeq, err := strconv.ParseFloat(lastLedgerSeqString, 64); err == nil {
			lastLedgerSeq = parsedLastLedgerSeq
		} else {
			return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to parse LastLedgerSequence string '%s': %w", lastLedgerSeqString, err)
		}
	} else if lastLedgerSeqNumber, ok := txResp.Tx["LastLedgerSequence"].(json.Number); ok {
		if parsedLastLedgerSeq, err := lastLedgerSeqNumber.Float64(); err == nil {
			lastLedgerSeq = parsedLastLedgerSeq
		} else {
			return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to parse LastLedgerSequence json.Number '%s': %w", lastLedgerSeqNumber, err)
		}
	} else {
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to extract LastLedgerSequence from transaction, got type: %T", txResp.Tx["LastLedgerSequence"])
	}

	// Try different types for Sequence
	var sequence float64
	if txResp.Tx["Sequence"] == nil {
		// Sequence should not be nil, but handle it gracefully
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("sequence is required but was nil")
	} else if sequenceFloat, ok := txResp.Tx["Sequence"].(float64); ok {
		sequence = sequenceFloat
	} else if sequenceString, ok := txResp.Tx["Sequence"].(string); ok {
		if parsedSequence, err := strconv.ParseFloat(sequenceString, 64); err == nil {
			sequence = parsedSequence
		} else {
			return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to parse Sequence string '%s': %w", sequenceString, err)
		}
	} else if sequenceNumber, ok := txResp.Tx["Sequence"].(json.Number); ok {
		if parsedSequence, err := sequenceNumber.Float64(); err == nil {
			sequence = parsedSequence
		} else {
			return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to parse Sequence json.Number '%s': %w", sequenceNumber, err)
		}
	} else {
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to extract Sequence from transaction, got type: %T", txResp.Tx["Sequence"])
	}

	signingPubKey, ok := txResp.Tx["SigningPubKey"].(string)
	if !ok {
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to extract SigningPubKey from transaction")
	}

	transactionType, ok := txResp.Tx["TransactionType"].(string)
	if !ok {
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to extract TransactionType from transaction")
	}

	txnSignature, ok := txResp.Tx["TxnSignature"].(string)
	if !ok {
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to extract TxnSignature from transaction")
	}

	baseTx = &transactions.BaseTx{
		Account:            types.Address(account),
		Fee:                types.XRPCurrencyAmount(uint64(fee)),
		Flags:              uint32(flags),
		LastLedgerSequence: uint32(lastLedgerSeq),
		Sequence:           uint32(sequence),
		SigningPubKey:      signingPubKey,
		TransactionType:    transactions.TxType(transactionType),
		TxnSignature:       txnSignature,
	}

	return &txResp, meta, baseTx, nil
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
	return b.Payment(b.w, types.Address(to), amount)
}

// PaymentToSystemAccount transfers XRP from the specified source wallet to the system account.
// This is typically used for reclaiming funds or collecting fees.
//
// Parameters:
// - from: The source wallet
// - amount: The amount to transfer in drops
//
// Returns the transaction hash if successful, or an error if the transfer fails.
func (b *Blockchain) PaymentToSystemAccount(from *wallet.Wallet, amount uint64) (hash string, err error) {
	return b.Payment(from, b.w.ClassicAddress, amount)
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
func (b *Blockchain) Payment(from *wallet.Wallet, to types.Address, amount uint64) (txHash string, err error) {
	payment := &transactions.Payment{
		Amount:      types.XRPCurrencyAmount(amount),
		Destination: to,
	}

	return b.SubmitTx(from, payment)
}

// MPTokenIssuanceCreate creates a new Multi-Purpose Token (MPT) on the XRPL network.
// This function handles the creation of token metadata and submission of the issuance transaction.
//
// Parameters:
// - w: The wallet that will own the token
// - mpt: The MPToken containing document hash and signature information
//
// Returns the transaction hash and issuance ID if successful, or an error if creation fails.
func (b *Blockchain) MPTokenIssuanceCreate(w *wallet.Wallet, mpt MPToken) (txHash, issuanceID string, err error) {
	md, err := mpt.CreateMetadata()
	if err != nil {
		return "", "", fmt.Errorf("failed to create metadata: %w", err)
	}

	blob, err := md.GetBlob()
	if err != nil {
		return "", "", fmt.Errorf("failed to get blob: %w", err)
	}

	maxAmount := types.XRPCurrencyAmount(1)
	tx := &transactions.MPTokenIssuanceCreate{
		MPTokenMetadata: &blob,
		MaximumAmount:   &maxAmount,
		TransferFee:     types.TransferFee(0),
	}
	tx.SetMPTCanEscrowFlag()
	tx.SetMPTCanTradeFlag()
	tx.SetMPTCanTransferFlag()

	hash, sequence, err := b.SubmitTxWithSequence(w, tx)
	if err != nil {
		return "", "", fmt.Errorf("failed to submit tx: %w", err)
	}

	issuanceID, err = mpt.CreateIssuanceID(string(w.ClassicAddress), sequence)
	if err != nil {
		return "", "", fmt.Errorf("failed to create issuance id: %w", err)
	}

	return hash, issuanceID, nil
}

// AuthorizeMPToken authorizes an MPT for use by the specified wallet.
// This is required before the token can be transferred or used in transactions.
//
// Parameters:
// - w: The wallet to authorize the token for
// - issuanceId: The ID of the token issuance to authorize
//
// Returns the transaction hash if successful, or an error if authorization fails.
func (b *Blockchain) AuthorizeMPToken(w *wallet.Wallet, issuanceId string) (txHash string, err error) {
	tx := &transactions.MPTokenAuthorize{
		MPTokenIssuanceID: issuanceId,
	}

	return b.SubmitTx(w, tx)
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
func (b *Blockchain) TransferMPToken(w *wallet.Wallet, issuanceId, to string) (txHash string, err error) {
	tx := &transactions.Payment{
		Amount: types.MPTCurrencyAmount{
			Value:         "1",
			MPTIssuanceID: issuanceId,
		},
		Destination: types.Address(to),
	}

	return b.SubmitTx(w, tx)
}

// GetIssuerAddressFromIssuanceID extracts the issuer's address from a token issuance ID.
// This is useful for determining the original creator of a token.
//
// Parameters:
// - issuanceId: The token issuance ID to extract the issuer from
//
// Returns the issuer's address as a string, or an error if extraction fails.
func (b *Blockchain) GetIssuerAddressFromIssuanceID(issuanceId string) (issuer string, err error) {
	if len(issuanceId) != 48 {
		return "", fmt.Errorf("invalid issuance ID length: expected 56 hex characters, got %d", len(issuanceId))
	}

	bytes, err := hex.DecodeString(issuanceId)
	if err != nil {
		return "", err
	}

	// Encode account ID bytes to classic address
	issuerAddr, err := addresscodec.EncodeAccountIDToClassicAddress(bytes[4:])
	if err != nil {
		return "", fmt.Errorf("failed to encode account id to classic address: %w", err)
	}

	return issuerAddr, nil
}

// MPToken represents a Multi-Purpose Token with associated metadata.
// It contains document hash and signature information for asset-backed tokens.
type MPToken struct {
	DocumentHash string
	Issuer       string
}

// NewMPToken creates and returns a new MPToken instance.
// It requires a document hash and signature for token creation.
func NewMPToken(docHash, issuer string) MPToken {
	return MPToken{
		DocumentHash: docHash,
		Issuer:       issuer,
	}
}

// CreateMetadata generates the metadata structure required for MPT creation.
// This includes token details, URLs, and additional information like document hash and signature.
//
// Returns the metadata structure or an error if creation fails.
func (m MPToken) CreateMetadata() (MPTokenMetadata, error) {
	addInfo, err := json.Marshal(map[string]string{
		"document_hash": m.DocumentHash,
	})
	if err != nil {
		return MPTokenMetadata{}, fmt.Errorf("failed to marshal additional info: %w", err)
	}

	return MPTokenMetadata{
		Ticker:        "FSWRNT",
		Name:          "FortStock Warrant",
		Desc:          "Digital representation of real-world asset-backed warrants",
		AssetClass:    "rwa",
		AssetSubclass: "commodity",
		IssuerName:    m.Issuer,
		Urls: []MPTokenMetadataUrl{
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
