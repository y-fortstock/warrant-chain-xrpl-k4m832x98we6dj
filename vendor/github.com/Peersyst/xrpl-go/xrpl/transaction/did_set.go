package transaction

import "errors"

var (
	ErrDIDSetMustSetEitherDataOrDIDDocumentOrURI = errors.New("did set: must set either Data, DIDDocument, or URI")
)

// (Requires the DID amendment)
// Creates a new DID ledger entry or updates the fields of an existing one.
//
// ```json
//
//	{
//	  "TransactionType": "DIDSet",
//	  "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
//	  "Fee": "10",
//	  "Sequence": 391,
//	  "URI": "697066733A2F2F62616679626569676479727A74357366703775646D37687537367568377932366E6634646675796C71616266336F636C67747179353566627A6469",
//	  "Data": "",
//	  "SigningPubKey":"0330E7FC9D56BB25D6893BA3F317AE5BCF33B3291BD63DB32654A313222F7FD020"
//	}
//
// ```
type DIDSet struct {
	BaseTx
	// The public attestations of identity credentials associated with the DID.
	Data string `json:",omitempty"`
	// The DID document associated with the DID.
	DIDDocument string `json:",omitempty"`
	// The URI associated with the DID.
	URI string `json:",omitempty"`
}

// TxType returns the type of the transaction.
func (tx *DIDSet) TxType() TxType {
	return DIDSetTx
}

// Flatten returns a flattened version of the transaction.
func (tx *DIDSet) Flatten() FlatTransaction {
	flattened := tx.BaseTx.Flatten()
	flattened["TransactionType"] = tx.TxType().String()

	if tx.Data != "" {
		flattened["Data"] = tx.Data
	}

	if tx.DIDDocument != "" {
		flattened["DIDDocument"] = tx.DIDDocument
	}

	if tx.URI != "" {
		flattened["URI"] = tx.URI
	}

	return flattened
}

// Validate validates the DIDSet struct.
func (tx *DIDSet) Validate() (bool, error) {

	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.Data == "" && tx.DIDDocument == "" && tx.URI == "" {
		return false, ErrDIDSetMustSetEitherDataOrDIDDocumentOrURI
	}

	return true, nil
}
