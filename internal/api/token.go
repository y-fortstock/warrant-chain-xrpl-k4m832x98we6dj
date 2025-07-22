package api

import (
	tokenv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/token/v1"
)

// Token is an implementation of tokenv1.TokenAPIServer.
type Token struct {
	tokenv1.UnimplementedTokenAPIServer
}

// NewToken returns a new Token implementation.
func NewToken() *Token {
	return &Token{}
}
