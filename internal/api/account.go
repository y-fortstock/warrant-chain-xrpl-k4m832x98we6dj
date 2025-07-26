package api

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

	accountv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/account/v1"
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
	a.logger.Debug("create account")
	address, err := a.bc.GetXRPLAddress(strings.Split(req.Password, "-")[0])
	if err != nil {
		a.logger.Error("failed to get XRPL address", "error", err)
		return nil, err
	}

	a.logger.Debug("account created", "address", address)
	return &accountv1.CreateResponse{
		Account: &accountv1.Account{
			Id: address,
		},
	}, nil
}

// Deposit deposits ETH in wei from system account.
func (a *Account) Deposit(ctx context.Context, req *accountv1.DepositRequest) (*accountv1.DepositResponse, error) {
	return nil, nil // TODO: implement
}

// ClearBalance clears the account balance.
func (a *Account) ClearBalance(ctx context.Context, req *accountv1.ClearBalanceRequest) (*accountv1.ClearBalanceResponse, error) {
	return nil, nil // TODO: implement
}

// GetBalance gets the account balance.
func (a *Account) GetBalance(ctx context.Context, req *accountv1.GetBalanceRequest) (*accountv1.GetBalanceResponse, error) {
	a.logger.Debug("get balance request", "account", req.AccountId)

	balance, err := a.bc.GetAccountBalance(req.AccountId)
	if err != nil {
		a.logger.Error("failed to get account balance", "error", err, "account", req.AccountId)
		return nil, err
	}

	a.logger.Debug("account balance response", "account", req.AccountId, "balance", balance)
	return &accountv1.GetBalanceResponse{
		Balance: strconv.FormatUint(balance, 10),
	}, nil
}
