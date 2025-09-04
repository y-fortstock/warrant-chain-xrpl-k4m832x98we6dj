package types

import (
	"errors"
)

// Maximum number of accepted credentials.
const MaxAcceptedCredentials int = 10

var (
	// Credential-specific errors

	ErrInvalidCredentialType   = errors.New("invalid credential type, must be a hexadecimal string between 1 and 64 bytes")
	ErrInvalidCredentialIssuer = errors.New("credential type: missing field Issuer")
)

// AuthorizeCredential represents an accepted credential for PermissionedDomainSet transactions.
type AuthorizeCredential struct {
	Credential Credential
}

// Validate checks if the AuthorizeCredential is valid.
func (a AuthorizeCredential) Validate() error {
	if a.Credential.Issuer.String() == "" {
		return ErrInvalidCredentialIssuer
	}
	if !a.Credential.CredentialType.IsValid() {
		return ErrInvalidCredentialType
	}
	return nil
}

// Flatten returns a flattened map representation of the AuthorizeCredential.
func (a AuthorizeCredential) Flatten() map[string]interface{} {
	m := make(map[string]interface{})
	m["Credential"] = a.Credential.Flatten()
	return m
}
