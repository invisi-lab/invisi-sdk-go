# Changelog

## 1.3.3

This release includes the complete eight-module cohort: root SDK, primitives, protocol, prover, rpc, session, ipp and paymentrouter. Upgrade their requirements together to `v1.3.3`.

### Core and Credit compatibility

- Replace Credit allocation shard/slot and revocation bitmaps with positive uint128 grant IDs, scoped by chain, Core, owner and token. `GrantID` is a `*big.Int`; JSON/RPC uses canonical decimal strings. Creation accepts an explicit ID or generates one cryptographically before encryption/proving, then preserves it through quotations, retries, initial-payment construction and recovery.
- Add canonical `GetCreditStatus`/contract status reads: 0 unknown, 1 active, 2 permanently disabled. Create requires unknown; draw/update require active; revoke accepts unknown or active without needing private grant state. Managed inventory keeps unknown grants visible but unselectable, preserves later anchored observations during historical replay, and retains permanent disables across successor commitments.
- Synchronize Credit commitments, EIP-712 authorization, calldata, receipts, RPC types, proving witnesses, session records, IPP terms and shared vectors with the grant-ID Core. Credit envelopes now use `0x84` (minimum 491 bytes); older Credit envelope/state models are rejected. Current Credit proving metadata is manifest 8 / format 9; Note manifest 4 / format 7 and aggregate format 10 retain their separate identities.
- Refresh Core/Factory ABI snapshots and generated bindings after address-screening removal and the Credit change. Core construction uses instance parameters and registry; Factory deployment/prediction uses the current five-argument interfaces. Generate the Credit ABI snapshot with the public bindings.
- **Upgrade requirement:** deploy/use the matching current Core contracts and a fresh Protocol RPC index/database for that deployment. Do not mix old shard/slot Credit data, signatures, prepared requests or proofs with the grant-ID ABI. Recreate affected Credit preparations and reauthorize/reprove them; coordinate contract addresses and service versions with the application. This is a protocol-data compatibility change despite retaining the SDK v1 version series. An offline creation snapshot must contain an exact status-0 observation for its preselected ID; it cannot establish absence for a new random ID.

### Payment Router and relay quotes

- Add canonical PaymentRouter and account ABIs, calldata codecs and validation for all four modes: self-submitted native, self-submitted ERC20, relayed ERC20 and sponsored native. Support exact-output Uniswap V3, validated native/AllowanceHolder 0x routes, same-token ERC20 and native-to-WETH wrapping. Exact output includes output Notes plus the Core-token Broadcaster fee. Preserve explicitly named historical native-only helpers.
- Add explicit ERC20 approvals, ERC-2612 permit helpers, Router/account typed-data signing and EIP-7702 delegation. Sponsored native spends the player's principal while the fixed relayer pays gas, using zero outer value. Validate pinned account implementation/code, independent payment/account/EOA nonces, sender, deadline and authorization bounds; simulate the complete outer call and authorization list before sending. Hosts must supply supported wallets/chains/RPCs and explicitly approve delegation. Router/account EOA signatures use low-s `r || s || v` with `v` 27/28; EIP-7702 `yParity` remains 0/1.
- Add `ExportRouterDraft`, `BuildPaymentDraft` and `BuildSponsoredDraft` so fee estimation and the formal signed quote precede Core proving and wallet signatures. Dedicated `EstimateRouterQuote`, `GetRouterQuote` and `SubmitRouterTransaction` methods bind the exact embedded Core calldata to the Router envelope; ordinary Core quote helpers remain separate.
- Introduce `payment-router-intent-v1`: completing proof, permit, Router/account and delegation signature bytes preserves quote identity, while outputs, fees, route, nonce/deadline and authorization metadata remain bound. A business-field change requires a new quote. Final submission rejects missing proof/authorization, validates the actual signed envelope and performs full simulation. `EnvelopeHash` remains the raw identity helper; Router scheduling and aggregation are unavailable.
- Add `payment-router-gas-v1` previews for four modes, V3/no-swap, permits and delegation, with source `model`. These are conservative engineering allowances rather than measured swap receipts. Opaque 0x requires an independently trusted route gas budget. Router relay pricing uses current network fees, an operational maximum gas cap and signed fee/gas caps; operator policy never trusts client gas hints. Ordinary Core cap-based pricing remains separate, and retained Core gas data keeps its historical provenance rather than claiming grant-ID remeasurement.
- Document frozen-request recovery: persist the literal submission body and returned ID, restore quote context only after trusted verification, and reconcile uncertain sends before creating another payment. An expired frozen quote requires the documented exact-body recovery path rather than bypassing SDK expiry checks.

