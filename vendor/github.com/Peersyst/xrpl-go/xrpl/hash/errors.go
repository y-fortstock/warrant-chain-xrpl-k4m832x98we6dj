package hash

import "errors"

var (
	ErrNonSignedTransaction = errors.New("transaction must have at least one of TxnSignature, Signers, or SigningPubKey")
)
