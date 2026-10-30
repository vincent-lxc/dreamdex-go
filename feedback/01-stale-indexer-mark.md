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

On 2026-10-30, all five sampled indexer marks and update timestamps stayed unchanged across three samples spanning 120 seconds, while each pool returned `ok=true` and a changing mark. The earlier three-market, two-sample observation also remains below. The [perps developer guide](https://app.testnet.dreamdex.io/docs/developers/perps-testnet/README.md) does not explain this cadence.

## Evidence

Re-verified 2026-10-30 03:37 UTC. Three samples started 60 seconds apart (03:35:23, 03:36:23, 03:37:23 UTC). All five indexer marks and timestamps stayed fixed while all five pool marks moved; every pool returned `ok=true`. This is a sampled delay, not a measured update cadence.

| UTC sample | Market | Indexer mark | `markPriceUpdatedAt` | Age | Live pool mark |
|---|---|---:|---:|---:|---:|
| 2026-10-30T03:35:23Z | BTC | 77353.2443620564865 | 1793329904 | 1419 s | 77661.0623116485045 |
| 2026-10-30T03:35:23Z | ETH | 2209.223360003374 | 1793330358 | 965 s | 2220.2553269837225 |
| 2026-10-30T03:35:23Z | SOL | 94.935428677001 | 1793328856 | 2467 s | 96.36147015697 |
| 2026-10-30T03:35:23Z | DOGE | 0.0749327873805 | 1793330082 | 1241 s | 0.0755561079285 |
| 2026-10-30T03:35:23Z | AVAX | 8.317315649009 | 1793330084 | 1239 s | 8.393319434046 |
| 2026-10-30T03:36:23Z | BTC | 77353.2443620564865 | 1793329904 | 1479 s | 77671.0413085312245 |
| 2026-10-30T03:36:23Z | ETH | 2209.223360003374 | 1793330358 | 1025 s | 2220.8485413094245 |
| 2026-10-30T03:36:23Z | SOL | 94.935428677001 | 1793328856 | 2527 s | 96.3489117623705 |
| 2026-10-30T03:36:23Z | DOGE | 0.0749327873805 | 1793330082 | 1301 s | 0.075569972991 |
| 2026-10-30T03:36:23Z | AVAX | 8.317315649009 | 1793330084 | 1299 s | 8.3955079182485 |
| 2026-10-30T03:37:23Z | BTC | 77353.2443620564865 | 1793329904 | 1539 s | 77787.824378077572 |
| 2026-10-30T03:37:23Z | ETH | 2209.223360003374 | 1793330358 | 1085 s | 2221.83814248459 |
| 2026-10-30T03:37:23Z | SOL | 94.935428677001 | 1793328856 | 2587 s | 96.348380243171 |
| 2026-10-30T03:37:23Z | DOGE | 0.0749327873805 | 1793330082 | 1361 s | 0.0755956062025 |
| 2026-10-30T03:37:23Z | AVAX | 8.317315649009 | 1793330084 | 1359 s | 8.393822509263 |

Raw first-sample responses (five-market GraphQL result and BTC `eth_call`):

```json
{"data":{"Market":[{"baseSymbol":"DOGE","poolAddress":"0xc410012e39765cd5f772e84bacb19554aec040df","markPrice":"74932787380500000","markPriceUpdatedAt":"1793330082"},{"baseSymbol":"AVAX","poolAddress":"0xde9762b6b441133d1314c68864ae4e63a6ee45c7","markPrice":"8317315649009000000","markPriceUpdatedAt":"1793330084"},{"baseSymbol":"SOL","poolAddress":"0xb483ec1fb34996638677c20ab087b16f1896fa53","markPrice":"94935428677001000000","markPriceUpdatedAt":"1793328856"},{"baseSymbol":"ETH","poolAddress":"0xd6cb1dd58580b84d76068ac21594219ce03aeddf","markPrice":"2209223360003374000000","markPriceUpdatedAt":"1793330358"},{"baseSymbol":"BTC","poolAddress":"0x4181dd08e73bbcdd1ba3e615a18f6abd941b94c0","markPrice":"77353244362056486500000","markPriceUpdatedAt":"1793329904"}]}}
{"jsonrpc":"2.0","id":1,"result":"0x000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000107203be628065cecf20"}
```

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

2026-10-30 03:35–03:37 UTC; Somnia Tokyo testnet (Hideki, chain 50383); wallet: n/a, read-only; GraphQL and pool `eth_call`; SDK 0.32.1 ABI reference; no transaction sent.