### Proving resources and Provenance

- Load recursive verification-key fields lazily from authenticated binary resources instead of embedding all field arrays. Native aggregation inherits local directory, mirror, cache and offline policy; dedicated routes load the required capacity, mixed routes C1–C5. Custom backends can provide `VerificationKeyLoader` or an explicit `AggregatorOptions` fallback. Canonical field checks, SHA256, Poseidon2 commitment and member/outer proof verification remain enforced.
- Pin current Core resources to `85a5d1dffe40fc965ab4345a28b632f463c3ad861c44d5845850fbc7447fd8f6` from Core `4940be21de91689b59ba1a2fe46945f767c3a34c`. Retain independent Provenance pin `8d5a3e60e3e5cbe315150511207d7aff67f4e2924148d396b67d82017ecd1f1c` from Periphery `33ba458ea11e07da282c21b8a2b091a0be1ebd01`. Objects load from immutable `/prover/sets/<index>/` catalogs. The split Core/Provenance loader is retained from the preceding release: each capability loads only its own catalog and shares authenticated content-addressed cache policy.
- Preserve historical combined/initial-payment pins for explicit legacy use; missing current resources never silently downgrade. Current grant-ID Credit statements reject historical combined mode. Provision matching current catalogs for offline deployments; the reviewed producer catalogs have been published and their public objects verified before this release.
- Retain generic Source recognition independently of spendability and onchain submission. Add Router-output recognition and later private-spend host recipes: public Router funding needs no input Provenance proof or Router enrollment; spending a recognized private Note uses real Core and separate Provenance proofs, authenticated Source trust, finality checks and persisted public recovery packets. Source may remain pending after chain finality. The existing 11 Provenance public inputs remain unchanged; generic Source APIs and split resources are not new interfaces introduced by this patch.

### History, inventory and recovery

- Optimize verified public history and Credit replay through shared receipt/header evidence and authenticated replay checkpoints. SQLite can persist intermediate replay metadata without rewriting all encrypted record rows; final replay still authenticates records and restores reservations. Cancellation/failure retains committed progress and revision/generation checks.
- Read trusted historical headers in ordered ranges of at most 256 blocks; validate completeness, sequence, hash/parent linkage and requested bindings. Fall back to bounded live header batches only for explicit unsupported-method/unavailable-history responses; malformed data and other failures remain errors. Keep native safe/finalized checks and canonical fences.
- Add batched membership authentication for up to 256 ordered paths, sharing status/root reads while checking every leaf/path and rereading distinct root anchors and the target before returning. Reuse evidence only within the authenticated operation. Add `RefreshStateWithOptions(... InventoryOnly: true)` to refresh known inventory without prewarming unrelated transaction material; discard the older preparation snapshot and authenticate selected inputs when subsequently preparing.
- Bound independent consumption/grant-status reads to four workers at one numeric block/hash and share exact duplicate reads. Merge only after all reads, private/public anchor fences, deployment identity, revision and generation checks succeed. Recheck current root acceptance even for paths retained at the same block; keep spent/consumed records subject to canonical refresh. Performance samples are loopback/provider fixtures, not live-chain speed guarantees.
- Check fresh canonical UTC day/month before first submission of Credit draws and creation with an initial payment. Period changes return `CREDIT_PERIOD_CHANGED` and require new authorization/proof. Detached/restored low-level preparations must install a chain timestamp reader; a historical snapshot alone cannot authorize a current payment.
- Recognize a first attempt as known-unsent only for the exact structured refusal pairs `-32024/CREDIT_PERIOD_CHANGED`, `-32025/INSUFFICIENT_EXECUTION_SLACK` or `-32026/CREDIT_PERIOD_ENDING`. Persist this state and block unchanged resubmission; refresh or safely abandon, requote and reauthorize/reprove, preserving a creation's selected grant ID. Ambiguous transport failures and refusals after an earlier uncertain attempt keep the frozen body and reservations. Terminal relay/cancellation recovery releases reservations only after independent canonical reconciliation, including after restart or when the Broadcaster is unavailable.

### Examples, documentation and validation

