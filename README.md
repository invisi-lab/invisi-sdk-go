# Invisi Go SDK

Go client for the current Invisi Protocol. Requires Go 1.26.1 or newer.

```sh
go get github.com/invisi-lab/invisi-sdk-go@v1.2.0
```

Release target: **1.2.0**. Module versions are repository tags; preparing this source does not create `v1.2.0` or publish a release. The `go 1.26.1` directive remains the toolchain requirement. The SDK remains in the v1 series. Source synchronization does not publish a new module version or upgrade deployed contracts.

## Six independently versioned modules

All six modules target **1.2.0**. Applications normally import the SDK's `sdk` package; backend services can depend only on the pure Protocol module and its Primitives dependency.

| Module | Responsibility | Internal dependencies |
| --- | --- | --- |
| root SDK | Stateless Note/Credit and transaction orchestration; `sdk`, `credit`, `wallet` and public Provenance facade | primitives, protocol, prover, rpc |
| `primitives` | Fields, hashes, key derivation, commitments, encryption and delivery codecs | none |
| `protocol` | Public ABI tuples, strict settlement codec, statements/digests, circuit identities, gas references and public Source contracts | primitives |
| `prover` | Private witnesses, authenticated proving assets, native proof generation and verification | primitives, protocol |
| `rpc` | Protocol RPC, Broadcaster and Provenance clients, deployment checks and generated network contract bindings | primitives, protocol |
| `session` | Note/Credit inventory, reservations, journals, encrypted storage and operation recovery | SDK and its public protocol/prover/client types |

Module paths use `github.com/invisi-lab/invisi-sdk-go/<name>`; the SDK uses the repository-root module path. Nested module release tags are `<name>/v1.2.0`; the root SDK tag is `v1.2.0`. The examples are a separate development module, not a public release package. No `core`, `runtime` or `protocolkit` compatibility module is retained.

Pure ABI/types are in `protocol/contracts`; network-bound geth clients are in `rpc/contracts`. Maintainers regenerate both from the same compiler ABIs with `go generate ./contracts` in the Protocol module. `session/credit` owns managed Credit state. `protocol/gas` owns measured references, while final `eth_estimateGas` calls belong to RPC and deployment/operator pricing remains outside the pure module.

For local multi-module development, use an untracked Go workspace. `GOWORK=off` builds use tagged module dependencies; production `go.mod` files contain no local replacement paths. Verify module boundaries with `python3 scripts/check-module-boundaries.py` in the configured workspace.

## Wallets and infrastructure

`sdk.Wallet` stores public identity and receiving material only. Initialize it with `sdk.WalletFromAddress(receivingAddress)`. The App owns account selection, key storage, unlocking, signing and decryption. No native transport or key-holding wallet is installed by the SDK.

Inject `sdk.DeliveryDecryptor` with `DecryptNote(ctx, ciphertext, context)` and `DecryptCredit(ctx, ciphertext, context, role)`. Return plaintext bytes; only `nil, nil` means nonrecipient. Locked accounts, cancellation and other operational failures return errors. Wallet decryption methods validate financial contents against the owner, token, deployment, output commitment and Credit receiving keys. The App authenticates ciphertext, including its AEAD tag; financial commitments do not independently authenticate opaque memo bytes.

```go
pending, err := managed.Refresh(ctx, options)
if err != nil { return err }
_, err = pending.Decrypt(ctx, appDecryptor)
if err != nil { return err }
request, err := prepared.ExportWalletAuthorization(ctx)
if err != nil { return err }
typed, err := request.TypedData()
if err != nil { return err }
digest, _, err := apitypes.TypedDataAndHash(typed)
if err != nil { return err }
signature, err := crypto.Sign(digest, signerKey) // Application-owned geth key.
if err != nil { return err }
prepared, err = prepared.WithWalletSignature(ctx, signature)
```

