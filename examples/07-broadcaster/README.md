# Broadcaster and independent Source delivery

These flat walkthroughs use the same managed transfer path. Configure public deployment, wallet, recipient and token through flags. Replace the App decryptor placeholder before discovery. EIP-712 authorization uses ordinary geth signing; `OWNER_PRIVATE_KEY` must match the Note owner.

| Example | Flow |
| --- | --- |
| `01-prepare-and-broadcast` | Discover registered relay, freeze the final fee quote, authorize and prove the payment. On explicit submit, the SDK generates required Source evidence and the relay submits it before broadcasting. |
| `02-pre-submit-provenance` | Prove the main payment, independently prove and submit Source evidence, verify its allowed tracking acknowledgement, then optionally send the unchanged payment with a geth signer. |
| `03-recover-and-submit-provenance` | Persist an ordinary main payment in the encrypted journal, send it, restart without an account key, reconcile canonical confirmation and generate/submit Source evidence. Optionally export or submit a standalone public proof packet. |

Source uses `https://provenance.invisipay.fi`. Its trust context contains chain/protocol, generation, provider and signing keys; no policy selector exists. Obtain the generation/provider/signing-key pins through an authenticated operator channel, then set `PROVENANCE_GENERATION`, `PROVENANCE_PROVIDER_ID`, `PROVENANCE_SIGNING_KEY_ID` (32-byte hex) and `PROVENANCE_SIGNING_PUBLIC_KEY` (raw Ed25519 hex). Production key rotation uses the verified history APIs in [PROVENANCE.md](../../PROVENANCE.md).

Synchronize authenticated artifacts with `python3 scripts/sync-provenance.py ../invisi-protocol`. Set `PROVENANCE_ASSET_ROOT` if different from `.cache/provenance`, and set `NOIR_EXECUTOR` and `BB_BINARY` to trusted local executables. Main and Source proofs run locally and are independently verified. Source replay is bounded to ten directory pages and 100,000 leaves.

```sh
export OWNER_PRIVATE_KEY='<Note owner EVM key>'
export BB_BINARY='/trusted/bin/bb'
export NOIR_EXECUTOR='/trusted/bin/noir-execute'
go run ./examples/07-broadcaster/01-prepare-and-broadcast \
  --wallet-address '<public receiving address>' --recipient '<recipient>' \
  --token '<token address>' --execute --submit
```

`--execute` permits discovery/preparation/proving. Example 01 sends only with `--submit`. Example 02 submits Source evidence during execution and sends the payment only with `--submit`. Source submission accepts the frozen transaction before or after payment confirmation and needs no transaction hash; tracking matches the confirmed operation automatically. A timeout does not authorize a replacement payment.

Example 03 uses `examples/wallet.sqlite` and the retained 32-byte hex `INVISI_EXAMPLE_STORAGE_KEY`. First run `--execute --send` with the public flags; retain its operation ID. Restart with `--execute --resume --operation-id '<original ID>'` and the same public wallet, deployment, journal and storage key. An account signing key is unnecessary on resume. Add `--reprove` to regenerate Source evidence under updated authenticated generation/signing keys or roots while retaining the exact main payment. If a crash occurred after sending but before recording the hash, supply `--transaction-hash '<original hash>'`. Resume reconciles the original call; it never sends payment again.

Add `--export-proof '<new private file path>'` during recovery to save the public `invisi-provenance-submission-v1` packet with exclusive creation. Another process can use `--execute --proof '<saved packet>'` with authenticated deployment/Source pins, without the journal, account key, decryptor or prover assets. `sdk.SubmitProvenanceProof` decodes exact calldata, reconstructs the native statement, verifies the operation digest and proof bindings before submission. A packet contains only public evidence; retain encrypted openings separately if future reproof is needed.

Example 02 calls `ready.EstimateGas(ctx, actualSender)` on the final main calldata before generating Source evidence. The result has `Source: "rpc"`; it is separate from a relay quote, `maxGas` ceiling and documented benchmark reference. Managed estimation uses the existing transaction export fence. Repeated `ProveProvenance(ctx)` reuses locally validated cached evidence with only a trusted Source client; `--reprove` explicitly queries and proves again.

The quote signs `execution` as its seventh EIP-712 field. Example 01 uses `Prove(ctx)` so the signed quote selects direct or recursive proving automatically. Aggregate relay submission is restricted to unfunded private Note C1..C10 transfers and sends public recursive `memberProof` with its single-operation envelope. Do not export this envelope as an EVM transaction. Source evidence binds the same bytes and is cached with the durable request. Shield and Credit use direct execution. After a lost response, replay the exact retained request; after obtaining an ID, use status lookup only, even when the service reports not found. Queued/proving/rejected/expired states do not release reservations without canonical reconciliation.
