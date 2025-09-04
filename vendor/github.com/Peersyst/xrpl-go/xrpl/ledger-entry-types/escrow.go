package ledger

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// (Added by the Escrow amendment.)
// An Escrow ledger entry represents an escrow, which holds XRP until specific conditions are met.
//
// ```json
//
//	{
//	    "Account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
//	    "Amount": "10000",
//	    "CancelAfter": 545440232,
//	    "Condition": "A0258020A82A88B2DF843A54F58772E4A3861866ECDB4157645DD9AE528C1D3AEEDABAB6810120",
//	    "Destination": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX",
//	    "DestinationTag": 23480,
//	    "FinishAfter": 545354132,
//	    "Flags": 0,
//	    "LedgerEntryType": "Escrow",
//	    "OwnerNode": "0000000000000000",
//	    "DestinationNode": "0000000000000000",
//	    "PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
//	    "PreviousTxnLgrSeq": 28991004,
//	    "SourceTag": 11747,
//	    "index": "DC5F3851D8A1AB622F957761E5963BC5BD439D5C24AC6AD7AC4523F0640244AC"
//	}
//
// ```
type Escrow struct {
	// The unique ID for this ledger entry. In JSON, this field is represented with different names depending on the
	// context and API method. (Note, even though this is specified as "optional" in the code, every ledger entry
	// should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// The type of ledger entry.
	LedgerEntryType EntryType
	// A set of boolean flags.
	Flags uint32
	// The address of the owner (sender) of this escrow. This is the account that provided the XRP,
	// and gets it back if the escrow is canceled.
	Account types.Address
	// The amount of XRP, in drops, currently held in the escrow.
	Amount types.XRPCurrencyAmount
	// The escrow can be canceled if and only if this field is present and the time it specifies has passed.
	// Specifically, this is specified as seconds since the Ripple Epoch and it "has passed" if it's
	// earlier than the close time of the previous validated ledger.
	CancelAfter uint32 `json:",omitempty"`
	// A PREIMAGE-SHA-256 crypto-condition, as hexadecimal. If present, the EscrowFinish transaction must
	// contain a fulfillment that satisfies this condition.
	Condition string `json:",omitempty"`
	// The destination address where the XRP is paid if the escrow is successful.
	Destination types.Address
	// A hint indicating which page of the destination's owner directory links to this object, in case the
	// directory consists of multiple pages. Omitted on escrows created before enabling the fix1523 amendment.
	DestinationNode string `json:",omitempty"`
	// An arbitrary tag to further specify the destination for this escrow, such as a hosted recipient
	// at the destination address.
	DestinationTag uint32 `json:",omitempty"`
	// The time, in seconds since the Ripple Epoch, after which this escrow can be finished. Any EscrowFinish
	// transaction before this time fails. (Specifically, this is compared with the close time of the previous validated ledger.)
	FinishAfter uint32 `json:",omitempty"`
	// A hint indicating which page of the sender's owner directory links to this entry, in case the directory consists of multiple pages.
	OwnerNode string
	// The identifying hash of the transaction that most recently modified this entry.
	PreviousTxnID types.Hash256
	// The index of the ledger that contains the transaction that most recently modified this entry.
	PreviousTxnLgrSeq uint32
	// An arbitrary tag to further specify the source for this escrow, such as a hosted recipient at the owner's address.
	SourceTag uint32 `json:",omitempty"`
}

// Returns the type of the ledger entry.
func (*Escrow) EntryType() EntryType {
	return EscrowEntry
}
