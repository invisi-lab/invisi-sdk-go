# Credit plans and optional Broadcaster delivery

The create, draw and update examples retain their existing managed encrypted inventory, canonical discovery and external native wallet authorization. To choose a relay, supply its authenticated HTTPS registration and fee budget before authorization:

```sh
go run ./examples/04-credit/02-alice-batch-draw \
  --chain-rpc '<primary>' --protocol-rpc '<protocol RPC>' \
  --wallet-address '<public long address>' --recipient '<recipient long address>' \
  --token '<token address>' --operation-id '<new stable journal ID>' \
  --broadcaster '<registered HTTPS endpoint>' \
  --quote-signer '<registered signer>' --fee-recipient '<registered recipient>' \
  --provenance-policy '<authenticated policy ID>' --max-fee '1' \
  --execute --submit
```

Use the same relay flags for `01-bob-create` and `03-bob-update`. An omitted policy selects an ordinary relay; a configured policy requires matching policy-bound signed quotes. Credit state transitions have no private Note inputs, so the source policy preflight needs no source circuit, source prover or manual sidecar. The relay validates the ordinary public funding facts. Credit draw creates a normal Note that can subsequently enter the Note flow under 03.

`WithBroadcaster` converges the fee and persists the quote before the external native wallet approves the final plan and before local main proving. Creation/update pay the fee through the owner's ERC20 wallet; draw debits each grant gross and emits a Note with the fee subtracted. `PrepareDrawToNote` retains recipient receiving keys so the SDK can reencrypt the final net Note after repricing. Inspect the logged safe preview and arrange required ERC20 allowance independently.

`--execute` is required when relay flags are present. Without `--submit`, the quoted/proved managed operation is retained but no transaction is broadcast. With `--submit`, `operation.SubmitBroadcaster` records exact calldata and uncertainty before calling the selected relay. Keep the same encrypted journal and operation ID; a known externalized operation queries that original relay ID without repricing, reproving, signing or sending again. Expiry, spent state or source downtime must not prevent this status lookup. Canonical primary reconciliation remains necessary to release reservations. Do not rerun creation using a new ID to retry an uncertain payment.

Without relay flags, examples retain their explicit unsigned-calldata export. Revocation currently demonstrates this direct-owner route; relay revocation separately requires the owner EIP-712 signature. No account signing key is accepted by these examples. Native wallet request arrays use stdout and responses use stdin; relay diagnostics use stderr.
