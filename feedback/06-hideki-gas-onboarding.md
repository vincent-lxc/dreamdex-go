# Public STT faucet guidance leads to Shannon, not Hideki

Severity: normal
Area: docs

## Steps to reproduce

1. Read the [Perps developer guide](https://app.testnet.dreamdex.io/docs/developers/perps-testnet/README.md), which says Hideki perps use chain 50383 and STT for gas. Its documented funding route is to sign into the app, link a bot, then use **Send it … for gas** or **Top up**.
2. Follow the public Somnia faucet link in the [general Quick Start](https://app.testnet.dreamdex.io/docs/welcome/quick-start.md).
3. Check the faucet page's network label and the Hideki RPC's chain ID.

```sh
curl -sS https://app.testnet.dreamdex.io/docs/welcome/quick-start.md | rg 'STT \(gas\)|50312'
curl -sS -L https://testnet.somnia.network/ | rg -o 'Testnet, Shannon|Shannon' | head
curl -sS -H 'content-type: application/json' --data '{"jsonrpc":"2.0","id":1,"method":"eth_chainId","params":[]}' https://api.hideki2.infra.testnet.somnia.network
```

## Expected

A developer funding a standalone Hideki bot address can identify an applicable chain-50383 STT faucet or a clearly documented alternative, with a warning that the general faucet instructions target another chain.

## Actual

The public faucet page identifies **Testnet, Shannon**, and Quick Start explicitly introduces it under Shannon chain 50312. The Hideki RPC returns `0xc4cf` (50383). The Perps guide's concrete route goes through the app and linked-wallet UI; it does not name a public Hideki faucet or a standalone non-app funding path. This does not prove that no such faucet exists.

## Evidence

At 2026-10-28T08:07:22Z, the faucet URL returned HTTP 200 with an accessible page label `Experience Somnia’s Testnet, Shannon`. Quick Start said `STT (gas): claim from the Somnia testnet faucet` within its Shannon 50312 section. The RPC `eth_chainId` response was `{"result":"0xc4cf"}` at 2026-10-28T08:05:43Z. The Perps guide separately documents the app route for Hideki gas.

## Environment

2026-10-28 UTC; public documentation and Somnia faucet page; Hideki chain 50383; REST GET and `eth_chainId` only; no faucet claim or transaction.
