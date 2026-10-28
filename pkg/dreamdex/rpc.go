package dreamdex

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type RPC struct{ Client *ethclient.Client }

func DialRPC(ctx context.Context) (*RPC, error) {
	c, err := ethclient.DialContext(ctx, RPCURL)
	if err != nil {
		return nil, err
	}
	return &RPC{c}, nil
}
func (r *RPC) Close() { r.Client.Close() }
func (r *RPC) CheckChain(ctx context.Context) error {
	n, err := r.Client.ChainID(ctx)
	if err != nil {
		return err
	}
	if n.Int64() != ChainID {
		return fmt.Errorf("refusing chain %s: expected %d", n, ChainID)
	}
	return nil
}

func selector(signature string) []byte { return crypto.Keccak256([]byte(signature))[:4] }
func word(b []byte, i int) (*big.Int, error) {
	if len(b) < (i+1)*32 {
		return nil, fmt.Errorf("short eth_call result: %d bytes, need %d", len(b), (i+1)*32)
	}
	return new(big.Int).SetBytes(b[i*32 : (i+1)*32]), nil
}
func addressWord(b []byte, i int) (common.Address, error) {
	if len(b) < (i+1)*32 {
		return common.Address{}, fmt.Errorf("short address result")
	}
	return common.BytesToAddress(b[i*32+12 : i*32+32]), nil
}
func (r *RPC) Call(ctx context.Context, to common.Address, data []byte, from common.Address) ([]byte, error) {
	return r.Client.CallContract(ctx, ethereum.CallMsg{From: from, To: &to, Data: data}, nil)
}
func (r *RPC) CallSelector(ctx context.Context, to common.Address, signature string) ([]byte, error) {
	return r.Call(ctx, to, selector(signature), common.Address{})
}

type BookLevel struct {
	Price    *big.Int
	Quantity *big.Int
}

func (r *RPC) Book(ctx context.Context, pool common.Address, bid bool, count uint64) ([]BookLevel, error) {
	if count == 0 || count > 50 {
		return nil, fmt.Errorf("levels must be 1..50")
	}
	data := make([]byte, 68)
	copy(data, selector("getBookLevels(bool,uint64)"))
	if bid {
		data[35] = 1
	}
	new(big.Int).SetUint64(count).FillBytes(data[36:68])
	b, err := r.Call(ctx, pool, data, common.Address{})
	if err != nil {
		return nil, err
	}
	offset, err := word(b, 0)
	if err != nil {
		return nil, err
	}
	if !offset.IsUint64() || offset.Uint64() > uint64(len(b)-32) || offset.Uint64()%32 != 0 {
		return nil, fmt.Errorf("invalid book offset")
	}
	start := int(offset.Uint64())
	length := new(big.Int).SetBytes(b[start : start+32])
	if !length.IsUint64() || length.Uint64() > count || length.Uint64() > uint64((len(b)-start-32)/64) {
		return nil, fmt.Errorf("invalid book length")
	}
	levels := make([]BookLevel, 0, length.Uint64())
	for i := 0; i < int(length.Uint64()); i++ {
		at := start + 32 + i*64
		levels = append(levels, BookLevel{new(big.Int).SetBytes(b[at : at+32]), new(big.Int).SetBytes(b[at+32 : at+64])})
	}
	return levels, nil
}

type BookParameters struct{ TickSize, MinQuantity, LotSize *big.Int }

func (r *RPC) BookParameters(ctx context.Context, pool common.Address) (BookParameters, error) {
	b, err := r.CallSelector(ctx, pool, "getOrderBookParameters()")
	if err != nil {
		return BookParameters{}, err
	}
	a, e := word(b, 0)
	if e != nil {
		return BookParameters{}, e
	}
	c, e := word(b, 1)
	if e != nil {
		return BookParameters{}, e
	}
	d, e := word(b, 2)
	return BookParameters{a, c, d}, e
}
func (r *RPC) Mark(ctx context.Context, pool common.Address) (*big.Int, bool, error) {
	b, err := r.CallSelector(ctx, pool, "tryGetMarkPrice()")
	if err != nil {
		return nil, false, err
	}
	ok, e := word(b, 0)
	if e != nil {
		return nil, false, e
	}
	price, e := word(b, 1)
	return price, ok.Sign() != 0, e
}

type SystemConfig struct {
	MarginBank, CollateralToken, PerpPoolFactory, LiquidationEngine, InsuranceFund, FeeRecipient common.Address
	MaxLeverageLimit                                                                             uint64
	FullyWired                                                                                   bool
}

func (r *RPC) SystemConfig(ctx context.Context, bank common.Address) (SystemConfig, error) {
	b, err := r.CallSelector(ctx, bank, "getSystemConfig()")
	if err != nil {
		return SystemConfig{}, err
	}
	var a [6]common.Address
	for i := range a {
		a[i], err = addressWord(b, i)
		if err != nil {
			return SystemConfig{}, err
		}
	}
	lev, err := word(b, 6)
	if err != nil {
		return SystemConfig{}, err
	}
	wired, err := word(b, 7)
	if err != nil {
		return SystemConfig{}, err
	}
	return SystemConfig{a[0], a[1], a[2], a[3], a[4], a[5], lev.Uint64(), wired.Sign() != 0}, nil
}
func ParseAddress(s string) (common.Address, error) {
	if !common.IsHexAddress(s) {
		return common.Address{}, fmt.Errorf("invalid address %q", s)
	}
	return common.HexToAddress(s), nil
}

// This rejects malformed decimal raw values before using them as a grid.
func RawInt(s string) (*big.Int, error) {
	n, ok := new(big.Int).SetString(strings.TrimSpace(s), 10)
	if !ok || n.Sign() < 0 {
		return nil, fmt.Errorf("invalid raw integer %q", s)
	}
	return n, nil
}
