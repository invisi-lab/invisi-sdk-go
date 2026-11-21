# Invisi Go SDK

Go client for the current Invisi Protocol. Requires Go 1.26.1 or newer.

```sh
go get github.com/invisi-labs/invisi-sdk-go@v1.3.3
```

Release target: **1.3.3**. Module versions are repository tags; preparing this source does not create `v1.3.3` or publish a release. The `go 1.26.1` directive remains the toolchain requirement. The SDK remains in the v1 series. Source synchronization does not publish a new module version or upgrade deployed contracts.

## Core contract bindings

Contract bindings follow the current `invisi-core` compiler artifacts: `InvisiCore` / `IInvisiCore` and `InvisiCoreFactory` / `IInvisiCoreFactory`. Update direct binding imports and generated tuple type names accordingly. Generic Protocol objects, package names, `protocolAddress` / `protocol_address`, deployment catalog records, EIP-712 domains and wire encodings retain their existing meanings.

## Eight independently versioned modules

All eight modules target **1.3.3**. Applications normally import the SDK's `sdk` package; backend services can depend only on the pure Protocol module and its Primitives dependency.

| Module | Responsibility | Internal dependencies |
| --- | --- | --- |
| root SDK | Stateless Note/Credit and transaction orchestration; `sdk`, `credit`, `wallet` and public Provenance facade | primitives, protocol, prover, rpc |
| `primitives` | Fields, hashes, key derivation, commitments, encryption and delivery codecs | none |
| `protocol` | Public ABI tuples, strict settlement codec, statements/digests, circuit identities, gas references and public Source contracts | primitives |
| `prover` | Private witnesses, authenticated proving assets, native proof generation and verification | primitives, protocol |
| `rpc` | Protocol RPC, Broadcaster and Provenance clients, deployment checks and generated network contract bindings | primitives, protocol |
| `session` | Note/Credit inventory, reservations, journals, encrypted storage and operation recovery | SDK and its public protocol/prover/client types |
| `paymentrouter` | Optional four-mode native/ERC20 PaymentRouter, V3/0x routing, authorization and relay envelopes | SDK; protocol; geth |
| `ipp` | Independent IPP protocol types, HTTP client and merchant helpers; Invisi transaction preparation through `ipp/invisi` | primitives; SDK for the Invisi adapter |

Module paths use `github.com/invisi-labs/invisi-sdk-go/<name>`; the SDK uses the repository-root module path. Nested module release tags are `<name>/v1.3.3`; the root SDK tag is `v1.3.3`. The examples are a separate development module, not a public release package. No `core`, `runtime` or `protocolkit` compatibility module is retained.

Pure ABI/types are in `protocol/contracts`; network-bound geth clients are in `rpc/contracts`. Maintainers regenerate both from the same compiler ABIs with `go generate ./contracts` in the Protocol module. `session/credit` owns managed Credit state. `protocol/gas` owns receipt-calibrated local estimates, while final transaction simulation belongs to the application and deployment/operator pricing remains outside the pure module.

For local multi-module development, use an untracked Go workspace. `GOWORK=off` builds use tagged module dependencies; production `go.mod` files contain no local replacement paths. Verify module boundaries with `python3 scripts/check-module-boundaries.py` in the configured workspace.

## Wallets and infrastructure

`sdk.Wallet` stores only an owner address and complete receiving keys. `GetAddress()` encodes a complete `0i` address on each call and errors when receiving keys are absent. `WalletFromAddress` accepts only complete `0i` addresses. Use independent EVM owner addresses with `wallet.Prepare` for identity-only requests, and `OwnerAddress()` to read an existing Wallet owner. Construction performs no RPC or registry lookup. Initialize it with `sdk.WalletFromAddress(receivingAddress)`. The App owns account selection, key storage, unlocking, signing and decryption. No native transport or key-holding wallet is installed by the SDK.

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

`credit.State` is authority to debit an owner's wallet, not a prepaid token balance. It keeps owner/spender identities, token, a positive uint128 grant ID, nonce, expiry, a grant-scoped shared key, three limits, available credit, unlimited mode and five usage fields. It has no payout, refund, remainder balance or Note funding fields.

