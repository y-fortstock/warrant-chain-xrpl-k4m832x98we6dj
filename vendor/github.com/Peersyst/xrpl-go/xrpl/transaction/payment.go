package transaction

import (
	"errors"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// Do not use the default path; only use paths included in the Paths field.
	// This is intended to force the transaction to take arbitrage opportunities.
	// Most clients do not need this.
	tfRippleNotDirect uint32 = 65536
	// If the specified Amount cannot be sent without spending more than SendMax,
	// reduce the received amount instead of failing outright. See Partial
	// Payments for more details.
	tfPartialPayment uint32 = 131072
	// Only take paths where all the conversions have an input:output ratio that
	// is equal or better than the ratio of Amount:SendMax. See Limit Quality for
	// details.
	tfLimitQuality uint32 = 262144
)

// Errors
var (
	// ErrPartialPaymentFlagRequired is returned when the tfPartialPayment flag is required but not set.
	ErrPartialPaymentFlagRequired = errors.New("tfPartialPayment flag required with DeliverMin")
)

// A Payment transaction represents a transfer of value from one account to another.
type Payment struct {
	BaseTx
	// API v1: Only available in API v1.
	// The maximum amount of currency to deliver.
	// For non-XRP amounts, the nested field names MUST be lower-case.
	// If the tfPartialPayment flag is set, deliver up to this amount instead.
	Amount types.CurrencyAmount

	// Set of Credentials to authorize a deposit made by this transaction.
	// Each member of the array must be the ledger entry ID of a Credential entry in the ledger.
	// For details see https://xrpl.org/docs/references/protocol/transactions/types/payment#credential-ids
	CredentialIDs types.CredentialIDs `json:",omitempty"`

	// API v2: Only available in API v2.
	// The maximum amount of currency to deliver.
	// For non-XRP amounts, the nested field names MUST be lower-case.
	// If the tfPartialPayment flag is set, deliver up to this amount instead.
	DeliverMax types.CurrencyAmount `json:",omitempty"`

	// (Optional) Minimum amount of destination currency this transaction should deliver.
	// Only valid if this is a partial payment.
	// For non-XRP amounts, the nested field names are lower-case.
	DeliverMin types.CurrencyAmount `json:",omitempty"`

	// The unique address of the account receiving the payment.
	Destination types.Address

	// (Optional) Arbitrary tag that identifies the reason for the payment to the destination, or a hosted recipient to pay.
	DestinationTag *uint32 `json:",omitempty"`

	// (Optional) Arbitrary 256-bit hash representing a specific reason or identifier for this payment
	InvoiceID types.Hash256 `json:",omitempty"`

	// (Optional, auto-fillable) Array of payment paths to be used for this transaction.
	// Must be omitted for XRP-to-XRP transactions.
	Paths [][]PathStep `json:",omitempty"`

	// (Optional) Highest amount of source currency this transaction is allowed to cost,
	// including transfer fees, exchange rates, and slippage.
	// Does not include the XRP destroyed as a cost for submitting the transaction.
	// For non-XRP amounts, the nested field names MUST be lower-case.
	// Must be supplied for cross-currency/cross-issue payments.
	// Must be omitted for XRP-to-XRP payments.
	SendMax types.CurrencyAmount `json:",omitempty"`
}

// TxType returns the type of the transaction (Payment).
func (Payment) TxType() TxType {
	return PaymentTx
}

