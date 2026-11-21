# Changelog

## 1.3.3

- Release all eight Go modules together at v1.3.3, with the Core resource catalog pinned to `85a5d1dffe40fc965ab4345a28b632f463c3ad861c44d5845850fbc7447fd8f6` (Core commit `4940be21de91689b59ba1a2fe46945f767c3a34c`) and the independent Provenance catalog pinned to `8d5a3e60e3e5cbe315150511207d7aff67f4e2924148d396b67d82017ecd1f1c` (Periphery commit `33ba458ea11e07da282c21b8a2b091a0be1ebd01`). Preserve historical combined and initial-payment resource identities.
- Synchronize uint128 Credit grant identifiers, validate UTC periods before submission, and recover unsent Credit draws and terminal relay jobs.
- Batch wallet evidence and refresh inventory on demand while retaining private anchor fences.

- Load recursive verification keys lazily from authenticated public binary resources instead of embedding complete field payloads. Native aggregation preserves the backend's resource policy; custom backends can implement `VerificationKeyLoader` or provide an explicit fallback through `AggregatorOptions`. Retain canonical field, SHA256, Poseidon2 commitment, member and outer proof verification.

- Add quote-before-proof Router drafts, prepared Core draft export and local `payment-router-gas-v1` previews with source `model`. Opaque 0x requires an independently trusted route gas budget; Broadcaster policy never trusts client gas hints.
- Introduce `payment-router-intent-v1` semantic quotes: completing proof and wallet signatures preserves quote identity while business fields remain bound. Final submission checks actual hashes/authorizations and simulates the final outer call. Keep `EnvelopeHash` as the raw identity helper.
- Update all four-mode Router guides and compiling host recipes to finalize the formal quote before proving/signing. Router dynamic network-fee pricing uses an operational gas maximum and signed caps; ordinary Core cap-based pricing and Core/circuit/contract interfaces remain unchanged.

- Add canonical PaymentRouter/account ABI codecs, four-mode payment construction and validation, native/ERC20 V3 and AllowanceHolder routes, approval/permit and explicit EIP-7702 signing helpers, full outer simulation and dedicated Broadcaster RPC methods. Preserve historical native-only APIs. Shared calldata/signature/envelope vectors cover cross-language parity.

- Read consecutive trusted historical headers through validated 256-block RPC ranges, falling back to live batches only when the server explicitly reports an unsupported method or unavailable stored history. Preserve fresh canonical and finality fences.

- Refresh Core and Factory ABI snapshots and generated RPC bindings after removal of address screening; Core construction now takes instance parameters and registry, and Factory deployment and prediction take five arguments.
- Keep the current Credit ABI snapshot generated alongside public bindings.

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
