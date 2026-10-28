package dreamdex

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

const orderABI = `[{"name":"placeOrder","type":"function","stateMutability":"nonpayable","inputs":[{"name":"isBid","type":"bool"},{"name":"userData","type":"uint64"},{"name":"price","type":"uint256"},{"name":"quantity","type":"uint256"},{"name":"expireTimestampNs","type":"uint64"},{"name":"orderType","type":"uint8"},{"name":"selfMatchingOption","type":"uint8"},{"name":"builder","type":"address"},{"name":"builderFeeBpsTimes1k","type":"uint96"}],"outputs":[]}]`

// LoadKey refuses group or world access before reading key material.
func LoadKey(path string) (*ecdsa.PrivateKey, error) {
	if path == "" {
		return nil, fmt.Errorf("DREAMDEX_KEY_FILE is required")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != 0600 {
		return nil, fmt.Errorf("key file must be regular and mode 0600")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return crypto.HexToECDSA(strings.TrimPrefix(strings.TrimSpace(string(b)), "0x"))
}

type OrderPreview struct {
	Account common.Address
	Voucher string
	Gas     uint64
	Data    []byte
	Pool    common.Address
}

// PreviewOrder validates on-chain grid and simulates voucher claim and the order.
// A successful eth_call is only a momentary check; it cannot reserve margin.
func (r *RPC) PreviewOrder(ctx context.Context, m Market, key *ecdsa.PrivateKey, bid bool, price, quantity *big.Int) (OrderPreview, error) {
	if err := r.CheckChain(ctx); err != nil {
		return OrderPreview{}, err
	}
	pool, err := ParseAddress(m.PoolAddress)
	if err != nil {
		return OrderPreview{}, err
	}
	bank, err := ParseAddress(m.MarginBank)
	if err != nil {
		return OrderPreview{}, err
	}
	params, err := r.BookParameters(ctx, pool)
	if err != nil {
		return OrderPreview{}, err
	}
	if !OnGrid(price, params.TickSize) {
		return OrderPreview{}, fmt.Errorf("price is off tick grid")
	}
	if !OnGrid(quantity, params.LotSize) || quantity.Cmp(params.MinQuantity) < 0 {
		return OrderPreview{}, fmt.Errorf("quantity is off lot grid or below minimum")
	}
	account := crypto.PubkeyToAddress(key.PublicKey)
	preview := OrderPreview{Account: account, Pool: pool, Voucher: "unavailable"}
	granterBytes, err := r.CallSelector(ctx, bank, "getCreditGranter()")
	if err == nil {
		granter, e := addressWord(granterBytes, 0)
		if e == nil && granter != (common.Address{}) {
			_, e = r.Call(ctx, granter, selector("claim()"), account)
			if e == nil {
				preview.Voucher = "eligible (claim eth_call succeeded)"
			} else {
				preview.Voucher = "ineligible (claim eth_call reverted)"
			}
		} else {
			preview.Voucher = "dormant"
		}
	}
	parsed, err := abi.JSON(strings.NewReader(orderABI))
	if err != nil {
		return OrderPreview{}, err
	}
	expiry := uint64(time.Now().Unix()+50*365*24*3600) * 1_000_000_000
	data, err := parsed.Pack("placeOrder", bid, uint64(0), price, quantity, expiry, uint8(0), uint8(0), common.Address{}, big.NewInt(0))
	if err != nil {
		return OrderPreview{}, err
	}
	preview.Data = data
	msg := ethereum.CallMsg{From: account, To: &pool, Data: data}
	if _, err = r.Client.CallContract(ctx, msg, nil); err != nil {
		return preview, fmt.Errorf("order eth_call rejected: %w", err)
	}
	gas, err := r.Client.EstimateGas(ctx, msg)
	if err != nil {
		return preview, fmt.Errorf("order gas estimation failed: %w", err)
	}
	preview.Gas = gas
	return preview, nil
}

// SendOrder is called only after the CLI's explicit --send gate.
func (r *RPC) SendOrder(ctx context.Context, key *ecdsa.PrivateKey, p OrderPreview) (common.Hash, error) {
	if err := r.CheckChain(ctx); err != nil {
		return common.Hash{}, err
	}
	nonce, err := r.Client.PendingNonceAt(ctx, p.Account)
	if err != nil {
		return common.Hash{}, err
	}
	gasPrice, err := r.Client.SuggestGasPrice(ctx)
	if err != nil {
		return common.Hash{}, err
	}
	gas := p.Gas + p.Gas/4
	if gas > 10_000_000 {
		return common.Hash{}, fmt.Errorf("gas estimate exceeds 10M cap")
	}
	tx := types.NewTransaction(nonce, p.Pool, big.NewInt(0), gas, gasPrice, p.Data)
	signed, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(ChainID)), key)
	if err != nil {
		return common.Hash{}, err
	}
	if err = r.Client.SendTransaction(ctx, signed); err != nil {
		return common.Hash{}, err
	}
	return signed.Hash(), nil
}
