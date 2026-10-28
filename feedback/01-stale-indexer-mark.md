# Indexer perp marks remain unchanged while live pool marks move

Severity: major
Area: API

## Steps to reproduce

1. Query `Market.markPrice`, `markPriceUpdatedAt`, and `poolAddress` for BTC, ETH, and SOL on the [Hideki GraphQL endpoint](https://hideki-dev.smk.somnia.host/v1/graphql).
2. Call `tryGetMarkPrice()` (`0xa8c485ca`) on each `poolAddress` through the [Hideki RPC](https://api.hideki2.infra.testnet.somnia.network) using `eth_call`.
3. Repeat at least one minute later. Divide both raw prices by `10^18`.

```sh
curl -sS -H 'content-type: application/json' --data '{"query":"query { Market(where: {marketType: {_eq: PERP}, baseSymbol: {_in: [\"BTC\",\"ETH\",\"SOL\"]}}) { baseSymbol poolAddress markPrice markPriceUpdatedAt } }"}' https://hideki-dev.smk.somnia.host/v1/graphql
curl -sS -H 'content-type: application/json' --data '{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0x4181dd08e73bbcdd1ba3e615a18f6abd941b94c0","data":"0xa8c485ca"},"latest"]}' https://api.hideki2.infra.testnet.somnia.network
```

## Expected

The perps guide explains the indexer mark update cadence and tells readers when to use `markPriceUpdatedAt` or the live pool read for freshness-sensitive work.

## Actual

All three indexer marks and update timestamps stayed unchanged across two samples 76 seconds apart, while each pool returned `ok=true` and a changing mark. The [perps developer guide](https://app.testnet.dreamdex.io/docs/developers/perps-testnet/README.md) does not explain this cadence.

## Evidence

| UTC sample | Market | Indexer mark | Age at sample | Pool mark |
|---|---|---:|---:|---:|
| 2026-10-28T08:05:44Z | BTC | 80823.962412778493 | 3241 s | 81074.3140977645175 |
| 2026-10-28T08:07:00Z | BTC | 80823.962412778493 | 3317 s | 81060.5165316030935 |
| 2026-10-28T08:05:44Z | ETH | 2350.025246845028 | 2786 s | 2363.31863639817 |
| 2026-10-28T08:07:00Z | ETH | 2350.025246845028 | 2862 s | 2360.9665340660255 |
| 2026-10-28T08:05:44Z | SOL | 102.094494293561 | 691 s | 102.3178544375455 |
| 2026-10-28T08:07:00Z | SOL | 102.094494293561 | 767 s | 102.2512638815365 |

The indexer `markPriceUpdatedAt` values remained `1793171503` (BTC), `1793171958` (ETH), and `1793174053` (SOL). This establishes sampled delay, not a universal update interval.

## Environment

2026-10-28 UTC; Hideki testnet chain 50383; SDK 0.32.1 ABI reference; GraphQL and `eth_call` only; no wallet or transaction.
