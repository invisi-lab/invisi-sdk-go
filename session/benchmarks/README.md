# Wallet history scan and Credit replay performance

This change keeps the public wallet API and the complete validation rules. It
changes how already authenticated work is scheduled and reused. No versions are
released by this change.

## Replay and state rules

- Public Credit activity alone no longer makes every wallet replay from the Core
  deployment. A newly decrypted owned grant schedules replay from its actual
  receipt block through the scan's frozen indexed target. A transition consuming
  a known parent nullifier also schedules its window, even without ciphertext.
- Dirty ranges merge with a pending obligation. Forward extensions preserve the
  committed cursor; a newly discovered historical parent rewinds it. Unresolved
  deliveries follow the same rule. Legacy snapshots without `CreditReplayFrom`
  retain the conservative deployment-wide replay obligation.
- Recovery runs chronologically and indexes parents by nullifier, retaining both
  owner and spender roles. Recovered children immediately become available to
  subsequent transitions. Child commitments and memberships are still checked;
  ciphertext-free recovery does not invent or inherit an unknown payload.
- Successful consumption and revocation reads can be reused within one serialized
  scan only at the same complete canonical block (including hash). Errors are not
  cached. The cache is bounded and discarded after each scan. Every canonicality
  check remains. The final live `RefreshState` deliberately ignores this cache so
  its preparation snapshot captures fresh checks for all retained assets.
- With `SkipStateRefresh`, empty intermediate Credit replay windows can use an
  optional adapter checkpoint capability. SQLite authenticates and rewrites only
  encrypted metadata; Memory changes only cursor metadata. The private revision,
  generation, replay revision, target, and monotonic cursor are checked atomically.
  Public generation is checked before and after. The shortcut cannot complete a
  replay, change records/reservations, or claim fresh spend/revocation checks.
  The final window and every Credit-bearing window use the full record path.
  Other storage adapters retain the existing full-update fallback.
- Reorgs still invalidate the generation and require reset/recovery. Public header,
  receipt, log, Merkle membership, and finality validation are not weakened.
  Journals and externalized reservations are preserved.

- Paired Note/Credit discovery shares one independently authenticated status for
  the request, while preserving per-stream row and canonical checks. The paired
  API rejects mismatched stream anchors, including a fork between individually
  valid reads. Batch errors retain their JSON-RPC error code and reject ambiguous
  result/error envelopes, including a null error field. Dense per-output
  membership validation remains fully enabled.
- Trusted Protocol RPC providers read full original headers using JSON-RPC
  batches of at most 20, validating every response ID, deployment and header hash.
  Other providers retain bounded eight-request reads. Every parent link and the
  fresh final canonical fence remain mandatory. A frozen state observation that
  precedes the discovery target is rejected before saving progress.

## Measurements (2026-11-20)

Local machine: Apple M2 Max, Darwin arm64, Go benchmark output `-12`. Compare the
unmodified production revision `10c200c` with this branch, using identical fixtures
and benchmark source. Figures below are medians of three runs; local fixture
measurements are not predictions of an entire production `collect` duration.

| Workload | Previous | Optimized | Work avoided |
| --- | ---: | ---: | --- |
| Retained foreign Credit, one new discovery block and 10 total history blocks | 1.759 ms | 0.706 ms | 4 unnecessary replay pages → 0; about 2.49× faster |
| 32 state-check passes, one Note and one Credit at one fixed block; 100 μs configured delay per contract call | 296.099 ms | 10.559 ms | 96 contract reads → 3; about 28.04× faster in this fixture |
| SQLite cursor checkpoint, 512 retained Notes | 59.090 ms | 3.156 ms | Full record update → metadata-only update; about 18.72× faster |
| Cold 256-block empty public page, 1 ms configured HTTP delay | 86.292 ms | 53.663 ms | 267 HTTP requests → 23; status validations 2 → 1 |

