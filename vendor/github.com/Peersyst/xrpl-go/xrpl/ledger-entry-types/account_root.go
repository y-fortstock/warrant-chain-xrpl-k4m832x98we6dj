package ledger

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// Enable Clawback for this account. (Requires the Clawback amendment.)
	lsfAllowTrustLineClawback uint32 = 0x80000000
	// Enable rippling on this addresses's trust lines by default. Required for issuing addresses; discouraged for others.
	lsfDefaultRipple uint32 = 0x00800000
	// This account has DepositAuth enabled, meaning it can only receive funds from transactions it sends, and from preauthorized accounts. (Added by the DepositAuth amendment)
	lsfDepositAuth uint32 = 0x01000000
	// Disallows use of the master key to sign transactions for this account.
	lsfDisableMaster uint32 = 0x00100000
	// This account blocks incoming Checks. (Added by the DisallowIncoming amendment.)
	lsfDisallowIncomingCheck uint32 = 0x08000000
	// This account blocks incoming NFTokenOffers. (Added by the DisallowIncoming amendment.)
	lsfDisallowIncomingNFTokenOffer uint32 = 0x04000000
	// This account blocks incoming Payment Channels. (Added by the DisallowIncoming amendment.)
	lsfDisallowIncomingPayChan uint32 = 0x10000000
	// This account blocks incoming trust lines. (Added by the DisallowIncoming amendment.)
	lsfDisallowIncomingTrustline uint32 = 0x20000000
	// Client applications should not send XRP to this account. (Advisory; not enforced by the protocol.)
	lsfDisallowXRP uint32 = 0x00080000
	// All assets issued by this account are frozen.
	lsfGlobalFreeze uint32 = 0x00400000
	// This account cannot freeze trust lines connected to it. Once enabled, cannot be disabled.
	lsfNoFreeze uint32 = 0x00200000
	// This account has used its free SetRegularKey transaction.
	lsfPasswordSpent uint32 = 0x00010000
	// This account must individually approve other users for those users to hold this account's tokens.
	lsfRequireAuth uint32 = 0x00040000
	// Requires incoming payments to specify a Destination Tag.
	lsfRequireDestTag uint32 = 0x00020000
)

// An AccountRoot ledger entry type describes a single account, its settings, and XRP balance.
//
// ```json
//
//	{
//	    "Account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
//	    "AccountTxnID": "0D5FB50FA65C9FE1538FD7E398FFFE9D1908DFA4576D8D7A020040686F93C77D",
//	    "Balance": "148446663",
//	    "Domain": "6D64756F31332E636F6D",
//	    "EmailHash": "98B4375E1D753E5B91627516F6D70977",
//	    "Flags": 8388608,
//	    "LedgerEntryType": "AccountRoot",
//	    "MessageKey": "0000000000000000000000070000000300",
//	    "OwnerCount": 3,
//	    "PreviousTxnID": "0D5FB50FA65C9FE1538FD7E398FFFE9D1908DFA4576D8D7A020040686F93C77D",
//	    "PreviousTxnLgrSeq": 14091160,
//	    "Sequence": 336,
//	    "TransferRate": 1004999999,
//	    "index": "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8"
//	}
//
// ```
type AccountRoot struct {
	// The unique ID for this ledger entry. In JSON, this field is represented with different names depending on the
	// context and API method. (Note, even though this is specified as "optional" in the code, every ledger entry
	// should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// Set of bit-flags for this ledger entry.
	Flags uint32
	// The type of ledger entry. Valid ledger entry types include AccountRoot, Offer, RippleState, and others.
	LedgerEntryType EntryType
	// The identifying (classic) address of this account.
	Account types.Address
	// The identifying hash of the transaction most recently sent by this account.
	// This field must be enabled to use the AccountTxnID transaction field.
	// To enable it, send an AccountSet transaction with the asfAccountTxnID flag enabled.
	AccountTxnID types.Hash256 `json:",omitempty"`
	// (Added by the AMM amendment) The ledger entry ID of the corresponding AMM ledger entry.
	// Set during account creation; cannot be modified. If present, indicates that this is a
	// special AMM AccountRoot; always omitted on non-AMM accounts.
	AMMID types.Hash256 `json:",omitempty"`
	// The account's current XRP balance in drops, represented as a string.
	Balance types.XRPCurrencyAmount `json:",omitempty"`
	// How many total of this account's issued non-fungible tokens have been burned.
	// This number is always equal or less than MintedNFTokens.
	BurnedNFTokens uint32 `json:",omitempty"`
	// A domain associated with this account. In JSON, this is the hexadecimal for the ASCII representation of the domain.
	// Cannot be more than 256 bytes in length.
	Domain string `json:",omitempty"`
	// The md5 hash of an email address. Clients can use this to look up an avatar through services such as Gravatar.
	EmailHash types.Hash128 `json:",omitempty"`
	// The account's Sequence Number at the time it minted its first non-fungible-token.
	// (Added by the fixNFTokenRemint amendment)
	FirstNFTokenSequence uint32 `json:",omitempty"`
	// A public key that may be used to send encrypted messages to this account. In JSON, uses hexadecimal.
	// Must be exactly 33 bytes, with the first byte indicating the key type: 0x02 or 0x03 for secp256k1
	// keys, 0xED for Ed25519 keys.
	MessageKey string `json:",omitempty"`
	// How many total non-fungible tokens have been minted by and on behalf of this account.
	// (Added by the NonFungibleTokensV1_1 amendment)
	MintedNFTokens uint32 `json:",omitempty"`
	// Another account that can mint non-fungible tokens on behalf of this account.
	// (Added by the NonFungibleTokensV1_1 amendment)
	NFTokenMinter types.Address `json:",omitempty"`
	// The number of objects this account owns in the ledger, which contributes to its owner reserve.
	OwnerCount uint32
	// The identifying hash of the transaction that most recently modified this object.
	PreviousTxnID types.Hash256
	// The index of the ledger that contains the transaction that most recently modified this object.
	PreviousTxnLgrSeq uint32
	// The address of a key pair that can be used to sign transactions for this account instead of the master key.
	// Use a SetRegularKey transaction to change this value.
	RegularKey types.Address `json:",omitempty"`
	// The sequence number of the next valid transaction for this account.
	Sequence uint32
	// How many Tickets this account owns in the ledger. This is updated automatically to ensure that the account
	// stays within the hard limit of 250 Tickets at a time. This field is omitted if the account has zero Tickets.
	// (Added by the TicketBatch amendment.)
	TicketCount uint32 `json:",omitempty"`
	// How many significant digits to use for exchange rates of Offers involving currencies issued by this address.
	// Valid values are 3 to 15, inclusive. (Added by the TickSize amendment.)
	TickSize uint8 `json:",omitempty"`
	// A transfer fee to charge other users for sending currency issued by this account to each other.
	TransferRate uint32 `json:",omitempty"`
	// An arbitrary 256-bit value that users can set.
	WalletLocator types.Hash256 `json:",omitempty"`
	// Unused. (The code supports this field but there is no way to set it.)
	WalletSize uint32 `json:",omitempty"`
}

