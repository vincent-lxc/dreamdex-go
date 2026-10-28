package dreamdex

import (
	"math/big"
	"testing"
)

func TestUnits(t *testing.T) {
	for _, tc := range []struct {
		value    string
		decimals int
		raw      string
	}{{"0.001", 8, "100000"}, {"100000.25", 18, "100000250000000000000000"}, {"0.000000000000000000000001", 24, "1"}, {"1", 0, "1"}} {
		n, err := ParseUnits(tc.value, tc.decimals)
		if err != nil || n.String() != tc.raw {
			t.Fatalf("parse %q: %v %v", tc.value, n, err)
		}
		got, err := FormatUnits(tc.raw, tc.decimals)
		if err != nil || got != tc.value {
			t.Fatalf("format %q: %q %v", tc.raw, got, err)
		}
	}
	if _, err := ParseUnits("0.000000001", 8); err == nil {
		t.Fatal("accepted excess precision")
	}
	if _, err := ParseUnits("-1", 18); err == nil {
		t.Fatal("accepted negative")
	}
}
func TestGrid(t *testing.T) {
	step := big.NewInt(25)
	if !OnGrid(big.NewInt(100), step) || OnGrid(big.NewInt(101), step) || OnGrid(big.NewInt(0), step) {
		t.Fatal("grid mismatch")
	}
}
