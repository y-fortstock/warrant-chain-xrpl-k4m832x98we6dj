package types

// (Optional) An arbitrary 256-bit value. If specified, the value is stored as part of the account but has no inherent meaning or requirements.
func WalletLocator(value Hash256) *Hash256 {
	return &value
}
