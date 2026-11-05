# Go examples

Run commands from the repository root. The seven categories follow the TypeScript SDK examples. Transaction examples require `--execute` before loading local accounts or making network requests. The offline wallet, session restore, and disclosure examples can run without that flag.

## Configure a fresh checkout

Edit the public values in [shared/settings.go](shared/settings.go): chain ID, Chain RPC, Protocol RPC, token address, native currency symbol and decimals, account/database paths, and the local prover binaries, thread count, and timeout. Use matching RPC endpoints and a Chain RPC that retains historical state. The checked-in defaults target chain 97; replace them together for another deployment.

Proof examples require Barretenberg 6.0.0 and a compatible local `noir-execute` for Noir 1.0.0-rc.6. Set `Settings.Prover.BBBinary` and `Settings.Prover.NoirBinary` to your installed binaries. Witnesses stay local.

Create ten fresh demonstration accounts:

```sh
go run ./examples/scripts/init-accounts
```

This exclusively creates `examples/accounts.json` with owner-only permissions and refuses to overwrite an existing file. Account order is Alice, Bob, Charlie, Dave, Eve, Frank, Grace, Heidi, Ivan, Judy (indices 0–9). To create a separate file, use `--out /your/path/accounts.json` and update `Settings.AccountsFile`.

The ignored account file stores `name`, `privateKey`, `address`, `receivingAddress`, and `viewingPrivateKey`; [accounts.example.json](accounts.example.json) documents the format with placeholders. Initialization derives the saved receiving/viewing material once. Loading constructs public wallets from the saved receiving addresses and validates the saved identities and keys. Recipients use their public receiving addresses; only the managed session constructs its own complete private wallet. Signing and viewing keys remain explicit local inputs.

Initialization never funds accounts. Prepare native currency for actual transaction submitters: Alice funds token approvals and payments; Bob and Charlie submit their Credit lifecycle transactions; Frank submits a relayed sweep when selected. `Settings.SuggestedSubmitterGasBalance` and `Settings.NativeCurrencySymbol` provide an editable preparation hint, not a fee estimate. Accounts used only as recipients need no native currency. Fund the token payers with your configured token as well. Keep account files and SQLite data local and outside version control.

## Run the categories

| Category | Command | Behavior |
| --- | --- | --- |
| Offline wallet | `go run ./examples/00-wallet` | Generate an ephemeral wallet; print only public addresses/keys and redacted capabilities. No account file or RPC. |
| Address registry | `go run ./examples/01-address-registry --execute` | Register Bob's saved receiving keys with an external EIP-712 signature. Add `--update` to update; `--actor 2` selects Charlie. |
| Transparent amount | `go run ./examples/02-transparent-amount --execute` | Alice approves public token funding and shields two 10-unit outputs to Bob. |
| Hidden amount | `go run ./examples/03-hidden-amount --execute --mode notes --actor 1 --recipient 0` | Discover Bob's real notes, authenticate them, select sufficient inputs, sign externally, prove, and pay Alice with automatic change. |
| Credit | `go run ./examples/04-credit --execute --mode create` | Bob creates a grant for Alice. Draw, top-up, and cancel modes use authenticated current Credit state. |
| Wallet session | `go run ./examples/05-wallet-session` | Restore Alice's encrypted SQLite inventory and print balance. Add `--execute` to select/reserve inputs, authorize, prove, send, and reconcile a transfer to Bob. |
| Disclosure | `go run ./examples/06-disclosure --mode note-opening --actor 1` | Disclose a real selected Bob note locally and verify canonical inclusion. No submission. |

Amounts are integer token base units. Long addresses come from the local account file; registry resolution is also demonstrated after registration. There are no invented note preimages or fixed fake tree positions.

### Transparent and hidden payments

Use `--actor N` for the funding/spending owner and `--recipient N` for the receiving public address. Set the same index to fund your own private notes. Transparent `--outputs N` accepts 1–14 outputs, each worth 10 base units; `--signature-funding` demonstrates a public funding grant.

Hidden `--amount N` sets the payment or partial withdrawal amount:

