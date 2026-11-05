# Invisi Go SDK

Go client for Invisi Protocol, aligned with the current TypeScript SDK workflows. Requires Go 1.26.1 or newer.

```sh
go get github.com/invisi-lab/invisi-sdk-go
```

Start with the [configurable examples](examples/README.md). They cover offline wallets, address registration, transparent payments, hidden transfers and withdrawals, Credit, managed wallet sessions, and selective disclosure. Example settings expose the network, native currency, token, RPC endpoints and prover paths. A local initializer creates ten fresh accounts without overwriting existing keys; no funded account is included.

## Public wallets and external signing

Use `sdk.PublicWalletFromAddress(receivingAddress)` for a public identity. A complete `0i` address contains the owner, viewing public key and clue key and resolves offline. An EVM destination needs registered receiving keys only when it requests an envelope. The output recipient selects Protocol RPC (default) or Chain RPC lookup during preparation; `InvisiOptions` accepts neither a global lookup selector nor caller resolver injection. Public wallets contain no spending or viewing secrets.

Create an `sdk.Invisi` with `Provider`, `ProtocolRPC` and optionally `Prover`. `sdk.ProtocolFromRPC` validates the selected deployment against Chain RPC. Applications select a trusted Protocol RPC and can pin deployment metadata independently.

Detached preparation never creates a Store or imports a private wallet. Supply a public wallet, explicit authenticated notes and payment outputs. The stages are:

1. Prepare and review the transaction.
2. Export a portable wallet signing request and sign it with the application's signing tool.
3. Attach the signature with `WithWalletSignature`.
4. Generate and independently verify the proof with `Prove`.
5. If public funding uses a signature, export its separate request and attach the payer's signature with `WithFundingSignature`.
6. Export calldata and submit it with ordinary go-ethereum APIs.

Requests use JSON-safe EIP-712 domain/message values and omit `EIP712Domain` from application types. Wallet spending authorization, public funding authorization, ERC20 allowance and gas submission are separate roles. Nonces, deadlines, signer identity and accepted roots are rechecked. EOA signatures support standard and compact formats; funding verification also supports ERC-1271 wallets.

`Prove` accepts strongly typed `sdk.TransactionProveOptions`, including an optional backend and single/aggregate execution. Recursive aggregation snapshots all members before invoking a backend. Errors from preparation use `sdk.InvisiError` with a stable code and operation/recipient details where applicable; inspect them with `errors.As`.

## Payments and Credit

`PrepareShield` creates transparent-amount deposits. `PrepareTransfer` supports private inputs, public funding, private outputs and an optional public `Withdrawal` in one entry point. All explicit inputs are consumed. Inputs must belong to the same spending owner. Automatic change uses a fresh private output.

A full withdrawal to each note's bound owner, without public funding or private outputs, needs viewing permission rather than an arbitrary-spending signature. Partial withdrawals and redirected payouts require owner authorization. There are no separate `PrepareWithdraw` or `PrepareSweep` APIs.

Credit supports creation, draw, top-up and cancellation, including private-note funding and change. A spender can consume grants from different owners together. Single/daily/monthly limits are enforced per grant; nil or zero creation limits disable that limit. Nil top-up limits inherit the current policy, while explicit zero disables it. Usage is preserved. Day/month checks use canonical chain timestamps and contract period rules. Recovered remainders retain financial state and historical receiving-key snapshots.

Registry preparation is exposed through `Invisi.PrepareRegistration`. Export the registration signing request and attach its signature, or explicitly use the direct owner-caller route. Registration updates are explicit; an existing registration does not silently become a first registration.

## Independent note outputs

Every Shield/Transfer business output is a `NoteOutput` created with `sdk.NewNoteOutput`; raw payment objects and zero-value outputs are rejected. The factory requires chain, Protocol, token and amount up front. Preparation rejects a different context instead of rewriting the output.

```go
output, err := sdk.NewNoteOutput(sdk.NoteOutputOptions{
    ChainID: chainID, ProtocolAddress: pool.Context().ProtocolAddress,
    TokenAddress: token, Amount: amount,
    Recipient: sdk.InvisiRecipient{Address: bobAddress, Nonce: nonce},
})
// Check err, then pass Outputs: []sdk.NoteOutput{output} to preparation.
```

Recipients are sealed typed variants:

- `InvisiRecipient{Address, Nonce}` accepts a complete `0i` address and encrypts offline. Nonce is optional; nil securely samples a canonical field element, including zero as a valid explicit value.
- `EVMRecipient{Address, Nonce, Envelope, AddressLookup}` accepts an EVM owner. A nil Envelope defaults to true. `DeliveryPending()` is true on the offline intent; preparation uses `"protocol-rpc"` (default) or `"chain"` to resolve registered keys and freezes delivery. Unregistered destinations or lookup failures reject preparation. A pointer to false skips registration and delivery entirely.
- `IdentityRecipient{Identity}` accepts a nonzero canonical per-note identity (`common.Hash`), with no nonce, wallet opening or envelope. `Note()` returns nil; the recipient must retain its own opening. Verify a merchant-supplied commitment independently with `output.AssertCommitment(expected)`.

