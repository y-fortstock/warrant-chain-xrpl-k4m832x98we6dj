// Package api provides the gRPC API implementations for the XRPL blockchain service.
// It includes implementations for account management, token operations, and blockchain interactions.
package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
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
	c *rpc.Client
	w *wallet.Wallet
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
	if txResp.Tx == nil {
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("transaction is nil")
	}

	if objMeta, ok := txResp.Meta.(transactions.TxObjMeta); ok {
		meta = objMeta
	} else {
		return nil, transactions.TxObjMeta{}, nil, fmt.Errorf("failed to cast metadata to TxObjMeta")
	}
	baseTx = &transactions.BaseTx{
		Account:            txResp.Tx["Account"].(types.Address),
		Fee:                txResp.Tx["Fee"].(types.XRPCurrencyAmount),
		Flags:              txResp.Tx["Flags"].(uint32),
		LastLedgerSequence: txResp.Tx["LastLedgerSequence"].(uint32),
		Sequence:           txResp.Tx["Sequence"].(uint32),
		SigningPubKey:      txResp.Tx["SigningPubKey"].(string),
		TransactionType:    transactions.TxType(txResp.Tx["TransactionType"].(string)),
		TxnSignature:       txResp.Tx["TxnSignature"].(string),
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

// SubmitTransactionAndWait submits a transaction to the XRPL network and waits for it to be confirmed.
//
// Parameters:
// - sender: The source wallet
// - tx: The transaction to submit
//
// Returns the transaction hash if successful, or an error if the submission fails.
func (b *Blockchain) SubmitTransactionAndWait(sender *wallet.Wallet, tx SubmittableTransaction) (hash string, err error) {
	if sender == nil {
		return "", fmt.Errorf("sender wallet cannot be nil")
	}

	txResponse, err := b.c.SubmitTxAndWait(tx.Flatten(), &rpctypes.SubmitOptions{
		Wallet:   sender,
		Autofill: true,
		FailHard: false,
	})

	if err != nil {
		return "", fmt.Errorf("failed to submit transaction: %w", err)
	}

	return txResponse.Hash.String(), nil
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

	hash, err := b.SubmitTx(w, tx)
	if err != nil {
		return "", "", fmt.Errorf("failed to submit tx: %w", err)
	}

	issuanceID, err = mpt.CreateIssuanceID(string(w.ClassicAddress), tx.Sequence)
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
func (m MPToken) CreateMetadata() (MPTokenMetadata, error) {
	addInfo, err := json.Marshal(map[string]string{
		"document_hash": m.DocumentHash,
		// "signature":     m.Signature,
	})
	if err != nil {
		return MPTokenMetadata{}, fmt.Errorf("failed to marshal additional info: %w", err)
	}

	return MPTokenMetadata{
		Ticker:        "FSWRNT",
		Name:          "FortStock Warrant",
		Desc:          "Digital representation of real-world asset-backed warrants",
		AssetClass:    "rwa",
		AssetSubclass: "other",
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

// Accounts already preconfigured to hold RLUSD on Devnet
const (
	// RLUSD Issuer https://dev.bithomp.com/explorer/rpBKbTPdestysw1jUkFxgcAH9pvxvWmzF8
	IssuerAddress = "rpBKbTPdestysw1jUkFxgcAH9pvxvWmzF8"
	IssuerSeed    = "sEdTHw4jqyZjwFz6wvnFLbYtQU2rXpn"

	// RLUSD Hex format for issued currency amount
	RLUSDHex = "524C555344000000000000000000000000000000"
)

// Loan Flow

// GetIssuerRLUSDWallet gets the pre-configured RLUSD issuer wallet on Devnet
func GetIssuerRLUSDWallet() (wallet.Wallet, error) {
	return wallet.FromSeed(IssuerSeed, "")
}

// CreateTrustline creates a trustline between an issuer and holder.
// If issuer is nil, it will use the default RLUSD issuer wallet.
func (b *Blockchain) CreateTrustline(issuer *wallet.Wallet, holder wallet.Wallet, trustSet transactions.TrustSet) (hash string, err error) {
	_, errValidation := trustSet.Validate()
	if errValidation != nil {
		return "", fmt.Errorf("failed to validate trustset: %w", err)
	}

	// Use provided issuer or get default RLUSD issuer wallet on Devnet
	var issuerWallet *wallet.Wallet
	if issuer != nil {
		issuerWallet = issuer
	} else {
		issuerRLUSD, err := GetIssuerRLUSDWallet()
		if err != nil {
			return "", fmt.Errorf("failed to get issuer wallet: %w", err)
		}
		issuerWallet = &issuerRLUSD
	}

	tx, err := b.c.SubmitTxAndWait(trustSet.Flatten(), &rpctypes.SubmitOptions{
		Wallet:   issuerWallet,
		Autofill: true,
		FailHard: false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create trustline: %w", err)
	}

	return tx.Hash.String(), nil
}

func (b *Blockchain) Deployment(borrower, lender wallet.Wallet, warrantMptIssuanceID string, mpt MPToken) (nil, err error) {
	// Borrower mints MPT
	_, issuanceID, err := b.MPTokenIssuanceCreate(&borrower, mpt)
	if err != nil {
		return nil, fmt.Errorf("failed to mint borrower MPT: %w", err)
	}

	// Lender authorizes MPT
	_, err = b.AuthorizeMPToken(&lender, issuanceID)
	if err != nil {
		return nil, fmt.Errorf("lender failed to authorize borrower MPT: %w", err)
	}

	// Borrower transfers MPT to lender
	_, err = b.TransferMPToken(&borrower, issuanceID, lender.ClassicAddress.String())
	if err != nil {
		return nil, fmt.Errorf("failed to transfer borrower MPT to lender: %w", err)
	}

	// Lender authorizes Warrant MPT
	_, err = b.AuthorizeMPToken(&lender, warrantMptIssuanceID)
	if err != nil {
		return nil, fmt.Errorf("lender failed to authorize warrant MPT: %w", err)
	}

	// Borrower transfers Warrant MPT to lender
	_, err = b.TransferMPToken(&borrower, warrantMptIssuanceID, lender.ClassicAddress.String())
	if err != nil {
		return nil, fmt.Errorf("borrower failed to transfer warrant MPT to lender: %w", err)
	}

	// Lender distribues 1000000 RLUSD to borrower
	payment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: lender.ClassicAddress,
		},
		Amount: types.IssuedCurrencyAmount{
			Issuer:   IssuerAddress,
			Currency: RLUSDHex,
			Value:    "1000000",
		},
		Destination: borrower.ClassicAddress,
	}
	paymentTx, err := b.c.SubmitTxAndWait(payment.Flatten(), &rpctypes.SubmitOptions{
		Wallet:   &lender,
		Autofill: true,
		FailHard: false,
	})
	if err != nil {
		return nil, fmt.Errorf("lender failed to distribute RLUSD to borrower: %w", err)
	}
	fmt.Printf("#️⃣ RLUSD Payment Hash: %+v\n", paymentTx.Hash)

	return nil, nil
}

type RepayFullLoan struct {
	Borrower             wallet.Wallet
	Lender               wallet.Wallet
	Warehouse            wallet.Wallet
	LoanAmount           types.IssuedCurrencyAmount
	WarrantMptIssuanceID string
}

// RepayFullLoan repays a full loan by sending RLUSD to the lender and burning the Debt MPT.
// It also returns the warrant MPT to the borrower and the warehouse.
//
// Parameters:
// - borrower: The borrower's wallet
// - lender: The lender's wallet
// - warehouse: The warehouse's wallet
// - loanAmount: The amount of RLUSD to repay
// - warrantMptIssuanceID: The issuance ID of the warrant MPT
//
// Returns an error if the repayment fails.
func (b *Blockchain) RepayFullLoan(rfl RepayFullLoan) (err error) {
	// Borrower repays full loan as payment RLUSD to lender
	payment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: rfl.Borrower.ClassicAddress,
		},
		Destination: rfl.Lender.ClassicAddress,
		Amount:      rfl.LoanAmount,
	}
	paymentTx, err := b.c.SubmitTxAndWait(payment.Flatten(), &rpctypes.SubmitOptions{
		Wallet:   &rfl.Borrower,
		Autofill: true,
		FailHard: false,
	})
	if err != nil {
		return fmt.Errorf("borrower failed to repay full loan: %w", err)
	}
	fmt.Printf("#️⃣ Repayment RLUSD Payment Hash: %+v\n", paymentTx.Hash)

	// Borrower burns Debt MPT
	burnMPT := &transactions.MPTokenIssuanceDestroy{
		BaseTx: transactions.BaseTx{
			Account: rfl.Lender.ClassicAddress,
		},
		MPTokenIssuanceID: rfl.WarrantMptIssuanceID,
	}

	burnMPTTx, err := b.c.SubmitTxAndWait(burnMPT.Flatten(), &rpctypes.SubmitOptions{
		Wallet:   &rfl.Borrower,
		Autofill: true,
		FailHard: false,
	})
	if err != nil {
		return fmt.Errorf("borrower failed to burn Debt MPT: %w", err)
	}
	fmt.Printf("#️⃣ Burn Debt MPT Hash: %+v\n", burnMPTTx.Hash)

	// Lender return warrant MPT to borrower
	returnWarrantMPT := transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: rfl.Lender.ClassicAddress,
		},
		Destination: rfl.Borrower.ClassicAddress,
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: rfl.WarrantMptIssuanceID,
			Value:         "1",
		},
	}
	returnWarrantMPTTx, err := b.c.SubmitTxAndWait(returnWarrantMPT.Flatten(), &rpctypes.SubmitOptions{
		Wallet:   &rfl.Lender,
		Autofill: true,
		FailHard: false,
	})
	if err != nil {
		return fmt.Errorf("lender failed to return warrant MPT to borrower: %w", err)
	}
	fmt.Printf("#️⃣ Return Warrant MPT Hash: %+v\n", returnWarrantMPTTx.Hash)

	// Borrower return warrant MPT to warehouse
	returnWarrantMPTToBorrower := transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: rfl.Borrower.ClassicAddress,
		},
		Destination: rfl.Warehouse.ClassicAddress,
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: rfl.WarrantMptIssuanceID,
			Value:         "1",
		},
	}
	returnWarrantMPTTxToBorrowerTx, err := b.c.SubmitTxAndWait(returnWarrantMPTToBorrower.Flatten(), &rpctypes.SubmitOptions{
		Wallet:   &rfl.Borrower,
		Autofill: true,
		FailHard: false,
	})
	if err != nil {
		return fmt.Errorf("borrower failed to return warrant MPT to warehouse: %w", err)
	}
	fmt.Printf("#️⃣ Return Warrant MPT Hash: %+v\n", returnWarrantMPTTxToBorrowerTx.Hash)

	// Warehouse burns warrant MPT
	burnWarrantMPT := &transactions.MPTokenIssuanceDestroy{
		BaseTx: transactions.BaseTx{
			Account: rfl.Warehouse.ClassicAddress,
		},
		MPTokenIssuanceID: rfl.WarrantMptIssuanceID,
	}

	burnWarrantMPTTx, err := b.c.SubmitTxAndWait(burnWarrantMPT.Flatten(), &rpctypes.SubmitOptions{
		Wallet:   &rfl.Warehouse,
		Autofill: true,
		FailHard: false,
	})
	if err != nil {
		return fmt.Errorf("warehouse failed to burn warrant MPT: %w", err)
	}
	fmt.Printf("#️⃣ Burn Warrant MPT Hash: %+v\n", burnWarrantMPTTx.Hash)

	return nil
}

