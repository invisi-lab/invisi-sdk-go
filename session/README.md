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
| `RetryUnresolved(ctx, decryptor, options...)` | Retry ambiguous deliveries without advancing discovery coverage; default online verifies canonical anchors and consumption, while `RetryUnresolvedOptions{Offline: true}` leaves recovered funds unchecked. A recovered Credit parent schedules its chronological successor pass. |
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

Preparation keeps the existing API: independent input checks and output work use at most four workers per group, with statement order preserved. The frame reads metadata instead of loading the full inventory twice. Block-bound public reads are operation-local, retained for at most 30 seconds and 1,024 entries; chain identity, canonical headers and latest funding nonces remain live. Broadcaster quotation reprices the frozen plan rather than rebuilding business outputs.

SQLite keeps an instance-local authenticated decode cache (512 records, 8 MiB weighted by ciphertext/plaintext size). Every read still checks the current encrypted row and authenticated manifest; returned values are detached, changed bytes are authenticated again, and cached plaintext is never persisted.

### Synchronized preparation and offline export

A successful online `Refresh`/`ScanHistory` state refresh retains preparation evidence in the existing encrypted wallet metadata. It includes the real canonical header, deployment binding, owned Note/Credit paths, consumption observations, revocation words, funding nonce and grant allocation. Paths and root checks already authenticated by normal sync are reused; incremental refresh checks root acceptance before reusing older paths. Public history keeps only a bounded transient copy of authenticated public paths, without private openings or a new persistent public schema.

Keep using `PrepareTransfer`, `PrepareShield` and `Credit().Prepare` as before. Cached facts are selected automatically; a missing fact retries the existing online path. Reopening a session with the same encrypted store and pinned Protocol supports offline preparation, EOA authorization, local proving and calldata export. Unknown tokens/payers and contract-wallet authorization may still require online data or verification. Canonical anchor timestamps stay historical; construction deadlines use the current local time when later than that anchor.

A cached operation exposes `SubmissionValidation: "required"`. Exported calldata is not permission to submit: call `operation.ValidateSubmission(ctx)` immediately before a host sends it. Credit operations expose the same method. Session-managed sending and first Broadcaster submission perform the live check automatically. These checks validate deployment and canonical state, accepted roots, consumed inputs, deadlines, funding nonce/signature, Credit allocation/revocation and any payer balance/Protocol allowance. They never treat a historical snapshot as current spending authority. A quote still requires the selected Broadcaster's online service.

### Scheduled broadcaster submissions

Pass `sdk.BroadcasterQuoteOptions{MaxFeeAmount: maxFee, Schedule:
&sdk.ScheduleOptions{PublishAfter: "2000000100"}}` to `client.Quote(ctx, operation, options)` before
signing or proving. `ExecuteBefore` is optional and defaults to the server's
advertised execution window (normally 300 seconds). Capabilities are bound to
the selected endpoint, chain and protocol. Prepare the operation, funding
authorization and permit with deadlines covering `executeBefore` plus the
advertised `minDeadlineSlackSeconds`; scheduling never silently extends signed
business deadlines. The server's advertised maximum delay
is honored, so deployments can configure a limit other than seven days.

The SDK generates a random 32-byte cancellation secret and persists it locally
with the signed quote and frozen broadcaster request before the first submit
attempt. Submission IDs are opaque and allocated by the server; the SDK stores
the actual response ID before returning it. If a response is lost, explicitly
retry `client.Submit(ctx, operation)` to replay the original frozen request idempotently and
recover that ID before calling `client.Cancel(ctx, operation)`. Cancellation never creates
a remote job for an operation that has not been submitted. Quotes and submission RPCs
carry only its Keccak-256 hash. Use encrypted storage for raw operation snapshots and backups; JSON deliberately
retains the secret for recovery, while normal Go formatting redacts it.
`client.Cancel(ctx, operation)` sends the secret to the same endpoint;
inputs remain reserved on errors or any response other than `cancelled`; an unacknowledged cancellation returns an error. An
explicit cancellation acknowledgement is persisted before canonical state refresh;
if refresh fails, a resumed operation can retry reconciliation while the
broadcaster is offline. Canonical state refresh must finish before reservations
are released. Reconcile submitted transactions normally if cancellation loses
the race to publication. Provenance remains the ordinary admission policy;
scheduling fields are never sent to Provenance.

Low-level callers can use `broadcaster.NewCancelToken()`,
`EstimateQuote(ctx, data, schedule)`, `GetQuote(ctx, data, trust, schedule)` and
`CancelTransaction(ctx, id, token)`. Store the secret before sending the transaction, and
retain it when a transport error leaves admission uncertain.

Credit scheduling supports create, update, draw and revoke. Low-level `credit/session.Options.CheckCancellation` must check anchored canonical consumption, revocation and allocation independently of expired authorization or payer allowance; `WalletSession.Credit()` supplies that checker. An acknowledged cancellation retains reservations until this check succeeds.
