package types

type NFTokenURI string

// String returns the string representation of a NFTokenURI.
func (n *NFTokenURI) String() string {
	return string(*n)
}
