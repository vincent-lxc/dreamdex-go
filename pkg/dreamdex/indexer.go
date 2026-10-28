package dreamdex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const IndexerURL = "https://hideki-dev.smk.somnia.host/v1/graphql"
const RPCURL = "https://api.hideki2.infra.testnet.somnia.network"
const SpotURL = "https://stg.api.dreamdex.io/v0/markets"
const ChainID int64 = 50383

type Market struct {
	ID            string `json:"id"`
	BaseSymbol    string `json:"baseSymbol"`
	QuoteSymbol   string `json:"quoteSymbol"`
	PoolAddress   string `json:"poolAddress"`
	MarginBank    string `json:"marginBank"`
	BaseDecimals  int    `json:"baseDecimals"`
	QuoteDecimals int    `json:"quoteDecimals"`
	TickSize      string `json:"tickSize"`
	LotSize       string `json:"lotSize"`
	MinQuantity   string `json:"minQuantity"`
	MarkPrice     string `json:"markPrice"`
}

func (m Market) Symbol() string { return m.BaseSymbol + ":" + m.QuoteSymbol }

type Indexer struct {
	URL  string
	HTTP *http.Client
}

func NewIndexer() *Indexer {
	return &Indexer{URL: IndexerURL, HTTP: &http.Client{Timeout: 15 * time.Second}}
}

// Markets pages over the indexer, because a fixed first page can miss older markets.
func (c *Indexer) Markets(ctx context.Context) ([]Market, error) {
	url := c.URL
	if url == "" {
		url = IndexerURL
	}
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	const query = `query ($limit: Int!, $offset: Int!) { Market(where: {marketType: {_eq: PERP}}, order_by: {createdAtTimestamp: desc}, limit: $limit, offset: $offset) { id baseSymbol quoteSymbol poolAddress marginBank baseDecimals quoteDecimals tickSize lotSize minQuantity markPrice } }`
	var all []Market
	for offset := 0; ; offset += 100 {
		body, _ := json.Marshal(map[string]any{"query": query, "variables": map[string]int{"limit": 100, "offset": offset}})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("content-type", "application/json")
		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		var result struct {
			Data struct {
				Market []Market `json:"Market"`
			} `json:"data"`
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("indexer HTTP %d", resp.StatusCode)
		}
		if len(result.Errors) > 0 {
			return nil, fmt.Errorf("indexer: %s", result.Errors[0].Message)
		}
		all = append(all, result.Data.Market...)
		if len(result.Data.Market) < 100 {
			break
		}
	}
	return all, nil
}

func FindMarket(markets []Market, symbol string) (Market, error) {
	for _, m := range markets {
		if strings.EqualFold(m.BaseSymbol, symbol) || strings.EqualFold(m.Symbol(), symbol) {
			return m, nil
		}
	}
	return Market{}, fmt.Errorf("perp market %q not found", symbol)
}
