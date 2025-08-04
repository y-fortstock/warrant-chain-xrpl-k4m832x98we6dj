package api

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

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
	a.logger.Info("create account")
	seeds := strings.Split(req.Password, "-")
	address, _, err := a.bc.GetXRPLWallet(seeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", seeds[1]))
	if err != nil {
		a.logger.Error("failed to get XRPL address", "error", err)
		return nil, err
	}

	a.logger.Info("account created", "address", address)
	return &accountv1.CreateResponse{
		Account: &accountv1.Account{
			Id: address,
		},
	}, nil
}

// Deposit deposits XRP in drops from system account.
func (a *Account) Deposit(ctx context.Context, req *accountv1.DepositRequest) (*accountv1.DepositResponse, error) {
	a.logger.Info("deposit request", "account", req.AccountId, "amount", req.WeiAmount)

	balance, err := a.bc.GetAccountBalance(a.bc.systemAccount)
	if err != nil {
		a.logger.Error("failed to get system account balance", "error", err)
		return nil, err
	}

	fee, err := a.bc.GetBaseFee()
	if err != nil {
		a.logger.Error("failed to get base fee", "error", err)
		return nil, err
	}
	fee = fee * 120 / 100 // 20% margin

	dropsToTransfer, err := strconv.ParseUint(req.WeiAmount, 10, 64)
	if err != nil {
		a.logger.Error("failed to parse wei amount", "error", err, "weiAmount", req.WeiAmount)
		return nil, fmt.Errorf("invalid wei amount: %s", req.WeiAmount)
	}

	if balance < dropsToTransfer+fee {
		a.logger.Error("system account balance is less than drops to transfer",
			"balance", balance,
			"dropsToTransfer", dropsToTransfer,
			"fee", fee)
		return nil, fmt.Errorf("system account balance is less than drops to transfer: %d < %d", balance, dropsToTransfer+fee)
	}

	txHash, err := a.bc.PaymentFromSystemAccount(req.AccountId, fee, dropsToTransfer)
	if err != nil {
		a.logger.Error("failed to payment from system account",
			"error", err,
			"account", req.AccountId,
			"fee", fee,
			"dropsToTransfer", dropsToTransfer)
		return nil, err
	}

	a.logger.Info("deposit response", "account", req.AccountId, "txHash", txHash)
	return &accountv1.DepositResponse{
		Transaction: &typesv1.Transaction{
			Id:          txHash,
			BlockNumber: []byte{0},
			BlockTime:   uint64(time.Now().Unix()),
		},
	}, nil
}

// ClearBalance clears the account balance.
func (a *Account) ClearBalance(ctx context.Context, req *accountv1.ClearBalanceRequest) (*accountv1.ClearBalanceResponse, error) {
	return nil, nil // TODO: implement
}

// GetBalance gets the account balance.
func (a *Account) GetBalance(ctx context.Context, req *accountv1.GetBalanceRequest) (*accountv1.GetBalanceResponse, error) {
	a.logger.Info("get balance request", "account", req.AccountId)

	balance, err := a.bc.GetAccountBalance(req.AccountId)
	if err != nil {
		a.logger.Error("failed to get account balance", "error", err, "account", req.AccountId)
		return nil, err
	}

	a.logger.Info("account balance response", "account", req.AccountId, "balance", balance)
	return &accountv1.GetBalanceResponse{
		Balance: strconv.FormatUint(balance, 10),
	}, nil
}
