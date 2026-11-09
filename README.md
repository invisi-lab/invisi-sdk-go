# Invisi Go SDK

Go client for the current Invisi Protocol. Requires Go 1.26.1 or newer.

```sh
go get github.com/invisi-lab/invisi-sdk-go
```

The SDK remains in the v1 series. Source synchronization does not publish a new module version or upgrade deployed contracts.

## Packages

| Package | Responsibility |
| --- | --- |
| `sdk` | Stateless Note payments, public/private wallet capabilities, registration, IPR/IPN and Note disclosure |
| `core` | Wallet derivation, addresses, ordinary Note cryptography, FMD and application payload codecs |
| `credit` | Separate postpaid grant State, policy/usage, deliveries, lifecycle preparation, proof requests, paid broadcasting and grant audit disclosure |
| `session` | Managed Note inventory, canonical refresh, encrypted persistence, journals and notifications |
| `credit/session` | Independent managed grant projection, cross-role reservations and recoverable Credit journals; composed through `WalletSession.Credit()` |
| `session/sqlite` | Authenticated encrypted wallet snapshots in an application-owned database |
| `protocol` | Note statements/calldata, deployment validation and registry contracts |
| `protocol/contracts` | Generated bindings for the six current contract ABIs |
| `rpc` | Separate typed Note and Credit data APIs, with independent Chain RPC checks |
| `prover`, `prover/native` | Authenticated statements, native recursion, local witness execution and independent verification |
| `broadcaster` | Registered operator clients, authenticated fee quotes, canonical intent hashes and idempotent submission |
| `gas` | Exact documented benchmark references with provenance and release identity |

All packages share one Go module and release. Import subpackages under `github.com/invisi-lab/invisi-sdk-go`.

## Wallets and infrastructure

`sdk.PublicWalletFromAddress` imports public receiving identity only. Complete `0i` addresses resolve offline; an EVM recipient requesting an envelope needs registered receiving keys. `sdk.WalletFromPrivateKey` restores local spending/viewing/detection capabilities from an application-owned EVM key. Mnemonic derivation is the application's responsibility.

Create an `sdk.Invisi` with `Provider`, `ProtocolRPC` and optionally `Prover`. `sdk.ProtocolFromRPC` authenticates the selected deployment against Chain RPC. Protocol RPC uses the `/v1` API; this does not change the independent on-chain `VERSION()` value. Applications choose trusted endpoints and can pin deployment metadata.

Spending authorization, public ERC20 funding, allowance/Permit, detection, viewing and gas submission are separate roles. A full Note viewing capability is not purely read-only: it also permits consumption tracking and a full owner-bound sweep.

## Ordinary Note payments

Every business output is an immutable `NoteOutput` created with `sdk.NewNoteOutput`. The factory accepts `PaymentRequest` or an IPR string and binds chain, Protocol, token, amount and recipient before preparation.

Recipients are typed variants:

- `InvisiRecipient`: complete receiving address; delivery can be encrypted offline.
- `EVMRecipient`: EVM owner; envelope lookup uses the selected Protocol RPC or Chain RPC registry. Explicitly disabling the envelope permits an unregistered destination.
- `IdentityRecipient`: canonical per-note identity, with no recoverable owner/nonce opening or envelope.

`sdk.Invisi.PrepareShield` creates transparent-amount deposits without a proof. `PrepareTransfer` combines explicit private inputs, optional public funding, private outputs and an optional `Withdrawal`. Every selected input is consumed. Automatic change uses a fresh Note.

The detached sequence is prepare → inspect → export portable wallet signing request → attach external signature → prove and independently verify → attach separate public funding signature if needed → export calldata → submit with ordinary geth APIs. Full owner-bound sweeps without funding, private outputs or fees need viewing capability; partial or redirected spending requires owner authorization.

Managed Note operations use `session.Connect`, reservations and durable journals. No Store is created or consulted by detached preparation.

## Postpaid Credit

`credit.State` is authority to debit an owner's wallet, not a prepaid token balance. It keeps owner/spender identities, token, the original revocation shard/slot, nonce, expiry, a grant-scoped shared key, four limits and five usage fields. It has no payout, refund, remainder balance or Note funding fields.

