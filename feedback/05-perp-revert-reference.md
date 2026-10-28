# Linked Errors reference omits perp revert selectors

Severity: normal
Area: docs

## Steps to reproduce

1. Open the [Perps developer guide](https://app.testnet.dreamdex.io/docs/developers/perps-testnet/README.md). Its raw-order guidance links the [Contracts Errors page](https://app.testnet.dreamdex.io/docs/developers/contracts/errors.md).
2. Search that page for the perp selectors `0x714b36d6` (`MaxPositionSizeExceeded()`) and `0x3a23d825` (`InsufficientCollateral()`), both named in the Perps guide. Compare the documented spot `0xd48c4403` (`ImmediateOrCancelNoFill()`).
3. Simulate an unfunded BTC perp `placeOrder` with a plain address as `from`. `eth_estimateGas` can be called with the same payload; neither method broadcasts.

```sh
curl -sS https://app.testnet.dreamdex.io/docs/developers/contracts/errors.md | rg '714b36d6|3a23d825|d48c4403'
python3 - <<'PY'
import json, time, urllib.request
rpc = 'https://api.hideki2.infra.testnet.somnia.network'
words = [1, 0, 80900 * 10**18, 100000, (int(time.time()) + 86400) * 10**9, 0, 0, 0, 0]
data = '0x4e978373' + ''.join(f'{value:064x}' for value in words)
tx = {'from': '0x7262D22e91D38D9883275664D1dAEbca70a4CCe0',
      'to': '0x4181dd08e73bbcdd1ba3e615a18f6abd941b94c0', 'data': data}
for method, params in [('eth_call', [tx, 'latest']), ('eth_estimateGas', [tx])]:
    payload = json.dumps({'jsonrpc': '2.0', 'id': 1, 'method': method, 'params': params}).encode()
    req = urllib.request.Request(rpc, payload, {'content-type': 'application/json'})
    print(method, urllib.request.urlopen(req).read().decode())
PY
```

## Expected

The linked error reference covers perp-specific errors and explains how to interpret standard ERC-20 `Error(string)` reverts seen during perp order preflight.

## Actual

The linked page lists `0xd48c4403` under SpotPool errors but omits both perp selectors. A live BTC perp order simulation returned a standard `Error(string)` payload decoding to `ERC20: insufficient allowance`, which the linked page does not discuss. This is a documentation gap; the return data is present and decodable.

## Evidence

During the 2026-10-28 08:05 UTC probe, the BTC pool `0x4181dd08e73bbcdd1ba3e615a18f6abd941b94c0` returned the same JSON-RPC error for `eth_call` and `eth_estimateGas`:

```json
{"code":3,"message":"execution reverted","data":"0x08c379a00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001d45524332303a20696e73756666696369656e7420616c6c6f77616e6365000000"}
```

The payload decodes as `Error("ERC20: insufficient allowance")`. Keccak-256 of the three signatures gives `0x714b36d6`, `0x3a23d825`, and `0xd48c4403`, respectively. At 2026-10-28T08:07:22Z the live Errors page contained only the third.

## Environment

2026-10-28 UTC; Hideki chain 50383; SDK 0.32.1 ABI; BTC perp; read-only `eth_call`, `eth_estimateGas`, and documentation GET; no private key or transaction.
