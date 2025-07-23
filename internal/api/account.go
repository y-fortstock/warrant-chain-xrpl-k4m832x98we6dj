package api

import (
	"context"

	accountv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/account/v1"
)

// Account is an implementation of accountv1.AccountAPIServer.
type Account struct {
	accountv1.UnimplementedAccountAPIServer
}

// NewAccount returns a new Account implementation.
func NewAccount() *Account {
	return &Account{}
}

// Create creates a new ETH account with a password.
func (a *Account) Create(ctx context.Context, req *accountv1.CreateRequest) (*accountv1.CreateResponse, error) {
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
