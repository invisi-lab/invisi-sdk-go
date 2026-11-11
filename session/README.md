# Invisi Session

Managed wallet state, persistence and operation recovery.

Module version: `1.2.0`. See the [SDK repository](https://github.com/invisi-lab/invisi-sdk-go) for the six-module dependency boundary.

# Managed wallet sessions

Import `github.com/invisi-lab/invisi-sdk-go/session` when the SDK should manage inventory, canonical refresh, reservations, private history and durable operations. Stateless callers use `sdk` without this package or a database driver.

```go
invisi := sdk.NewInvisi(sdk.InvisiOptions{
    Provider: provider,
    ProtocolRPC: protocolRPC,
    Prover: localProver,
})
managed, err := session.Connect(ctx, session.Options{
    Invisi: invisi,
    Wallet: wallet, // Public material only.
    Protocol: pool,
    Store: store,
})
```

`Connect` validates the pinned Protocol and RPC deployment before returning. Wallet contains only public material. Connect takes no wallet executor, viewing provider or authorization callback. Shared infrastructure contains no default signing authority. The geth signer and optional public-funding signer remain separate from private wallet authorization.

`pending, err := managed.Refresh(ctx, options)` stages a bounded public page without advancing private inventory. `pending.Decrypt(ctx, appDecryptor)` invokes the explicitly supplied `sdk.DeliveryDecryptor` and commits each validated account batch with its coverage. SDK validation binds plaintext to owner, token, deployment, commitment and Credit receiving keys. Concurrent application, stale pending revisions and reuse after success reject. See shared history below for ambiguity, resumption and current-state checks.

Authorization exports `operation.ExportWalletAuthorization(ctx)`, accepts an ordinary geth EIP-712 signature through `WithWalletSignature`, then proves. Public funding and revocation have their own frozen typed-data requests. Refresh unseen deliveries before Credit reconciliation; already authenticated stored openings can be reused.

`session.NewMemoryWalletStore` is explicitly ephemeral. For persistence, import `session/sqlite`, provide an application-owned database, and construct `session.NewEncryptedWalletCodec` with an application-owned encryption key. Back up the database and key separately. Failed local work never advances the store revision.

Managed `Operation` uses the same preparation, statements, signatures, witnesses, proof validation and calldata as `sdk.PreparedTransaction`, and additionally tracks persistent worker leases, reservations, submission and canonical reconciliation. `SendTransaction` delegates to an optional geth signer and does not wait. Both paths can export calldata for external submission.

All packages belong to the existing Go module and share its version. No compatibility exports remain under `sdk` or `sdk/sqlite`.

## Shared public history

A deployment-wide `PublicHistory` caches canonical headers, successful receipts, parsed Note/Credit deliveries (including clues), insertion positions and validated membership. Its scope binds chain, protocol address, runtime code hash and deployment height; it contains no account key or private opening. Inject the same instance in multiple `Connect` calls. Sessions using one `sdk.Invisi` and Protocol handle receive an ephemeral shared instance by default.

```go
publicStore, err := sqlite.NewPublicHistoryStore(ctx, database, session.PublicRetention{})
if err != nil { return err }
history, err := session.NewPublicHistory(session.PublicHistoryOptions{
    Store: publicStore, Protocol: protocol, Provider: provider,
    ProtocolRPC: protocolRPC.WithProvider(provider),
})
if err != nil { return err }
privateStore, err := sqlite.NewWalletStore(ctx, database, accountCodec)
if err != nil { return err }
a, err := session.Connect(ctx, session.Options{
    Invisi: invisi, Wallet: walletA, Protocol: protocol,
    Store: privateStore, PublicHistory: history,
})
if err != nil { return err }
pending, err := a.Refresh(ctx, session.RefreshOptions{})
if err != nil { return err }
_, err = pending.Decrypt(ctx, appDecryptorA)
```

SQLite uses separate `invisi_public_history` metadata and `invisi_public_history_pages` payload tables in the same host-owned database as encrypted account records. Public rows need no account encryption key. A second account can use another private Store/codec and the same history. `ScanHistory(ctx, decryptor, ScanOptions{Offline:true})` reuses retained public receipt parsing and position validation without querying discovery, receipts or membership again. Newly recovered offline assets have no fresh consumption observation, so `GetFundsStatus` excludes them until an online state refresh.

| Method | Responsibility |
| --- | --- |
| `SynchronizeHistory(ctx, options)` | Fetch, validate and persist public pages, including empty ranges; no decryption. An explicit FromBlock pointer requests a forward range. |
| `ScanHistory(ctx, decryptor, options)` | Scan account gaps and atomically commit each page's inventory, unresolved candidates and private coverage. Online mode downloads missing payloads as needed. |
| `Refresh(ctx, options)` / `pending.Decrypt(ctx, decryptor)` | Stage one recent bounded public page, then interleave downloads and per-page private commits through the same scan pipeline. |
| `RefreshState(ctx)` | Check known nullifiers, Credit revocation and externalized in-flight operations at the current primary-chain head. |
| `RetryUnresolved(ctx, decryptor)` | Retry ambiguous deliveries without advancing discovery coverage. A recovered Credit parent schedules its chronological successor pass. |
| `GetScanStatus(ctx)` | Read checkpoint metadata: public/private coverage, gaps, recovery boundary, suggested work and exact block progress. |
| `GetFundsStatus(ctx, token)` | Report known fresh unspent, unreserved funds, independently of missing historical ranges. Preparation revalidates eligibility. |

The default page size is 100 blocks, maximum 256. Recovery first extends an existing covered window to the indexed head, then scans older gaps backwards to deployment. An empty cache starts with its most recent bounded page. Progress callbacks distinguish processing/committed phases and forward/backward/Credit recovery directions. Every call has a fixed indexed target. A cancellation or failed batch preserves earlier committed batches; start a new Refresh/ScanHistory call to resume persisted gaps.

Public coverage survives eviction. Defaults retain 128 MiB of page payloads, at most 4,096 pages, with an 8 MiB per-page bound. The lowest block pages are evicted first. `history.PruneBefore(ctx, height)` removes whole old payload pages while keeping validated coverage. Online scans refetch evicted history one page at a time; offline scans fail explicitly for missing or corrupt data. Dense pages exceeding the per-page byte limit require a smaller page size. These limits bound cache payloads, not all live wallet assets or total process RSS. SQLite file maintenance belongs to the host.

Tasks sharing a PublicHistoryStore instance serialize public downloads and reuse overlapping retained ranges. Cancelling one subscriber ends its wait while other subscribers continue to use that download. Independently opened processes use transactional revision checks, but do not coalesce network requests. The cache trusts host storage integrity; payload digests detect corruption and are not a proof against a malicious host rewriting metadata. Chain RPC authenticates receipts and identities; Protocol RPC still supplies candidate discovery, not a cryptographic proof that no transactions were omitted.

On a changed canonical anchor, public generation advances and orphan coverage/payloads are discarded. Account reads and preparation fence mismatched generations. `ScanHistory(ctx, decryptor, ScanOptions{Reset:true})` rebuilds from deployment, removes old unresolved candidates, and restores archived externalized journals before scanning. Reconciliation and reservations remain available. This is a fenced reset/rebuild, not automatic incremental rollback of private history.

Native detection is optional through `wallet.BatchDeliveryDetector` implemented alongside `DeliveryDecryptor`. Each batch item carries owner, chain, protocol, token, commitment, role, clue and ciphertext; native responses echo the account and return `candidate`, `not-recipient` or `unresolved` for every ID. Private detection keys remain native. Decryption receives clue and owner context too. Return nil plaintext with nil error only for an explicit nonrecipient. Return `wallet.ErrAuthenticationFailed` for ambiguous authentication: the encrypted account store retains it for retry, scanning continues, and asset completeness remains unknown. Locked, cancelled, expired, wrong-account and invalid authenticated plaintext failures abort the batch. Generic errors must never be translated to nonrecipient decisions.

`maxScannedHeight` can exceed `fullyScannedHeight`: coverage 100–300 and 900–1,000 has max 1,000, contiguous prefix 300 and gap 301–899. Status progress counts blocks; an empty denominator has a nil ratio and is complete. `ScanComplete` requires contiguous canonical coverage through the known primary tip; indexer lag reports `wait-for-indexer`. `IsSynced` additionally requires chronological Credit recovery to finish. Unresolved ambiguity is reported separately. Credit parents stay active while a persisted target/revision/cursor controls chronological successor replay; an older snapshot cannot clear a newer replay obligation.

## Bounded private history

SQLite migrates existing encrypted single-blob accounts on the next successful write. Changed records are individually encrypted with scope/locator associated data; encrypted metadata binds active row digests, counts, account revision, scan coverage and Credit replay metadata. Failed transactions preserve the old blob/rows. Missing, replayed or tampered active rows fail instead of silently appearing as an empty wallet. Account-scoped encrypted routing uses an independent random index secret, not public commitments. Database routing still reveals categories, row counts and discovery order.

Memory and SQLite stores retain 128 terminal records per collection by default. Set `MemoryWalletStoreOptions{RecentRecords:n}` or SQLite `WalletStoreOptions{RecentRecords:n}` explicitly, including zero. Older terminal Notes, Credits and operations move to cold archives. Live inventory, unresolved deliveries, reservations, pending journals and uncertain externalized authority stay active; these categories may grow and are never silently discarded. Consumed Credit parents are pinned while successor recovery is pending.

Use `GetHistoryPage(ctx, HistoryPageOptions{Direction:"incoming", Limit:50})` or `Direction:"outgoing"` for normal screens. Entries expose the existing business history view, plus `Archived` and `Canonical`. A token filter applies to the selected underlying page: an empty filtered result can still have a continuation cursor. Cursors bind account, kind and wallet revision; restart paging when state changes. Raw adapter users can `ReadRecords` for `notes`, `credits`, `operations`, `credit-operations` or `unresolved`. SQLite pages archives directly without loading the entire archive. `GetHistory` remains an explicitly unbounded compatibility view and rejects mixed revisions. Custom WalletStore adapters retain their existing contract; paging fails explicitly unless they implement `RecordWalletStore`.

Archived records are not chain finality. Records from a previous public generation report canonical state `unknown`. The host owns account export, backup, deliberate deletion and database lifecycle.
