package types

import (
	"encoding/hex"
	"strings"

	"github.com/CreatureDev/xrpl-go/binary-codec/serdes"
	"github.com/CreatureDev/xrpl-go/model/transactions/types"
)

var _ hashI = (*Hash192)(nil)

// Hash192 struct represents a 192-bit hash.
type Hash192 struct {
}

// NewHash192 is a constructor for creating a new 192-bit hash.
func NewHash192() *Hash192 {
	return &Hash192{}
}

// getLength method for hash returns the hash's length.
func (h *Hash192) getLength() int {
	return 24
}

// FromJson method for hash converts a hexadecimal string from JSON to a byte array.
// It returns an error if the conversion fails or the length of the decoded byte array is not as expected.
func (h *Hash192) FromJson(json any) ([]byte, error) {
	var s string
	switch json := json.(type) {
	case string:
		s = json
	case types.Hash192:
		s = string(json)
	default:
		return nil, ErrInvalidHashType
	}
	v, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if h.getLength() != len(v) {
		return nil, &ErrInvalidHashLength{Expected: h.getLength()}
	}
	return v, nil
}

// ToJson method for hash reads a certain number of bytes from a BinaryParser and converts it into a hexadecimal string.
// It returns an error if the read operation fails.
func (h *Hash192) ToJson(p *serdes.BinaryParser, opts ...int) (any, error) {
	b, err := p.ReadBytes(h.getLength())
	if err != nil {
		return nil, err
	}
	return strings.ToUpper(hex.EncodeToString(b)), nil
}
