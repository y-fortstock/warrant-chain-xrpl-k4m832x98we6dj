package types

import (
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

type SubmitOptions struct {
	Autofill bool
	Wallet   *wallet.Wallet
	FailHard bool
}
