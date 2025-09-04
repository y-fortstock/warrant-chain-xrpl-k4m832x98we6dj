package transaction

import (
	"errors"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// Batch transaction flags
	tfAllOrNothing uint32 = 0x00010000
	tfOnlyOne      uint32 = 0x00020000
	tfUntilFailure uint32 = 0x00040000
	tfIndependent  uint32 = 0x00080000
)

var (
	// General batch validation errors

	// ErrBatchRawTransactionsEmpty is returned when the RawTransactions array is empty or nil.
	// This validates that a batch transaction contains at least one inner transaction to execute.
	ErrBatchRawTransactionsEmpty = errors.New("RawTransactions must be a non-empty array")

	// ErrBatchSignersNotArray is returned when BatchSigners field is present but not an array type.
	// BatchSigners must be an array of signer objects for multi-signing validation.
	ErrBatchSignersNotArray = errors.New("BatchSigners must be an array")

	// RawTransactions validation errors

	// ErrBatchRawTransactionNotObject is returned when a RawTransaction array element is not an object.
	// Each element in the RawTransactions array must be a valid transaction object.
	ErrBatchRawTransactionNotObject = errors.New("batch RawTransaction element is not an object")

	// ErrBatchRawTransactionMissing is returned when the RawTransaction field is missing from an array element.
	// Each RawTransactions array element must contain a RawTransaction field.
	ErrBatchRawTransactionMissing = errors.New("batch RawTransaction field is missing")

	// ErrBatchRawTransactionFieldNotObject is returned when the RawTransaction field is not an object.
	// The RawTransaction field must contain a valid transaction object structure.
	ErrBatchRawTransactionFieldNotObject = errors.New("batch RawTransaction field is not an object")

	// ErrBatchNestedTransaction is returned when trying to include a Batch transaction within another Batch.
	// Nested batch transactions are not allowed to prevent infinite recursion and complexity.
	ErrBatchNestedTransaction = errors.New("batch cannot contain nested Batch transactions")

	// ErrBatchMissingInnerFlag is returned when an inner transaction lacks the TfInnerBatchTxn flag.
	// All transactions within a batch must have the TfInnerBatchTxn flag set to indicate they are inner transactions.
	ErrBatchMissingInnerFlag = errors.New("batch RawTransaction must contain the TfInnerBatchTxn flag")

	// Inner transaction validation errors

	// ErrBatchInnerTransactionInvalid is returned when an inner transaction fails its own validation.
	// Each inner transaction must pass its individual validation rules.
	ErrBatchInnerTransactionInvalid = errors.New("batch inner transaction validation failed")

	// BatchSigners validation errors

	// ErrBatchSignerNotObject is returned when a BatchSigner array element is not an object.
	// Each element in the BatchSigners array must be a valid signer object.
	ErrBatchSignerNotObject = errors.New("batch BatchSigner element is not an object")

	// ErrBatchSignerMissing is returned when the BatchSigner field is missing from an array element.
	// Each BatchSigners array element must contain a BatchSigner field.
	ErrBatchSignerMissing = errors.New("batch BatchSigner field is missing")

	// ErrBatchSignerFieldNotObject is returned when the BatchSigner field is not an object.
	// The BatchSigner field must contain a valid signer object structure.
	ErrBatchSignerFieldNotObject = errors.New("batch BatchSigner field is not an object")

	// ErrBatchSignerAccountMissing is returned when a BatchSigner lacks the required Account field.
	// Each BatchSigner must specify an Account for the signing operation.
	ErrBatchSignerAccountMissing = errors.New("batch BatchSigner Account is missing")

	// ErrBatchSignerAccountNotString is returned when a BatchSigner Account field is not a string.
	// The Account field must be a valid string representing an XRPL account address.
	ErrBatchSignerAccountNotString = errors.New("batch BatchSigner Account must be a string")

	// ErrBatchSignerInvalid is returned when a BatchSigner fails its validation rules.
	// Each BatchSigner must pass its individual validation requirements.
	ErrBatchSignerInvalid = errors.New("batch signer validation failed")
)

// Batch represents a Batch transaction that can execute multiple transactions atomically.
//
// Example:
//
// ```json
//
//	{
//	    "TransactionType": "Batch",
//	    "Account": "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
//	    "Fee": "100",
//	    "Flags": 65536,
//	    "Sequence": 1,
//	    "RawTransactions": [
//	        {
//	            "RawTransaction": {
//	                "TransactionType": "Payment",
//	                "Account": "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
//	                "Amount": "1000000",
//	                "Destination": "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
//	                "Flags": 1073741824,
//	                "Fee": "0",
//	                "SigningPubKey": ""
//	            }
//	        }
//	    ]
//	}
//
// ```
type Batch struct {
	BaseTx
	// Array of transactions to be executed as part of this batch.
	RawTransactions []types.RawTransaction `json:"RawTransactions"`
	// Optional array of batch signers for multi-signing the batch.
	BatchSigners []types.BatchSigner `json:"BatchSigners,omitempty"`
}

// TxType returns the type of the transaction (Batch).
func (*Batch) TxType() TxType {
	return BatchTx
}

// **********************************
// Batch Flags
// **********************************

// SetAllOrNothingFlag sets the AllOrNothing flag.
//
// AllOrNothing: Execute all transactions in the batch or none at all.
// If any transaction fails, the entire batch fails.
func (b *Batch) SetAllOrNothingFlag() {
	b.Flags |= tfAllOrNothing
}

// SetOnlyOneFlag sets the OnlyOne flag.
//
// OnlyOne: Execute only the first transaction that succeeds.
// Stop execution after the first successful transaction.
func (b *Batch) SetOnlyOneFlag() {
	b.Flags |= tfOnlyOne
}

// SetUntilFailureFlag sets the UntilFailure flag.
//
// UntilFailure: Execute transactions until one fails.
// Stop execution at the first failed transaction.
func (b *Batch) SetUntilFailureFlag() {
	b.Flags |= tfUntilFailure
}

// SetIndependentFlag sets the Independent flag.
//
// Independent: Execute all transactions independently.
// The failure of one transaction does not affect others.
func (b *Batch) SetIndependentFlag() {
	b.Flags |= tfIndependent
}

// Flatten returns the flattened map of the Batch transaction.
func (b *Batch) Flatten() FlatTransaction {
	flattenedTx := b.BaseTx.Flatten()

	flattenedTx["TransactionType"] = b.TxType().String()

	rawTxs := make([]map[string]any, len(b.RawTransactions))
	for i, rtw := range b.RawTransactions {
		rawTxs[i] = rtw.Flatten()
	}
	flattenedTx["RawTransactions"] = rawTxs

	if len(b.BatchSigners) > 0 {
		signers := make([]map[string]any, len(b.BatchSigners))
		for i, bs := range b.BatchSigners {
			signers[i] = bs.Flatten()
		}
		flattenedTx["BatchSigners"] = signers
	}

	return flattenedTx
}

// Validate validates the Batch transaction.
func (b *Batch) Validate() (bool, error) {
	_, err := b.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	if len(b.RawTransactions) == 0 {
		return false, ErrBatchRawTransactionsEmpty
	}

	// Validate each RawTransaction
	for _, rawTx := range b.RawTransactions {
		if valid, err := rawTx.Validate(); !valid {
			return false, err
		}
	}

	for _, batchSigner := range b.BatchSigners {
		if err := batchSigner.Validate(); err != nil {
			return false, err
		}
	}

	return true, nil
}