// Returns the type of this ledger entry.
func (*AccountRoot) EntryType() EntryType {
	return AccountRootEntry
}

// Set the AllowTrustLineClawback flag.
func (a *AccountRoot) SetLsfAllowTrustLineClawback() {
	a.Flags |= lsfAllowTrustLineClawback
}

// Set the DefaultRipple flag.
func (a *AccountRoot) SetLsfDefaultRipple() {
	a.Flags |= lsfDefaultRipple
}

// Set the DepositAuth flag.
func (a *AccountRoot) SetLsfDepositAuth() {
	a.Flags |= lsfDepositAuth
}

// Set the DisableMaster flag.
func (a *AccountRoot) SetLsfDisableMaster() {
	a.Flags |= lsfDisableMaster
}

// Set the DisallowIncomingCheck flag.
func (a *AccountRoot) SetLsfDisallowIncomingCheck() {
	a.Flags |= lsfDisallowIncomingCheck
}

// Set the DisallowIncomingNFTokenOffer flag.
func (a *AccountRoot) SetLsfDisallowIncomingNFTokenOffer() {
	a.Flags |= lsfDisallowIncomingNFTokenOffer
}

// Set the DisallowIncomingPayChan flag.
func (a *AccountRoot) SetLsfDisallowIncomingPayChan() {
	a.Flags |= lsfDisallowIncomingPayChan
}

// Set the DisallowIncomingTrustline flag.
func (a *AccountRoot) SetLsfDisallowIncomingTrustline() {
	a.Flags |= lsfDisallowIncomingTrustline
}

// Set the DisallowXRP flag.
func (a *AccountRoot) SetLsfDisallowXRP() {
	a.Flags |= lsfDisallowXRP
}

// Set the GlobalFreeze flag.
func (a *AccountRoot) SetLsfGlobalFreeze() {
	a.Flags |= lsfGlobalFreeze
}

// Set the NoFreeze flag.
func (a *AccountRoot) SetLsfNoFreeze() {
	a.Flags |= lsfNoFreeze
}

// Set the PasswordSpent flag.
func (a *AccountRoot) SetLsfPasswordSpent() {
	a.Flags |= lsfPasswordSpent
}

// Set the RequireAuth flag.
func (a *AccountRoot) SetLsfRequireAuth() {
	a.Flags |= lsfRequireAuth
}

// Set the RequireDestTag flag.
func (a *AccountRoot) SetLsfRequireDestTag() {
	a.Flags |= lsfRequireDestTag
}