- `PrepareCreate` registers private authority without depositing principal or spending Notes.
- `PrepareDraw` consumes N grant states, groups sorted owner payer receipts, creates N successor states and one ordinary Note for any recipient identity.
- `PrepareDrawToNote` additionally freezes that Note opening and public receiving keys, so fee quotation can preserve owner/nonce/payload while rebuilding its net amount and envelope.
- `PrepareUpdate` changes policy/expiry while preserving usage and immutable grant identity/namespace.
- `PrepareRevoke` uses the original owner/token/grant ID reference. It needs no latest merchant state and no ZK proof.

The shared Note Tree is a storage/membership primitive. A Credit state remains a distinct business model from an ordinary spendable Note.

Policy limits are single/day/month; zero disables only that check. State carries uint128 `AvailableCredit` and an independent `Unlimited` flag. Finite zero is exhausted; a finite draw subtracts gross debit. Unlimited grants keep zero balance and still accumulate daily/monthly/total usage. Draw and recovery use canonical chain time. State recovery authenticates the masked **gross debit**, deterministic next nonce and successor commitment; unknown child payload is not copied from the parent.

Initial-payment creation and draws freeze their UTC day and month into the authorization and proof. Before a new submission the SDK reads the latest canonical chain timestamp. A changed period returns `credit.ErrCreditPeriodChanged` (use `errors.Is`), with machine code `CREDIT_PERIOD_CHANGED`: prepare again, reauthorize, and regenerate the proof. The business deadline is independent and cannot extend the proof into another usage period. Scheduled execution must remain in the frozen UTC day; the exclusive `ExecuteBefore` may equal the next midnight. Already externalized or uncertain submissions retain their exact replay and reconciliation path. SDK-created detached preparations use the configured live provider automatically. Low-level or restored detached preparations require `prepared.SetChainTimestampReader` before first broadcaster submission; standalone `session/credit` callers configure `Options.ChainTimestamp`. These process-local readers are never serialized, and unavailable chain time fails closed for submission. Cached local authorization and proving retain their frozen preparation period without network reads; cached construction may use the later of local time and its canonical anchor, so that candidate period is not evidence of live execution time. Fresh canonical chain time is mandatory before submission. Broadcaster status preserves `ReasonCode`, `PublishAfter`, and `ExecuteBefore`; a queued payment rejected on rollover reports `ReasonCode == "CREDIT_PERIOD_CHANGED"`.


`CreatePreparation.GrantID` accepts an explicit `*big.Int` in `1..2^128-1`; otherwise creation selects one with cryptographic randomness before encryption or proving. The ID and shared key stay frozen through quotes, retries and snapshot fallback. Grant identity is scoped by chain, Core, owner and token, and survives update, draw and recovery unchanged. `CreditStatus` returns 0 (unknown), 1 (active), or 2 (permanently disabled). Create requires 0; update/draw require 1; revoke accepts 0 or 1 and needs no private grant. Managed submission rechecks canonical status. Offline creation requires an exact captured status-0 observation for the preselected ID; a snapshot cannot establish absence for a newly generated ID.

`credit.DeliveryState` retains both receiving-key snapshots. Current Credit envelopes use `0x84` (minimum 491 bytes); older financial models/envelopes are rejected. `Payload` is opaque authenticated application data outside commitments. `credit.MemoV1` is an optional memo codec independent of the envelope version.

### Authorization, fees and proving

Create/update/draw use v4 wallet authorization domains and direct local proofs. `ExportSigningRequest` returns portable JSON-safe EIP-712 data without `EIP712Domain` in application types; `WithWalletSignature` accepts standard or compact EOA signatures. `WithAuthority` is the explicit raw public-key/circuit-signature boundary.

Revoke uses its own fee-bound owner EIP-712 signature, including EOA/ERC-1271 support through `WithRevocationSignature`. A direct owner caller may omit the signature. Revocation remains available during protocol and circuit pauses and cannot claw back an already issued Note.

