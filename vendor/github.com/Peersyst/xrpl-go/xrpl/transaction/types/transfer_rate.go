package types

// (Optional) Sets the TransferRate. The fee to charge when users transfer this account's tokens, represented as billionths of a unit.
// Cannot be more than 2000000000 or less than 1000000000, except for the special case 0 meaning no fee.
func TransferRate(value uint32) *uint32 {
	return &value
}
