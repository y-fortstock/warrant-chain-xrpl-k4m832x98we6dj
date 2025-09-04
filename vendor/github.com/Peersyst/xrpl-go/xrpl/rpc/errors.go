package rpc

import "errors"

// Static errors
var (
	ErrIncorrectID                            = errors.New("incorrect id")
	ErrMissingTxSignatureOrSigningPubKey      = errors.New("transaction must have a TxSignature or SigningPubKey set")
	ErrSignerDataIsEmpty                      = errors.New("signer data is empty")
	ErrCannotFundWalletWithoutClassicAddress  = errors.New("cannot fund wallet without classic address")
	ErrMissingLastLedgerSequenceInTransaction = errors.New("missing LastLedgerSequence in transaction")
	ErrMissingWallet                          = errors.New("wallet must be provided when submitting an unsigned transaction")

	ErrRawTransactionsFieldIsNotAnArray = errors.New("RawTransactions field is not an array")
	ErrRawTransactionFieldIsNotAnObject = errors.New("RawTransaction field is not an object")

	ErrSigningPubKeyFieldMustBeEmpty = errors.New("SigningPubKey field must be empty")
	ErrTxnSignatureFieldMustBeEmpty  = errors.New("TxnSignature field must be empty")
	ErrSignersFieldMustBeEmpty       = errors.New("Signers field must be empty")
	ErrAccountFieldIsNotAString      = errors.New("Account field is not a string")
)

// Dynamic errors

type ClientError struct {
	ErrorString string
}

func (e *ClientError) Error() string {
	return e.ErrorString
}
