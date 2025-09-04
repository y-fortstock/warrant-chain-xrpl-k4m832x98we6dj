package account

import (
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	accounttypes "github.com/Peersyst/xrpl-go/xrpl/queries/account/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// The `account_info` command retrieves information about an account, its
// activity, and its XRP balance. All information retrieved is relative to a
// particular version of the ledger.
type InfoRequest struct {
	common.BaseRequest
	Account     types.Address          `json:"account"`
	LedgerIndex common.LedgerSpecifier `json:"ledger_index,omitempty"`
	LedgerHash  common.LedgerHash      `json:"ledger_hash,omitempty"`
	Queue       bool                   `json:"queue,omitempty"`
	SignerLists bool                   `json:"signer_lists,omitempty"`
	Strict      bool                   `json:"strict,omitempty"`
}

func (*InfoRequest) Method() string {
	return "account_info"
}

func (*InfoRequest) APIVersion() int {
	return version.RippledAPIV2
}

// TODO: Implement (V2)
func (*InfoRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// The expected response from the account_info method.
type InfoResponse struct {
	AccountData        ledger.AccountRoot     `json:"account_data"`
	SignerLists        []ledger.SignerList    `json:"signer_lists,omitempty"`
	LedgerCurrentIndex common.LedgerIndex     `json:"ledger_current_index,omitempty"`
	LedgerIndex        common.LedgerIndex     `json:"ledger_index,omitempty"`
	QueueData          accounttypes.QueueData `json:"queue_data,omitempty"`
	Validated          bool                   `json:"validated"`
}
