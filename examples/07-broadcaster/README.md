# Broadcaster and independent Source delivery

These flat walkthroughs use managed Note transfers. Broadcaster flows use `relay.Quote(ctx, operation, options)`, `relay.Submit(ctx, operation)` and `relay.Cancel(ctx, operation)` on the selected client; the same API accepts detached Note/Credit preparations and managed Session handles. Configure public deployment, wallet, recipient and token through flags. Replace the App decryptor placeholder before discovery. EIP-712 authorization uses ordinary geth signing; `OWNER_PRIVATE_KEY` must match the Note owner.

| Example | Flow |
| --- | --- |
| `01-prepare-and-broadcast` | Discover registered relay, freeze the final fee quote, authorize and prove the payment. On explicit submit, the SDK generates required Source evidence and the relay submits it before broadcasting. |
| `02-pre-submit-provenance` | Prove the main payment, independently prove and submit Source evidence, verify its allowed tracking acknowledgement, then optionally send the unchanged payment with a geth signer. |
| `03-recover-and-submit-provenance` | Persist an ordinary main payment in the encrypted journal, send it, restart without an account key, reconcile canonical confirmation and generate/submit Source evidence. Optionally export or submit a standalone public proof packet. |
| `04-scheduled-cancel` | Prepare and prove a scheduled Note transfer, submit, resume the encrypted operation by its local ID, query the returned Broadcaster ID, then cancel and reconcile canonical input state. |

Source defaults to scoped official HTTPS bootstrap at `https://provenance.invisipay.fi`: chain/protocol, release manifest, generation and signed key-rotation history are checked, then identity is pinned for the client instance. No policy selector exists. Private deployments retain explicit authenticated trust: set `PROVENANCE_ENDPOINT`, `PROVENANCE_GENERATION`, `PROVENANCE_PROVIDER_ID`, `PROVENANCE_SIGNING_KEY_ID` and `PROVENANCE_SIGNING_PUBLIC_KEY` together. An endpoint alone never establishes private trust.

The native prover automatically composes local Source proving with all seven authenticated packaged circuits. Set `NOIR_EXECUTOR` and `BB_BINARY` to trusted local executables. An explicit `PROVENANCE_ASSET_ROOT` selects an authenticated local asset override. Main and Source proofs run locally and are independently verified. Source replay is bounded to ten directory pages and 100,000 leaves.

Example 01 accepts `--max-fee` in token base units. All three examples accept `--finality-timeout` (default two minutes); native finalized confirmation is awaited with an independent bounded deadline. Source tracking has its own bounded wait and never authorizes resubmission on timeout.

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

Example 02 calls `prepared.EstimateGas()` before authorization and proof. All valid direct shapes return positive gas with `Source: "benchmark"`; `Statistic` distinguishes measured medians from calibrated estimates. Recursive aggregate outer previews remain unavailable. This local operation does not export transactions or write the journal. Application broadcasters own final simulation.

Receipt success is inclusion, not finality. Observe native confirmation through the managed operation before treating payment as finalized. Provenance publishes recognized outputs only after primary native finalized confirmation; node outages do not trigger resubmission.


Example 04 runs its real submit-and-cancel flow directly, without execution or submission switches. Configure recognized spendable Notes covering `--amount` plus the fee limit, `OWNER_PRIVATE_KEY`, the App decryptor, local prover, a recipient receiving address and `INVISI_EXAMPLE_STORAGE_KEY`. The example submits, resumes the saved operation, checks status and cancels immediately, with a ten-minute default publication delay. If cancellation fails or the process stops, retain the printed operation ID and encrypted journal for recovery. It never prints the cancellation secret.

```sh
export OWNER_PRIVATE_KEY='<Note owner EVM key>'
export BB_BINARY='/trusted/bin/bb'
export NOIR_EXECUTOR='/trusted/bin/noir-execute'
export INVISI_EXAMPLE_STORAGE_KEY='<retained 32-byte hex storage key>'
go run ./examples/07-broadcaster/04-scheduled-cancel \
  --chain-rpc '<Chain RPC>' --protocol-rpc '<Protocol RPC>' \
  --wallet-address '<public receiving address>' --recipient '<recipient>' \
  --token '<token address>' --amount 10 --max-fee 1
```

The flow authenticates the chain/protocol and registry endpoint, reads live scheduling capabilities, and covers the execution window plus minimum slack in the Note authorization deadline. The SDK stores the random secret, signed quote and frozen request before sending. The submission ID is opaque and allocated by the server. A lost response is recovered by explicitly replaying the same frozen submission before status or cancellation calls. An acknowledged cancellation is durably recorded before canonical refresh, so a refresh outage preserves input reservations and can be reconciled after reconnecting without another relay response. Source receives its ordinary evaluations and no scheduling fields.
