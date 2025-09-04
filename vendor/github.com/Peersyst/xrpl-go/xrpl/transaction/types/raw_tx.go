package types

import "errors"

const (
	// TfInnerBatchTxn flag that must be set on inner transactions within a batch
	TfInnerBatchTxn uint32 = 0x40000000
)

var (
	// ErrBatchRawTransactionMissing is returned when the RawTransaction field is missing from an array element.
	ErrBatchRawTransactionMissing = errors.New("batch RawTransaction field is missing")

	// ErrBatchRawTransactionFieldNotObject is returned when the RawTransaction field is not an object.
	ErrBatchRawTransactionFieldNotObject = errors.New("batch RawTransaction field is not an object")

	// ErrBatchNestedTransaction is returned when trying to include a Batch transaction within another Batch.
	ErrBatchNestedTransaction = errors.New("batch cannot contain nested Batch transactions")

	// ErrBatchMissingInnerFlag is returned when an inner transaction lacks the TfInnerBatchTxn flag.
	ErrBatchMissingInnerFlag = errors.New("batch RawTransaction must contain the TfInnerBatchTxn flag")

	// ErrBatchInnerTransactionInvalid is returned when an inner transaction fails its own validation.
	ErrBatchInnerTransactionInvalid = errors.New("batch inner transaction validation failed")
)

// RawTransactionWrapper represents the wrapper structure for transactions within a Batch.
type RawTransaction struct {
	RawTransaction map[string]any `json:"RawTransaction"`
}

// Flatten returns the flattened map representation of the RawTransaction.
func (r *RawTransaction) Flatten() map[string]any {
	return map[string]any{
		"RawTransaction": r.RawTransaction,
	}
}

// Validate validates the RawTransaction and its wrapped transaction.
func (r *RawTransaction) Validate() (bool, error) {
	// Validate RawTransaction field exists
	if r.RawTransaction == nil {
		return false, ErrBatchRawTransactionMissing
	}

	return validateRawTransaction(r.RawTransaction)
}

func validateRawTransaction(rawTx map[string]any) (bool, error) {
	// Check that TransactionType is not "Batch" (no nesting)
	if txType, ok := rawTx["TransactionType"].(string); ok && txType == "Batch" {
		return false, ErrBatchNestedTransaction
	}

	// Check for the TfInnerBatchTxn flag in the inner transactions
	if flags, ok := rawTx["Flags"].(uint32); !ok || !IsFlagEnabled(flags, TfInnerBatchTxn) {
		return false, ErrBatchMissingInnerFlag
	}

	// Fee must be "0" for inner transactions (or missing, which means 0)
	if feeField, exists := rawTx["Fee"]; exists {
		if feeStr, ok := feeField.(string); !ok || feeStr != "0" {
			return false, ErrBatchInnerTransactionInvalid
		}
	}

	// SigningPubKey must be empty for inner transactions (or missing, which means empty)
	if signingPubKeyField, exists := rawTx["SigningPubKey"]; exists {
		if signingPubKey, ok := signingPubKeyField.(string); !ok || signingPubKey != "" {
			return false, ErrBatchInnerTransactionInvalid
		}
	}

	// Check for disallowed fields in inner transactions
	if _, exists := rawTx["LastLedgerSequence"]; exists {
		return false, ErrBatchInnerTransactionInvalid
	}
	if _, exists := rawTx["Signers"]; exists {
		return false, ErrBatchInnerTransactionInvalid
	}
	if _, exists := rawTx["TxnSignature"]; exists {
		return false, ErrBatchInnerTransactionInvalid
	}

	return true, nil
}
