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

Re-verified 2026-10-30 03:35 UTC. The exact REST requests still returned HTTP 400 for the Hideki perp and a three-market spot list. Raw response snippets:

```json
{"context":{"invalid_symbols":["BTC:USDso"]},"description":"unknown market symbols: BTC:USDso","name":"invalid_param","status":400}
{"markets":[{"base":"0x28f34DeFd2b4CB48d9eE6d89f2Be4Bc601694c00","baseDecimals":18,"contract":"0x259fD6559214dd5aD3752322426eA9F9fABEFff4","kind":"spot","lotSize":"0.01","minQuantity":"1","quote":"0x9c32F3827A1a99f0cf9B213de8b53eC3d57bb171","quoteDecimals":18,"stopRegistry":"0xEb97349Aa62A68507c0bE535eD88B0d028a47E1e","symbol":"SOMI:USDso","tickSize":"0.0001"},{"base":"0x4e85DC48a70DA1298489d5B6FC2492767d98f384","baseDecimals":8,"contract":"0x3605f28aA7C50e7441211e77Cb0762d49539326C","kind":"spot","lotSize":"0.00001","minQuantity":"0.0001","quote":"0x9c32F3827A1a99f0cf9B213de8b53eC3d57bb171","quoteDecimals":18,"stopRegistry":"0x53d5B2b0791b3992a1F3b5e0b0277Ee2e08B7aaD","symbol":"WBTC:USDso","tickSize":"0.1"},{"base":"0x4d8E02BBfCf205828A8352Af4376b165E123D7b0","baseDecimals":18,"contract":"0xD180195da5459C7a0DEA188ed61216ec43682b50","kind":"spot","lotSize":"0.0001","minQuantity":"0.001","quote":"0x9c32F3827A1a99f0cf9B213de8b53eC3d57bb171","quoteDecimals":18,"stopRegistry":"0xf822D4Cb94902d667c9650e702aA5f096cc7598F","symbol":"WETH:USDso","tickSize":"0.01"}]}
```

The Perps guide still gives Hideki GraphQL and RPC endpoints without explaining the staging REST split.


At 2026-10-28T08:06:06Z, `GET /v0/orderbooks?symbols=BTC:USDso` returned HTTP 400:

```json
{"context":{"invalid_symbols":["BTC:USDso"]},"description":"unknown market symbols: BTC:USDso","name":"invalid_param","status":400}
```

`GET /v0/markets` returned HTTP 200 with `SOMI:USDso`, `WBTC:USDso`, and `WETH:USDso`, each `kind=spot`. The Hideki GraphQL indexer simultaneously returned `BTC` as a PERP market.

## Environment

2026-10-30 03:35 UTC; Somnia Tokyo testnet (Hideki, chain 50383); staging REST serves Shannon chain 50312; wallet: n/a, read-only; REST GET, documentation GET, and GraphQL; no transaction sent.
