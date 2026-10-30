# Voucher eligibility checks pass but claim simulation reverts

Severity: major
Area: SDK

## Steps to reproduce

1. Read the [Perps developer guide's voucher claim](https://app.testnet.dreamdex.io/docs/developers/perps-testnet/README.md): any flat, unclaimed address can claim the 1,000 USDso credit, and the example gates `claimPerpVoucher` on `getPerpVoucherEligibility().eligible`.
2. At one block, apply the SDK 0.32.1 `getPerpVoucherEligibility` reads to `0x7262D22e91D38D9883275664D1dAEbca70a4CCe0` and `0x000000000000000000000000000000000000dEaD`: `getCreditGranter()` on MarginBank, `getConfig()` / `budgetRemaining()` / `isClaimed(address)` on the distributor, and `getActivePerpPools(address)` / `getMainFundedPrincipal(address)` on MarginBank.
3. Simulate `claim()` from each plain address at that block. No private key is needed for `eth_call`.

```sh
curl -sS -H 'content-type: application/json' --data '{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"from":"0x000000000000000000000000000000000000dEaD","to":"0x2ACD9EFdA721450906bfE1bef3edE269636E42C4","data":"0x4e71d92d"},"0x2195ef5"]}' https://api.hideki2.infra.testnet.somnia.network
```

`0x4e71d92d` is `claim()`; the block tag above is the fixed block from Evidence. To repeat at the head, replace it with `latest`.

The following runs the eligibility reads and both claim simulations at that same block with Python's standard library:

```sh
python3 - <<'PY'
import json, urllib.request
rpc = 'https://api.hideki2.infra.testnet.somnia.network'
bank = '0x478769E1177C0773c43A593DCCc4Ad15F15F1394'
voucher = '0x2ACD9EFdA721450906bfE1bef3edE269636E42C4'
block = '0x2195ef5'
def call(to, data, sender=None):
    tx = {'to': to, 'data': data}
    if sender: tx['from'] = sender
    body = json.dumps({'jsonrpc':'2.0','id':1,'method':'eth_call','params':[tx,block]}).encode()
    request = urllib.request.Request(rpc, body, {'content-type':'application/json'})
    response = json.load(urllib.request.urlopen(request))
    return response.get('result', response.get('error'))
print('granter', call(bank, '0x6a5c0a70'))
print('config', call(voucher, '0xc3f909d4'))
print('budget', call(voucher, '0x0417204b'))
for account in ['0x7262D22e91D38D9883275664D1dAEbca70a4CCe0',
                '0x000000000000000000000000000000000000dEaD']:
    arg = account[2:].lower().zfill(64)
    print(account, 'active', call(bank, '0xde3ba54d' + arg),
          'mainFunded', call(bank, '0x7856c5bf' + arg),
          'claimed', call(voucher, '0x8cc08025' + arg),
          'claim', call(voucher, '0x4e71d92d', account))
PY
```

## Expected

The SDK eligibility result agrees with whether the distributor would accept a claim at the same block, or reports every additional precondition and its revert reason.

## Actual

Both accounts satisfy every check in SDK 0.32.1's eligibility implementation, implying `eligible: true` with a 1,000 USDso grant. Both `claim()` simulations revert with the same unlisted custom selector, `0x3e31dcc2`. The selector's meaning is unknown from the published SDK ABI and linked Errors page; an undocumented requirement should not be inferred from it.

## Evidence

Re-verified 2026-10-30 03:36 UTC. The fixed-block command still reverts for both addresses: `{"jsonrpc":"2.0","id":1,"error":{"code":3,"message":"execution reverted","data":"0x3e31dcc2"}}`. Repeating all eligibility reads and both `claim()` calls at `latest` returned the same contradiction. At 03:41 UTC, `eth_blockNumber` returned `0x2314937`; pinning every read and both claims to that head block gave the same result. Raw response snippets:

```text
{"jsonrpc":"2.0","id":1,"result":"0x2314937"}
0x7262D22e91D38D9883275664D1dAEbca70a4CCe0 active 0x00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000 mainFunded 0x0000000000000000000000000000000000000000000000000000000000000000 claimed 0x0000000000000000000000000000000000000000000000000000000000000000 claim {'code': 3, 'message': 'execution reverted', 'data': '0x3e31dcc2'}
0x000000000000000000000000000000000000dEaD active 0x00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000 mainFunded 0x0000000000000000000000000000000000000000000000000000000000000000 claimed 0x0000000000000000000000000000000000000000000000000000000000000000 claim {'code': 3, 'message': 'execution reverted', 'data': '0x3e31dcc2'}
```

The npm registry returned `"version":"0.32.1"`, `"latest":"0.32.1"`, and `"canary":"0.32.2-canary.38575926931.ge4550933d36d"`. `active` decodes to an empty array; the config still has an active campaign and a 1,000 USDso grant, and the latest budget exceeds the grant. A GET of `llms-full.txt` and searches of both npm 0.32.1 and 0.32.2 canary package contents found no literal `0x3e31dcc2`. The canary eligibility implementation still checks the same six conditions as 0.32.1. The selector's meaning remains unidentified.


At 2026-10-28T08:08:00Z, fixed block `35217141` (`0x2195ef5`; block timestamp `1793174880`), MarginBank's `getCreditGranter()` returned `0x2ACD9EFdA721450906bfE1bef3edE269636E42C4`. The campaign window was `[1790553600, 1798477478)`, grant `1000000000000000000000`, and remaining budget `219234000000000000000000000`.

| Address | Active pools | Main-funded principal | Claimed | STT balance | SDK checks | `claim()` |
|---|---:|---:|---:|---:|---|---|
| `0x7262D22e91D38D9883275664D1dAEbca70a4CCe0` | 0 | 0 | false | 0 | eligible | `{"code":3,"message":"execution reverted","data":"0x3e31dcc2"}` |
| `0x000000000000000000000000000000000000dEaD` | 0 | 0 | false | 0 | eligible | same `0x3e31dcc2` |

All eligibility reads and claim calls used the same block tag. Zero STT balance does not itself prevent `eth_call`; it would matter for broadcasting, which was not attempted.

## Environment

2026-10-30 03:36–03:41 UTC; Somnia Tokyo testnet (Hideki, chain 50383); wallets: `0x7262D22e91D38D9883275664D1dAEbca70a4CCe0` and `0x000000000000000000000000000000000000dEaD` (plain `from` addresses); read-only `eth_call` and documentation GET. npm `latest` remains SDK 0.32.1; a newer prerelease is tagged `canary` at 0.32.2-canary.38575926931.ge4550933d36d. No transaction sent.
