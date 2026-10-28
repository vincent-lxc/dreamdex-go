package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vincent-lxc/dreamdex-go/pkg/dreamdex"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: dreamdex markets | book SYMBOL [--levels N] | config | spot-markets | order SYMBOL --side buy|sell --price P --quantity Q [--send]")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	switch args[0] {
	case "markets":
		if len(args) != 1 {
			return fmt.Errorf("usage: dreamdex markets")
		}
		ms, err := dreamdex.NewIndexer().Markets(ctx)
		if err != nil {
			return err
		}
		for _, m := range ms {
			tick, _ := dreamdex.FormatUnits(m.TickSize, m.QuoteDecimals)
			lot, _ := dreamdex.FormatUnits(m.LotSize, m.BaseDecimals)
			min, _ := dreamdex.FormatUnits(m.MinQuantity, m.BaseDecimals)
			fmt.Printf("%-18s pool=%s baseDecimals=%d quoteDecimals=%d tick=%s lot=%s minQty=%s marginBank=%s\n", m.Symbol(), m.PoolAddress, m.BaseDecimals, m.QuoteDecimals, tick, lot, min, m.MarginBank)
		}
		fmt.Printf("%d perp markets\n", len(ms))
		return nil
	case "book":
		fs := flag.NewFlagSet("book", flag.ContinueOnError)
		n := fs.Uint64("levels", 5, "depth per side")
		if len(args) < 2 {
			return fmt.Errorf("usage: dreamdex book SYMBOL [--levels N]")
		}
		symbol := args[1]
		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
		if *n == 0 || *n > 50 {
			return fmt.Errorf("levels must be 1..50")
		}
		ms, err := dreamdex.NewIndexer().Markets(ctx)
		if err != nil {
			return err
		}
		m, err := dreamdex.FindMarket(ms, symbol)
		if err != nil {
			return err
		}
		r, err := dreamdex.DialRPC(ctx)
		if err != nil {
			return err
		}
		defer r.Close()
		if err = r.CheckChain(ctx); err != nil {
			return err
		}
		pool := common.HexToAddress(m.PoolAddress)
		mark, live, markErr := r.Mark(ctx, pool)
		fmt.Printf("%s pool=%s\n", m.Symbol(), m.PoolAddress)
		if markErr != nil {
			fmt.Printf("mark: unavailable (%v)\n", markErr)
		} else if !live {
			fmt.Println("mark: stale / unavailable")
		} else {
			s, _ := dreamdex.FormatUnits(mark.String(), m.QuoteDecimals)
			fmt.Println("mark:", s)
		}
		for _, side := range []struct {
			name string
			bid  bool
		}{{"bids", true}, {"asks", false}} {
			levels, e := r.Book(ctx, pool, side.bid, *n)
			if e != nil {
				return fmt.Errorf("%s: %w", side.name, e)
			}
			fmt.Println(side.name + ":")
			for _, v := range levels {
				p, _ := dreamdex.FormatUnits(v.Price.String(), m.QuoteDecimals)
				q, _ := dreamdex.FormatUnits(v.Quantity.String(), m.BaseDecimals)
				fmt.Printf("  price=%s qty=%s\n", p, q)
			}
			if len(levels) == 0 {
				fmt.Println("  (empty)")
			}
		}
		return nil
	case "config":
		if len(args) != 1 {
			return fmt.Errorf("usage: dreamdex config")
		}
		ms, err := dreamdex.NewIndexer().Markets(ctx)
		if err != nil {
			return err
		}
		if len(ms) == 0 {
			return fmt.Errorf("no perp markets")
		}
		r, err := dreamdex.DialRPC(ctx)
		if err != nil {
			return err
		}
		defer r.Close()
		if err = r.CheckChain(ctx); err != nil {
			return err
		}
		bank := common.HexToAddress(ms[0].MarginBank)
		cfg, err := r.SystemConfig(ctx, bank)
		if err != nil {
			return err
		}
		fmt.Printf("chainId=%d marginBank=%s collateral=%s factory=%s liquidationEngine=%s insuranceFund=%s feeRecipient=%s maxLeverage=%d fullyWired=%t\n", dreamdex.ChainID, cfg.MarginBank, cfg.CollateralToken, cfg.PerpPoolFactory, cfg.LiquidationEngine, cfg.InsuranceFund, cfg.FeeRecipient, cfg.MaxLeverageLimit, cfg.FullyWired)
		m := ms[0]
		params, err := r.BookParameters(ctx, common.HexToAddress(m.PoolAddress))
		if err != nil {
			return err
		}
		tick, _ := dreamdex.FormatUnits(params.TickSize.String(), m.QuoteDecimals)
		min, _ := dreamdex.FormatUnits(params.MinQuantity.String(), m.BaseDecimals)
		lot, _ := dreamdex.FormatUnits(params.LotSize.String(), m.BaseDecimals)
		fmt.Printf("pool=%s symbol=%s tick=%s minQty=%s lot=%s\n", m.PoolAddress, m.Symbol(), tick, min, lot)
		return nil
	case "spot-markets":
		if len(args) != 1 {
			return fmt.Errorf("usage: dreamdex spot-markets")
		}
		req, err := http.NewRequestWithContext(ctx, "GET", dreamdex.SpotURL, nil)
		if err != nil {
			return err
		}
		client := http.Client{Timeout: 15 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("spot HTTP %d", resp.StatusCode)
		}
		var data struct {
			Markets []struct {
				Symbol   string `json:"symbol"`
				Kind     string `json:"kind"`
				Contract string `json:"contract"`
			} `json:"markets"`
		}
		if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return err
		}
		for _, m := range data.Markets {
			fmt.Printf("%s kind=%s contract=%s\n", m.Symbol, m.Kind, m.Contract)
		}
		fmt.Printf("%d spot markets (Shannon testnet REST API)\n", len(data.Markets))
		return nil
	case "order":
		if len(args) < 2 {
			return fmt.Errorf("usage: dreamdex order SYMBOL --side buy|sell --price P --quantity Q [--send]")
		}
		symbol := args[1]
		fs := flag.NewFlagSet("order", flag.ContinueOnError)
		side := fs.String("side", "", "buy or sell")
		priceText := fs.String("price", "", "human quote units")
		qtyText := fs.String("quantity", "", "human base units")
		send := fs.Bool("send", false, "broadcast testnet transaction")
		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
		if *side != "buy" && *side != "sell" {
			return fmt.Errorf("--side must be buy or sell")
		}
		key, err := dreamdex.LoadKey(os.Getenv("DREAMDEX_KEY_FILE"))
		if err != nil {
			return err
		}
		ms, err := dreamdex.NewIndexer().Markets(ctx)
		if err != nil {
			return err
		}
		m, err := dreamdex.FindMarket(ms, symbol)
		if err != nil {
			return err
		}
		price, err := dreamdex.ParseUnits(*priceText, m.QuoteDecimals)
		if err != nil {
			return fmt.Errorf("price: %w", err)
		}
		qty, err := dreamdex.ParseUnits(*qtyText, m.BaseDecimals)
		if err != nil {
			return fmt.Errorf("quantity: %w", err)
		}
		r, err := dreamdex.DialRPC(ctx)
		if err != nil {
			return err
		}
		defer r.Close()
		preview, err := r.PreviewOrder(ctx, m, key, *side == "buy", price, qty)
		fmt.Printf("account=%s symbol=%s side=%s price=%s quantity=%s voucher=%s\n", preview.Account, m.Symbol(), strings.ToUpper(*side), *priceText, *qtyText, preview.Voucher)
		if err != nil {
			return err
		}
		fmt.Printf("order eth_call=accepted estimatedGas=%d\n", preview.Gas)
		if !*send {
			fmt.Println("dry run only; pass --send to broadcast")
			return nil
		}
		hash, err := r.SendOrder(ctx, key, preview)
		if err != nil {
			return err
		}
		fmt.Println("tx=", hash.Hex())
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
