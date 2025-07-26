package ledger

import (
	"encoding/json"

	"github.com/CreatureDev/xrpl-go/model/transactions/types"
)

type NFTokenOffer struct {
	Amount            types.CurrencyAmount `json:",omitempty"`
	Destination       types.Address        `json:",omitempty"`
	Expiration        uint                 `json:",omitempty"`
	Flags             *types.Flag          `json:",omitempty"`
	LedgerEntryType   LedgerEntryType      `json:",omitempty"`
	NFTokenID         types.Hash256        `json:",omitempty"`
	NFTokenOfferNode  string               `json:",omitempty"`
	Owner             types.Address        `json:",omitempty"`
	OwnerNode         string               `json:",omitempty"`
	PreviousTxnID     types.Hash256        `json:",omitempty"`
	PreviousTxnLgrSeq uint32               `json:",omitempty"`
	Index             types.Hash256        `json:"index,omitempty"`
}

func (*NFTokenOffer) EntryType() LedgerEntryType {
	return NFTokenOfferEntry

}

func (n *NFTokenOffer) UnmarshalJSON(data []byte) error {
	type nftHelper struct {
		Amount            json.RawMessage
		Destination       types.Address
		Expiration        uint
		Flags             *types.Flag
		LedgerEntryType   LedgerEntryType
		NFTokenID         types.Hash256
		NFTokenOfferNode  string
		Owner             types.Address
		OwnerNode         string
		PreviousTxnID     types.Hash256
		PreviousTxnLgrSeq uint32
		Index             types.Hash256 `json:"index,omitempty"`
	}
	var h nftHelper
	if err := json.Unmarshal(data, &h); err != nil {
		return err
	}
	*n = NFTokenOffer{
		Destination:       h.Destination,
		Expiration:        h.Expiration,
		Flags:             h.Flags,
		LedgerEntryType:   h.LedgerEntryType,
		NFTokenID:         h.NFTokenID,
		NFTokenOfferNode:  h.NFTokenOfferNode,
		Owner:             h.Owner,
		OwnerNode:         h.OwnerNode,
		PreviousTxnID:     h.PreviousTxnID,
		PreviousTxnLgrSeq: h.PreviousTxnLgrSeq,
		Index:             h.Index,
	}
	amnt, err := types.UnmarshalCurrencyAmount(h.Amount)
	if err != nil {
		return err
	}
	n.Amount = amnt
	return nil
}
