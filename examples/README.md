# Go examples

Each numbered directory is one runnable example with one `main.go`. Read it from top to bottom: configuration, connection, inputs and outputs, preparation, authorization, proof, submission, and confirmation are shown directly. Setup code is intentionally repeated so a reader does not need to follow a shared business workflow or choose a `--mode`.

Run commands from the repository root. Transaction examples require `--execute` before loading accounts or making network requests. Wallet and output examples are offline. Session restore, operation reconciliation, and disclosures read chain/RPC state but do not submit transactions. Only category 05 imports `session` and `session/sqlite`; other workflows use stateless `sdk` preparation.

## Setup

Edit [shared/settings.go](shared/settings.go) for the chain, Chain RPC, Protocol RPC, token, native currency, account/database paths, and prover binaries. Defaults target chain 97. Use a Chain RPC that retains historical state.

Local proofs require Barretenberg 6.0.0 and `noir-execute` compatible with Noir 1.0.0-rc.6. Set `Settings.Prover.BBBinary` and `Settings.Prover.NoirBinary`. Witnesses stay local.

```sh
go run ./examples/scripts/init-accounts
```

The initializer creates ten fresh accounts in `examples/accounts.json`, with owner-only permissions, and refuses to overwrite a file. Their order is Alice, Bob, Charlie, Dave, Eve, Frank, Grace, Heidi, Ivan, Judy. `--out /your/path/accounts.json` selects a different new file; update `Settings.AccountsFile` accordingly.

[accounts.example.json](accounts.example.json) shows the saved fields. Initialization derives receiving/viewing material; loading validates owner addresses, receiving addresses, viewing keys and duplicate accounts. Keep private account files and encrypted SQLite data outside version control.

Initialization does not fund accounts. Give public token payers enough tokens and actual transaction submitters enough native currency. Amounts in the examples are **integer token base units**. Edit the named accounts, amount or output construction directly in the example to change its scenario.

## Reading order

Each path below can be run with `go run ./examples/<path>`. Add `--execute` for transaction examples.

| Path | Scenario |
| --- | --- |
| `00-wallet/01-wallet` | Generate an offline wallet and inspect public addresses and capabilities. |
| `00-wallet/02-note-output` | Create offline outputs for `0i`, registered `0x`, identity, and no-envelope recipients. |
| `01-address-registry/01-register-bob` | Register Bob's receiving keys with external EIP-712 authorization. |
| `01-address-registry/02-update-bob` | Update Bob's registration using a fresh authorization. |
| `02-transparent-amount/01-alice-to-bob` | Alice publicly funds one private note for Bob, using caller funding. |
| `02-transparent-amount/02-alice-to-nine` | Alice funds nine outputs, one each for Bob through Judy. |
| `02-transparent-amount/03-bob-funded-by-signature` | Bob approves tokens and signs funding; Alice receives the note and submits the transaction. |
| `02-transparent-amount/04-shield-with-memo` | Alice shields a note for Bob with an explicit application memo. |
| `02-transparent-amount/05-shield-via-broadcaster` | Discover services, quote Shield, approve gross funding, sign and submit through a broadcaster. |
| `03-hidden-amount/01-cash-to-bob` | Alice funds a proved transfer to Bob entirely from public tokens. |
| `03-hidden-amount/01b-cash-funded-by-signature` | Alice signs public funding; Frank submits the proved transfer to Bob. |
| `03-hidden-amount/02-notes-and-cash` | Bob combines existing notes with one public token unit and pays Dave. |
| `03-hidden-amount/03-notes-with-change` | Bob pays Dave from existing notes, returning any excess as private change. |
| `03-hidden-amount/04-sweep-multiple-notes` | Bob withdraws all selected notes to his bound EVM owner without a spend signature. |
| `03-hidden-amount/04b-partial-withdraw-with-change` | Bob withdraws a chosen amount and retains private change when selected notes exceed it. |
| `03-hidden-amount/05-frank-sweep-for-bob` | Frank pays gas to sweep Bob's notes to Bob's bound EVM owner. |
| `03-hidden-amount/06-aggregate` | Bob and Charlie independently authorize four transfers; Alice submits their aggregate. |
| `04-credit/01-bob-create` | Bob creates a Credit for Alice. |
| `04-credit/01b-charlie-create` | Charlie creates a separate Credit for Alice. |
| `04-credit/02-alice-batch-draw` | Alice draws from grants belonging to two distinct owners in one transaction. |
| `04-credit/03-bob-top-up` | Bob tops up an unexpired Credit and extends expiry. |
| `04-credit/04-charlie-cancel` | Charlie cancels his remaining Credit, including after expiry. |
| `05-wallet-session/01-restore-session` | Restore Alice's encrypted SQLite inventory, refresh and inspect balance. |
| `05-wallet-session/02-notes-to-bob` | Alice selects/reserves notes, authorizes, proves, submits and reconciles through a session. |
| `05-wallet-session/03-recover-operation` | Resume an existing persisted operation and reconcile its transaction. |
| `06-disclosure/01-note-opening` | Export and verify a selected Bob note opening. |
| `06-disclosure/02-note-payload` | Disclose one selected Bob note's envelope and decode its application payload. |
| `06-disclosure/03-credit-opening` | Export and verify a selected Bob Credit opening. |

## Typical runs

Start with an offline example, then fund Bob with a real note:

```sh
go run ./examples/00-wallet/01-wallet
go run ./examples/00-wallet/02-note-output
go run ./examples/02-transparent-amount/01-alice-to-bob --execute
go run ./examples/03-hidden-amount/03-notes-with-change --execute
```

