package types

type Credential struct {
	// The issuer of the credential.
	Issuer Address
	// A hex-encoded value to identify the type of credential from the issuer.
	CredentialType CredentialType
}

func (c Credential) Flatten() map[string]interface{} {
	m := make(map[string]interface{})
	m["Issuer"] = c.Issuer.String()
	m["CredentialType"] = c.CredentialType.String()
	return m
}
