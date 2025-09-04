package types

import (
	"encoding/json"
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
	Flatten() interface{}
}

func UnmarshalCurrencyAmount(data []byte) (CurrencyAmount, error) {
	if len(data) == 0 {
		return nil, nil
	}
	switch data[0] {
	case '{':
		var raw map[string]interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, err
		}

		if _, hasMPTID := raw["mpt_issuance_id"]; hasMPTID {
			var m MPTCurrencyAmount
			if err := json.Unmarshal(data, &m); err != nil {
				return nil, err
			}
			return m, nil
		}

		var i IssuedCurrencyAmount
		if err := json.Unmarshal(data, &i); err != nil {
			return nil, err
		}
		return i, nil
	default:
		var x XRPCurrencyAmount
		if err := json.Unmarshal(data, &x); err != nil {
			return nil, err
		}
		return x, nil
	}
}

type IssuedCurrencyAmount struct {
	Issuer   Address `json:"issuer,omitempty"`
	Currency string  `json:"currency,omitempty"`
	Value    string  `json:"value,omitempty"`
}

func (IssuedCurrencyAmount) Kind() CurrencyKind {
	return ISSUED
}

func (i IssuedCurrencyAmount) Flatten() interface{} {
	json := make(map[string]interface{})

	if i.Issuer != "" {
		json["issuer"] = i.Issuer.String()
	}

	if i.Currency != "" {
		json["currency"] = i.Currency
	}

	if i.Value != "" {
		json["value"] = i.Value
	}
	return json
}

// IsZero returns true if the IssuedCurrencyAmount is the zero value (empty object).
func (i IssuedCurrencyAmount) IsZero() bool {
	return i == IssuedCurrencyAmount{}
}

type XRPCurrencyAmount uint64

func (a XRPCurrencyAmount) Uint64() uint64 {
	return uint64(a)
}

func (a XRPCurrencyAmount) String() string {
	return strconv.FormatUint(uint64(a), 10)
}

func (XRPCurrencyAmount) Kind() CurrencyKind {
	return XRP
}

func (a XRPCurrencyAmount) Flatten() interface{} {
	return a.String()
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
	MPTIssuanceID string `json:"mpt_issuance_id"`
	Value         string `json:"value"`
}

func (MPTCurrencyAmount) Kind() CurrencyKind {
	return MPT
}

func (m MPTCurrencyAmount) Flatten() interface{} {
	json := make(map[string]interface{})
	if m.MPTIssuanceID != "" {
		json["mpt_issuance_id"] = m.MPTIssuanceID
	}
	if m.Value != "" {
		json["value"] = m.Value
	}
	return json
}

func (m MPTCurrencyAmount) IsValid() bool {
	return m.MPTIssuanceID != "" && m.Value != ""
}