Optional low-level `wallet` wire codecs validate application-defined integrations. Normal examples and Session refresh do not use wallet request/response arrays or retain an executor. See [examples](examples/README.md) for explicit App callbacks and ordinary geth signatures.

Spending authorization, public ERC20 funding, allowance/Permit, detection, viewing and gas submission are separate roles. A full Note viewing capability is not purely read-only: it also permits consumption tracking and a full owner-bound sweep.

## Ordinary Note payments

Every business output is an immutable `NoteOutput` created with `sdk.NewNoteOutput`. The factory accepts `PaymentRequest` or an IPR string and binds chain, Protocol, token, amount and recipient before preparation.

Recipients are typed variants:

- `InvisiRecipient`: complete receiving address; delivery can be encrypted offline.
- `EVMRecipient`: EVM owner; envelope lookup uses the selected Protocol RPC or Chain RPC registry. Explicitly disabling the envelope permits an unregistered destination.
- `IdentityRecipient`: canonical per-note identity, with no recoverable owner/nonce opening or envelope.

`sdk.Invisi.PrepareShield` creates transparent-amount deposits without a proof. `PrepareTransfer` combines explicit private inputs, optional public funding, private outputs and an optional `Withdrawal`. Every selected input is consumed. Automatic change uses a fresh Note.

The detached sequence is prepare → inspect → export EIP-712 authorization → attach geth signature → prove and independently verify → attach separate public funding signature if needed → export calldata → submit with ordinary geth APIs. Full owner-bound sweeps without funding, private outputs or fees need viewing capability; partial or redirected spending requires owner authorization.

Managed Note operations use `session.Connect`, reservations and durable journals. No Store is created or consulted by detached preparation.

## Postpaid Credit

`credit.State` is authority to debit an owner's wallet, not a prepaid token balance. It keeps owner/spender identities, token, the original revocation shard/slot, nonce, expiry, a grant-scoped shared key, four limits and five usage fields. It has no payout, refund, remainder balance or Note funding fields.

- `PrepareCreate` registers private authority without depositing principal or spending Notes.
- `PrepareDraw` consumes N grant states, groups sorted owner payer receipts, creates N successor states and one ordinary Note for any recipient identity.
- `PrepareDrawToNote` additionally freezes that Note opening and public receiving keys, so fee quotation can preserve owner/nonce/payload while rebuilding its net amount and envelope.
- `PrepareUpdate` changes policy/expiry while preserving usage and immutable grant identity/namespace.
- `PrepareRevoke` uses the original owner/token/shard/slot reference. It needs no latest merchant state and no ZK proof.

The shared Note Tree is a storage/membership primitive. A Credit state remains a distinct business model from an ordinary spendable Note.

Policy limits are single/day/month; zero disables only that check. State carries uint128 `AvailableCredit` and an independent `Unlimited` flag. Finite zero is exhausted; a finite draw subtracts gross debit. Unlimited grants keep zero balance and still accumulate daily/monthly/total usage. Draw and recovery use canonical chain time. State recovery authenticates the masked **gross debit**, deterministic next nonce and successor commitment; unknown child payload is not copied from the parent.

`credit.DeliveryState` retains both receiving-key snapshots. Current Credit envelopes use `0x82`; older financial models/envelopes are rejected. `Payload` is opaque authenticated application data outside commitments. `credit.MemoV1` is an optional memo codec independent of the envelope version.

### Authorization, fees and proving

Create/update/draw use v4 wallet authorization domains and direct local proofs. `ExportSigningRequest` returns portable JSON-safe EIP-712 data without `EIP712Domain` in application types; `WithWalletSignature` accepts standard or compact EOA signatures. `WithAuthority` is the explicit raw public-key/circuit-signature boundary.

Revoke uses its own fee-bound owner EIP-712 signature, including EOA/ERC-1271 support through `WithRevocationSignature`. A direct owner caller may omit the signature. Revocation remains available during protocol and circuit pauses and cannot claw back an already issued Note.