Every Credit request carries `FeeRecipient` and `FeeAmount`: either a complete zero pair or a nonzero recipient and positive uint128 fee. Create/update/revoke pull fees from the owner's ERC20 wallet. Draw debits and usage are gross; the ordinary Note receives `totalAmount - feeAmount`. Allowance and token balance remain separate from wallet authorization.

`Prove` creates a local witness, uses an authenticated statement, and independently verifies the resulting proof before `Calldata` becomes available. Recovery intents omit private wallet authorization and separate revoke signing material. Final public calldata, including a submitted revoke signature, is retained in the encrypted journal for exact reconciliation and idempotency. Explicit intent exports contain private grant data and belong only in authenticated encrypted storage.

## Paid broadcaster flow

Ordinary Core Note and Credit operations, including managed Session handles, use `*broadcaster.Client`. The same client owns `Quote(ctx, target, options)`, `Submit(ctx, target)` and `Cancel(ctx, target)` for detached Prepared and managed Session operations. Successful managed cancellation means the remote acknowledgement and canonical reconciliation have been persisted and reservations released; a pending acknowledgement returns an error. Raw RPC callers use `SubmitTransaction(ctx, data, quote)` and `CancelTransaction(ctx, id, token)` without SDK lifecycle or inventory dependencies. `client.Quote(ctx, operation, broadcaster.QuoteOptions{MaxFeeAmount: limit})` atomically installs the quote on the existing unsigned handle; it returns an error rather than a replacement handle. Failed quotation preserves the original request and reservations. Do not copy prepared/session handles by value: copied quote capabilities are rejected. `EstimateQuote(ctx, data)` estimates fees and `GetQuote(ctx, data, trust)` freezes the signed public intent. `client.Quote(ctx, prepared, options)` finalizes fees before authorization and main proving; it does not build Source proofs. If the final quote requires evidence, `client.Submit(ctx, operation)` generates and attaches Source proof from the final ready transaction. Broadcaster verifies/preflights it before sending. Configure optional `sdk.InvisiOptions.Provenance` with the trusted Source client, authenticated manifest and local offline prover. Managed sessions persist public attachments and reservations. See the [Broadcaster examples](examples/07-broadcaster/README.md).
The seventh EIP-712 Quote field, `execution`, is mandatory (`direct` or `aggregate`). After `client.Quote`, `Prove(ctx)` follows that signed choice; an explicit conflicting execution is rejected. Aggregate delivery accepts unfunded private Note transfers with native capacity C1 through C10, sends a single-operation public recursive envelope plus `memberProof`, and cannot be exported as a direct EVM transaction. Shield and Credit remain direct. `broadcaster_getStatus.provenance.aggregateSupported` declares recursive delivery support.

All four Credit operations support:

1. Prepare the business request.
2. Select `client.Quote(ctx, prepared, credit.BroadcasterQuoteOptions{MaxFeeAmount: limit})` before authorization/proving.
3. Inspect `Preview()`, including fee and ordinary Note net amount; prepare the required owner/payer ERC20 allowance.
4. Authorize and prove, or attach the revoke owner signature.
5. Call `client.Submit(ctx, operation)` with the original client and persist the service-assigned submission ID.
6. Authenticate canonical receipts independently of the operator's status response.

Fee convergence is bounded and fails without modifying the original request on over-limit/nonconverging quotes. Grant gross debits and successor state commitments remain fixed during draw repricing. Enveloped draws need `PrepareDrawToNote` to reencrypt the final net Note safely.

Quotes bind chain, Protocol, token, registered signer, fee recipient/amount, deadline and canonical public intent. Only proofs and the operation's separate signature slots are stripped from the intent hash. Final calldata must match the quote. Submission IDs are opaque and assigned by the service. If the initial response is lost, retry the exact retained request; if an ID is known, query only that ID, including after a not-found response. Queued, proving, rejected and expired relay states retain reservations until canonical reconciliation. Do not create a conflicting spend.

## Managed grant state and recovery

`walletSession.Credit()` composes `session/credit` with the same encrypted WalletStore. Note and grant inventories/journals remain separate. Both roles of one physical grant reserve together. Current scoped grant status, consumed nullifiers, registered historical roots, UTC draw window and ERC20 gross/fee allowance are checked through the session's pinned provider.

