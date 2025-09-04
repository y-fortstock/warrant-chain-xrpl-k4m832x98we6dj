package types

// Sets an alternate account that is allowed to mint NFTokens on this
// account's behalf using NFTokenMint's `Issuer` field.
func NFTokenMinter(value string) *string {
	return &value
}
