package transaction

import (
	"errors"
)

var (
	// Credential-specific errors
	ErrMissingDomainID = errors.New("missing required field: DomainID")
)

// Delete a permissioned domain that you own.
// (Requires the PermissionedDomains amendment)
//
// ```json
//
//	{
//	  "TransactionType": "PermissionedDomainDelete",
//	  "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
//	  "Fee": "10",
//	  "Sequence": 392,
//	  "DomainID": "77D6234D074E505024D39C04C3F262997B773719AB29ACFA83119E4210328776"
//	}
//
// ```
type PermissionedDomainDelete struct {
	BaseTx
	// The ledger entry ID of the Permissioned Domain entry to delete.
	DomainID string
}

// TxType returns the transaction type.
func (p *PermissionedDomainDelete) TxType() TxType {
	return PermissionedDomainDeleteTx
}

// Flatten returns a flattened map representation of the PermissionedDomainDelete transaction.
func (p *PermissionedDomainDelete) Flatten() FlatTransaction {
	flattened := p.BaseTx.Flatten()
	flattened["TransactionType"] = p.TxType().String()
	flattened["DomainID"] = p.DomainID
	return flattened
}

// Validate validates the PermissionedDomainDelete transaction.
// It ensures that the base transaction is valid and that the required DomainID field is present.
func (p *PermissionedDomainDelete) Validate() (bool, error) {
	// Validate common transaction fields.
	if ok, err := p.BaseTx.Validate(); !ok {
		return false, err
	}
	// Ensure DomainID is provided.
	if p.DomainID == "" {
		return false, ErrMissingDomainID
	}
	return true, nil
}
