package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	MPTokenMetadataMaxSize = 1024
)

type MPTokenMetadataUrl struct {
	Url   string `json:"url,omitempty"`
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
}

type MPTokenMetadata struct {
	Ticker         string               `json:"ticker,omitempty"`
	Name           string               `json:"name,omitempty"`
	Desc           string               `json:"desc,omitempty"`
	Icon           string               `json:"icon,omitempty"`
	AssetClass     string               `json:"asset_class,omitempty"`
	AssetSubclass  string               `json:"asset_subclass,omitempty"`
	IssuerName     string               `json:"issuer_name,omitempty"`
	Urls           []MPTokenMetadataUrl `json:"urls,omitempty"`
	AdditionalInfo json.RawMessage      `json:"additional_info,omitempty"`
}

func NewMPTokenMetadataFromBlob(blob string) (*MPTokenMetadata, error) {
	b, err := hex.DecodeString(blob)
	if err != nil {
		return nil, fmt.Errorf("decode from blob in hex: %w", err)
	}
	m := MPTokenMetadata{}

	err = json.Unmarshal(b, &m)
	if err != nil {
		// https://github.com/XRPLF/XRPL-Standards/tree/master/XLS-0089d-multi-purpose-token-metadata-schema
		return nil, fmt.Errorf("metadata is not in XLS-0089d schema: %w", err)
	}
	return &m, nil
}

func (m MPTokenMetadata) GetBlob() (string, error) {
	json, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("marshal to json for blob: %w", err)
	}

	if len(json) > MPTokenMetadataMaxSize {
		return "", fmt.Errorf("blob is too large: %d", len(json))
	}

	h := hex.EncodeToString(json)
	return h, nil
}

func (m MPTokenMetadata) Validate() error {
	switch m.AssetClass {
	case "rwa", "memes", "wrapped", "gaming", "defi", "other":
		// ok
	default:
		return fmt.Errorf("invalid asset class: %s", m.AssetClass)
	}

	switch m.AssetSubclass {
	case "stablecoin", "commodity", "real_estate", "private_credit", "equity", "treasury", "other":
		// ok
	default:
		return fmt.Errorf("invalid asset subclass: %s", m.AssetSubclass)
	}

	return nil
}
