package transaction

import "fmt"

const (
	// Minimum TicketCount value.
	// https://xrpl.org/docs/references/protocol/transactions/types/ticketcreate#ticketcreate-fields
	MinTicketCount = 1
	// Maximum TicketCount value.
	// https://xrpl.org/docs/references/protocol/transactions/types/ticketcreate#ticketcreate-fields
	MaxTicketCount = 250
)

var (
	ErrTicketCreateInvalidTicketCount = fmt.Errorf("ticket count must be between %d and %d", MinTicketCount, MaxTicketCount)
)

// A TicketCreate transaction sets aside one or more sequence numbers as Tickets.
//
// Example:
//
// ```json
//
//	{
//	    "TransactionType": "TicketCreate",
//	    "Account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
//	    "Fee": "10",
//	    "Sequence": 381,
//	    "TicketCount": 10
//	}
//
// ```
type TicketCreate struct {
	// Base transaction fields
	BaseTx
	// How many Tickets to create. This must be a positive number and cannot cause
	// the account to own more than 250 Tickets after executing this transaction.
	TicketCount uint32
}

// TxType returns the type of the transaction (TicketCreate).
func (*TicketCreate) TxType() TxType {
	return TicketCreateTx
}

// Flatten returns the flattened map of the AMMVote transaction.
func (t *TicketCreate) Flatten() FlatTransaction {
	flattened := t.BaseTx.Flatten()

	flattened["TransactionType"] = "TicketCreate"

	if t.TicketCount != 0 {
		flattened["TicketCount"] = t.TicketCount
	}

	return flattened
}

// Validates the TicketCreate transaction and makes sure all the fields are correct.
func (t *TicketCreate) Validate() (bool, error) {
	_, err := t.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	if t.TicketCount < MinTicketCount || t.TicketCount > MaxTicketCount {
		return false, ErrTicketCreateInvalidTicketCount
	}

	return true, nil
}
