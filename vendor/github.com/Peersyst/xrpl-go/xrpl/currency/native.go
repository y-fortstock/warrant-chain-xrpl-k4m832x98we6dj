package currency

import (
	"errors"
	"strconv"
	"strings"
)

const (
	DropsPerXrp          float64 = 1000000
	MaxFractionLength    uint    = 6
	NativeCurrencySymbol string  = "XRP"
)

// Convert an amount in XRP to an amount in drops.
func XrpToDrops(value string) (string, error) {
	if i := strings.IndexByte(value, '.'); i != -1 && len(value[i+1:]) > int(MaxFractionLength) {
		return "", errors.New("xrp to drops: value has too many decimals")
	}

	xrpFloat, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "", err
	}

	dropsFloat := xrpFloat * DropsPerXrp
	return strconv.FormatFloat(dropsFloat, 'f', -1, 64), nil

}

// Convert an amount of drops into an amount of xrp
func DropsToXrp(value string) (string, error) {
	dropUint, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return "", err
	}

	xrpFloat := float64(dropUint) / DropsPerXrp

	return strconv.FormatFloat(xrpFloat, 'f', -1, 64), nil
}