Payload is an outer `NoteOutputOptions` field allowed only with envelopes. Getter results are detached copies. Factory objects are in-memory intents; retain business parameters/complete notes to recreate them after a process restart. Managed operation plans use durable Store snapshots independently of factory objects. Empty Funds delivery always means both ciphertext and clue are empty. Decryption skips these slots without changing sibling leaf positions; active Credit owner/spender envelopes remain mandatory.

## Optional broadcaster submission

Discover registered operators through `pool.GetBroadcasters(ctx, sdk.BroadcasterOptions{TokenAddress: token})`. It returns clients that advertise support for that token; unreachable or unsupported operators are skipped. Discovery uses the Protocol's pinned registry and Protocol RPC index, not an endorsement of an operator. `sdk.NewBroadcasterClient` also accepts an explicit public registration configuration. The service uses JSON-RPC 2.0 positional parameters, matching Protocol RPC.

The stage sequence matches TypeScript:

1. Prepare business intent, then inspect `Preview()` or `EstimateGas()`.
2. Call `prepared.WithBroadcaster(ctx, client, sdk.WithBroadcasterOptions{MaxFeeAmount: limit})` before signing/proving. The returned handle contains the fee-bearing frozen plan and authenticated quote.
3. Review `quoted.Preview().FeeAmount`; existing private change pays first when sufficient. Otherwise an explicitly selected public payer authorizes the gross funding amount. Private-only managed sessions may select additional available notes; detached inputs remain explicit. Cancellation deducts the fee from its public refund.
4. Complete wallet authorization and proving when required, then sign the final EIP-712 public funding request externally. Shield requires no ZK proof.
5. Call `ready.SubmitBroadcaster(ctx, client)`, retaining the operation/request ID before sending. On timeout, recover through the same client's status API rather than issuing a conflicting spend. Quote and final calldata remain bound to chain, Protocol, token, registered signer, fee recipient and deadline.

Transfer, Shield, Credit creation, top-up and cancellation support this flow. Credit draw, registration and aggregates do not. Fee estimation rebuilds the plan in bounded attempts and rejects a non-converging or over-limit fee. An optional `FundingPermit` callback signs a permit for the final gross amount; funding signatures do not replace ERC20 allowance/permit. Business output identities and delivery stay frozen during repricing. Existing self-submission via geth remains available. Successful fees accrue in Protocol custody; generated contract bindings expose the balance and withdrawal methods.

Managed sessions reserve additional quoted inputs atomically; failed quotes restore the original reservations. Potentially externalized submissions remain uncertain until canonical receipt reconciliation. Both managed Operation and detached PreparedTransaction expose quote/intent exports and the same broadcaster attachment stages.

See the complete [Shield broadcaster example](examples/02-transparent-amount/05-shield-via-broadcaster/main.go) and the [offline output variants](examples/00-wallet/02-note-output/main.go).

## Envelopes and application payloads

Funds and Credit deliveries use current version-1 envelopes, identified by header `0x01` and `0x81`. Deliveries are bounded by 4096 bytes. Historical formats are rejected by local decryption. Protocol RPC transports opaque deliveries rather than interpreting application payloads.

`core.Note.Payload` and `core.CreditNote.Payload` contain raw application bytes. Decryption authenticates the SDK envelope and financial fields and preserves these bytes without choosing an application codec. Applications explicitly call `core.FundsV1.Encode` / `Decode` or `core.CreditV1.Encode` / `Decode` for the built-in memo payload, or use their own codec. The built-in payload contains only memo metadata; it does not assert sender identity or a return address. Payloads do not change note commitments.

Top-up and draw payload overrides preserve the parent payload when nil; a non-nil empty slice clears it. Selective disclosures can reveal one note opening or authenticate its encrypted payload without exporting an entire viewing key. Verify disclosures against canonical receipts; explicit disclosure exports are sensitive.

## Managed wallet sessions

The stateless SDK does not import the session package or a database driver. Both paths use one internal transaction engine; managed persistence and lease checks are injected by the session package. There is no separate Go module or independently versioned session release.

`sdk.WalletFromPrivateKey` restores a complete local wallet for a managed session. It derives separate viewing and detection capabilities from the application's existing EVM key; mnemonic derivation remains the application's responsibility. Private wallet and capability objects redact default formatting and JSON.

Import `github.com/invisi-lab/invisi-sdk-go/session` for managed wallets. Compose `session.Connect(ctx, session.Options{Invisi: invisi, Wallet: wallet, Protocol: pool, Store: store})`. Supply private capabilities, a geth signer and optional funding signer on these session options; the shared `sdk.InvisiOptions` contains infrastructure only. Session operations provide local authorization, input reservations, submission tracking, reconciliation and recovery. Only this workflow uses a Store.

