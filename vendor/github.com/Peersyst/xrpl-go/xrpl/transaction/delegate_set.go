package transaction

import (
	"errors"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// DelegateSet-specific errors
var (
	// ErrDelegateSetAuthorizeAccountConflict is returned when the Authorize account matches the Account.
	ErrDelegateSetAuthorizeAccountConflict = errors.New("authorize account cannot be the same as the Account")
	// ErrDelegateSetPermissionMalformed is returned when the Permissions array is empty or malformed.
	ErrDelegateSetPermissionMalformed = errors.New("permissions array is required and cannot be empty")
	// ErrDelegateSetPermissionsMaxLength is returned when the Permissions array exceeds the maximum length.
	ErrDelegateSetPermissionsMaxLength = errors.New("permissions array cannot exceed maximum length")
	// ErrDelegateSetEmptyPermissionValue is returned when a permission value is empty or undefined.
	ErrDelegateSetEmptyPermissionValue = errors.New("permission value cannot be empty")
	// ErrDelegateSetNonDelegatableTransaction is returned when trying to delegate a non-delegatable transaction type.
	ErrDelegateSetNonDelegatableTransaction = errors.New("cannot delegate non-delegatable transaction types")
	// ErrDelegateSetDuplicatePermissions is returned when the same permission is specified multiple times.
	ErrDelegateSetDuplicatePermissions = errors.New("duplicate permissions are not allowed")
)

const (
	// Maximum number of permissions that can be delegated in a single transaction
	PermissionsMaxLength = 10
)

// Set of transaction types that cannot be delegated
var NonDelegatableTransactionsMap = map[string]uint8{
	string(AccountSetTx):    0,
	string(SetRegularKeyTx): 0,
	string(SignerListSetTx): 0,
	string(DelegateSetTx):   0,
	string(AccountDeleteTx): 0,
	// Pseudo transactions below:
	"EnableAmendment": 0,
	"SetFee":          0,
	"UNLModify":       0,
}

// DelegateSet allows an account to delegate a set of permissions to another account.
//
// Example:
//
// ```json
//
//	{
//	    "Account": "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
//	    "TransactionType": "DelegateSet",
//	    "Authorize": "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
//	    "Permissions": [
//	        {
//	            "Permission": {
//	                "PermissionValue": "Payment"
//	            }
//	        },
//	        {
//	            "Permission": {
//	                "PermissionValue": "TrustlineAuthorize"
//	            }
//	        }
//	    ],
//	    "Fee": "12",
//	    "Sequence": 1
//	}
//
// ```
type DelegateSet struct {
	BaseTx
	// The authorized account.
	Authorize types.Address `json:"Authorize"`
	// The transaction permissions that the account has been granted.
	Permissions []types.Permission `json:"Permissions"`
}

// TxType returns the type of the transaction (DelegateSet).
func (*DelegateSet) TxType() TxType {
	return DelegateSetTx
}

// Flatten returns the flattened map of the DelegateSet transaction.
func (d *DelegateSet) Flatten() FlatTransaction {
	flattened := d.BaseTx.Flatten()

	flattened["TransactionType"] = "DelegateSet"

	if d.Authorize != "" {
		flattened["Authorize"] = d.Authorize.String()
	}

	if len(d.Permissions) > 0 {
		flattenedPermissions := make([]interface{}, len(d.Permissions))
		for i, permission := range d.Permissions {
			flattenedPermissions[i] = permission.Flatten()
		}
		flattened["Permissions"] = flattenedPermissions
	}

	return flattened
}

// Validate validates the DelegateSet transaction and ensures all fields are correct.
func (d *DelegateSet) Validate() (bool, error) {
	_, err := d.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	if !addresscodec.IsValidAddress(d.Authorize.String()) {
		return false, ErrInvalidDestination
	}

	// Authorize and Account must be different
	if d.Authorize == d.Account {
		return false, ErrDelegateSetAuthorizeAccountConflict
	}

	// Validate Permissions array
	if len(d.Permissions) == 0 {
		return false, ErrDelegateSetPermissionMalformed // Permissions array is required
	}

	if len(d.Permissions) > PermissionsMaxLength {
		return false, ErrDelegateSetPermissionsMaxLength
	}

	// Track permission values to check for duplicates
	permissionValueSet := make(map[string]uint8)

	for _, permission := range d.Permissions {
		// Validate permission structure using the Permission's own validation
		if !permission.Permission.IsValid() {
			return false, ErrDelegateSetEmptyPermissionValue
		}

		permissionValue := permission.Permission.PermissionValue

		// Check if it's a non-delegatable transaction
		if _, isNonDelegatable := NonDelegatableTransactionsMap[permissionValue]; isNonDelegatable {
			return false, ErrDelegateSetNonDelegatableTransaction
		}

		// Add to set for duplicate detection
		permissionValueSet[permissionValue] = 0
	}

	// Check for duplicates by comparing lengths (like JS implementation)
	if len(d.Permissions) != len(permissionValueSet) {
		return false, ErrDelegateSetDuplicatePermissions
	}

	return true, nil
}