```sh
# Fresh private notes for Bob and Charlie, funded by Alice.
go run ./examples/02-transparent-amount --execute --recipient 1 --outputs 4
go run ./examples/02-transparent-amount --execute --recipient 2 --outputs 4

# Ordinary private payment, public cash funding, and mixed private/public funding.
go run ./examples/03-hidden-amount --execute --mode notes --actor 1 --recipient 3 --amount 10
go run ./examples/03-hidden-amount --execute --mode cash --recipient 1 --amount 10 --signature-funding
go run ./examples/03-hidden-amount --execute --mode mixed --actor 1 --recipient 3 --amount 10

# Both withdrawals use PrepareTransfer with Withdrawal; full sweep needs no wallet signature.
go run ./examples/03-hidden-amount --execute --mode withdraw --actor 1 --amount 5
go run ./examples/03-hidden-amount --execute --mode sweep --actor 1 --submitter 5
```

Mixed mode adds one public token unit to selected existing notes. Partial withdrawal returns private change; sweep withdraws all discovered selected funds to their bound owner. `--submitter 5` makes Frank the final geth sender while the spending owner remains `--actor`. With a separate sender, public funding requires `--signature-funding`; its token approval is still signed by the payer. The example limits selected private inputs to 12; consolidate or narrow selection for larger inventories.

The aggregate command requires two distinct unspent notes of at least two units for each of Bob and Charlie:

```sh
go run ./examples/03-hidden-amount/aggregate --execute
```

Each owner authorizes two transfers independently. `Prove` selects `Execution: "aggregate"`; the four exported public members are combined and Alice submits the aggregate. Applications must reserve the selected notes until the aggregate outcome is resolved.

### Two-owner Credit draw

Start with enough real private notes for Bob and Charlie. The following sequence uses them as two grant owners and Alice as their common spender:

```sh
go run ./examples/04-credit --execute --mode create --actor 1 --amount 20
go run ./examples/04-credit --execute --mode create --actor 2 --amount 20
go run ./examples/04-credit --execute --mode draw --amount 10
go run ./examples/04-credit --execute --mode top-up --actor 1 --amount 5
go run ./examples/04-credit --execute --mode cancel --actor 2
```

Create/top-up/cancel default to Bob (`--actor 1`). Draw defaults to Alice (`--actor 0`), discovers her spender envelopes, and requires grants from two different owners in one batch. A grant starts with a one-day expiry and per-grant single/daily/monthly limits equal to its amount. Draw validates current policy usage and chain time. Top-up preserves raw payload, policy usage, shared keys, and historical receiving-key snapshots; cancellation also supports expired owner grants.

### Session recovery and disclosures

The managed session injects private capabilities and a geth signer only at `Connect`. It persists reservations and operation status in the configured encrypted SQLite file. Resume an interrupted operation with its persisted ID; an uncertain send must keep its inputs reserved:

```sh
go run ./examples/05-wallet-session --recover OPERATION_ID --transaction 0xTRANSACTION_HASH
```

Disclosure modes are `note-opening`, `note-payload`, and `credit-opening`; select the owner with `--actor`. Payload decoding is explicit through `core.FundsV1`, separate from envelope decryption. The sensitive export is round-tripped locally and never printed. A canonical receipt and confirmation check verifies the selected output.

Discovery helpers page genuine RPC outputs in ranges of at most `sdk.MaxUpdateBlocks` (10000 blocks), authenticate owner/role, token and commitment during decryption, reject conflicting records, and check spent nullifiers. Credit discovery additionally binds role ciphertext and historical key snapshots to canonical receipts. Preparation independently rechecks current spent state and membership before proving. Stateless examples leave durable reservations and transaction-hash persistence to the application; their geth submission is visible directly in each `main` function.

## Local verification

```sh
go test ./examples/...
go run ./examples/00-wallet
```

The initializer tests use temporary fresh accounts and verify ten-account metadata, file permissions, identity binding, and refusal to overwrite. No funded account file or network is needed for these checks.

## Output factories and broadcaster submission

Every payment example constructs complete outputs with `sdk.NewNoteOutput` before preparation. Run `go run ./examples/00-wallet/note-output` for the fully offline `0i`, lazy registered `0x`, precomputed identity and explicit no-envelope variants. Default EVM envelope resolution happens during preparation; no application resolver function is needed.

The full broadcaster walkthrough is `go run ./examples/02-transparent-amount/broadcaster --execute --max-fee 1000`. Set the fee limit in token base units for your deployment. It discovers compatible services from Protocol RPC, prepares Shield, obtains the fee-bearing quote, approves final gross funding, signs EIP-712 externally, and submits through the chosen broadcaster. It does not create a signer inside Invisi or require a proof for Shield. You may quote multiple discovered clients before selecting an operator. The same attachment stages apply to private transfers, Credit creation/top-up/cancellation; Credit draw and registration remain self-submitted.
