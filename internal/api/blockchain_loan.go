package api

import (
	"strconv"
	"time"

	transactions "github.com/Peersyst/xrpl-go/xrpl/transaction"
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

func (b *Blockchain) CreateTrustline(from, to *wallet.Wallet, amount float64) (txHash string, err error) {
	trustline := &transactions.TrustSet{
		LimitAmount: types.IssuedCurrencyAmount{
			Issuer:   from.ClassicAddress,
			Currency: RLUSDHex,
			Value:    strconv.FormatFloat(amount, 'f', -1, 64),
		},
	}

	return b.SubmitTxAndWait(to, trustline)
}

func (b *Blockchain) CreateTrustlineFromSystemAccount(to *wallet.Wallet, amount float64) (txHash string, err error) {
	return b.CreateTrustline(b.w, to, amount)
}

func (b *Blockchain) PaymentRLUSDFromSystemAccount(to *wallet.Wallet, amount float64) (txHash string, err error) {
	return b.PaymentRLUSD(b.w, to, amount)
}

func (b *Blockchain) PaymentRLUSDToSystemAccount(from *wallet.Wallet, amount float64) (txHash string, err error) {
	return b.PaymentRLUSD(from, b.w, amount)
}

func (b *Blockchain) PaymentRLUSD(from, to *wallet.Wallet, amount float64) (txHash string, err error) {
	payment := &transactions.Payment{
		Amount: types.IssuedCurrencyAmount{
			Issuer:   from.ClassicAddress,
			Currency: RLUSDHex,
			Value:    strconv.FormatFloat(amount, 'f', -1, 64),
		},
		Destination: to.ClassicAddress,
	}

	return b.SubmitTxAndWait(from, payment)
}