// Flatten returns the flattened map of the Payment transaction.
func (p *Payment) Flatten() FlatTransaction {
	// Add BaseTx fields
	flattened := p.BaseTx.Flatten()

	// Add Payment-specific fields
	flattened["TransactionType"] = "Payment"

	if p.Amount != nil {
		flattened["Amount"] = p.Amount.Flatten()
	}

	if len(p.CredentialIDs) > 0 {
		flattened["CredentialIDs"] = p.CredentialIDs.Flatten()
	}

	if p.DeliverMax != nil {
		flattened["DeliverMax"] = p.DeliverMax.Flatten()
	}

	if p.DeliverMin != nil {
		flattened["DeliverMin"] = p.DeliverMin.Flatten()
	}

	if p.Destination != "" {
		flattened["Destination"] = p.Destination.String()
	}

	if p.DestinationTag != nil {
		flattened["DestinationTag"] = *p.DestinationTag
	}

	if p.InvoiceID != "" {
		flattened["InvoiceID"] = p.InvoiceID.String()
	}

	if len(p.Paths) > 0 {
		flattenedPaths := make([][]interface{}, len(p.Paths))
		for i, path := range p.Paths {
			flattenedPath := make([]interface{}, len(path))
			for j, step := range path {
				flattenedStep := step.Flatten()
				if flattenedStep != nil {
					flattenedPath[j] = flattenedStep
				}
			}
			flattenedPaths[i] = flattenedPath
		}
		flattened["Paths"] = flattenedPaths
	}

	if p.SendMax != nil {
		flattened["SendMax"] = p.SendMax.Flatten()
	}

	return flattened
}

// SetRippleNotDirectFlag sets the RippleNotDirect flag.
//
// RippleNotDirect: Do not use the default path; only use paths included in the Paths field.
// This is intended to force the transaction to take arbitrage opportunities.
// Most clients do not need this.
func (p *Payment) SetRippleNotDirectFlag() {
	p.Flags |= tfRippleNotDirect
}

// SetPartialPaymentFlag sets the PartialPayment flag.
//
// PartialPayment: If the specified Amount cannot be sent without spending more than SendMax,
// reduce the received amount instead of failing outright. See Partial
// Payments for more details.
func (p *Payment) SetPartialPaymentFlag() {
	p.Flags |= tfPartialPayment
}

// SetLimitQualityFlag sets the LimitQuality flag.
//
// LimitQuality: Only take paths where all the conversions have an input:output ratio that
// is equal or better than the ratio of Amount:SendMax. See Limit Quality for
// details.
func (p *Payment) SetLimitQualityFlag() {
	p.Flags |= tfLimitQuality
}

// ValidatePayment validates the Payment struct and make sure all the fields are correct.
func (p *Payment) Validate() (bool, error) {
	// Validate the base transaction
	_, err := p.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	// Check if the field Amount is valid
	if ok, err := IsAmount(p.Amount, "Amount", true); !ok {
		return false, err
	}

	// Check if Destination is a valid xrpl address
	if !addresscodec.IsValidAddress(p.Destination.String()) {
		return false, ErrInvalidDestination
	}

	// Check if the field Paths is valid
	if p.Paths != nil {
		if ok, err := IsPaths(p.Paths); !ok {
			return false, err
		}
	}

	// Check if the field SendMax is valid
	if ok, err := IsAmount(p.SendMax, "SendMax", false); !ok {
		return false, err
	}

	// Check if the field DeliverMax is valid
	if ok, err := IsAmount(p.DeliverMax, "DeliverMax", false); !ok {
		return false, err
	}

	// Check if the field DeliverMin is valid
	if ok, err := IsAmount(p.DeliverMin, "DeliverMin", false); !ok {
		return false, err
	}

	// Check if the field CredentialIDs is valid
	if p.CredentialIDs != nil && !p.CredentialIDs.IsValid() {
		return false, ErrInvalidCredentialIDs
	}

	// Check partial payment fields
	if ok, err := checkPartialPayment(p); !ok {
		return false, err
	}

	return true, nil
}

func checkPartialPayment(tx *Payment) (bool, error) {
	if tx.DeliverMin == nil {
		return true, nil
	}

	if tx.Flags == 0 {
		return false, ErrPartialPaymentFlagRequired
	}

	if !types.IsFlagEnabled(tx.Flags, tfPartialPayment) {
		return false, ErrPartialPaymentFlagRequired
	}

	return true, nil

}