The public-page samples average three cold page operations each. The repeated
samples (ns/op) are baseline `[86321764, 86292389, 84752139]` and optimized
`[54370375, 53662806, 51736430]` (about 1.61× at their medians). Timing varies with
machine load; the request reduction is deterministic. See
[raw public-page samples](public-page.json).

The state-check fixture includes the mock provider's ABI handling overhead in both
versions. The SQLite comparison invokes both paths on this branch with identical
inventories: `FullRecordUpdate` uses the previous full `WalletStore.Update` path,
whereas `MetadataOnly` uses the new optional checkpoint. It does not compare a
weaker storage or validation mode.

A separate read-only run against the official BSC testnet RPC measured full headers
for blocks 143821381–143821636. The median changed from 12.036 s to 4.860 s (2.48×),
and HTTP requests including the fresh end fence changed from 257 to 14. Both paths
returned the same complete original headers and passed number/hash/parent checks.
See [raw measurements](../../rpc/benchmarks/headers/bsc97-20261120.json) and the
[reproducible read-only benchmark](../../rpc/benchmarks/headers/main.go).
This benchmark makes no financial transactions and is not a full wallet or proof
benchmark.

No before/after claim is made for end-to-end collection, proof generation,
broadcasting, or confirmation time. The previously observed 575-page deployment-wide replay is the motivating case,
not a measured new end-to-end latency. Relevant newly scanned transition windows
can still require replay; the new scheduler avoids deployment-wide obligations
caused only by unrelated public Credit activity.

## Reproduce local benchmarks

From a local Go workspace that selects this repository's root and module folders:

```sh
go test ./session -run '^$' -bench '^BenchmarkWallet(RetainedForeignCredit|FixedAnchorStateChecks)$' -benchtime=10x -count=3
go test ./session/sqlite -run '^$' -bench '^BenchmarkReplayCheckpoint$' -benchtime=5x -count=3
go test ./session -run '^$' -bench '^BenchmarkPublicHistoryColdPage$' -benchtime=3x -count=3
```

For the old production baseline, extract `git archive 10c200c` into a separate
scratch directory, copy the local `go.work` and the benchmark files, and change
only the `testSession` and `sharedEmptyFixture` test-helper arguments from
`*testing.T` to `testing.TB`. The fixed-anchor benchmark needs this test-only
compatibility shim on the old revision, which has no scan-cache constructor:

```go
package session
import "context"
func withScanStateCache(ctx context.Context) context.Context { return ctx }
```

The retained-foreign fixture asserts that its cached public page really contains
the foreign transition before timing. This avoids accidentally benchmarking an
empty downloaded page after a malformed cached fixture was discarded.

## Regression coverage

The tests cover foreign activity, ciphertext-free owned nullifiers, historical
parent recovery for both roles using a canonical CreditDraw receipt, interrupted
replay resume, unchanged-state no-op replay, legacy pending obligations, exact
anchor reuse, next-scan fresh consumption/revocation reads, same-height changed
hash rejection, empty-window record-load counts, metadata CAS and encrypted-row
stability, corrupted/missing records at full validation, and a public-generation
race during checkpoint commit. Existing canonical receipt, membership, reset,
reorg, consumption/revocation, preparation, and externalized journal tests remain
part of the suite.

## Online read-only comparison

The standalone tool accepts a pinned public deployment JSON (the same deployment
pins used by a normal client). It reads original headers and verifies their
parent chain and fresh final anchor; it does not decrypt inventory or submit a
transaction:

```sh
go run ./rpc/benchmarks/headers \
  --rpc https://rpc.invisipay.fi/v1/chains/97 \
  --deployment /path/to/pinned-deployment.json \
  --from 143821381 --count 256 --runs 3
```

## Validation commands

```sh
go test -race -count=1 ./rpc/... ./session/...
go test -count=1 ./... ./primitives/... ./protocol/... ./prover/... ./ipp/...
go vet ./... ./primitives/... ./protocol/... ./prover/... ./rpc/... ./session/... ./ipp/...
```
