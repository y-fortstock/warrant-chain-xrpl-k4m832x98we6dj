package api

import (
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
