// Package api provides the gRPC API implementations for the XRPL blockchain service.
// It includes implementations for account management, token operations, and blockchain interactions.
package api

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"gitlab.com/warrant1/warrant/chain-xrpl/internal/crypto"
	accountv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/account/v1"
	typesv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/types/v1"
)

// Account implements the accountv1.AccountAPIServer interface.
// It provides methods for creating, managing, and querying XRPL accounts.
type Account struct {
	accountv1.UnimplementedAccountAPIServer
	bc     *Blockchain
	logger *slog.Logger
}

// NewAccount creates and returns a new Account API server instance.
// It requires a logger and blockchain instance for operation.
func NewAccount(l *slog.Logger, bc *Blockchain) *Account {
	return &Account{logger: l, bc: bc}
}

// Create creates a new XRPL account using the provided password.
// The password should be in the format "hexSeed-derivationIndex" where:
// - hexSeed is a 64-character hexadecimal string representing the master seed
// - derivationIndex is the BIP-44 derivation path index
//
// Returns the created account information or an error if creation fails.
func (a *Account) Create(ctx context.Context, req *accountv1.CreateRequest) (*accountv1.CreateResponse, error) {
	l := a.logger.With("method", "Create")
	l.Debug("start")
	seeds := strings.Split(req.GetPassword(), "-")
	if len(seeds) != 2 {
		l.Error("invalid password format", "password", req.GetPassword())
		return nil, fmt.Errorf("invalid password format: %s", req.GetPassword())
	}
	w, err := crypto.NewWalletFromHexSeed(seeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", seeds[1]))
	if err != nil {
		l.Error("failed to get XRPL address", "error", err)
		return nil, err
	}

	l.Info("account created", "address", w.ClassicAddress)
	return &accountv1.CreateResponse{
		Account: &accountv1.Account{
			Id: string(w.ClassicAddress),
		},
	}, nil
}

// Deposit transfers XRP from the system account to the specified account.
// The amount is specified in drops (the smallest unit of XRP, where 1 XRP = 1,000,000 drops).
//
// Parameters:
// - req.AccountId: The destination account address
// - req.WeiAmount: The amount to deposit in drops (as a string)
//
// Returns transaction details including the transaction hash and timestamp.
func (a *Account) Deposit(ctx context.Context, req *accountv1.DepositRequest) (*accountv1.DepositResponse, error) {
	l := a.logger.With("method", "Deposit", "account", req.GetAccountId())
	l.Debug("start", "amount", req.GetWeiAmount())
	a.bc.Lock()
	defer a.bc.Unlock()

	dropsToTransfer, err := strconv.ParseUint(req.GetWeiAmount(), 10, 64)
	if err != nil {
		l.Error("failed to parse amount", "error", err, "amount", req.GetWeiAmount())
		return nil, fmt.Errorf("invalid amount: %s", req.GetWeiAmount())
	}

	l.Info("payment from system account", "dropsToTransfer", dropsToTransfer)
	txHash, err := a.bc.PaymentFromSystemAccount(req.AccountId, dropsToTransfer)
	if err != nil {
		l.Error("failed to payment from system account",
			"error", err,
			"account", req.GetAccountId(),
			"dropsToTransfer", dropsToTransfer)
		return nil, err
	}

	l.Info("deposit response", "txHash", txHash)
	return &accountv1.DepositResponse{
		Transaction: &typesv1.Transaction{
			Id:        txHash,
			BlockTime: uint64(time.Now().Unix()),
		},
	}, nil
}

// ClearBalance transfers all available XRP from the specified account back to the system account,
// leaving only the minimum reserve and transaction fee.
//
// The account password must match the account ID to authorize the operation.
// The function calculates the available balance by subtracting the reserve and estimated fee.
//
// Parameters:
// - req.AccountId: The account address to clear
// - req.AccountPassword: The password in format "hexSeed-derivationIndex"
//
// Returns transaction details if successful, or an error if the balance is insufficient.
func (a *Account) ClearBalance(ctx context.Context, req *accountv1.ClearBalanceRequest) (*accountv1.ClearBalanceResponse, error) {
	l := a.logger.With("method", "ClearBalance", "account", req.GetAccountId())
	l.Debug("start")
	a.bc.Lock()
	defer a.bc.Unlock()

	seeds := strings.Split(req.GetAccountPassword(), "-")
	if len(seeds) != 2 {
		l.Error("invalid password format", "password", req.GetAccountPassword())
		return nil, fmt.Errorf("invalid password format: %s", req.GetAccountPassword())
	}
	w, err := crypto.NewWalletFromHexSeed(seeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", seeds[1]))
	if err != nil {
		l.Error("failed to get XRPL address", "error", err)
		return nil, err
	}
	if string(w.ClassicAddress) != req.GetAccountId() {
		l.Error("account id mismatch", "address", w.ClassicAddress, "accountId", req.GetAccountId())
		return nil, fmt.Errorf("account id mismatch: %s != %s", w.ClassicAddress, req.GetAccountId())
	}

	info, err := a.bc.GetAccountInfo(req.GetAccountId())
	if err != nil {
		l.Error("failed to get account balance", "error", err)
		return nil, err
	}
	balance := uint64(info.AccountData.Balance)

	srvInfo, err := a.bc.GetBaseFeeAndReserve()
	if err != nil {
		l.Error("failed to get base fee and reserve", "error", err)
		return nil, err
	}

	mptCnt, err := a.bc.GetMPTokenCount(req.GetAccountId())
	if err != nil {
		l.Error("failed to get mp token count", "error", err)
		return nil, err
	}

	fee := uint64(srvInfo.BaseFeeXRP * xrpToDrops * 120 / 100) // 20% margin
	reserve := uint64((srvInfo.ReserveBaseXRP + srvInfo.ReserveIncXRP*float32(mptCnt)) * xrpToDrops)
	l.Debug("reserves",
		"count", mptCnt,
		"baseReserve", srvInfo.ReserveBaseXRP,
		"incReserve", srvInfo.ReserveIncXRP,
	)

	if balance <= (fee + reserve) {
		l.Warn("account balance is less or equal than fee + reserve", "balance", balance, "fee", fee, "reserve", reserve)

		return &accountv1.ClearBalanceResponse{
			Transaction: &typesv1.Transaction{
				Id:        "0",
				BlockTime: uint64(time.Now().Unix()),
			},
		}, nil
	}

	amount := balance - (fee + reserve)

	l.Info("payment to system account", "fee", fee, "reserve", reserve, "amount", amount)
	txHash, err := a.bc.PaymentToSystemAccount(w, amount)
	if err != nil {
		l.Error("failed to payment to system account",
			"error", err,
			"fee", fee,
			"reserve", reserve,
			"amount", amount,
			"hash", txHash)
		return nil, err
	}

	return &accountv1.ClearBalanceResponse{
		Transaction: &typesv1.Transaction{
			Id:        txHash,
			BlockTime: uint64(time.Now().Unix()),
		},
	}, nil
}

// GetBalance retrieves the current XRP balance of the specified account.
// The balance is returned in drops (the smallest unit of XRP).
//
// Parameters:
// - req.AccountId: The account address to query
//
// Returns the account balance as a string representation of drops.
func (a *Account) GetBalance(ctx context.Context, req *accountv1.GetBalanceRequest) (*accountv1.GetBalanceResponse, error) {
	l := a.logger.With("method", "GetBalance", "account", req.GetAccountId())
	l.Debug("start")

	info, err := a.bc.GetAccountInfo(req.GetAccountId())
	if err != nil {
		if strings.Contains(err.Error(), "actNotFound") {
			return &accountv1.GetBalanceResponse{
				Balance: "0",
			}, nil
		}
		l.Error("failed to get account balance", "error", err)
		return nil, err
	}
	balance := uint64(info.AccountData.Balance)

	l.Info("account balance response", "balance", balance)
	return &accountv1.GetBalanceResponse{
		Balance: strconv.FormatUint(balance, 10),
	}, nil
}
