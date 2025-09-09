package api

import (
	"encoding/json"
	"fmt"
	"time"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
)

type MPToken interface {
	CreateMetadata() (MPTokenMetadata, error)
}

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
	InterestRate      uint64
	Period            time.Duration
	CollateralTokenID string
	OwnerAddress      string
	CreditorAddress   string
}

func NewDebtMPToken(collateralTokenID string, ownerAddress string, creditorAddress string) DebtMPToken {
	return DebtMPToken{
		Currency:          LoanCurrency,
		Amount:            LoanAmount,
		InterestRate:      LoanInterestRate,
		Period:            LoanPeriod,
		CollateralTokenID: collateralTokenID,
		OwnerAddress:      ownerAddress,
		CreditorAddress:   creditorAddress,
	}
}

func (d DebtMPToken) CreateMetadata() (MPTokenMetadata, error) {
	return MPTokenMetadata{}, nil
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
