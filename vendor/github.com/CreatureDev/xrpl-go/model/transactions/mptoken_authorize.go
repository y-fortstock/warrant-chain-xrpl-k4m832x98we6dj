package transactions

import (
	"github.com/CreatureDev/xrpl-go/model/transactions/types"
)

type MPTokenAuthorize struct {
	BaseTx
	MPTokenIssuanceID types.Hash192 `json:",omitempty"`
	Holder            types.Address `json:",omitempty"`
	Flags             *types.Flag   `json:",omitempty"`
}

func (*MPTokenAuthorize) TxType() TxType {
	return MPTokenAuthorizeTx
}