type DefaultRepayFullLoan struct {
	Borrower             wallet.Wallet
	Lender               wallet.Wallet
	Warehouse            wallet.Wallet
	LoanAmount           types.IssuedCurrencyAmount
	DebtMptIssuanceID    string
	WarrantMptIssuanceID string
}

// DefaultRepayFullLoan executes the step in the case the loan can't be repaid on time.
// It returns the warrant MPT to the warehouse and burns the Debt MPT.
//
// Parameters:
// - dfl: The DefaultRepayFullLoan struct containing the necessary information
//
// Returns an error if the repayment fails.
func (b *Blockchain) DefaultRepayFullLoan(dfl DefaultRepayFullLoan) (err error) {
	// Lender returns Debt MPT to borrower
	returnDebtMPT := transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: dfl.Lender.ClassicAddress,
		},
		Destination: dfl.Borrower.ClassicAddress,
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: dfl.DebtMptIssuanceID,
			Value:         "1",
		},
	}
	returnDebtMPTTx, err := b.c.SubmitTxAndWait(returnDebtMPT.Flatten(), &rpctypes.SubmitOptions{
		Wallet:   &dfl.Lender,
		Autofill: true,
		FailHard: false,
	})
	if err != nil {
		return fmt.Errorf("lender failed to return debt MPT to borrower: %w", err)
	}
	fmt.Printf("#️⃣ Return Debt MPT Hash: %+v\n", returnDebtMPTTx.Hash)

	// Borrower burns Debt MPT
	burnDebtMPT := &transactions.MPTokenIssuanceDestroy{
		BaseTx: transactions.BaseTx{
			Account: dfl.Borrower.ClassicAddress,
		},
		MPTokenIssuanceID: dfl.DebtMptIssuanceID,
	}

	burnDebtMPTTx, err := b.c.SubmitTxAndWait(burnDebtMPT.Flatten(), &rpctypes.SubmitOptions{
		Wallet:   &dfl.Borrower,
		Autofill: true,
		FailHard: false,
	})
	if err != nil {
		return fmt.Errorf("borrower failed to burn Debt MPT: %w", err)
	}
	fmt.Printf("#️⃣ Burn Debt MPT Hash: %+v\n", burnDebtMPTTx.Hash)

	// Lender returns warrant MPT to warehouse
	returnWarrantMPT := transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: dfl.Lender.ClassicAddress,
		},
		Destination: dfl.Warehouse.ClassicAddress,
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: dfl.WarrantMptIssuanceID,
			Value:         "1",
		},
	}
	returnWarrantMPTTx, err := b.c.SubmitTxAndWait(returnWarrantMPT.Flatten(), &rpctypes.SubmitOptions{
		Wallet:   &dfl.Lender,
		Autofill: true,
		FailHard: false,
	})
	if err != nil {
		return fmt.Errorf("lender failed to return warrant MPT to warehouse: %w", err)
	}
	fmt.Printf("#️⃣ Return Warrant MPT Hash: %+v\n", returnWarrantMPTTx.Hash)

	// Warehouse burns warrant MPT
	burnWarrantMPT := transactions.MPTokenIssuanceDestroy{
		BaseTx: transactions.BaseTx{
			Account: dfl.Warehouse.ClassicAddress,
		},
		MPTokenIssuanceID: dfl.WarrantMptIssuanceID,
	}
	burnWarrantMPTTx, err := b.c.SubmitTxAndWait(burnWarrantMPT.Flatten(), &rpctypes.SubmitOptions{
		Wallet:   &dfl.Warehouse,
		Autofill: true,
		FailHard: false,
	})
	if err != nil {
		return fmt.Errorf("warehouse failed to burn warrant MPT: %w", err)
	}
	fmt.Printf("#️⃣ Burn Warrant MPT Hash: %+v\n", burnWarrantMPTTx.Hash)

	return
}