The managed flow provides `Prepare`, explicit EIP-712 authorization, `Prove`, `Calldata`, paid broadcaster submission, `Cancel`, `Resume` and canonical `Reconcile`. Externalization and exact calldata are committed before submission; transport failures retain reservations. Resume replaces the old worker lease and independently verifies saved proofs. A canonical reverted transaction can release reservations; an unconfirmed operator result cannot.

After `credit.ErrCreditPeriodChanged`, `credit.ErrCreditExecutionExpired`, or a persisted definitive first-attempt submission refusal, a known-unsent managed draw can call `operation.ReprepareDraw(ctx, credit.DrawPreparation{...})`. Refresh canonical inventory first and preserve the intended token, input debits and recipient. Repreparation reads fresh chain time and rebuilds the statement through the normal builder, keeping the same operation ID and reserved input set. A zero deadline asks the builder for a fresh chain-time deadline. The journal replacement clears the old quote, wallet authorization, proof and calldata atomically and fences old worker handles; quote, authorize and prove again before submission. A failed rebuild retains the reservations and prior journal. An explicitly exported Core call remains protected from quote-only abandonment: expiration of a Broadcaster quote does not invalidate the Core authorization. Externalized operations, transaction hashes or retained Broadcaster requests cannot be rebuilt and must resume their exact original submission for reconciliation. The protocol and API versions remain 1.

Use `walletSession.CreditReconciler("latest")` to bind the actual transaction destination, chain and exact calldata to canonical receipts and successor events. `WalletSession.Refresh` stages Note/Credit discovery without writes. Inject App decryption through `pending.Decrypt(ctx, decryptor)`; only nil plaintext is a nonrecipient. Apply rechecks canonical receipts, membership and consumption, guards the captured store revision, and commits the complete inventory/checkpoint atomically. App failures and invalid plaintext never advance the real store. Reconciliation reuses authenticated stored openings; refresh unseen deliveries before reconciling them. Reorganizations fence old workers without discarding externalized journals.

`GetHistory` on each component returns business summaries without proof witnesses or signing secrets. Credit summaries omit grant recovery keys and nonces. Top-level `Subscribe` notifies after successful Store commits, serially outside operation locks. The application owns refresh scheduling and cross-process notifications.

## Typed Protocol RPC and receiving

Note methods are `GetNote`, `GetNoteUpdates`, `GetTransactionNoteUpdates` and `IsNoteSpent`. Credit methods are `GetCreditState`, `GetCreditTransitions`, `GetTransactionCreditUpdates`, `IsCreditConsumed`, `GetCreditStatus`.

One indexed CreditState is one physical leaf with two opaque role deliveries. Credit records never pass through the Note APIs. Public grant identifiers/statuses are distinct from private spender, terms and usage. Index availability (`not_found`, `pending_index`, `indexed`) is separate from chain receipt success/reversion.

RPC candidates are authenticated against canonical receipts, insertion positions, membership and chain state before becoming spendable inventory. FMD is a filter, not proof of completeness or payment. Discovery coverage is separate from blocks merely observed during reconciliation. A reorganization needs an explicit reset refresh.

`credit.Disclose` creates a grant-scoped audit disclosure without wallet signing keys or recipient Note openings. Its explicit `Export` is sensitive; `Verify` requires an independently accepted commitment anchor.

## IPR and IPN

`sdk.EncodePaymentRequest` / `DecodePaymentRequest` use the cross-language `ipr:` wire format for ordinary Note output intent. `NewNoteOutput` freezes execution-specific nonce/lookup choices without rewriting the request's deployment context.

`TransactionOutput.ExportPaymentNote` and `NoteOutput.ExportPaymentNote` export one definite ordinary Note opening as `ipn:`. `PaymentNoteFromIPN` recomputes identity/commitment offline. `Verify` checks canonical confirmed membership; `VerifySpendable` also checks the actual owner's spending capability and current nullifier status, then provides `ToInput`.