`session.NewMemoryWalletStore` is ephemeral. For persistence, combine `session.NewEncryptedWalletCodec` with `session/sqlite.NewWalletStore` and an application-owned database and encryption key. Successful atomic commits advance a revision; failed commits do not. SQLite authenticates the stored revision. Private authorization signatures are not persisted.

Refresh defaults to `sdk.MaxUpdateBlocks` / `rpc.MaxUpdateBlocks` (10000 inclusive blocks); `PageSize` may reduce that limit. Empty intermediate pages skip repeated inventory work, while an entirely empty range retains one final spent-state check. Every page still checks its block anchor, and output count, response size and detection delivery limits remain separate.

Refresh authenticates RPC candidates against canonical receipts, membership and spent state. Complete discovery coverage is separate from blocks observed during reconciliation. Partial or failed scans do not claim complete coverage. Refresh is serialized per session; revision checks prevent a stale scan from overwriting another session's reset. Reorganizations require an explicit reset refresh.

`GetHistory` returns copied incoming notes and outgoing business summaries, including public funding, private outputs and withdrawals, without proof witnesses or recovery secrets. `Subscribe` queues notifications only after successful commits. Callbacks run serially outside operation locks and may read the session; unsubscribe cancels pending callbacks. Notifications are local to that session instance. Applications own refresh scheduling and UI updates.

A timeout does not release potentially externalized inputs. Persist a broadcast hash, reconcile the exact calldata and canonical receipt, and resume the saved operation after interruption. Use ordinary geth receipt waiting; the SDK does not hide a transaction submission service.

## Packages

| Package | Responsibility |
| --- | --- |
| `sdk` | Stateless wallets, preparation, signing requests, registry and disclosure |
| `session` | Managed operations, inventory, history, refresh, reservations and recovery |
| `session/sqlite` | Encrypted persistent wallet state |
| `core` | Wallet KDF, addresses, note encryption, payload codecs and FMD |
| `protocol` | Deployment context, statements, calldata and generated contract bindings |
| `rpc` | Deployment discovery, registered keys, broadcaster profiles, outputs, candidates and membership |
| `broadcaster` | Token capability probes, bound fee quotes, delegated submission and status |
| `prover` | Authenticated proof statements and recursive aggregation |
| `prover/native` | Local witness execution, generation and independent verification |

Chain RPC supplies contract reads, canonical blocks/receipts and submission. Protocol RPC supplies indexed public records and candidate filtering. Detection keys only delegate filtering; they do not decrypt notes or authorize spending. Applications remain responsible for trusted endpoints, keys, reservations in detached workflows and durable broadcast tracking.

## Single-transaction indexed updates

Protocol RPC service 1.1.4+ supports one-transaction lookup with an independently configured Chain RPC provider:

```go
updates, err := protocolRPC.GetTransactionUpdates(ctx, protocol.Context(), transactionHash)
if err != nil { return err }
if updates.Status == sdk.TransactionIndexed {
    fmt.Println(updates.Block.Number, *updates.ReceiptStatus)
    fmt.Println(updates.Outputs, updates.CreditTransitions)
}
```

`ProtocolTransactionUpdates` distinguishes `not_found` (no mined receipt; unknown or still pending), `pending_index` and `indexed`. Outputs and Credit transitions stay nil until indexed; indexed empty slices mean no matching updates. `Block` and `ReceiptStatus` are nil only for `not_found`. Receipt status 0/1 independently describes EVM reversion/success.

For unavailable data, `IndexedHead` is an observed checkpoint and may remain on a stale fork. Mined transaction blocks are checked through the caller's Chain RPC; indexed results additionally require canonical checkpoint/output references and a final block recheck. The ordinary `GetStatus` call retains its strict canonical checks. Opaque outputs preserve both Credit role envelopes and deduplicate exact deliveries. Transition `Data` contains raw public JSON fields without floating-point conversion; decode it explicitly and authenticate receipts before applying financial state.

The method does not poll, fall back to ranges, decrypt payloads or mutate a wallet Store. Context controls cancellation. Result types and availability constants are exported from both `rpc` and `sdk`.

## Proving and verification

Native proving requires BB 6.0.0 and `noir-execute` for Noir 1.0.0-rc.6. Configure `native.NativeOptions.BBBinary` and `NoirBinary`, or put matching tools on PATH. Tool versions and circuit/VK identities are checked. There is no Node runtime dependency and no automatic tool download.

```sh
go test ./...
go test -race ./...
go vet ./...
```

Real native integration tests additionally use `BB_BINARY_PATH` and `NOIR_EXECUTE_PATH`. Cross-language fixtures check current TypeScript signing, wallet, encryption and disclosure formats. [CHANGELOG.md](CHANGELOG.md) describes the breaking migration from 1.1.3.
