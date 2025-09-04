package types

import (
	"errors"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
)

var (

	// ErrBatchSignerAccountMissing is returned when a BatchSigner lacks the required Account field.
	ErrBatchSignerAccountMissing = errors.New("batch BatchSigner Account is missing")

	// ErrBatchSignerSigningPubKeyMissing is returned when a BatchSigner lacks the required SigningPubKey field.
	ErrBatchSignerSigningPubKeyMissing = errors.New("batch BatchSigner SigningPubKey is missing")

	// ErrBatchSignerInvalidTxnSignature is returned when a BatchSigner has an invalid TxnSignature field.
	ErrBatchSignerInvalidTxnSignature = errors.New("batch BatchSigner TxnSignature is invalid")
)

// BatchSigner represents a single batch signer entry.
type BatchSigner struct {
	BatchSigner BatchSignerData `json:"BatchSigner"`
}

// BatchSignerData contains the actual batch signer information.
type BatchSignerData struct {
	Account       Address  `json:"Account"`
	SigningPubKey string   `json:"SigningPubKey,omitempty"`
	TxnSignature  string   `json:"TxnSignature,omitempty"`
	Signers       []Signer `json:"Signers,omitempty"`
}

// Flatten returns the flattened map of the BatchSigner.
func (bs *BatchSigner) Flatten() map[string]any {
	signer := map[string]any{
		"Account": bs.BatchSigner.Account.String(),
	}

	if bs.BatchSigner.SigningPubKey != "" {
		signer["SigningPubKey"] = bs.BatchSigner.SigningPubKey
	}
	if bs.BatchSigner.TxnSignature != "" {
		signer["TxnSignature"] = bs.BatchSigner.TxnSignature
	}
	if len(bs.BatchSigner.Signers) > 0 {
		innerSigners := make([]map[string]any, len(bs.BatchSigner.Signers))
		for i, innerSigner := range bs.BatchSigner.Signers {
			innerSigners[i] = innerSigner.Flatten()
		}
		signer["Signers"] = innerSigners
	}

	return map[string]any{
		"BatchSigner": signer,
	}
}

func (bs *BatchSigner) Validate() error {
	if !addresscodec.IsValidAddress(bs.BatchSigner.Account.String()) {
		return ErrBatchSignerAccountMissing
	}

	if bs.BatchSigner.SigningPubKey == "" {
		return ErrBatchSignerSigningPubKeyMissing
	}

	if bs.BatchSigner.TxnSignature != "" {
		return ErrBatchSignerInvalidTxnSignature
	}

	return nil
}