IPN contains no viewing/spending key, ciphertext, payload or membership path. Export is not a claim of payment. Identity-only outputs and Credit states cannot export IPN. Deliver only the recipient's final output after confirmation, never the payer's change.

## Gas estimates and references

`prepared.EstimateGas()` returns `(GasEstimate, error)` before authorization or proof, on Note and managed operations and all four Credit operations. It reads only the prepared structure and embedded benchmarks: no provider, calldata export, store write or submission is involved.

Gas data retains its recorded historical source and manifest provenance; current grant-ID layouts have not been remeasured by this SDK synchronization. All valid direct shapes return `Source: "benchmark"`, `Available: true` and positive `*uint64 GasUnits`: Shield1–10 outputs, all canonical transact shapes through190 total slots (funding/withdrawal/fee included), Credit create/update/revoke and draw1–80 with any payer grouping. `Statistic` is `median` for a measured shape or `estimate` for receipt-calibrated estimation. `ProtocolCommit` and `SourceSHA256` identify the same data as TS; payload, extra source trees, Permit and funding signatures use the same allowances. Unknown revocation callers use a conservative contract-wallet allowance; estimating never queries a provider. Recursive aggregate outer previews remain explicitly unavailable.

```go
estimate, err := prepared.EstimateGas()
if err != nil { return err }
if estimate.Available { log.Printf("benchmark gas units: %d", *estimate.GasUnits) }
```

`gas.EstimateGas(shape)` is the pure protocol entry point; `gas.EstimateGasChecked(shape)` additionally returns invalid-shape errors. Prepared/session methods propagate those errors. A benchmark-derived gas count is an estimate, not a guaranteed transaction gas limit or monetary fee. Applications supply current fee prices and own final transaction simulation. Broadcaster quotes and `maxGas` have separate meanings.

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

Maintainers regenerate bindings with `cd protocol && INVISI_PROTOCOL_ROOT=/path/to/invisi-core go generate ./contracts`. `python3 scripts/sync-reference.py --proving-only` authenticates the current 28 daily Note circuits, 34 aggregate outers and 16 direct Credit routes, with recursive certificates for C1..C10. User manifest4/format7, aggregate receipts10 and Credit manifest8/format9 have separate identities.

`State.Update(Change, deadline)` applies mutually exclusive uint128 increase/decrease deltas to the authenticated input balance. `Change.Unlimited == nil` preserves the mode; entering unlimited clears the balance and requires zero deltas, and leaving unlimited starts at zero plus this increase. Update preserves historical usage. Owner-pulled create/update/revoke fees do not consume available credit. `PrepareUpdate` freezes the original change; quotes never apply the delta again. An explicit new preparation applies that change to the latest accepted state; an externalized journal always resumes its original calldata.

Current configurable examples live under [examples](examples/README.md). Real native proof tests establish cryptographic interoperability; mocked operator/provider tests establish stage and journal behavior and do not claim real token accounting.

## Optional Provenance tracking

Independent Provenance keeps source recognition separate from protocol spendability. Relay use is optional and limited to operator-supported tokens. Source evidence is generated from the final ready statement; Broadcaster verifies and preflights it before sending. Direct payment submission and `SubmitProvenance` are independent App-controlled steps, in either order. The service automatically matches confirmed payments and publishes recognized outputs.

`ExportProvenanceProof` returns a public recovery packet without private Note openings. Persist it before sending; a fresh client can call `sdk.SubmitProvenanceProof` without the original wallet, prover or transaction hash. Repeated `ProveProvenance(ctx)` validates and reuses cached public evidence locally using the trusted Source client; it needs no Chain RPC, Source RPC, prover or manifest assets. If trust or chain generation changes, use `ProveProvenance(ctx, sdk.ProvenanceProveOptions{Reprove:true})` with the retained encrypted operation. Reproof preserves original payment calldata and the main proof. See [examples](examples/README.md).

## Module checks

