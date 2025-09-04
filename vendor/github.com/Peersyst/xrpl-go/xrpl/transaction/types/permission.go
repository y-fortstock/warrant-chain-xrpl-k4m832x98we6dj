package types

import "errors"

var (
	// ErrInvalidPermissionValue is returned when PermissionValue is empty or undefined.
	ErrInvalidPermissionValue = errors.New("permission value cannot be empty or undefined")
)

// Permission represents a transaction permission that can be delegated to another account.
// This matches the xrpl.js Permission interface structure.
type Permission struct {
	Permission PermissionValue `json:"Permission"`
}

// PermissionValue represents the inner permission value structure.
type PermissionValue struct {
	PermissionValue string `json:"PermissionValue"`
}

// Flatten returns the flattened map representation of the Permission.
func (p *Permission) Flatten() map[string]interface{} {
	flattened := make(map[string]interface{})
	flattened["Permission"] = p.Permission.Flatten()
	return flattened
}

// Validate validates the Permission structure.
func (p *Permission) IsValid() bool {
	return p.Permission.IsValid()
}

// Flatten returns the flattened map representation of the PermissionValue.
func (pv *PermissionValue) Flatten() map[string]interface{} {
	flattened := make(map[string]interface{})
	flattened["PermissionValue"] = pv.PermissionValue
	return flattened
}

// IsValid validates the PermissionValue structure.
func (pv *PermissionValue) IsValid() bool {
	return pv.PermissionValue != ""
}
