package transaction

// Deletes a price oracle for the Account.
// ```json
//
//	{
//	  "TransactionType": "OracleDelete",
//	  "Account": "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
//	  "OracleDocumentID": 34
//	}
//
// ```
type OracleDelete struct {
	BaseTx
	// A unique identifier of the price oracle for the Account. By default, it is 0.
	OracleDocumentID uint32
}

// Returns the type of the transaction.
func (tx *OracleDelete) TxType() TxType {
	return OracleDeleteTx
}

// Returns a flattened transaction.
func (tx *OracleDelete) Flatten() FlatTransaction {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	flattened["OracleDocumentID"] = tx.OracleDocumentID

	return flattened
}

// Validates the transaction.
func (tx *OracleDelete) Validate() (bool, error) {
	return tx.BaseTx.Validate()
}
