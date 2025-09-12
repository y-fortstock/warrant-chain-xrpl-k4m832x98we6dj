package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
)

// MPToken represents a Multi-Purpose Token with associated metadata.
// It contains document hash and signature information for asset-backed tokens.
type WarrantMPToken struct {
	DocumentHash string
	Issuer       string
}

// NewMPToken creates and returns a new MPToken instance.
// It requires a document hash and signature for token creation.
func NewWarrantMPToken(docHash, issuer string) WarrantMPToken {
	return WarrantMPToken{
		DocumentHash: docHash,
		Issuer:       issuer,
	}
}

// CreateMetadata generates the metadata structure required for MPT creation.
// This includes token details, URLs, and additional information like document hash and signature.
//
// Returns the metadata structure or an error if creation fails.
func (m WarrantMPToken) CreateMetadata() (MPTokenMetadata, error) {
	addInfo, err := json.Marshal(map[string]string{
		"document_hash": m.DocumentHash,
	})
	if err != nil {
		return MPTokenMetadata{}, fmt.Errorf("failed to marshal additional info: %w", err)
	}

	return MPTokenMetadata{
		Ticker:        "FSWRNT",
		Name:          "FortStock Warrant",
		Desc:          "Digital representation of real-world asset-backed warrants",
		AssetClass:    "rwa",
		AssetSubclass: "commodity",
		IssuerName:    m.Issuer,
		Urls: []MPTokenMetadataUrl{
			{
				Url:   "https://fortstock.io",
				Type:  "website",
				Title: "Home",
			},
			{
				Url:   "https://fortstock.io/rulebook/",
				Type:  "document",
				Title: "Legal framework",
			},
		},
		AdditionalInfo: addInfo,
	}, nil
}

// Currency: RLUSD
// Loan amount: 1,000,000 RLUSD
// Term: 3 days
// Annual rate: 36.5%
// Servicing cadence: daily payment 0.1% (executed per schedule defined in Debt token metadata)

// loan_id,
// amount,
// currency,
// due_date
// collateral_token_id = Warrant Token ID
// borrower and lender addresses

type DebtMPToken struct {
	Currency          string
	Amount            uint64
	InterestRate      float64
	Period            time.Duration
	CollateralTokenID string
	OwnerAddress      string
	CreditorAddress   string
}

func NewDebtMPToken(collateralTokenID string, ownerAddress string, creditorAddress string) DebtMPToken {
	return DebtMPToken{
		Currency:          LoanCurrency,
		Amount:            uint64(LoanAmount),
		InterestRate:      float64(LoanInterestRate),
		Period:            LoanPeriod,
		CollateralTokenID: collateralTokenID,
		OwnerAddress:      ownerAddress,
		CreditorAddress:   creditorAddress,
	}
}

func (d DebtMPToken) CreateMetadata() (MPTokenMetadata, error) {
	addInfo, err := json.Marshal(map[string]string{
		"currency":             d.Currency,
		"notional":             strconv.FormatUint(d.Amount, 10),
		"apr_percent":          strconv.FormatFloat(d.InterestRate, 'f', -1, 64),
		"term_days":            strconv.FormatInt(int64(d.Period.Hours()/24), 10),
		"servicing":            "daily",
		"rate_percent_per_day": strconv.FormatFloat(d.InterestRate/365, 'f', -1, 64),
		"origination_ts":       time.Now().Format(time.RFC3339),
		"maturity_ts":          time.Now().Add(d.Period).Format(time.RFC3339),
		"borrower_account":     d.OwnerAddress,
		"lender_account":       d.CreditorAddress,
		"warrant_token_id":     d.CollateralTokenID,
		"warrant_ticker":       "FSWRNT",
	})
	if err != nil {
		return MPTokenMetadata{}, fmt.Errorf("failed to marshal additional info: %w", err)
	}

	return MPTokenMetadata{
		Ticker:        "FSDEBT",
		Name:          "FortStock Debt Token",
		Icon:          "https://cdn.fortstock.io/app/fortstock.png",
		AssetClass:    "rwa",
		AssetSubclass: "credit",
		IssuerName:    d.OwnerAddress,
		Urls: []MPTokenMetadataUrl{
			{
				Url:   "https://fortstock.io",
				Type:  "website",
				Title: "Home",
			},
			{
				Url:   "https://fortstock.io/rulebook/",
				Type:  "document",
				Title: "Rulebook",
			},
			{
				Url:   "https://fortstock.io/terms/<contract_id>.pdf",
				Type:  "document",
				Title: "Loan Agreement",
			},
		},
		AdditionalInfo: addInfo,
	}, nil
}

// CreateIssuanceID generates a unique issuance ID for the token.
// This ID combines the issuer's account ID with the transaction sequence number.
//
// Parameters:
// - issuer: The issuer's account address
// - sequence: The transaction sequence number
//
// Returns the issuance ID as a string, or an error if generation fails.
func CreateIssuanceID(issuer string, sequence uint32) (string, error) {
	_, accountID, err := addresscodec.DecodeClassicAddressToAccountID(issuer)
	if err != nil {
		return "", fmt.Errorf("failed to decode classic address to account id: %w", err)
	}
	accountIDHex := fmt.Sprintf("%X", accountID)
	return fmt.Sprintf("%08X%s", sequence, accountIDHex), nil
}