- `PrepareCreate` registers private authority without depositing principal or spending Notes.
- `PrepareDraw` consumes N grant states, groups sorted owner payer receipts, creates N successor states and one ordinary Note for any recipient identity.
- `PrepareDrawToNote` additionally freezes that Note opening and public receiving keys, so fee quotation can preserve owner/nonce/payload while rebuilding its net amount and envelope.
- `PrepareUpdate` changes policy/expiry while preserving usage and immutable grant identity/namespace.
- `PrepareRevoke` uses the original owner/token/shard/slot reference. It needs no latest merchant state and no ZK proof.

The shared Note Tree is a storage/membership primitive. A Credit state remains a distinct business model from an ordinary spendable Note.

Policy limits are single/day/month; zero disables only that check. State carries uint128 `AvailableCredit` and an independent `Unlimited` flag. Finite zero is exhausted; a finite draw subtracts gross debit. Unlimited grants keep zero balance and still accumulate daily/monthly/total usage. Draw and recovery use canonical chain time. State recovery authenticates the masked **gross debit**, deterministic next nonce and successor commitment; unknown child payload is not copied from the parent.

`credit.DeliveryState` retains both receiving-key snapshots. Current Credit envelopes use `0x82`; older financial models/envelopes are rejected. `Payload` is opaque authenticated application data outside commitments. `credit.MemoV1` is an optional memo codec independent of the envelope version.

### Authorization, fees and proving

Create/update/draw use v4 wallet authorization domains and direct local proofs. `ExportSigningRequest` returns portable JSON-safe EIP-712 data without `EIP712Domain` in application types; `WithWalletSignature` accepts standard or compact EOA signatures. `WithAuthority` is the explicit raw public-key/circuit-signature boundary.

Revoke uses its own fee-bound owner EIP-712 signature, including EOA/ERC-1271 support through `WithRevocationSignature`. A direct owner caller may omit the signature. Revocation remains available during protocol and circuit pauses and cannot claw back an already issued Note.

Every Credit request carries `FeeRecipient` and `FeeAmount`: either a complete zero pair or a nonzero recipient and positive uint128 fee. Create/update/revoke pull fees from the owner's ERC20 wallet. Draw debits and usage are gross; the ordinary Note receives `totalAmount - feeAmount`. Allowance and token balance remain separate from wallet authorization.

`Prove` creates a local witness, uses an authenticated statement, and independently verifies the resulting proof before `Calldata` becomes available. Recovery intents omit private wallet authorization and separate revoke signing material. Final public calldata, including a submitted revoke signature, is retained in the encrypted journal for exact reconciliation and idempotency. Explicit intent exports contain private grant data and belong only in authenticated encrypted storage.

## Paid broadcaster flow

Note operations retain their `WithBroadcaster` flow. All four Credit operations support:

1. Prepare the business request.
2. Select `prepared.WithBroadcaster(ctx, client, credit.WithBroadcasterOptions{MaxFeeAmount: limit})` before authorization/proving.
3. Inspect `Preview()`, including fee and ordinary Note net amount; prepare the required owner/payer ERC20 allowance.
4. Authorize and prove, or attach the revoke owner signature.
5. Persist a submission ID, then `SubmitBroadcaster` with the original client.
6. Authenticate canonical receipts independently of the operator's status response.

Fee convergence is bounded and fails without modifying the original request on over-limit/nonconverging quotes. Grant gross debits and successor state commitments remain fixed during draw repricing. Enveloped draws need `PrepareDrawToNote` to reencrypt the final net Note safely.

Quotes bind chain, Protocol, token, registered signer, fee recipient/amount, deadline and canonical public intent. Only proofs and the operation's separate signature slots are stripped from the intent hash. Final calldata must match the quote. After a timeout, query the same operator with the same request ID rather than creating a conflicting spend.

## Managed grant state and recovery

`walletSession.Credit()` composes `credit/session` with the same encrypted WalletStore. Note and grant inventories/journals remain separate. Both roles of one physical grant reserve together. Current allocation, bitmap, consumed nullifiers, registered historical roots, UTC draw window and ERC20 gross/fee allowance are checked through the session's pinned provider.

The managed flow provides `Prepare`, local or external authorization, `Prove`, `Calldata`, paid broadcaster submission, `Cancel`, `Resume` and canonical `Reconcile`. Externalization and exact calldata are committed before submission; transport failures retain reservations. Resume replaces the old worker lease and independently verifies saved proofs. A canonical reverted transaction can release reservations; an unconfirmed operator result cannot.

Use `walletSession.CreditReconciler(confirmations)` to bind the actual transaction destination, chain and exact calldata to canonical receipts and successor events. `WalletSession.Refresh` separately discovers Notes and Credit states, checks current consumption/revocation and restores unresolved reservations after reset. Reorganizations fence old workers without discarding externalized journals.