- Add compiling Router funding/relay and Provenance recognition/private-spend host workflows; update four-mode, quote-before-proof, signatures, deployment trust and frozen-request recovery guides. Keep application decryption, wallet signing, persistence and sending explicit. Offline help performs no network or submission.
- Remove obsolete SDK benchmark sources/reports while retaining gas reference data and its recorded provenance. Document bounded wallet/membership performance fixtures without presenting them as production measurements.
- Verify all eight modules plus private examples with module-boundary checks, race tests, vet and builds; verify all eight independent module archives without sibling checkouts. Validate release identities/notes, authenticated Core metadata reproducibility, independent catalog export and native asset/prefetch/Provenance authentication.
- Shared cross-language fixtures cover calldata, commitments, signatures, semantic quotes and recovery boundaries. Compiling host recipes do not execute live RPC, DEX or wallet workflows. Local compiled-contract/Anvil and genuine Noir/Barretenberg proof tests remain opt-in; no proof regeneration or live mainnet swap/Source recognition is claimed by release preparation.

## 1.3.2

- Add optional paymentrouter module with authenticated deployment snapshots, Router-funded SDK preparation, exact-output Uniswap V3 and native 0x quote adapters, canonical calldata, and full-call simulation. Include cross-language vectors and opt-in local compiled-contract execution tests.

## 1.3.1

- Accept unified CreditDraw events in indexed Credit RPC responses, including post-debit state deliveries.
- Preserve canonical block, range, membership, duplicate and ciphertext checks; restore automatic IPP Credit reconciliation.

## 1.3.0

- Support optional owner-only Credit initial payments, post-debit Credit state, merchant net Notes, gross funding checks and UTC-period validation.
- Align contract bindings, unified Note operation events and CreditDraw/CreditStateTransition discovery and recovery with the redeployed Core contracts.
- Recompute operation IDs and validate physical insertions and aggregate membership against shared real-receipt fixtures.
- Add independently pinned lazy resources and witness metadata for the 18-input combined circuit without bundling large assets.
- Refresh measured Gas models and TS parity vectors; expose uncovered paths explicitly.
- Extend IPP initial-payment terms and receipts; remove the obsolete Invisi action field.
- Prepare the complete SDK module cohort at 1.3.0; no registry publication or contract deployment is performed.

## 1.2.6

- Move all Go module paths and public imports to `github.com/invisi-labs/invisi-sdk-go`. Consumers must update imports together; existing release tags remain unchanged.

## 1.2.5

- Restore required non-semantic compiler metadata when executing stripped resources with the official Noir CLI, preserving the pinned artifact and verification-key identities.

- Add the independent IPP module and Invisi preparation adapter.
- Load fixed-hash public proving resources lazily, with local directory and offline support, verified cache reads and shared cache naming.
- Expose authenticated resource prefetch for deployment images without requiring local proving executables.
- Align contract bindings with Invisi Core.


## 1.2.4

- Refresh receipt-calibrated gas estimates for optimized Protocol contracts and validate authorization and unsigned funding boundaries.
- Add trusted RPC Note confirmation and collection with broadcaster fees deducted from the payout.
- Support scheduled broadcaster submissions and cancellation; align quote lifecycle, Credit proof validation and managed recovery.
- Bound parallel public-history header reads during wallet synchronization.
- Preserve Protocol ABI and authenticated circuit formats. Gas estimates remain distinct from transaction gas limits.

## 1.2.3

- Align transaction, Credit, history and Provenance workflows across the three SDKs.
- Reuse synchronized preparation snapshots and decoded records, including offline preparation.
- Bundle authenticated native Source proving assets and configure official Provenance automatically.
- Preserve cancellation and recovery fences, final-fee authorization and request snapshots.

## 1.2.2

This release includes all changes since 1.2.0; Go SDK 1.2.1 was skipped.

- Store Wallet public material as owner address plus required receiving keys; encode complete `0i` addresses on demand, reject owner-only Wallets and missing keys, and copy material across constructors and runtime snapshots.

- Follow the latest Protocol and RPC contracts; remove delegated RPC detection and perform native batch detection in managed account scans.
- Follow signed broadcaster execution quotes, validate quote-bound submission, and relay aggregate members.

- Estimate prepared Note and Credit gas locally from exact measured benchmark shapes before proof; report unavailable shapes explicitly and remove SDK RPC gas simulation.

