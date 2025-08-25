package types

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
)

type CurrencyKind int

const (
	XRP CurrencyKind = iota
	ISSUED
	MPT
)

type CurrencyAmount interface {
	Kind() CurrencyKind
	Validate() error
}

func UnmarshalCurrencyAmount(data []byte) (CurrencyAmount, error) {
	if len(data) == 0 {
		return nil, nil
	}

	// Try to parse as JSON object first to determine the type
	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err == nil {
		// It's a JSON object, determine which type based on fields
		if _, hasMPTIssuanceID := temp["mp_issuance_id"]; hasMPTIssuanceID {
			var mpt MPTCurrencyAmount
			if err := json.Unmarshal(data, &mpt); err != nil {
				return nil, err
			}
			return mpt, nil
		} else {
			// Object without mp_issuance_id, treat as IssuedCurrencyAmount
			var issued IssuedCurrencyAmount
			if err := json.Unmarshal(data, &issued); err != nil {
				return nil, err
			}
			return issued, nil
		}
	}

	// If not a JSON object, try to parse as XRPCurrencyAmount
	var x XRPCurrencyAmount
	if err := json.Unmarshal(data, &x); err != nil {
		return nil, err
	}
	return x, nil
}

type IssuedCurrencyAmount struct {
	Issuer   Address `json:"issuer,omitempty"`
	Currency string  `json:"currency,omitempty"`
	Value    string  `json:"value,omitempty"`
}

func (i IssuedCurrencyAmount) Validate() error {
	if i.Currency == "" {
		return fmt.Errorf("issued currency: missing currency code")
	}
	if i.Currency == "XRP" && i.Issuer != "" {
		return fmt.Errorf("issued currency: xrp cannot be issued")

	}
	/*
		// Issuer not required for source currencies field (path find request)
		if i.Currency != "XRP" && i.Issuer == "" {
			return fmt.Errorf("issued currency: non-xrp currencies require and issuer")
		}
	*/
	if err := i.Issuer.Validate(); i.Issuer != "" && err != nil {
		return fmt.Errorf("issued currency: %w", err)
	}
	return nil
}

func (IssuedCurrencyAmount) Kind() CurrencyKind {
	return ISSUED
}

type XRPCurrencyAmount uint64

func (XRPCurrencyAmount) Kind() CurrencyKind {
	return XRP
}

func (XRPCurrencyAmount) Validate() error {
	return nil
}

func XRPDropsFromFloat(f float32) XRPCurrencyAmount {
	d := f * 1000000
	return XRPCurrencyAmount(d)
}

func (a XRPCurrencyAmount) MarshalJSON() ([]byte, error) {
	s := strconv.FormatUint(uint64(a), 10)
	return json.Marshal(s)
}

func (a *XRPCurrencyAmount) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*a = XRPCurrencyAmount(v)
	return nil
}

func (a *XRPCurrencyAmount) UnmarshalText(data []byte) error {

	v, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return err
	}
	*a = XRPCurrencyAmount(v)
	return nil
}

type MPTCurrencyAmount struct {
	MPTIssuanceID Hash192 `json:"mp_issuance_id,omitempty"`
	Value         string  `json:"value,omitempty"`
}

func (MPTCurrencyAmount) Kind() CurrencyKind {
	return MPT
}

func (a MPTCurrencyAmount) Validate() error {
	if err := a.MPTIssuanceID.Validate(); err != nil {
		return fmt.Errorf("mp_issuance_id: %w", err)
	}
	return nil
}

func (a MPTCurrencyAmount) IssuerAccountID() ([]byte, error) {
	if err := a.Validate(); err != nil {
		return nil, err
	}

	issuanceID, err := hex.DecodeString(string(a.MPTIssuanceID))
	if err != nil {
		return nil, err
	}

	return issuanceID[4:], nil
}

func (a MPTCurrencyAmount) SequenceNumber() (uint32, error) {
	if err := a.Validate(); err != nil {
		return 0, err
	}

	issuanceID, err := hex.DecodeString(string(a.MPTIssuanceID))
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(issuanceID[:4]), nil
}
