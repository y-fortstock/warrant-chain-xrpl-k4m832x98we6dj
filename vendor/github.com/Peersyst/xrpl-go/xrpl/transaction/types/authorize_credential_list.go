package types

import "errors"

var (
	ErrEmptyCredentials       = errors.New("credentials list cannot be empty")
	ErrInvalidCredentialCount = errors.New("accepted credentials list must contain at least one and no more than the maximum allowed number of items")
	ErrDuplicateCredentials   = errors.New("credentials list cannot contain duplicate elements")
)

type AuthorizeCredentialList []AuthorizeCredential

func (ac *AuthorizeCredentialList) Validate() error {
	if len(*ac) == 0 {
		return ErrEmptyCredentials
	}
	if len(*ac) > MaxAcceptedCredentials {
		return ErrInvalidCredentialCount
	}
	seen := make(map[string]bool)
	for _, cred := range *ac {
		key := cred.Credential.Issuer.String() + cred.Credential.CredentialType.String()
		if seen[key] {
			return ErrDuplicateCredentials
		}
		seen[key] = true

		if err := cred.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (ac *AuthorizeCredentialList) Flatten() []map[string]interface{} {
	acs := make([]map[string]interface{}, len(*ac))
	for i, c := range *ac {
		acs[i] = c.Flatten()
	}
	return acs
}
