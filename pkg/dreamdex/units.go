package dreamdex

import (
	"errors"
	"math/big"
	"strings"
)

// FormatUnits formats an unsigned raw token amount without floating-point rounding.
func FormatUnits(raw string, decimals int) (string, error) {
	if decimals < 0 || decimals > 255 {
		return "", errors.New("invalid decimals")
	}
	n, ok := new(big.Int).SetString(raw, 10)
	if !ok || n.Sign() < 0 {
		return "", errors.New("invalid raw amount")
	}
	s := n.String()
	if decimals == 0 {
		return s, nil
	}
	if len(s) <= decimals {
		s = strings.Repeat("0", decimals-len(s)+1) + s
	}
	i := len(s) - decimals
	return strings.TrimRight(strings.TrimRight(s[:i]+"."+s[i:], "0"), "."), nil
}

// ParseUnits converts an exact nonnegative decimal string to raw token units.
func ParseUnits(value string, decimals int) (*big.Int, error) {
	if decimals < 0 || decimals > 255 {
		return nil, errors.New("invalid decimals")
	}
	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" || len(parts) == 2 && (parts[1] == "" || len(parts[1]) > decimals) {
		return nil, errors.New("invalid decimal amount")
	}
	for _, p := range parts {
		for _, c := range p {
			if c < '0' || c > '9' {
				return nil, errors.New("invalid decimal amount")
			}
		}
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	n, ok := new(big.Int).SetString(parts[0]+fraction+strings.Repeat("0", decimals-len(fraction)), 10)
	if !ok {
		return nil, errors.New("invalid decimal amount")
	}
	return n, nil
}

// OnGrid reports whether a raw amount is a positive multiple of step.
func OnGrid(amount, step *big.Int) bool {
	return amount != nil && step != nil && amount.Sign() > 0 && step.Sign() > 0 && new(big.Int).Mod(amount, step).Sign() == 0
}
