package faucet

import (
	"fmt"

	"github.com/CreatureDev/xrpl-go/model/transactions/types"
)

type FundAccountRequest struct {
	Destination  types.Address `json:"destination"`
	UsageContext string        `json:"usageContext"`
	UserAgent    string        `json:"userAgent"`
}

func (*FundAccountRequest) Method() string {
	return "fund_account"
}

func (f *FundAccountRequest) Validate() error {
	if err := f.Destination.Validate(); err != nil {
		return fmt.Errorf("faucet fund account: %w", err)
	}
	return nil
}
