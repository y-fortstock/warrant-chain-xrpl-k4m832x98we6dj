package ledger

import "github.com/CreatureDev/xrpl-go/model/transactions/types"

type MPToken struct {
	MPTAmount         types.MPTCurrencyAmount `json:",omitempty"`
	Flags             *types.Flag             `json:",omitempty"`
	MPTokenIssuanceID string                  `json:"mpt_issuance_id,omitempty"`
	LedgerEntryType   LedgerEntryType         `json:",omitempty"`
	LockedAmount      string                  `json:",omitempty"`
	OwnerNode         string                  `json:",omitempty"`
	PreviousTxnID     types.Hash256           `json:",omitempty"`
	PreviousTxnLgrSeq uint                    `json:",omitempty"`
	Index             types.Hash256           `json:"index,omitempty"`
}

func (*MPToken) EntryType() LedgerEntryType {
	return MPTokenEntry
}
