package api

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

const (
	LoanAmount       = 1_000_000
	LoanCurrency     = "RLUSD"
	LoanInterestRate = 36.5
	LoanPeriod       = 10 * time.Minute

	// RLUSD Hex format for issued currency amount
	RLUSDHex = "524C555344000000000000000000000000000000"
)

func (b *Blockchain) SystemAccountInit() error {
	accountSet := &transaction.AccountSet{}
	accountSet.SetAsfDefaultRipple()

	return b.SubmitTxAndWait(b.w, accountSet)
}

func (b *Blockchain) CreateTrustline(from, to *wallet.Wallet, amount float64) error {
	trustline := &transaction.TrustSet{
		LimitAmount: types.IssuedCurrencyAmount{
			Issuer:   from.ClassicAddress,
			Currency: RLUSDHex,
			Value:    strconv.FormatFloat(amount, 'f', -1, 64),
		},
	}
	trustline.SetClearNoRippleFlag()

	return b.SubmitTxAndWait(to, trustline)
}

func (b *Blockchain) CreateTrustlineFromSystemAccount(to *wallet.Wallet, amount float64) error {
	if err := b.CreateTrustline(b.w, to, amount); err != nil {
		return fmt.Errorf("failed to create trustline from system account: %v", err)
	}

	return b.CreateTrustline(to, b.w, 0)
}

func (b *Blockchain) PaymentRLUSDFromSystemAccount(to *wallet.Wallet, amount float64) error {
	return b.PaymentRLUSD(b.w, to, amount)
}

func (b *Blockchain) PaymentRLUSDToSystemAccount(from *wallet.Wallet, amount float64) error {
	return b.PaymentRLUSD(from, b.w, amount)
}

func (b *Blockchain) PaymentRLUSD(from, to *wallet.Wallet, amount float64) error {
	payment := &transaction.Payment{
		Amount: types.IssuedCurrencyAmount{
			Issuer:   b.w.ClassicAddress,
			Currency: RLUSDHex,
			Value:    strconv.FormatFloat(amount, 'f', -1, 64),
		},
		Destination: to.ClassicAddress,
	}

	return b.SubmitTxAndWait(from, payment)
}
