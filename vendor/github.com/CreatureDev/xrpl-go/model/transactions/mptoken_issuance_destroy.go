package transactions

import "github.com/CreatureDev/xrpl-go/model/transactions/types"

type MPTokenIssuanceDestroy struct {
	BaseTx
	MPTokenIssuanceID types.Hash192 `json:",omitempty"`
}

func (*MPTokenIssuanceDestroy) TxType() TxType {
	return MPTokenIssuanceDestroyTx
}
