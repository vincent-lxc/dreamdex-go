# Staging REST rejects Hideki perp markets

Severity: normal
Area: docs

## Steps to reproduce

1. Read the [Perps developer guide](https://app.testnet.dreamdex.io/docs/developers/perps-testnet/README.md), which identifies Hideki chain 50383 and `BTC:USDso` as a perp market.
2. Query the staging REST API described by the general [HTTP API guide](https://app.testnet.dreamdex.io/docs/developers/http-api/).

```sh
curl -sS -i 'https://stg.api.dreamdex.io/v0/orderbooks?symbols=BTC:USDso'
curl -sS 'https://stg.api.dreamdex.io/v0/markets'
```

## Expected

The Perps guide explicitly says the staging REST API serves Shannon spot markets and directs readers to Hideki GraphQL and RPC for perps.

## Actual

The REST order-book request rejects the valid Hideki perp symbol. Its markets list contains only Shannon spot markets. The general HTTP API guide identifies Shannon chain 50312, but the Perps guide does not explain the split at the point where a developer chooses a data source.

## Evidence

At 2026-10-28T08:06:06Z, `GET /v0/orderbooks?symbols=BTC:USDso` returned HTTP 400:

```json
{"context":{"invalid_symbols":["BTC:USDso"]},"description":"unknown market symbols: BTC:USDso","name":"invalid_param","status":400}
```

`GET /v0/markets` returned HTTP 200 with `SOMI:USDso`, `WBTC:USDso`, and `WETH:USDso`, each `kind=spot`. The Hideki GraphQL indexer simultaneously returned `BTC` as a PERP market.

## Environment

2026-10-28 UTC; Hideki 50383 vs Shannon 50312; SDK 0.32.1 reference; REST GET and GraphQL only; no wallet or transaction.