Every Credit request carries `FeeRecipient` and `FeeAmount`: either a complete zero pair or a nonzero recipient and positive uint128 fee. Create/update/revoke pull fees from the owner's ERC20 wallet. Draw debits and usage are gross; the ordinary Note receives `totalAmount - feeAmount`. Allowance and token balance remain separate from wallet authorization.

`Prove` creates a local witness, uses an authenticated statement, and independently verifies the resulting proof before `Calldata` becomes available. Recovery intents omit private wallet authorization and separate revoke signing material. Final public calldata, including a submitted revoke signature, is retained in the encrypted journal for exact reconciliation and idempotency. Explicit intent exports contain private grant data and belong only in authenticated encrypted storage.

## Paid broadcaster flow

Note operations use `*broadcaster.Client`. `EstimateQuote(ctx, data)` estimates fees and `GetQuote(ctx, data, trust)` freezes the signed public intent. `prepared.WithBroadcaster` finalizes fees before authorization and main proving; it does not build Source proofs. If the final quote requires evidence, `SubmitBroadcaster` generates and attaches Source proof from the final ready transaction. Broadcaster verifies/preflights it before sending. Configure optional `sdk.InvisiOptions.Provenance` with the trusted Source client, authenticated manifest and local offline prover. Managed sessions persist public attachments and reservations. See [Provenance](PROVENANCE.md) and [Broadcaster examples](examples/README.md).
The seventh EIP-712 Quote field, `execution`, is mandatory (`direct` or `aggregate`). After `WithBroadcaster`, `Prove(ctx)` follows that signed choice; an explicit conflicting execution is rejected. Aggregate delivery accepts unfunded private Note transfers with native capacity C1 through C10, sends a single-operation public recursive envelope plus `memberProof`, and cannot be exported as a direct EVM transaction. Shield and Credit remain direct. `broadcaster_getStatus.provenance.aggregateSupported` declares recursive delivery support.

All four Credit operations support:

1. Prepare the business request.
2. Select `prepared.WithBroadcaster(ctx, client, credit.WithBroadcasterOptions{MaxFeeAmount: limit})` before authorization/proving.
3. Inspect `Preview()`, including fee and ordinary Note net amount; prepare the required owner/payer ERC20 allowance.
4. Authorize and prove, or attach the revoke owner signature.
5. Call `SubmitBroadcaster` with the original client and persist the service-assigned submission ID.
6. Authenticate canonical receipts independently of the operator's status response.

Fee convergence is bounded and fails without modifying the original request on over-limit/nonconverging quotes. Grant gross debits and successor state commitments remain fixed during draw repricing. Enveloped draws need `PrepareDrawToNote` to reencrypt the final net Note safely.

Quotes bind chain, Protocol, token, registered signer, fee recipient/amount, deadline and canonical public intent. Only proofs and the operation's separate signature slots are stripped from the intent hash. Final calldata must match the quote. Submission IDs are opaque and assigned by the service. If the initial response is lost, retry the exact retained request; if an ID is known, query only that ID, including after a not-found response. Queued, proving, rejected and expired relay states retain reservations until canonical reconciliation. Do not create a conflicting spend.

## Managed grant state and recovery

`walletSession.Credit()` composes `session/credit` with the same encrypted WalletStore. Note and grant inventories/journals remain separate. Both roles of one physical grant reserve together. Current allocation, bitmap, consumed nullifiers, registered historical roots, UTC draw window and ERC20 gross/fee allowance are checked through the session's pinned provider.

The managed flow provides `Prepare`, explicit EIP-712 authorization, `Prove`, `Calldata`, paid broadcaster submission, `Cancel`, `Resume` and canonical `Reconcile`. Externalization and exact calldata are committed before submission; transport failures retain reservations. Resume replaces the old worker lease and independently verifies saved proofs. A canonical reverted transaction can release reservations; an unconfirmed operator result cannot.

