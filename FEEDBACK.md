# Builder feedback

I tested the dreamDEX Go client's Hideki perp reads, voucher and order simulations, staging REST API, and public developer documentation on **2026-10-28, 08:05–08:08 UTC**. Evidence comes from live GraphQL, REST GET, `eth_call`, `eth_estimateGas`, and documentation GET requests. All chain work was on **testnet only** (Hideki 50383; Shannon 50312 for the REST comparison). No transaction was sent or broadcast.

Tools: Go 1.27.3 (project target Go 1.22), go-ethereum 1.14.13, curl 8.14.1, Python 3.13.5; `@somnia-chain/markets-sdk` 0.32.1 supplied the reference ABIs and eligibility logic. All six reports below reproduced today.

| Item | Severity | Area | Summary | Suggestion |
|---|---|---|---|---|
| [01 · Indexer mark freshness](feedback/01-stale-indexer-mark.md) | Major | API | Three indexer marks stayed fixed while pool marks moved over 76 seconds. | Document cadence and when to check timestamps or read pools. |
| [02 · Perps guide navigation](feedback/02-developers-overview-omits-perps.md) | Normal | docs | Developers overview omits its live Perps guide. | Add the guide to the overview list. |
| [03 · REST network split](feedback/03-rest-network-split.md) | Normal | docs | Staging REST rejects Hideki perps and lists Shannon spot only. | State the split and link Hideki GraphQL/RPC from the Perps guide. |
| [04 · Voucher eligibility](feedback/04-voucher-eligibility-false-positive.md) | Major | SDK | Two flat, unclaimed addresses pass SDK checks but `claim()` reverts. | Align the eligibility check with contract policy and publish the selector. |
| [05 · Perp revert reference](feedback/05-perp-revert-reference.md) | Normal | docs | Linked Errors page omits perp selectors and an observed ERC-20 revert path. | Add a perp error table and decoding guidance. |
| [06 · Hideki STT onboarding](feedback/06-hideki-gas-onboarding.md) | Normal | docs | Public faucet guidance targets Shannon, while Hideki gas route uses the app. | Identify a Hideki funding route for standalone bot addresses. |

## Checks that found no issue

All 21 live perp markets matched the documented base-decimals table; their indexer tick, lot, and minimum quantities matched pool reads. The documented core addresses matched on-chain system config, and the tested relative links in the Perps guide returned HTTP 200. The Hideki RPC reported chain 50383 and its Tokyo explorer link resolved. An unfunded order reverted during `eth_estimateGas`, so the guide's successful-order gas figures could not be retested; that is not a reported mismatch.

## How to reproduce

Run the read-only [CLI commands in README.md](README.md#quickstart) for markets, books, config, and Shannon spot comparison. Each report includes its exact GraphQL, REST, or RPC reproduction command and UTC response. The full investigation trail, including checks without findings, is in [notes/dx-log.md](notes/dx-log.md).
