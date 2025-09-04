package types

type NFTokenID Hash256

// String returns the string representation of a NFTokenID.
func (n *NFTokenID) String() string {
	return string(*n)
}
