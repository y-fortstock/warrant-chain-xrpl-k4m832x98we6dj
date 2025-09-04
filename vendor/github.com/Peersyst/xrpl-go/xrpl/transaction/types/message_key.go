package types

// Public key for sending encrypted messages to this account.
func MessageKey(value string) *string {
	return &value
}
