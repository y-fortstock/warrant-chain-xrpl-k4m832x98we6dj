package types

// BatchAccount is a type that represents a batch account.
// It is used to sign a batch transaction.
type BatchAccount struct {
	value string
}

// NewBatchAccount creates a new batch account.
func NewBatchAccount(value string) *BatchAccount {
	return &BatchAccount{
		value: value,
	}
}

// String returns the string representation of the batch account.
func (b *BatchAccount) String() string {
	return b.value
}
