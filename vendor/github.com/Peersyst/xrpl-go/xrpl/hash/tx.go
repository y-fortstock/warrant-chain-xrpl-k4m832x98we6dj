package hash

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

var (
	// ErrMissingSignature is returned when a transaction lacks the required signature fields.
	// A transaction must have at least one of: TxnSignature, Signers, or SigningPubKey,
	// unless it's an inner batch transaction (has TfInnerBatchTxn flag set).
	ErrMissingSignature = errors.New("transaction must have at least one of TxnSignature, Signers, or SigningPubKey")
)

// SignTxBlob hashes a signed transaction blob
// It takes a transaction blob and returns the hash of the signed transaction.
// It returns an error if the transaction blob is invalid.
func SignTxBlob(txBlob string) (string, error) {
	tx, err := binarycodec.Decode(txBlob)
	if err != nil {
		return "", err
	}

	if valid, err := isTxValid(tx); !valid {
		return "", err
	}

	return encodeSignedTxBlob(txBlob)
}

// SignTx hashes a signed transaction
// It takes a signed transaction and returns the hash of the signed transaction.
// It returns an error if the transaction is invalid.
func SignTx(tx map[string]interface{}) (string, error) {
	if valid, err := isTxValid(tx); !valid {
		return "", err
	}

	txBlob, err := binarycodec.Encode(tx)
	if err != nil {
		return "", err
	}

	return encodeSignedTxBlob(txBlob)
}

func encodeSignedTxBlob(txBlob string) (string, error) {
	// Create a byte slice with the correct capacity
	payload := make([]byte, 4+len(txBlob)/2)

	// Convert TRANSACTION_PREFIX to big-endian bytes
	binary.BigEndian.PutUint32(payload[:4], TransactionPrefix)

	// Decode the txBlob into the rest of the payload
	_, err := hex.Decode(payload[4:], []byte(txBlob))
	if err != nil {
		return "", err
	}

	return strings.ToUpper(hex.EncodeToString(crypto.Sha512Half(payload))), nil
}

func isTxValid(tx map[string]interface{}) (bool, error) {
	isInnerBatchTxn := false
	if flags, ok := tx["Flags"].(uint32); ok {
		isInnerBatchTxn = (flags & types.TfInnerBatchTxn) != 0
	}

	hasTxnSignature := tx["TxnSignature"] != nil
	hasSigners := tx["Signers"] != nil
	hasSigningPubKey := tx["SigningPubKey"] != nil

	if !hasTxnSignature && !hasSigners && !hasSigningPubKey && !isInnerBatchTxn {
		return false, ErrNonSignedTransaction
	}

	return true, nil
}