`GetHistory` on each component returns business summaries without proof witnesses or signing secrets. Credit summaries omit grant recovery keys and nonces. Top-level `Subscribe` notifies after successful Store commits, serially outside operation locks. The application owns refresh scheduling and cross-process notifications.

## Typed Protocol RPC and receiving

Note methods are `GetNote`, `GetNoteUpdates`, `GetTransactionNoteUpdates`, `DetectNotes` and `IsNoteSpent`. Credit methods are `GetCreditState`, `GetCreditTransitions`, `GetTransactionCreditUpdates`, `DetectCreditStates`, `IsCreditConsumed`, `GetCreditAllocation` and `GetCreditRevocationBitmap`.

One indexed CreditState is one physical leaf with two opaque role deliveries. Credit records never pass through the Note APIs. Public grant references/bitmap are distinct from private spender, terms and usage. Index availability (`not_found`, `pending_index`, `indexed`) is separate from chain receipt success/reversion.

RPC candidates are authenticated against canonical receipts, insertion positions, membership and chain state before becoming spendable inventory. FMD is a filter, not proof of completeness or payment. Discovery coverage is separate from blocks merely observed during reconciliation. A reorganization needs an explicit reset refresh.

`credit.Disclose` creates a grant-scoped audit disclosure without wallet signing keys or recipient Note openings. Its explicit `Export` is sensitive; `Verify` requires an independently accepted commitment anchor.

## IPR and IPN

`sdk.EncodePaymentRequest` / `DecodePaymentRequest` use the cross-language `ipr:` wire format for ordinary Note output intent. `NewNoteOutput` freezes execution-specific nonce/lookup choices without rewriting the request's deployment context.

`TransactionOutput.ExportPaymentNote` and `NoteOutput.ExportPaymentNote` export one definite ordinary Note opening as `ipn:`. `PaymentNoteFromIPN` recomputes identity/commitment offline. `Verify` checks canonical confirmed membership; `VerifySpendable` also checks the actual owner's spending capability and current nullifier status, then provides `ToInput`.

IPN contains no viewing/spending key, ciphertext, payload or membership path. Export is not a claim of payment. Identity-only outputs and Credit states cannot export IPN. Deliver only the recipient's final output after confirmation, never the payer's change.

## Gas references

`gas.Reference(shape)` matches a complete documented benchmark shape and returns a detached result with source hash, docs commit, conditions and measured median. Unsupported shapes return unavailable; there is no arbitrary linear fallback or current aggregate gas claim based on historical private-pin experiments.

Credit measurements must match the packaged manifest/format identity, raw Credit manifest SHA256 (`creditManifestSha256`) and fee mechanism. Rows without the exact manifest hash are unavailable. The current catalog contains 58 measured Credit references for the available-quota model and 55 Note references. Historical rows remain identifiable; they cannot impersonate a new release. Use Chain RPC `eth_estimateGas` on the final calldata and real caller for actual submission.

## Development and validation

```sh
go test ./...
go test -race ./...
go vet ./...
python3 scripts/test-sync-reference.py
```

Native integration tests additionally use `BB_BINARY_PATH` and `NOIR_EXECUTE_PATH`. The backend checks exact Noir/BB versions and authenticated artifact/VK identity, does not automatically download tools and never uploads witnesses.

Maintainers regenerate bindings with `INVISI_PROTOCOL_ROOT=/path/to/invisi-protocol go generate ./protocol/contracts`. `python3 scripts/sync-reference.py --proving-only` authenticates the current 28 daily Note circuits, 34 aggregate outers and 15 direct Credit routes, with recursive certificates for C1..C10. User manifest4/format7, aggregate receipts10 and Credit manifest7/format8 have separate identities.

`State.Update(Change, deadline)` applies mutually exclusive uint128 increase/decrease deltas to the authenticated input balance. `Change.Unlimited == nil` preserves the mode; entering unlimited clears the balance and requires zero deltas, and leaving unlimited starts at zero plus this increase. Update preserves historical usage. Owner-pulled create/update/revoke fees do not consume available credit. `PrepareUpdate` freezes the original change; quotes never apply the delta again. An explicit new preparation applies that change to the latest accepted state; an externalized journal always resumes its original calldata.

Current configurable examples live under [examples](examples/README.md). Real native proof tests establish cryptographic interoperability; mocked operator/provider tests establish stage and journal behavior and do not claim real token accounting.