- Share bounded canonical public history across managed accounts, with independent head-first private recovery coverage and chronological Credit successor replay.
- Add native batch detection, retained ambiguous-delivery retries, exact scan/funds status, and current-head known-state reconciliation.
- Migrate SQLite accounts to authenticated encrypted per-record rows and cold archive paging; add persistent public cache tables in the same host database.

## 1.2.0

- Split the SDK into independently versioned primitives, protocol, prover, rpc, session and root SDK modules; rename core to primitives without an old-name module.
- Share strict settlement decoding, public statements, quote/submission hashes and gas references with RPC, Broadcaster and Provenance.
- Separate pure ABI tuples from network contract bindings and place Provenance contracts, local proving, clients and encrypted recovery in their owning modules.

- Check historical Source root eligibility before replaying a tree for a new proof.
- Reuse locally validated cached provenance evidence without network calls, local manifest assets or a prover; explicit reproof refreshes trust and preserves the original payment on failure or cancellation.
- Add final provider-backed gas estimation for prepared and managed Note transactions and all four Credit operations, returning an explicit RPC source.
- Verify token-filtered broadcaster discovery against live token, chain, protocol and registered signer identity.
- Document release preparation and distinguish actual gas estimation from measured references, relay fees and quote ceilings.

## 1.1.7

- Add canonical cross-language IPR payment requests and fixed-size IPN Note delivery.
- Rename NoteOutputOptions to PaymentRequest and preserve original IPR through nonce selection and output snapshots.
- Add commitment-based RPC lookup, confirmed Note verification and owner spendability verification without signing.
- Export final Note outputs and generated change as IPN, while rejecting identity-only and Credit exports.

- Organize public SDK entry files by capability, isolate cryptographic and gas implementations, and split transaction, Credit, RPC and session workflows into focused source files.
- Align reference synchronization and fixture generators with the current TypeScript functional module layout.

- Separate managed wallet sessions, history, inventory, operations and storage into `session` and `session/sqlite`.
- Keep `sdk` preparation stateless, remove `Invisi.Connect` and the infrastructure Store option, and share transaction algorithms through an internal engine.
- Preserve numbered flat examples with explicit session composition and external transaction submission.

## 1.1.6

- Align note outputs, optional encrypted delivery and EVM receiving-key resolution with the current contracts.
- Support broadcaster discovery, quotes, fee-funded preparation and transaction submission workflows.
- Refresh authenticated circuit artifacts, gas references and cross-language workflow coverage.

## 1.1.5

- Add `ProtocolRpcClient.GetTransactionUpdates` for Protocol RPC service 1.1.4+, matching the TypeScript SDK 1.1.5 availability semantics.
- Return typed index status, canonical mined block, EVM receipt status, encrypted outputs and public Credit transitions without range scanning or Store updates.
- Preserve stale observed checkpoints for pending indexing, retain strict checks for indexed results, and reject mismatched transactions, invalid fields, conflicting deliveries and reorganizations.
- Export result types and status constants from both RPC and SDK packages.

## 1.1.4

This release aligns Go with TypeScript SDK 1.1.4 and the current Protocol RPC interface.

### Breaking changes

- Separate public wallet identities from private session capabilities. Remove address-only private-wallet construction and implicit detached signing.
- Use exported EIP-712 requests and externally attached wallet/funding/registration signatures. Private signing helpers remain in managed sessions.
- Unify transfer, withdrawal and owner-only sweep through `PrepareTransfer` with optional `Withdrawal`.
- Replace note memo fields with raw `Payload` bytes and explicit application codecs. Use version-1 Funds/Credit envelopes; reject historical formats.
- Use strongly typed proving options and remove the dedicated withdrawal calldata builder.

### Added and improved

- Protocol RPC registered-key and output discovery, with strict identity, block and duplicate/conflict validation.
- Stable preparation error codes and recipient context.
- Session business history, post-commit subscriptions, authenticated Store revisions and contiguous discovery coverage.
- Serialized refresh and revision checks across resets; reentrant subscription callbacks run outside operation locks.
- Multi-owner Credit examples, editable network/native-currency settings, ten-account local setup, seven example categories and explicit execution guards.
- Compact signatures, redacted private wallet serialization, defensive aggregate snapshots and authenticated reference synchronization.

### Validation

- Full Go tests, race detection and vet.
- Current TypeScript cross-language vectors and reference synchronization tests.
- Genuine native transfer, withdrawal, sweep, recursive aggregation and Credit lifecycle proof generation and independent verification.