`python3 scripts/test-modules.py` creates a temporary workspace and runs boundary checks, race tests, vet and builds for all eight modules plus the private examples. `python3 scripts/check-packages.py` constructs isolated local-proxy archives and tests them without the workspace or sibling checkouts; unpublished versions never enter the global module cache. GitHub CI runs both commands. Native Noir/BB integration tests are opt-in and require explicit local tool paths.

Workspace modules currently declare v1.3.3; no remote tags are created by these checks. Before standalone consumer builds can resolve the new version from GitHub, publish the required module tags in dependency order and refresh dependency checksums.

Managed `SendTransaction` requires a positive caller-supplied gas limit in the signer or `TransactionOptions.GasLimit`; it never triggers geth automatic gas simulation. Applications choose the submission budget.

## Native confirmation progress

Chain confirmation uses native `latest`, `safe`, and `finalized` tags. Explicit payment Note and disclosure verification default to `finalized`; choose `Finality: "latest"` or `"safe"` explicitly when the application needs another boundary. No confirmation-count or elapsed-time fallback is applied. Spendability and live wallet reconciliation still use current canonical state.

Protocol RPC indexed progress is independent of native finality. Its claimed finality heads are checked against the configured primary provider. Broadcaster receipt success/revert is separate from its optional confirmation progress; wallet operations independently query the chain rather than trust relay claims. Unsupported tags, transient outages, inconsistent headers, and finality still pending are distinct outcomes. SDK observations are stateless and do not replace the services' durable accepted-finality fences.

`rpc.ProtocolRpcClient.GetNoteFromTrustedRPC(ctx, scope, commitment, "safe" | "finalized")`
queries the confirmation selector of `invisi_getNote`. Construct the client with
`rpc.ClientOptions{Deployment: &deployment}` to explicitly pin the full deployment;
a deployment discovered from the server is insufficient. This path makes one Protocol
RPC request and needs no chain provider. Its distinct `TrustedNoteEvidence` result
validates the envelope, deployment, block bindings and Merkle proof locally, while
receipt success, canonicality, rootHistory and native finality remain trusted RPC
claims. Pending evidence has a reason and no output, membership or confirmation;
it does not prove that a payment is absent. Existing `GetNote` continues to perform
independent chain authentication.

## Contributing

Open a focused GitHub issue or pull request describing the problem, scope, and validation. Run `python3 scripts/test-modules.py` and `python3 scripts/check-packages.py` for your change. Never include private keys, credentials, or private witnesses in issues, pull requests, logs, or fixtures.

Release notes use the local versioned CHANGELOG as the only maintained source.
`scripts/release-notes.py vMAJOR.MINOR.PATCH --output NOTES.md` renders the complete
current entry (all public package entries for TypeScript) into the GitHub Release
body. Missing, duplicate, empty or link-only entries fail validation before
publication. Historical entries are retained locally; a release body includes only
its matching version. Release retries update the same body from the same source.

After broadcasting an address registration, call `TrackTransaction` and persist
the tracked operation's `ConfirmationSnapshot` as JSON. On restart, load it
as `sdk.RegistrationSnapshot` and call
`registry.RestoreRegistrationConfirmation(ctx, wallet, snapshot)`.
If the result is uncertain, call `Reconcile` again. Restored operations cannot
export a transaction.


Credit creation supports `InitialPayment: &credit.InitialPayment{Amount: gross, Payload: payload}`. The SDK encrypts one net payment Note to the grant spender and two postdebit grant deliveries, then obtains only the owner's combined authorization. The gross debit includes the actual fee. Set optional `NetAmount` to pin the merchant amount: initial `Amount` must equal `NetAmount + Fee.Amount`, and fee quotations rebuild gross debit, usage, grant deliveries and the Note while preserving that net. Without `NetAmount`, `Amount` remains a fixed gross debit and fees reduce the Note. IPP's `credit.initialPayment.amount` is net; its preparation adapter always pins `NetAmount` and adds the locally selected fee. The pin survives encrypted recovery. Omit `InitialPayment` for pure creation. Both pure and combined creation default to the current ABI (`00b6fa56`). Use `CreationABI: "legacy"` for an older deployment, or `PrepareCreateLegacy` / `LegacyContractABI` at the low level.

