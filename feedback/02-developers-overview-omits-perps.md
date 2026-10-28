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

At 2026-10-28T08:06:06Z, all three URLs returned HTTP 200. The overview response contained no `Perps` text; `llms.txt` contained `Perps`; the Perps guide contained its title. The overview's REST and WebSocket guidance also gives no obvious route to the separate Hideki guide.

## Environment

2026-10-28 UTC; public testnet documentation; curl 8.14.1; no wallet or transaction.