Use `walletSession.CreditReconciler(confirmations)` to bind the actual transaction destination, chain and exact calldata to canonical receipts and successor events. `WalletSession.Refresh` stages Note/Credit discovery without writes. Inject App decryption through `pending.Decrypt(ctx, decryptor)`; only nil plaintext is a nonrecipient. Apply rechecks canonical receipts, membership and consumption, guards the captured store revision, and commits the complete inventory/checkpoint atomically. App failures and invalid plaintext never advance the real store. Reconciliation reuses authenticated stored openings; refresh unseen deliveries before reconciling them. Reorganizations fence old workers without discarding externalized journals.

`GetHistory` on each component returns business summaries without proof witnesses or signing secrets. Credit summaries omit grant recovery keys and nonces. Top-level `Subscribe` notifies after successful Store commits, serially outside operation locks. The application owns refresh scheduling and cross-process notifications.

## Typed Protocol RPC and receiving

Note methods are `GetNote`, `GetNoteUpdates`, `GetTransactionNoteUpdates` and `IsNoteSpent`. Credit methods are `GetCreditState`, `GetCreditTransitions`, `GetTransactionCreditUpdates`, `IsCreditConsumed`, `GetCreditAllocation` and `GetCreditRevocationBitmap`.

One indexed CreditState is one physical leaf with two opaque role deliveries. Credit records never pass through the Note APIs. Public grant references/bitmap are distinct from private spender, terms and usage. Index availability (`not_found`, `pending_index`, `indexed`) is separate from chain receipt success/reversion.

RPC candidates are authenticated against canonical receipts, insertion positions, membership and chain state before becoming spendable inventory. FMD is a filter, not proof of completeness or payment. Discovery coverage is separate from blocks merely observed during reconciliation. A reorganization needs an explicit reset refresh.

`credit.Disclose` creates a grant-scoped audit disclosure without wallet signing keys or recipient Note openings. Its explicit `Export` is sensitive; `Verify` requires an independently accepted commitment anchor.

## IPR and IPN

`sdk.EncodePaymentRequest` / `DecodePaymentRequest` use the cross-language `ipr:` wire format for ordinary Note output intent. `NewNoteOutput` freezes execution-specific nonce/lookup choices without rewriting the request's deployment context.

`TransactionOutput.ExportPaymentNote` and `NoteOutput.ExportPaymentNote` export one definite ordinary Note opening as `ipn:`. `PaymentNoteFromIPN` recomputes identity/commitment offline. `Verify` checks canonical confirmed membership; `VerifySpendable` also checks the actual owner's spending capability and current nullifier status, then provides `ToInput`.

IPN contains no viewing/spending key, ciphertext, payload or membership path. Export is not a claim of payment. Identity-only outputs and Credit states cannot export IPN. Deliver only the recipient's final output after confirmation, never the payer's change.

## Gas estimates and references

`ready.EstimateGas(ctx, from)` on `sdk.PreparedTransaction` or `session.Operation` calls the configured provider's `eth_estimateGas` with the exported final destination, calldata, zero value and actual caller. It returns `GasEstimate{Source:"rpc", GasUnits:...}` and requires a ready EVM transaction. The configured provider must additionally implement `protocol.GasEstimator` (geth `ethclient.Client` does); the base `sdk.Provider` contract is unchanged. Managed estimation uses `ExportTransaction`, including its durable externalization fence.

For all four Credit operations (create, draw, update and revoke), use `prepared.EstimateGas(ctx, provider, from)` or `managedCreditOperation.EstimateGas(ctx, provider, from)`. These use `Calldata(from)` so proof requirements and revocation caller authorization remain enforced. Managed Credit estimation journals the final calldata through the existing ready transition. Estimation does not send a transaction; provider errors are returned unchanged.

```go
estimate, err := ready.EstimateGas(ctx, actualSender)
if err != nil { return err }
log.Printf("final RPC gas estimate: %d", estimate.GasUnits)
```

