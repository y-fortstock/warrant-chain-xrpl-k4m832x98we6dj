package ledger

import (
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	ledgertypes "github.com/Peersyst/xrpl-go/xrpl/queries/ledger/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
)

// ############################################################################
// Request
// ############################################################################

// Retrieve information about the public ledger.
type Request struct {
	common.BaseRequest
	// A 32-byte hex string for the ledger version to use. (See Specifying Ledgers).
	LedgerHash common.LedgerHash `json:"ledger_hash,omitempty"`
	// The ledger index of the ledger to use, or a shortcut string to choose a ledger automatically. (See Specifying Ledgers)
	LedgerIndex common.LedgerSpecifier `json:"ledger_index,omitempty"`
	Full        bool                   `json:"full,omitempty"`
	Accounts    bool                   `json:"accounts,omitempty"`
	// Provide full JSON-formatted information for transaction/account information instead of only hashes. The default is false. Ignored unless you request transactions, accounts, or both.
	Expand bool `json:"expand,omitempty"`
	// If true, return information on transactions in the specified ledger version. The default is false. Ignored if you did not specify a ledger version.
	Transactions bool `json:"transactions,omitempty"`
	// If true, include owner_funds field in the metadata of OfferCreate transactions in the response. The default is false. Ignored unless transactions are included and expand is true.
	OwnerFunds bool `json:"owner_funds,omitempty"`
	// If true, and transactions and expand are both also true, return transaction information in binary format (hexadecimal string) instead of JSON format. The default is false. Ignored unless transactions and expand are both true.
	Binary bool `json:"binary,omitempty"`
	// If true, and the command is requesting the current ledger, includes an array of queued transactions in the results. The default is false.
	Queue bool             `json:"queue,omitempty"`
	Type  ledger.EntryType `json:"type,omitempty"`
}

func (*Request) Method() string {
	return "ledger"
}

func (*Request) APIVersion() int {
	return version.RippledAPIV2
}

// TODO: Implement V2
func (*Request) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// The expected response from the ledger method.
type Response struct {
	Ledger      ledgertypes.BaseLedger  `json:"ledger"`
	LedgerHash  string                  `json:"ledger_hash"`
	LedgerIndex common.LedgerIndex      `json:"ledger_index"`
	Validated   bool                    `json:"validated,omitempty"`
	QueueData   []ledgertypes.QueueData `json:"queue_data,omitempty"`
}
