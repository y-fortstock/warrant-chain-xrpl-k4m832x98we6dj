package transaction

import (
	"errors"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

var (
	// ErrEmptyNFTokenOffers is returned when the NFTokenOffers array contains no entries.
	ErrEmptyNFTokenOffers = errors.New("the NFTokenOffers array must have at least one entry")
)

// The NFTokenCancelOffer transaction can be used to cancel existing token offers created using NFTokenCreateOffer.
//
// Example:
//
// ```json
//
//	{
//		"TransactionType": "NFTokenCancelOffer",
//		"Account": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX",
//		"NFTokenOffers": [
//			"9C92E061381C1EF37A8CDE0E8FC35188BFC30B1883825042A64309AC09F4C36D"
//		]
//	}
//
// ```
type NFTokenCancelOffer struct {
	BaseTx
	// An array of IDs of the NFTokenOffer objects to cancel (not the IDs of NFToken objects, but the IDs of the NFTokenOffer objects).
	// Each entry must be a different object ID of an NFTokenOffer object; the transaction is invalid if the array contains duplicate entries.
	NFTokenOffers []types.NFTokenID
}

// TxType returns the type of the transaction (NFTokenCancelOffer).
func (*NFTokenCancelOffer) TxType() TxType {
	return NFTokenCancelOfferTx
}

// Flatten returns a map of the NFTokenCancelOffer transaction fields.
func (n *NFTokenCancelOffer) Flatten() FlatTransaction {
	flattened := n.BaseTx.Flatten()

	flattened["TransactionType"] = "NFTokenCancelOffer"

	if len(n.NFTokenOffers) > 0 {
		flattenedOffers := make([]string, len(n.NFTokenOffers))
		for i, offer := range n.NFTokenOffers {
			flattenedOffers[i] = offer.String()
		}
		flattened["NFTokenOffers"] = flattenedOffers
	}

	return flattened
}

// Validate checks the validity of the NFTokenCancelOffer fields.
func (n *NFTokenCancelOffer) Validate() (bool, error) {
	ok, err := n.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	if len(n.NFTokenOffers) == 0 || n.NFTokenOffers == nil {
		return false, ErrEmptyNFTokenOffers
	}

	return true, nil
}