The combined circuit uses the same independent Core resource set as other Core circuits. The historical `9161f92f5789d8eedae15b23fa2651d5399f64ff8fbfc8de6f3f683f0d93be36` set is selected only by explicit `LegacyCombined` mode. Loading remains lazy and honors the existing local, cache, mirror and offline options. A directory containing both resource sets can retain each index as `<index-sha256>.index.json` alongside flat content-addressed objects.

## Independent proving resource sets

Core (Transact/Credit, including initial-payment) and Provenance use separate immutable catalogs. The default Core index is `85a5d1dffe40fc965ab4345a28b632f463c3ad861c44d5845850fbc7447fd8f6`; the Provenance index is `8d5a3e60e3e5cbe315150511207d7aff67f4e2924148d396b67d82017ecd1f1c`. Credit objects use the current grant-ID identity; unchanged Note and Provenance objects retain their authenticated identities. Downloads use `/prover/sets/<index-hash>/`, and verified objects share the same content-addressed cache.

Core operations do not require the Provenance catalog; Provenance operations do not load Core. Missing or corrupt split resources fail without a legacy downgrade. Historical catalogs remain available through explicit legacy configuration.

**Release prerequisite:** these split catalogs have been built and verified locally but are not claimed published. Publish each producer's authenticated CI artifact before releasing SDKs with these new defaults. Until then, provision matching local catalogs or select legacy mode explicitly. No production pins are fetched from an untrusted URL or inferred from a failed download.

`native.ResourceOptions` accepts `LegacyCombined: true` for historical compatibility. Use `Directory`/`Sources` for Core and `ProvenanceDirectory`/`ProvenanceSources` for Provenance. Cache, HTTP-client and offline settings are shared. `PrefetchResources` defaults to Core; `PrefetchProvenanceResources` preloads only Provenance.

### Optional Payment Router

The canonical four-mode API is included in v1.3.3. See [paymentrouter/README.md](paymentrouter/README.md) for self-submitted native/ERC20 and relayed ERC20/EIP-7702 native payments, exact-output V3/0x and same-token/wrap routes. The exact output includes the Core-token fee. Router relay uses `EstimateRouterQuote`, `GetRouterQuote` and `SubmitRouterTransaction` with each phase’s exact embedded Core data and zero-value outer envelope. Export `prepared.ExportRouterDraft(ctx)`, build with `BuildPaymentDraft` / `BuildSponsoredDraft`, and finalize the semantic quote before proving and wallet signing. `EstimateRouterGas` previews conservative local Router allowances above the prepared Core estimate. Quotes use `payment-router-intent-v1`. Proof and wallet signature changes retain `RouterIntentHash`; fee, route, nonce or deadline changes require a new quote. Final submission validates and simulates the actual signed envelope. See [compiling workflows](paymentrouter/workflow_example_test.go). The host owns wallet/proof lifecycle and submission; the core SDK does not import this optional module. The preview source is `model`, not receipt-calibrated swap measurements; opaque 0x is unavailable without an independent trusted route budget. The service uses operator policy rather than client hints and prices Router quotes from current network fees under an operational gas maximum and signed caps. Ordinary Core cap-based pricing and Core/circuit interfaces remain unchanged. Historical native-only helpers remain explicitly available.


For Credit submission paths, a first transport attempt is known not to have been sent only when the structured Broadcaster RPC error matches both its numeric code and machine reason: `-32024/CREDIT_PERIOD_CHANGED`, `-32025/INSUFFICIENT_EXECUTION_SLACK`, or `-32026/CREDIT_PERIOD_ENDING`. Managed journals persist that local refusal state; detached callers retain their handle and own durable recovery storage. The refused handle blocks an unchanged SDK submission. Refresh or safely abandon the known-unsent operation, obtain a fresh quote, and authorize/prove a new statement. When rebuilding creation with an initial payment, preserve its originally selected grant ID; choosing a different ID creates a separate authorization. Ordinary transport failures, malformed or mismatched errors, and a refusal received after an earlier uncertain attempt retain the exact original request and reservations. A missing transaction hash alone never permits rebuilding.
