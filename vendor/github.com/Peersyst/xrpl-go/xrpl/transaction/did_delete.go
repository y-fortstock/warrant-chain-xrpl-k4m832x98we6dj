package transaction

// ```json
//
//	{
//	    "TransactionType": "DIDDelete",
//	    "Account": "rp4pqYgrTAtdPHuZd1ZQWxrzx45jxYcZex",
//	    "Fee": "12",
//	    "Sequence": 391,
//	    "SigningPubKey":"0293A815C095DBA82FAC597A6BB9D338674DB93168156D84D18417AD509FFF5904",
//	    "TxnSignature":"3044022011E9A7EE3C7AE9D202848390522E6840F7F3ED098CD13E..."
//	}
//
// ```
type DIDDelete struct {
	BaseTx
}

// TxType returns the type of the transaction.
func (tx *DIDDelete) TxType() TxType {
	return DIDDeleteTx
}

// Flatten returns a flattened version of the transaction.
func (tx *DIDDelete) Flatten() FlatTransaction {
	flattened := tx.BaseTx.Flatten()
	flattened["TransactionType"] = tx.TxType().String()
	return flattened
}

// Validate validates the transaction.
func (tx *DIDDelete) Validate() (bool, error) {
	return tx.BaseTx.Validate()
}
