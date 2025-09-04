package types

import "math/big"

// Perform bitwise AND (&) to check if a flag is enabled within Flags (as a number).
// @param Flags - A number that represents flags enabled.
// @param checkFlag - A specific flag to check if it's enabled within Flags.
// @returns True if checkFlag is enabled within Flags.
func IsFlagEnabled(flags, checkFlag uint32) bool {
	flagsBigInt := new(big.Int).SetUint64(uint64(flags))
	checkFlagBigInt := new(big.Int).SetUint64(uint64(checkFlag))
	return new(big.Int).And(flagsBigInt, checkFlagBigInt).Cmp(checkFlagBigInt) == 0
}
