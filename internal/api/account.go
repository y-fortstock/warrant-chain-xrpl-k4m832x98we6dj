package api

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"

	bip32 "github.com/tyler-smith/go-bip32"
	accountv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/account/v1"
)

// Account is an implementation of accountv1.AccountAPIServer.
type Account struct {
	accountv1.UnimplementedAccountAPIServer
	logger *slog.Logger
}

// NewAccount returns a new Account implementation.
func NewAccount(logger *slog.Logger) *Account {
	return &Account{logger: logger}
}

// Create creates a new ETH account with a password.
func (a *Account) Create(ctx context.Context, req *accountv1.CreateRequest) (*accountv1.CreateResponse, error) {
	a.logger.Debug("create account")
	seed, err := hex.DecodeString(strings.Split(req.Password, "-")[0])
	if err != nil {
		a.logger.Error("failed to decode seed", "error", err)
		return nil, err
	}
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		a.logger.Error("failed to create master key", "error", err)
		return nil, err
	}
	fmt.Println(masterKey)

	return nil, nil // TODO: implement
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
	return nil, nil // TODO: implement
}
