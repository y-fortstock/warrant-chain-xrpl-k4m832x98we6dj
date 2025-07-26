package ledger

import (
	"encoding/json"

	"github.com/CreatureDev/xrpl-go/model/transactions/types"
)

type OfferFlags uint32

func (f OfferFlags) ToUint() uint32 {
	return uint32(f)
}

const (
	PassiveOffer OfferFlags = 0x00010000
	SellOffer    OfferFlags = 0x00020000
)

type Offer struct {
	Account           types.Address        `json:",omitempty"`
	BookDirectory     types.Hash256        `json:",omitempty"`
	BookNode          string               `json:",omitempty"`
	Expiration        uint                 `json:",omitempty"`
	Flags             *types.Flag          `json:",omitempty"`
	LedgerEntryType   LedgerEntryType      `json:",omitempty"`
	OwnerNode         string               `json:",omitempty"`
	PreviousTxnID     types.Hash256        `json:",omitempty"`
	PreviousTxnLgrSeq uint32               `json:",omitempty"`
	Sequence          uint32               `json:",omitempty"`
	TakerPays         types.CurrencyAmount `json:",omitempty"`
	TakerGets         types.CurrencyAmount `json:",omitempty"`
	Index             types.Hash256        `json:"index,omitempty"`
}

func (*Offer) EntryType() LedgerEntryType {
	return OfferEntry
}

func (o *Offer) UnmarshalJSON(data []byte) error {
	type offerHelper struct {
		Account           types.Address
		BookDirectory     types.Hash256
		BookNode          string
		Expiration        uint
		Flags             *types.Flag
		LedgerEntryType   LedgerEntryType
		OwnerNode         string
		PreviousTxnID     types.Hash256
		PreviousTxnLgrSeq uint32
		Sequence          uint32
		TakerPays         json.RawMessage
		TakerGets         json.RawMessage
		Index             types.Hash256 `json:"index,omitempty"`
	}
	var h offerHelper
	if err := json.Unmarshal(data, &h); err != nil {
		return err
	}
	*o = Offer{
		Account:           h.Account,
		BookDirectory:     h.BookDirectory,
		BookNode:          h.BookNode,
		Expiration:        h.Expiration,
		Flags:             h.Flags,
		LedgerEntryType:   h.LedgerEntryType,
		OwnerNode:         h.OwnerNode,
		PreviousTxnID:     h.PreviousTxnID,
		PreviousTxnLgrSeq: h.PreviousTxnLgrSeq,
		Sequence:          h.Sequence,
		Index:             h.Index,
	}
	pays, err := types.UnmarshalCurrencyAmount(h.TakerPays)
	if err != nil {
		return err
	}
	gets, err := types.UnmarshalCurrencyAmount(h.TakerGets)
	if err != nil {
		return err
	}
	o.TakerPays = pays
	o.TakerGets = gets
	return nil
}
