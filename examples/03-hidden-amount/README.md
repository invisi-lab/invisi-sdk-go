# Normal prepared payment with Provenance

`07-via-broadcaster` uses the ordinary SDK flow: discover authenticated Notes through Protocol RPC and the external native wallet, prepare a private payment, call `WithBroadcaster`, authorize the final plan through the native wallet, prove the main transaction locally, then explicitly `SubmitBroadcaster`. The SDK obtains signed source checkpoints, checks their primary-chain anchors, rebuilds recognized trees locally, selects the source circuit capacity, proves locally and attaches public sidecars automatically. The selected Broadcaster evaluates the source policy and returns tracking with its quote; the SDK independently confirms that tracking through the pinned source client without evaluating again. The application supplies no prepared JSON, opening, main public-input array or sidecar.

Copy `broadcaster.config.example.json` outside the repository and replace the public deployment, wallet, recipient, token, base-unit amount and fee budget. Trust policy/provider/generation/signing-key pins must come from an authenticated operator channel. Production key rotation uses the verified history APIs in [PROVENANCE.md](../../PROVENANCE.md). `protocolRpcUrl` and `chainRpcUrl` must describe the same deployment.

Synchronize authenticated source artifacts with `python3 scripts/sync-provenance.py ../invisi-protocol`; `assetRoot` points at `.cache/provenance`. Set `NOIR_EXECUTOR` and `BB_BINARY` to trusted local Noir/Barretenberg executables and provision the local SRS. Main and source proofs run locally and are independently verified; no remote proving URL is accepted.

```sh
export PROVENANCE_CONFIG='/protected/provenance-config.json'
export PROVENANCE_RECORD='/protected/new-payment-public.json'
export NOIR_EXECUTOR='/trusted/bin/noir-execute'
export BB_BINARY='/trusted/bin/bb'
go run ./examples/03-hidden-amount/07-via-broadcaster --execute --submit
```

The host reads wallet request arrays from stdout and writes the matching response arrays to stdin, using the existing JSON boundary from examples 03/05. Discovery/decryption and final authorization use that boundary; no wallet key enters the SDK process. Diagnostics go to stderr. Without `--execute` the program does no network/proving work; without `--submit` it does not broadcast.

The fee is finalized before source proving and authorization. The example selects private funds for the payment plus maximum fee and returns excess as change. Input selection is local; source membership is found in locally rebuilt trees, with no per-Note source API lookup. Source replay is bounded to ten root-directory pages and 100,000 leaves. Missing capability, unknown membership, noncanonical anchors, denied/stale tracking or budget exhaustion fails closed and retains the original prepared operation. Fee changes require explicit preparation again; submit never reprices, reproves or resigns.

The exclusively created, fsynced `PROVENANCE_RECORD` contains only final public calldata, the stable SDK broadcast ID and the private lookup capability `trackingId` with its digest. Keep it private. Existing files abort rather than overwrite an uncertain payment. This flat explicit example is not a production wallet journal: a managed `session.Connect` with encrypted storage persists the complete public sidecar/quote/tracking and reservations using the same transaction engine. Supply the same runtime Provenance capability when resuming; it is never serialized. After timeout query the same Broadcaster with the retained broadcast ID and reconcile the original transaction on primary; never create another payment to retry.

After actual confirmation, set `TRANSACTION_HASH` and run:

```sh
go run ./examples/03-hidden-amount/08-confirm-source --execute
```

The second program verifies the exact original calldata, target, receipt and canonical block through primary RPC before waiting on the stored tracking ID. It requires the matched locator/digest and a published recognized operation. Normal indexing adopts the verified preflight evidence automatically; no repeated `submitProof` is needed. Explicit low-level self-broadcast recovery remains available through `RecoveryCoordinator.PrepareDirect`, independently of this normal prepared path. Never silently switch modes after an uncertain broadcast.


Relay use is optional. Ordinary direct Note preparation/export remains independent of source submission. An application choosing direct main delivery can explicitly evaluate source evidence before broadcast and retain its tracking ID, or broadcast/confirm the main transaction first and explicitly submit source proof using the canonical locator. These are separate SDK steps, with order chosen by the application. All source verification and recognition is offchain; no source proof is added to the onchain main request.
