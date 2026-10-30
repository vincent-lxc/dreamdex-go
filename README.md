# dreamdex-go

A small Go client and CLI for dreamDEX perpetual markets on the **Somnia Hideki testnet (chain 50383)**. It is for builders who want to inspect live perp markets, books, marks, and contract configuration without running the TypeScript SDK. The optional order path is a guarded testnet scaffold.

## Quickstart

Requires Go 1.22 or newer and access to the public Hideki indexer and RPC.

```sh
go run ./cmd/dreamdex markets
go run ./cmd/dreamdex book BTC --levels 3
go run ./cmd/dreamdex config
go test ./...
```

The first run downloads `go-ethereum`. The CLI uses the documented public endpoints; it needs no key for reads.

## Commands

The output below was captured from live endpoints on **2026-10-28 at 08:06 UTC**. Prices, depth, and market counts change. Long lists are trimmed where marked.

### `markets`

Lists Hideki perp markets from the GraphQL indexer. Tick sizes use quote decimals; lot and minimum quantities use base decimals.

```text
$ go run ./cmd/dreamdex markets
ZRO:USDso          pool=0x2a1822b02039ddf0e293f3f7243673fc03eaf16c baseDecimals=18 quoteDecimals=18 tick=0.0001 lot=0.1 minQty=5.1 marginBank=0x478769e1177c0773c43a593dccc4ad15f15f1394
LIT:USDso          pool=0xdcb0ad226727b2ce92857f0af5751389d42b2a3f baseDecimals=18 quoteDecimals=18 tick=0.0001 lot=0.1 minQty=2.8 marginBank=0x478769e1177c0773c43a593dccc4ad15f15f1394
... 18 market rows omitted ...
BTC:USDso          pool=0x4181dd08e73bbcdd1ba3e615a18f6abd941b94c0 baseDecimals=8 quoteDecimals=18 tick=0.1 lot=0.00001 minQty=0.00013 marginBank=0x478769e1177c0773c43a593dccc4ad15f15f1394
21 perp markets
```

### `book <SYMBOL>`

Reads the mark and top bid/ask levels directly from the market pool. The mark has a live/stale signal, and all raw values are scaled by that market's own decimals.

```text
$ go run ./cmd/dreamdex book BTC --levels 3
BTC:USDso pool=0x4181dd08e73bbcdd1ba3e615a18f6abd941b94c0
mark: 81084.804939359058
bids:
  price=81277.8 qty=0.01233
  price=81277.4 qty=0.14799
  price=81277.3 qty=0.01849
asks:
  price=81290.7 qty=0.30216
  price=81291.7 qty=0.02466
  price=81293.2 qty=0.01233
```

### `config`

Checks `eth_chainId == 50383`, calls `MarginBank.getSystemConfig()`, then reads `getOrderBookParameters()` on a market pool.

```text
$ go run ./cmd/dreamdex config
chainId=50383 marginBank=0x478769E1177C0773c43A593DCCc4Ad15F15F1394 collateral=0xed7993A74dAe1e6656B704A424DcD8711aAf4F72 factory=0xE631fAEE547baeE07353E290f0562ea67d856560 liquidationEngine=0xA66d3C24Ad4a389d863846905D8173e1017e9F4D insuranceFund=0x50826894D47C6D4f3BeA38BFDC7Dc5c5be171b80 feeRecipient=0xd58596620Ee6951B2958a931949e4ff7ce6222be maxLeverage=20 fullyWired=true
pool=0x2a1822b02039ddf0e293f3f7243673fc03eaf16c symbol=ZRO:USDso tick=0.0001 minQty=5.1 lot=0.1
```

### `spot-markets`

For comparison only: this calls the documented staging REST API, which serves **Shannon testnet (50312) spot**, not Hideki perps.

```text
$ go run ./cmd/dreamdex spot-markets
SOMI:USDso kind=spot contract=0x259fD6559214dd5aD3752322426eA9F9fABEFff4
WBTC:USDso kind=spot contract=0x3605f28aA7C50e7441211e77Cb0762d49539326C
WETH:USDso kind=spot contract=0xD180195da5459C7a0DEA188ed61216ec43682b50
3 spot markets (Shannon testnet REST API)
```

### `order <SYMBOL>`

This scaffold requires `DREAMDEX_KEY_FILE` to point to a **regular file with mode 0600** containing a hex private key. It checks chain 50383, validates price and quantity against the pool's tick/lot/minimum grid, simulates a voucher claim for eligibility, simulates the order with `eth_call`, and estimates gas. It broadcasts only when `--send` is supplied. The implementation uses the key's **own** perp account; it does not support linked operator wallets or order management yet.

```sh
chmod 600 /path/to/testnet.key
export DREAMDEX_KEY_FILE=/path/to/testnet.key
go run ./cmd/dreamdex order BTC --side buy --price 80900 --quantity 0.001
# Add --send only after reviewing the preview and funding this testnet account.
```

An unfunded test-only key produced this real dry-run result; the rejected simulation sent no transaction:

```text
account=0x7262D22e91D38D9883275664D1dAEbca70a4CCe0 symbol=BTC:USDso side=BUY price=80900 quantity=0.001 voucher=ineligible (claim eth_call reverted)
error: order eth_call rejected: execution reverted
```

**Safety:** The CLI never prints key material. It checks the chain ID before an order simulation or send, uses zero transaction value, and requires `--send` to broadcast. A successful preview is not a guarantee that the transaction will succeed later; book, mark, margin, and nonce can change. This project is testnet only. No transaction was sent while building or validating it.

## Architecture

`pkg/dreamdex` contains the paginated GraphQL market client, RPC reads, exact integer unit conversion, and guarded order preview/send code. `cmd/dreamdex` is a thin CLI. GraphQL supplies market discovery and per-market decimals; pool `eth_call`s supply book levels and live mark. MarginBank supplies the system address book. The SDK 0.32.1 source and ABIs were used as the reference for query fields and method signatures.

## Feedback

The [feedback index](FEEDBACK.md) summarizes the live, read-only investigations. All six reports were re-verified on 2026-10-30 and still reproduce. Individual reports:

- [01 · Indexer mark freshness](feedback/01-stale-indexer-mark.md) — major; reproduced
- [02 · Perps guide navigation](feedback/02-developers-overview-omits-perps.md) — normal; reproduced
- [03 · REST network split](feedback/03-rest-network-split.md) — normal; reproduced
- [04 · Voucher eligibility](feedback/04-voucher-eligibility-false-positive.md) — major; reproduced
- [05 · Perp revert reference](feedback/05-perp-revert-reference.md) — normal; reproduced
- [06 · Hideki STT onboarding](feedback/06-hideki-gas-onboarding.md) — normal; reproduced

The full investigation, including checks that found no issue, is in [notes/dx-log.md](notes/dx-log.md).

## Roadmap

- Add margin and position reads, order history, and explicit freshness checks.
- Build a testnet perps bot/trader tool with bounded risk, cancellations, and linked operator wallet support.

## X posts

