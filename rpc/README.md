# Invisi Rpc

Public chain, Broadcaster and Provenance clients.

Module version: `1.2.0`. See the [SDK repository](https://github.com/invisi-lab/invisi-sdk-go) for the six-module dependency boundary.

# Protocol RPC client

`ProtocolRpcClient` reads public indexed data and checks canonical block/tree facts through an independently configured `ChainProvider`. Discovery, deployment runtime pins, chain/version validation, depth-16 membership verification, transport limits and cancellation remain shared infrastructure. Business methods and response models separate ordinary Notes from postpaid Credit authority.

| Method | Response |
| --- | --- |
| `GetNoteUpdates(ctx, scope, range)` | `ProtocolNoteUpdates{Block, Notes}` |
| `GetCreditTransitions(ctx, scope, range)` | `ProtocolCreditUpdates{Block, Transitions}` |
| `GetNote(ctx, scope, commitment, block)` | `*NoteLookup{Output, Membership}` or nil |
| `GetCreditState(ctx, scope, commitment, block)` | `*CreditStateLookup{State, Membership}` or nil |
| `GetTransactionNoteUpdates(ctx, scope, hash)` | Availability plus `Notes` |
| `GetTransactionCreditUpdates(ctx, scope, hash)` | Availability plus `States` and `Transitions` |
| `IsNoteSpent(ctx, scope, nullifier, block)` | Historical indexed Note consumption observation |
| `IsCreditConsumed(ctx, scope, nullifier, block)` | Historical indexed Credit state consumption observation |
| `GetCreditAllocation(ctx, scope, owner, token, block)` | Historical next original-grant shard/slot |
| `GetCreditRevocationBitmap(ctx, scope, owner, token, shard, word, block)` | Historical uint256 bitmap |

`ProtocolNoteOutput` has one optional encrypted envelope. `IndexedCreditState` is one physical authority leaf with exactly two owner/spender `IndexedDelivery` envelopes in on-chain order. Responses preserve both without inferring private recipient roles. Lookups never decide the business type from a ciphertext-header bit. `GetNote` cannot return Credit and `GetCreditState` cannot return a Note.

Creation registers authority without principal funding; an optional authorized fee is pulled from the owner wallet. A Credit draw consumes N states, produces N successor states and one ordinary receiver Note. Payer totals and grant usage are gross; the Note amount is gross minus the fee accrued to the broadcaster ledger. Update consumes/replaces a state while preserving its original grant namespace and may charge an authorized owner fee. Revocation may charge an authorized owner fee and changes the public bitmap without a new state, Note, nullifier or ZK proof. Supported transition names are `CreditCreated`, `CreditDraw`, `CreditUpdated` and `CreditRevoked`; obsolete prepaid top-up/cancellation models are rejected.

Mixed `GetUpdates`, `GetOutputs`, `GetCreditOutputs`, `GetEncryptedOutputs`, `GetTransactionUpdates`, `ProtocolOutput`, `ProtocolUpdates` and `ProtocolTransactionUpdates` are removed. Migrate callers to the explicit methods/types. Range pages accept up to `MaxUpdateBlocks = 10000` inclusive deployment blocks. Detection and authenticated decryption run locally; RPC clients never transmit account detection keys. Status contains only the indexed head after deployment/version validation. Context cancellation applies to all network requests; no failure falls back to chain-log scanning.

Transaction availability is separate from EVM success. `not_found` has no mined block, receipt status or collections. `pending_index` has a canonical mined block and receipt status but nil collections. `indexed` has non-nil typed arrays, empty for unrelated or reverted transactions. Indexed results require a canonical checkpoint; unindexed states may preserve a stale observed checkpoint. Every collection item must match the requested transaction and canonical block. Conflicts, unsafe wire integers and reorganizations fail explicitly.

Public event `Data` remains `json.RawMessage`, preserving decimal-string ABI uint64/uint256 values without float conversions. Allocation shard is a `uint64`; bitmap is `*big.Int` limited to 256 bits. Wire shard/bitmap use canonical decimal strings; slot is uint16 and bitmap word uint8. Allocation begins at shard 1, slot 0. Grant revocation and state consumption are different facts.

Block binding authenticates the block reference, not the truth/completeness of arbitrary indexed values. Authenticate receipts, decrypt locally, recompute commitments and check current contract state before financial decisions. Indexed grant and consumption observations must be compared with independent contract reads when authorizing operations. Local detection adapters retain detection keys; no wallet secret is sent to the indexed service.

The server migration uses fresh index format 3 and snapshot format 2 with explicit backup/reindex/rollback. This client changes neither server databases nor wallet stores.

Verification: from the repository root, run `python3 scripts/test-modules.py` to check every module and the examples.

### High-level broadcaster quoting

`client.Quote(ctx, preparedOrSession, broadcaster.QuoteOptions{MaxFeeAmount: limit})`
asks the operation's bound SDK lifecycle to converge and authenticate the fee,
then atomically installs the quote on that unsigned handle. It returns an error;
subsequent authorization and proving continue on the same typed handle. Managed
Note and Credit targets retain their own Store transactions and worker checks.
`EstimateQuote` and `GetQuote` remain the low-level calldata RPC methods.

The RPC module does not import SDK or Session packages. SDK constructors bind a
private quote capability to the exact handle address; copied, unbound and nil
handles are rejected before networking. The capability is excluded from JSON,
never serialized into a journal, and is rebound when restoring an operation.
Custom capabilities can run only their own callback; they do not expose another
operation's private state or bypass its authorization/persistence checks.
