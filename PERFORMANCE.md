# Wallet state refresh performance

The default `WalletSession.RefreshState(ctx)` still retains transaction material
for offline preparation. Hosts that only need inventory may call
`RefreshStateWithOptions(ctx, StateRefreshOptions{InventoryOnly: true})`.
This preserves deployment/version/code checks, every Note/Credit consumption
check, Credit revocation, pending reconciliation and final chain identity and
canonical fences. It discards older preparation material. Later preparation
authenticates only the selected inputs online; submission checks remain live.

Independent consumption and Credit status reads use at most four workers at one
numeric block/hash. Duplicate nullifiers and exact owner/token/GrantID namespaces share a
read within that operation. Workers never mutate the wallet or scan cache;
results merge only after all reads and the canonical fence succeed. Revisions
and public-history generations remain checked before committing. Historical
spent/consumed records are still checked, so reorganization recovery does not
depend on irreversible cached booleans.

`ProtocolRpcClient.GetMemberships(ctx, scope, inputs, block)` verifies at most
256 ordered paths. Status/finality and roots are shared within the call, with
one common root per tree covering the highest requested leaf. Explicit root
conflicts fail. Each response binding, field and Merkle path is authenticated.
All distinct used root anchors and the target are reread before returning any
result. Evidence is not cached across calls. Single `GetMembership` remains
available. Both full snapshot prewarming and live selected-input preparation
use batches; cached paths require current on-chain root acceptance even at the
same block. Offline preparation continues to replay its historical snapshot
and requires live submission validation before spending.

## Reproduce

Use the repository's local workspace so all modules, including `paymentrouter`,
resolve to this checkout:

```sh
go test ./rpc -run '^$' -bench BenchmarkMembershipBatch -benchtime=20x
go test ./session -run '^$' -bench BenchmarkStateRefreshInventory -benchtime=5x
go test -race ./rpc/... ./session/...
```

The membership fixture requests two distinct leaves in one tree and uses a real
HTTP client/server plus real Merkle hashing. The baseline invokes the existing
single API twice; the batch authenticates the same leaves together. Indexed
status and root-list requests fall from two of each to one of each; both paths
are still checked. The state-refresh fixture contains 32 Notes in separate
trees, includes real wallet storage cloning/validation, and clears retained
preparation material before each sample. Full refresh fetches one shared status
per sample; inventory-only fetches none for transaction material. Every
nullifier remains checked in both modes.

Example local Apple M2 Max samples on 2026-11-21:

| Fixture | Baseline/full | Batch/inventory-only |
| --- | ---: | ---: |
| Two membership paths | 1.245 ms/op | 0.620 ms/op |
| Status requests for those paths | 2/op | 1/op |
| Root-list requests for those paths | 2/op | 1/op |
| 32-Note state refresh | 197.022 ms/op | 39.248 ms/op |
| Prewarm status requests | 1/op | 0/op |

These loopback/provider fixtures include local CPU and transport overhead, not
remote-chain latency. Timings are observations, never pass/fail thresholds or
production speed claims. Full-CLI BSC97 measurements are maintained by the
Facilitator host separately; native node latency, indexing lag, proving and
Broadcaster quotation remain part of that end-to-end workload.