RPC estimates reflect current chain execution. Broadcaster quotes price a relay service, quote `maxGas` is a ceiling, and the catalog is a measured reference; these values serve different purposes.

`gas.Reference(shape)` matches a complete documented benchmark shape and returns a detached result with source hash, docs commit, conditions and measured median. Unsupported shapes return unavailable; there is no arbitrary linear fallback or current aggregate gas claim based on historical private-pin experiments.

Credit measurements must match the packaged manifest/format identity, raw Credit manifest SHA256 (`creditManifestSha256`) and fee mechanism. Rows without the exact manifest hash are unavailable. The current catalog contains 58 measured Credit references for the available-quota model and 55 Note references. Historical rows remain identifiable; they cannot impersonate a new release. Use Chain RPC `eth_estimateGas` on the final calldata and real caller for actual submission.

## Development and validation

```sh
go test ./...
go test -race ./...
go vet ./...
go list ./...
go mod verify
python3 scripts/test-sync-reference.py
```

Native integration tests additionally use `BB_BINARY_PATH` and `NOIR_EXECUTE_PATH`. The backend checks exact Noir/BB versions and authenticated artifact/VK identity, does not automatically download tools and never uploads witnesses.

Maintainers regenerate bindings with `cd protocol && INVISI_PROTOCOL_ROOT=/path/to/invisi-protocol go generate ./contracts`. `python3 scripts/sync-reference.py --proving-only` authenticates the current 28 daily Note circuits, 34 aggregate outers and 15 direct Credit routes, with recursive certificates for C1..C10. User manifest4/format7, aggregate receipts10 and Credit manifest7/format8 have separate identities.

`State.Update(Change, deadline)` applies mutually exclusive uint128 increase/decrease deltas to the authenticated input balance. `Change.Unlimited == nil` preserves the mode; entering unlimited clears the balance and requires zero deltas, and leaving unlimited starts at zero plus this increase. Update preserves historical usage. Owner-pulled create/update/revoke fees do not consume available credit. `PrepareUpdate` freezes the original change; quotes never apply the delta again. An explicit new preparation applies that change to the latest accepted state; an externalized journal always resumes its original calldata.

Current configurable examples live under [examples](examples/README.md). Real native proof tests establish cryptographic interoperability; mocked operator/provider tests establish stage and journal behavior and do not claim real token accounting.

## Optional Provenance tracking

The independent [Provenance module](PROVENANCE.md) keeps source recognition separate from protocol spendability. Relay use is optional and limited to operator-supported tokens. Source evidence is generated from the final ready statement; Broadcaster verifies and preflights it before sending. Direct payment submission and `SubmitProvenance` are independent App-controlled steps, in either order. The service automatically matches confirmed payments and publishes recognized outputs.

`ExportProvenanceProof` returns a public recovery packet without private Note openings. Persist it before sending; a fresh client can call `sdk.SubmitProvenanceProof` without the original wallet, prover or transaction hash. Repeated `ProveProvenance(ctx)` validates and reuses cached public evidence locally using the trusted Source client; it needs no Chain RPC, Source RPC, prover or manifest assets. If trust or chain generation changes, use `ProveProvenance(ctx, sdk.ProvenanceProveOptions{Reprove:true})` with the retained encrypted operation. Reproof preserves original payment calldata and the main proof. See [examples](examples/README.md).

## Module checks

`python3 scripts/test-modules.py` creates a temporary workspace and runs boundary checks, race tests, vet and builds for all six modules plus the private examples. `python3 scripts/check-packages.py` constructs isolated local-proxy archives and tests them without the workspace or sibling checkouts; unpublished versions never enter the global module cache. GitHub CI runs both commands. Native Noir/BB integration tests are opt-in and require explicit local tool paths.

The modules are prepared at 1.2.0; no remote tags are created by these checks. Before standalone consumer builds can resolve the new version from GitHub, publish the required module tags in dependency order and refresh dependency checksums.
