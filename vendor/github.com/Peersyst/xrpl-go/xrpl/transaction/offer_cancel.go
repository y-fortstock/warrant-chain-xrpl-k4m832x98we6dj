package transaction

import "errors"

var (
	ErrOfferCancelMissingOfferSequence = errors.New("missing offer sequence")
)

// An OfferCancel transaction removes an Offer object from the XRP Ledger.
//
// Example:
//
// ```json
//
//	{
//	    "TransactionType": "OfferCancel",
//	    "Account": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX",
//	    "Fee": "12",
//	    "Flags": 0,
//	    "LastLedgerSequence": 7108629,
//	    "OfferSequence": 6,
//	    "Sequence": 7
//	}
//
// ```
type OfferCancel struct {
	BaseTx
	// The sequence number (or Ticket number) of a previous OfferCreate transaction.
	// If specified, cancel any offer object in the ledger that was created by that transaction. It is not considered an error if the offer specified does not exist.
	OfferSequence uint32
}

func (*OfferCancel) TxType() TxType {
	return OfferCancelTx
}

// Flatten returns the flattened map of the OfferCancel transaction.
func (o *OfferCancel) Flatten() FlatTransaction {
	flattened := o.BaseTx.Flatten()

	flattened["TransactionType"] = o.TxType().String()

	flattened["OfferSequence"] = o.OfferSequence
	return flattened
}

// Validates the OfferCancel struct and makes sure all fields are correct.
func (o *OfferCancel) Validate() (bool, error) {
	return o.BaseTx.Validate()
}
