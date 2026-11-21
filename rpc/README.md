# Invisi Rpc

Public chain, Broadcaster and Provenance clients.

Module version: `1.2.0`. See the [SDK repository](https://github.com/invisi-labs/invisi-sdk-go) for the six-module dependency boundary.

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

### Ordinary Core broadcaster quoting

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

### PaymentRouter relay quoting

For Router payments use `EstimateRouterQuote(ctx, coreData, envelope)`,
`GetRouterQuote(ctx, coreData, envelope, trust)` and
`SubmitRouterTransaction(ctx, coreData, envelope, signedQuote)`.
`coreData` must exactly match the Core `transact` calldata embedded in that
phase's `RouterEnvelope`. The envelope contains `to`, canonical `data`,
`value: "0"` and optional ordered `authorizationList`. The existing JSON-RPC
methods take an optional `router` field; submission sends `quote` and `signature`
separately. New Router quotes use `requestHashFormat: "payment-router-intent-v1"`
and direct execution. Ordinary Core quote verification does not apply.

Prepare the Core business request and fee, export `prepared.ExportRouterDraft(ctx)`,
then call `BuildPaymentDraft` and, for native sponsorship, `BuildSponsoredDraft`.
`EstimateRouterQuote` and the formal `GetRouterQuote` accept empty proof and wallet
signature fields. Complete fee/route convergence before the formal quote. Native
drafts still bind the player, enabled/pinned account implementation, account and
Router nonces, deadlines, and any EIP-7702 chain/delegate/EOA nonce metadata.

After quotation, prove the same Core business request, rebuild its actual Core
request hash, sign the final Router/account payment and delegation, and rebuild
the final envelope. `RouterIntentHash` excludes proof and wallet signature bytes
while binding all business fields and authorization order/metadata. `EnvelopeHash`
remains the raw full-envelope identity helper. Proof/signature completion preserves
the semantic quote identity; fee, route, nonce, deadline or other business changes
require a new quote. Each phase must provide its own exact embedded Core calldata.

ERC20 submission targets the enabled Router with signed `payTokenFor`; native
submission targets the delegated player with signed `executeNativePayment`
wrapping `payNative`. The player's balance supplies principal; the relayer pays
gas with zero outer value. Installation carries a signed EIP-7702 authorization;
an existing approved delegation can omit it. The service validates the actual
proof-dependent hashes, wallet authorization and live state, then estimates and
simulates the complete final outer call at submission within the signed caps.
Draft quotation does not require that executable-call simulation.

`EstimateRouterGas` and `prepared.EstimateRouterGas` expose
`payment-router-gas-v1`, source `model`, as conservative engineering allowances
above the Core preview. These are not receipt-calibrated swap measurements. Opaque
0x returns `zeroex-route-gas-unavailable` without a trusted independent route gas
budget; the service uses operator policy and never trusts a client gas hint.
Router pricing uses current network fees and an operational maximum gas cap;
`maxGas` is a signed execution bound. Ordinary Core cap-based pricing policy and
Core/circuit/contract interfaces remain unchanged.

Fees are denominated in the Core output token and included in exact output.
Scheduling/aggregation are unsupported. Pure public Router funding has no
private sources/nullifiers and needs no Provenance source proofs: keep
`proofSidecars` empty under enabled evaluation even when `provenanceRequired`
reports enabled service policy. The Go helper intentionally omits sidecars. See the
[Router guide](../paymentrouter/README.md) and
[compiling workflows](../paymentrouter/workflow_example_test.go).

### Explicitly trusted chain provider

`NewTrustedProvider(client, DeploymentContext{ChainID: chainID, ProtocolAddress: address})`
requires a client constructed with `ClientOptions{Deployment: &pins}`. It implements
SDK `Provider`, native `HeaderByTag` and transaction-by-hash reads using only the
Protocol RPC URL. Use `client.WithProvider(provider)` when constructing a session.
This trusts the service's chain data; it does not independently authenticate it.
Deployment discovery alone does not authorize this mode. Every response checks
API `v1`, Protocol version `1`, and the complete pinned deployment. Calls are
fresh requests, with no provider cache.

Wire method: `invisi_readChainState`, one object parameter. `kind` is `header`,
`code`, `call`, `receipt`, or `transaction`. Optional fields are `block` (`latest`,
`safe`, `finalized`, or a canonical hex quantity), `address`, `from`, `data` (hex),
`gas` (integer), and `transactionHash`. Responses contain `apiVersion`,
`protocolVersion`, `deployment`, and the requested `header` + `headerHash`, `data`,
`receipt`, or `transaction` + `pending`. A missing receipt or transaction becomes
`ethereum.NotFound`. Full geth headers preserve canonical hash computation;
receipts are checked against their canonical header on both sides.

The gateway has no submission, signing, endpoint selection, or generic Ethereum
RPC forwarding. Calls allow fixed-length Protocol state selectors and ERC20
metadata/balance/allowance reads; configured address and broadcaster registries
have narrow read selectors. Simulation permits only canonical ABI-encoded
`shield`, `transact`, and `crossTreeTransact` calls to the pinned Protocol.
Calldata is capped at 128 KiB, gas at 30 million, and returned code/data at 1 MiB.
Nonzero value, fee overrides, access lists, blobs, and authorizations are not
supported by the provider.

## Contributing

See the [contribution guidelines](../README.md#contributing) for issues, pull requests, validation, and handling sensitive material.
