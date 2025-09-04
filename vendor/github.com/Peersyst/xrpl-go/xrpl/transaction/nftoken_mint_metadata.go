package transaction

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

type NFTokenMintMetadata struct {
	TxObjMeta
	// rippled 1.11.0 or later
	NFTokenID *types.NFTokenID `json:"nftoken_id,omitempty"`
	// if Amount is present
	OfferID *types.Hash256 `json:"offer_id,omitempty"`
}

func (NFTokenMintMetadata) TxMeta() {}
