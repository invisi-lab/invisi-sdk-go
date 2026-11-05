# Changelog

## Unreleased

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
