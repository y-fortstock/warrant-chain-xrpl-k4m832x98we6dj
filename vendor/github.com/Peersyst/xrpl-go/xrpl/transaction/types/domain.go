package types

// The domain that owns this account, as a string of hex representing the.
// ASCII for the domain in lowercase.
func Domain(value string) *string {
	return &value
}
