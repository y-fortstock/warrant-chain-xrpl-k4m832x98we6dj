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

// Account is an implementation of accountv1.AccountAPIServer.
type Account struct {
	accountv1.UnimplementedAccountAPIServer
	bc     *Blockchain
	logger *slog.Logger
}

// NewAccount returns a new Account implementation.
func NewAccount(l *slog.Logger, bc *Blockchain) *Account {
	return &Account{logger: l, bc: bc}
}

// Create creates a new ETH account with a password.
func (a *Account) Create(ctx context.Context, req *accountv1.CreateRequest) (*accountv1.CreateResponse, error) {
	l := a.logger.With("method", "Create")
	l.Debug("start")
	seeds := strings.Split(req.Password, "-")
	w, err := crypto.NewWalletFromHexSeed(seeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", seeds[1]))
	if err != nil {
		l.Error("failed to get XRPL address", "error", err)
		return nil, err
	}

	l.Info("account created", "address", w.Address)
	return &accountv1.CreateResponse{
		Account: &accountv1.Account{
			Id: string(w.Address),
		},
	}, nil
}

// Deposit deposits XRP in drops from system account.
func (a *Account) Deposit(ctx context.Context, req *accountv1.DepositRequest) (*accountv1.DepositResponse, error) {
	l := a.logger.With("method", "Deposit", "account", req.AccountId)
	l.Debug("start", "amount", req.WeiAmount)

	dropsToTransfer, err := strconv.ParseUint(req.WeiAmount, 10, 64)
	if err != nil {
		l.Error("failed to parse amount", "error", err, "amount", req.WeiAmount)
		return nil, fmt.Errorf("invalid amount: %s", req.WeiAmount)
	}

	l.Info("payment from system account", "dropsToTransfer", dropsToTransfer)
	txHash, err := a.bc.PaymentFromSystemAccount(req.AccountId, dropsToTransfer)
	if err != nil {
		l.Error("failed to payment from system account",
			"error", err,
			"account", req.AccountId,
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

// ClearBalance clears the account balance.
func (a *Account) ClearBalance(ctx context.Context, req *accountv1.ClearBalanceRequest) (*accountv1.ClearBalanceResponse, error) {
	l := a.logger.With("method", "ClearBalance", "account", req.AccountId)
	l.Debug("start")

	seeds := strings.Split(req.AccountPassword, "-")
	w, err := crypto.NewWalletFromHexSeed(seeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", seeds[1]))
	if err != nil {
		l.Error("failed to get XRPL address", "error", err)
		return nil, err
	}
	if string(w.Address) != req.AccountId {
		l.Error("account id mismatch", "address", w.Address, "accountId", req.AccountId)
		return nil, fmt.Errorf("account id mismatch: %s != %s", w.Address, req.AccountId)
	}

	info, err := a.bc.GetAccountInfo(req.AccountId)
	if err != nil {
		l.Error("failed to get account balance", "error", err)
		return nil, err
	}
	balance := uint64(info.AccountData.Balance)

	feeRaw, reserveRaw, err := a.bc.GetBaseFeeAndReserve()
	if err != nil {
		l.Error("failed to get base fee and reserve", "error", err)
		return nil, err
	}
	fee := uint64(feeRaw * xrpToDrops * 120 / 100) // 20% margin
	reserve := uint64(reserveRaw * xrpToDrops)

	if balance <= (fee + reserve) {
		l.Warn("account balance is less or equal than fee + reserve", "balance", balance, "fee", fee, "reserve", reserve)
		return nil, fmt.Errorf("account balance is less or equal than fee + reserve: %d <= %d", balance, fee+reserve)
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

// GetBalance gets the account balance.
func (a *Account) GetBalance(ctx context.Context, req *accountv1.GetBalanceRequest) (*accountv1.GetBalanceResponse, error) {
	l := a.logger.With("method", "GetBalance", "account", req.AccountId)
	l.Debug("start")

	info, err := a.bc.GetAccountInfo(req.AccountId)
	if err != nil {
		l.Error("failed to get account balance", "error", err)
		return nil, err
	}
	balance := uint64(info.AccountData.Balance)

	l.Info("account balance response", "balance", balance)
	return &accountv1.GetBalanceResponse{
		Balance: strconv.FormatUint(balance, 10),
	}, nil
}
