package types

import (
	"errors"
	"fmt"
	"slices"

	"github.com/Peersyst/xrpl-go/xrpl/hash"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

var (
	// ErrBatchSignableInvalid is returned when the batch signable is invalid.
	ErrBatchSignableInvalid = errors.New("batch signable is invalid")
	// ErrFlagsFieldIsNotAnUint32 is returned when the flags field is not an uint32.
	ErrFlagsFieldIsNotAnUint32 = errors.New("flags field is not an uint32")
	// ErrRawTransactionsFieldIsNotAnArray is returned when the raw transactions field is not an array.
	ErrRawTransactionsFieldIsNotAnArray = errors.New("raw transactions field is not an array")
	// ErrRawTransactionFieldIsNotAnObject is returned when the raw transaction field is not an object.
	ErrRawTransactionFieldIsNotAnObject = errors.New("raw transaction field is not an object")
)

// BatchSignable contains the fields needed to perform a Batch transactions signature.
// It contains the Flags of all transactions in the batch and the IDs of the transactions.
type BatchSignable struct {
	Flags uint32
	TxIDs []string
}

// FromFlatBatchTransaction creates a BatchSignable from a Batch transaction.
// It returns an error if the transaction is invalid.
func FromFlatBatchTransaction(transaction *transaction.FlatTransaction) (*BatchSignable, error) {
	flags, ok := (*transaction)["Flags"].(uint32)
	if !ok {
		return nil, ErrFlagsFieldIsNotAnUint32
	}

	rawTxs, ok := (*transaction)["RawTransactions"].([]map[string]any)
	if !ok {
		return nil, ErrRawTransactionsFieldIsNotAnArray
	}

	batchSignable := &BatchSignable{
		Flags: flags,
		TxIDs: make([]string, len(rawTxs)),
	}

	for i, rawTx := range rawTxs {
		innerRawTx, ok := rawTx["RawTransaction"].(map[string]any)
		if !ok {
			return nil, ErrRawTransactionFieldIsNotAnObject
		}
		txID, err := hash.SignTx(innerRawTx)
		if err != nil {
			return nil, fmt.Errorf("failed to get txID from raw transaction: %w", ErrBatchSignableInvalid)
		}
		batchSignable.TxIDs[i] = txID
	}

	return batchSignable, nil
}

// FromFlatBatchTransaction creates a BatchSignable from a Batch transaction.
// It returns an error if the transaction is invalid.
func FromBatchTransaction(transaction *transaction.Batch) (*BatchSignable, error) {
	rawTxs := transaction.RawTransactions

	batchSignable := &BatchSignable{
		Flags: transaction.Flags,
		TxIDs: make([]string, len(rawTxs)),
	}

	for i, rawTx := range rawTxs {
		txID, err := hash.SignTx(rawTx.RawTransaction)
		if err != nil {
			return nil, fmt.Errorf("failed to get txID from raw transaction: %w", ErrBatchSignableInvalid)
		}
		batchSignable.TxIDs[i] = txID
	}

	return batchSignable, nil
}

// Equals checks if the BatchSignable is equal to another BatchSignable.
// It returns true if the flags and txIDs are equal, false otherwise.
func (b *BatchSignable) Equals(other *BatchSignable) bool {
	return b.Flags == other.Flags && slices.Equal(b.TxIDs, other.TxIDs)
}

// Flatten returns the BatchSignable as a map[string]interface{} for encoding.
func (b *BatchSignable) Flatten() map[string]interface{} {
	flattened := make(map[string]interface{})

	flattened["flags"] = b.Flags

	if len(b.TxIDs) > 0 {
		flattened["txIDs"] = b.TxIDs
	}

	return flattened
}
