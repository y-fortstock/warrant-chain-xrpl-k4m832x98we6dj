package transactions

import (
	"github.com/CreatureDev/xrpl-go/model/transactions/types"
)

type MPTokenIssuanceCreate struct {
	BaseTx
	AssetScale      uint8       `json:",omitempty"`
	MaximumAmount   string      `json:",omitempty"`
	TransferFee     uint16      `json:",omitempty"`
	MPTokenMetadata string      `json:",omitempty"`
	Flags           *types.Flag `json:",omitempty"`
}

func (*MPTokenIssuanceCreate) TxType() TxType {
	return MPTokenIssuanceCreateTx
}
