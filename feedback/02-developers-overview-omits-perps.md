# Developers overview omits the live Perps guide

Severity: normal
Area: docs

## Steps to reproduce

```sh
curl -sS https://app.testnet.dreamdex.io/docs/developers/README.md
curl -sS https://app.testnet.dreamdex.io/llms.txt
curl -sS -o /dev/null -w '%{http_code}\n' https://app.testnet.dreamdex.io/docs/developers/perps-testnet/README.md
```

## Expected

The Developers overview lists its live Perps on Testnet for Developers guide beside Contracts, HTTP API, WebSocket API, and Libraries.

## Actual

The overview lists those four sections but no Perps guide. `llms.txt` lists Perps (Testnet), and the guide itself is live.

## Evidence

Re-verified 2026-10-30 03:35 UTC. The live Developers overview still omits the Perps guide, while `llms.txt` lists it and the guide returned HTTP 200. Raw response snippets:

```text
* [**Contracts**](contracts/) — Solidity reference for the on-chain order book: functions for placing and managing orders, events emitted during matching, and the types used across both.
* [**HTTP API**](http-api/) — REST endpoints for market data, authentication, trading, and vault operations. The API returns unsigned transactions that you sign and broadcast yourself.
* [**WebSocket API**](websocket-api/) — Real-time streaming interface for order book updates, trade feeds, and order status changes.
* [**Libraries**](libraries/) — Client libraries wrapping the dreamDEX APIs, including the CCXT integration for TypeScript/JavaScript.
### Perps (Testnet)
- [Overview](/docs/developers/perps-testnet/README.md)
200
```

The first four lines are from the overview; the next two are from `llms.txt`, followed by the guide's HTTP status.


At 2026-10-28T08:06:06Z, all three URLs returned HTTP 200. The overview response contained no `Perps` text; `llms.txt` contained `Perps`; the Perps guide contained its title. The overview's REST and WebSocket guidance also gives no obvious route to the separate Hideki guide.

## Environment

2026-10-30 03:35 UTC; Somnia Tokyo testnet (Hideki, chain 50383); wallet: n/a, read-only; documentation GETs; no transaction sent.
