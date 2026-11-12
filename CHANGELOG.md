# Changelog

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