An exact-value input produces no change. For a partial withdrawal demonstration, first give Bob more notes than the amount withdrawn. Hidden examples select at most 12 inputs; creation/top-up examples select at most 10. These limits keep the walkthrough focused, and preparation independently validates available funds and circuit capacity.

The examples are separate scenarios, not a script to run all transactions against the same finite inventory. To try the Credit lifecycle, first give Bob and Charlie enough unspent private notes. Repeating `02-transparent-amount/02-alice-to-nine` supplies both owners. Then:

```sh
go run ./examples/04-credit/01-bob-create --execute
go run ./examples/04-credit/01b-charlie-create --execute
go run ./examples/04-credit/02-alice-batch-draw --execute
go run ./examples/04-credit/03-bob-top-up --execute
go run ./examples/04-credit/04-charlie-cancel --execute
```

Each grant starts with 20 base units, a one-day expiry, and single/daily/monthly limits equal to the grant. Draw consumes 10 units from each owner. Top-up adds 5 units, preserving payload, policy usage, shared keys and historical receiving-key snapshots.

Aggregation needs two distinct unspent notes of at least two units for each of Bob and Charlie. Each owner authorizes independently; the four transfers are proved with `Execution: "aggregate"` before Alice submits the combined public members. Applications must reserve inputs until the aggregate outcome is resolved.

Run broadcaster Shield with an appropriate maximum token fee:

```sh
go run ./examples/02-transparent-amount/05-shield-via-broadcaster --execute --max-fee 1000
```

The example discovers token-compatible clients from Protocol RPC. Applications may quote several clients and select one. The final quote includes its fee before funding authorization. Private transfers and Credit creation/top-up/cancellation support the same attachment stages; Credit draw and registration remain self-submitted.

For managed recovery, use the operation ID persisted by the session and optionally its already submitted hash:

```sh
go run ./examples/05-wallet-session/03-recover-operation --recover OPERATION_ID --transaction 0xTRANSACTION_HASH
```

This command only tracks/reconciles; it never signs or resubmits. Uncertain operations retain reservations. To supply Alice's inventory for the managed examples, Bob can fund her through `02-transparent-amount/03-bob-funded-by-signature`.

## Shared helpers and verification

Only public settings, account loading and authenticated note/Credit discovery remain under `shared`. Discovery pages real RPC outputs, validates owner/role, token and commitment, rejects conflicts, and checks nullifiers. Credit discovery binds historical keys and ciphertext to canonical receipts. Preparation rechecks membership and current spent state.

Each transaction example shows geth submission directly. Persist hashes before waiting in a production application. Stateless examples leave durable reservations and recovery to the application. Disclosure exports stay local and are never logged.

```sh
go test ./examples/...
go vet ./examples/...
go run ./examples/00-wallet/01-wallet
go run ./examples/00-wallet/02-note-output
```

These checks do not execute chain transactions. Account initializer tests use temporary fresh files and verify metadata, permissions, identity binding and overwrite refusal.

## Payment requests and offline Note delivery

IPR describes a requested output. IPN delivers one actual Note opening. These are independent data steps around the existing payment workflows; no additional executable example is needed. Continue using the existing transparent-payment or hidden-amount example for external signing, proving and transaction submission.

```go
// Recipient: publish a reusable request, for example as a QR code.
ipr, err := sdk.EncodePaymentRequest(sdk.PaymentRequest{
    ChainID: chainID, ProtocolAddress: protocolAddress, TokenAddress: token,
    Amount: amount, Recipient: sdk.InvisiRecipient{Address: bobAddress},
})
if err != nil { return err }

// Payer: use this output in the existing PrepareShield/PrepareTransfer workflow.
output, err := sdk.NewNoteOutput(ipr)
if err != nil { return err }
// Omitted request nonce means a fresh nonce is selected per payment.
// Keep the final prepared handle after any broadcaster fee/change reconstruction.

// Payer: export only Bob's output, identified by its commitment.
var ipn string
for _, final := range prepared.GetOutputs() {
    if final.Commitment() == output.Commitment() {
        ipn, err = final.ExportPaymentNote()
        if err != nil { return err }
        break
    }
}
if ipn == "" { return fmt.Errorf("payment output is missing from the final plan") }
// Submit through the existing workflow, then privately deliver ipn after confirmation.
// Generated change belongs to the payer and is not delivered to Bob.

// Bob: reconstruct and verify the exact requested payment.
note, err := sdk.PaymentNoteFromIPN(ipn)
if err != nil { return err }
if note.TokenAddress() != token || note.Amount().Cmp(amount) != 0 {
    return fmt.Errorf("payment does not match the requested token and amount")
}
verified, err := note.Verify(ctx, pool) // Correct opening and confirmed membership.
if err != nil { return err }
_ = verified
spendable, err := note.VerifySpendable(ctx, pool, bobWallet) // Also owner capability and current unspent state.
if err != nil { return err }
input, err := spendable.ToInput()
if err != nil { return err }
_ = input // Use in the next explicit PrepareTransfer; authorization remains separate.
```

The sections are separate recipient/payer steps, using the chosen protocol, token and amount from the existing examples. `pool` comes from `sdk.ProtocolFromRPC(ctx, rpc, provider)`. Simple verification also accepts spent Notes; detailed verification does not. Neither method signs. EVM recipients may export IPN with or without a chain envelope; identity recipients cannot. See the root README for wire layouts, nonce execution overrides and explicit verification connections.
