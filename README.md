# Invisi Go SDK

The Go SDK follows the TypeScript SDK's client model: restore a wallet, select a Protocol through Protocol RPC, connect local capabilities, refresh owned notes, then prepare, review, authorize, prove and export a transaction for geth sending. The module is `github.com/invisi-lab/invisi-sdk-go` and requires Go 1.26.1 or newer.

```go
provider, err := ethclient.DialContext(ctx, chainRPCURL)
if err != nil { return err }
defer provider.Close()

protocolRPC, err := sdk.NewProtocolRpcClient(protocolRPCURL)
if err != nil { return err }
protocol, err := sdk.ProtocolFromRPC(ctx, protocolRPC, provider)
if err != nil { return err }
wallet, err := sdk.WalletFromPrivateKey(accountPrivateKey)
if err != nil { return err }

localProver, err := native.NewNativeProver(native.NativeOptions{})
if err != nil { return err }

invisi := sdk.NewInvisi(sdk.InvisiOptions{
    ProtocolRPC: protocolRPC,
    Provider: provider,
    Store: store,
    Signer: signer,
    Prover: localProver,
})
session, err := invisi.Connect(ctx, sdk.ConnectionOptions{
    Wallet: wallet,
    Protocol: protocol,
    Namespace: "my-app",
})
if err != nil { return err }
```

`provider` is an ordinary go-ethereum client. Applications own its connection and the EVM signer. `ProtocolFromRPC` obtains the chain ID from that provider and verifies the selected deployment's runtime code and VERSION before returning. Select a trusted Protocol RPC service; consistency checks cannot establish that a deployment is official. An independently published deployment can be pinned with `rpc.ClientOptions.Deployment`. There is no built-in deployment catalog or Factory/operator discovery.

## Per-Credit spending limits

`CreateCreditPreparation.SingleLimit`, `DailyLimit` and `MonthlyLimit` are optional `*big.Int` amounts below 2^126; nil or zero disables a limit. Each input Credit independently checks its debit, including full consumption. Owners may set any of the three limits during top-up. Nil top-up limits inherit their previous value; an explicit zero disables the check. All four usage fields are preserved exactly, including across period boundaries. Lowering a daily or monthly limit below current usage freezes further draws in that period; the next UTC period starts fresh.

`core.CreditNote` embeds `core.CreditSpendingState`, containing those three limits and `LastDay`, `DailySpent`, `LastMonth`, `MonthlySpent`. Daily/monthly spent totals are checked uint128 values and record every debit, even with a zero limit. Disabling or re-enabling a check never clears usage. The low-level `ComputeCreditCommitment` requires that explicit state; `ComputeCreditNoteCommitment(note)` hashes the complete note. Omitting new JSON fields initializes unrestricted state, while commitments and Credit delivery always use the current 18-field/v5 formats.

Credit draw preparation reads the Chain RPC header timestamp. UTC Gregorian day/month identifiers are proof-bound public inputs reconstructed by the contract, with no caller date fields in `CreditDrawParams`. A prepared operation crossing a UTC period must be prepared and authorized again. `core.RecoverCreditRemainder(context, previous, event, confirmedTimestamp)` requires the confirmed receipt block timestamp; managed refresh verifies its block hash against the Chain RPC and reconstructs both usage counters without relying on ciphertext. No period computation defaults to the computer's clock.

## Encrypted output memos

`core.Note` and `core.CreditNote` carry optional `Memo []byte` metadata. Set `Memo` on a `sdk.PaymentOutput` or `sdk.CreateCreditPreparation` to encrypt it with that output. Text is UTF-8 by convention; `core.MemoFromText(text)` validates UTF-8 and the 256-byte limit, while `core.MemoText(memo)` explicitly decodes text. Raw bytes may contain NUL or invalid UTF-8 and are never automatically parsed as JSON.

```go
memo, err := core.MemoFromText("Invoice 123")
if err != nil { return err }
output := sdk.PaymentOutput{To: recipient, Amount: big.NewInt(50), Memo: memo}
// Use output in ShieldPreparation.Outputs or TransferPreparation.Outputs.
```

`TopUpCreditPreparation.Memo` and each `CreditDrawInput.Memo` preserve the parent memo when nil. A non-nil empty slice (`[]byte{}`) clears it; another slice replaces it in both owner and spender deliveries. Preparation and storage copy memo bytes. Proof-only Credit remainder recovery restores financial fields with no memo when the child delivery cannot be authenticated. Commitment-only disclosures omit memos; payload-key disclosures authenticate and reveal them (`GetScope().RevealsMemo`). Memos do not change commitments or nullifiers.

Factory and Protocol share protocol version 1 (`ProtocolVersion.CURRENT = 1`); the Go SDK requires Protocol `VERSION = 1` and Protocol RPC `/v1` with protocol version 1 metadata. Funds envelopes use version 2 (160–416 bytes); Credit envelopes use version 5 (498–754 bytes). Historical deployments are incompatible and their records remain unchanged.

## TypeScript to Go

| TypeScript | Go |
| --- | --- |
| `Wallet.fromPrivateKey(key)` | `sdk.WalletFromPrivateKey(key)` |
| `Wallet.fromAddress(address)` | `sdk.WalletFromAddress(address)` |
| `wallet.getAddress()` | `wallet.GetAddress()` |
| `wallet.ownerAddress` | `wallet.OwnerAddress()` |
| `wallet.getViewer()` / `getDetector()` | `wallet.GetViewer()` / `GetDetector()` |
| `wallet.exportAuthorization()` | `wallet.ExportAuthorization()` |
| `new ProtocolRpcClient(url)` | `sdk.NewProtocolRpcClient(url)` |
| `Protocol.fromRpc(rpc, provider)` | `sdk.ProtocolFromRPC(ctx, rpc, provider)` |
| `rpc.withProvider(provider)` | `rpc.WithProvider(provider)` |
| `new Invisi(options)` | `sdk.NewInvisi(options)` |
| `invisi.connect(options)` | `invisi.Connect(ctx, options)` |
| `invisi.prepareTransfer(options)` | `invisi.PrepareTransfer(ctx, options)` |
| `prepared.exportTransaction()` | `prepared.ExportTransaction(ctx)` |
| `session.refresh(options)` | `session.Refresh(ctx, options)` |
| `session.balance(token)` | `session.Balance(ctx, token)` |
| `session.prepareTransfer(options)` | `session.PrepareTransfer(ctx, options)` |
| `prepared.preview()` | `prepared.Preview()` |
| `prepared.estimateGas()` | `prepared.EstimateGas()` |
| `prepared.authorize()` | `prepared.Authorize(ctx)` |
| `authorized.prove(options)` | `authorized.Prove(ctx, options)` |
| `proved.authorizeFunding()` | `proved.AuthorizeFunding(ctx)` |
| `ready.exportTransaction()` | `ready.ExportTransaction(ctx)` |
| `ready.trackTransaction(hash)` | `ready.TrackTransaction(ctx, hash)` |
| `session.sendTransaction(ready)` | `session.SendTransaction(ctx, ready)` |
| `ready.reconcile()` | `ready.Reconcile(ctx)` |
| `session.resume(id)` | `session.Resume(ctx, id)` |
| `Disclosure.fromNote(note)` | `sdk.DisclosureFromNote(note)` |
| `new DisclosureVerifier({ protocol, provider })` | `sdk.NewDisclosureVerifier(options)` |

Go uses exported names, `context.Context`, explicit errors and `*big.Int` amounts. All eight direct prepare methods support both explicit and managed workflows; business roles do not change RPC or cryptographic rules. These language conventions do not change operation permissions or protocol rules. Constructors do not open databases, scan events or start background tasks.

## Packages and responsibility

| Package | Responsibility |
| --- | --- |
| `sdk` | Wallet permissions, sessions, inventory, reservations, operations, registry and disclosure |
| `sdk/sqlite` | Explicit encrypted SQLite wallet storage; the application owns the database connection |
| `core` | Versioned wallet KDF, checked receiving addresses, Poseidon2, note cryptography and FMD |
| `protocol` | Protocol objects, current calldata, public statements and registry builders |
| `protocol/contracts` | Generated bindings for the exact shared ABI |
| `rpc` | Protocol RPC discovery, deployment metadata, candidates and authenticated membership |
| `prover` | Local witness boundary, pinned proof statements and aggregation |
| `prover/native` | Local BB 6.0.0 generation and verification with authenticated circuit resources |

Chain RPC supplies blocks, canonical receipts, contract reads and transaction submission. Protocol RPC supplies public history, candidate filtering and membership paths. The SDK decrypts and verifies candidates, saves only its own wallet state, reserves inputs and authorizes local proofs. It has no public-tree implementation, full-chain scanner, IPP transport, payer/receiver packages or legacy compatibility layer.

## Wallets and permissions

The application's existing 32-byte EVM account key is the only wallet root. Invisi defines a versioned KDF, not a mnemonic or HD derivation scheme. The EVM spending owner is also the wallet identity. Viewing and detection keys derive independently using the same salts and rejection sampling as the TS SDK. No independent nullifier key is required.

A complete receiving address is a strict 138-character `0i` string containing the 20-byte owner, 33-byte viewing public key and 32-byte clue key. Its custom Base32 case checksum and public keys are validated. The short form is the 42-character EVM owner address and requires a selected receiving-key registry when used as a payment destination. Full addresses resolve offline.

`GetViewer` returns viewing permission; `GetDetector` returns detection permission. Exported key slices are defensive copies. A wallet restored only from an address has neither permission and cannot authorize spending. Secret-bearing wallet, capability, disclosure and witness objects redact default formatting/JSON. Explicit key or disclosure exports are sensitive and application-owned.

## Refresh and local storage

`Refresh` pages through Protocol RPC in ranges of at most 100 blocks, independently fetches complete candidate receipts from Chain RPC, authenticates insertion positions, decrypts owned deliveries, checks historical membership paths against chain-known roots, and updates local spent state. Candidate transaction order is reconstructed from canonical receipts. Credit recovery retains both historical receiving-key snapshots and verifies deterministic remainder recovery.

```go
result, err := session.Refresh(ctx, sdk.RefreshOptions{
    FromBlock: protocol.Deployment().Protocol.FromBlock,
})
```

`ToBlock`, `PageSize`, `Reset` and an explicitly exported `DetectionKey` are optional. Detection delegates candidate filtering only. Neither FMD candidates nor a saved checkpoint prove complete history; a malicious sender can provide an unrelated valid clue. Applications manage refresh scheduling and complete-history recovery. Reorgs invalidate authenticated checkpoints and require an explicit reset refresh. No scan fallback silently queries every chain log.

`Balance` sums locally known, unspent funds notes for one token. `ConfirmedKnown` includes known unspent notes; `Reserved` is reported separately and `AvailableKnown` excludes those reservations. `AvailableKnown` is nil until a checked inventory exists and after a canonical-history invalidation. It is neither an ERC20 wallet balance nor the Protocol's total holdings; use standard geth/ERC20 APIs for those.

`NewMemoryWalletStore` is an explicit ephemeral choice. Persistent applications use `sdk.NewEncryptedWalletCodec` with a caller-owned 32-byte AES key and `sqlite.NewWalletStore` with a caller-owned `*sql.DB`. SQLite callbacks commit atomically, roll back on failure and encrypt note preimages and operation plans with deployment/wallet/namespace-bound authenticated encryption. Private authorization signatures are not persisted. `Resume` atomically replaces the previous operation worker; stale handles cannot continue signing, proving or submitting.

## Funds and Credit operations

- `PrepareShield`: 1–10 public outputs; `Ready` encodes caller-funded deposits without a Noir proof.
- `PrepareTransfer`: select stored notes, create private outputs and independent random change.
- `PrepareWithdraw`: a public payout with private change, authorized by the wallet owner.
- `PrepareSweep`: whole-note payouts to the bound owner with viewing permission alone; no redirection, change or fee output.
- `PrepareCreateCredit`: private/public funding, spender/owner identities, bound payouts, expiry and shared recovery key.
- `PrepareDrawCredit`: spender-authorized public payouts, with 1/12/50/120 draw capacities and private remainder recovery.
- `PrepareCancelCredit`: owner-authorized complete public refund; expiry does not prohibit cancellation.
- `PrepareTopUpCredit`: owner-authorized increase before expiry; the replacement expiry may stay unchanged or increase.

ERC20 assets require no token registration, token pause configuration or deposit limits. Global pauses, address screening and funding/spend authorization remain enforced. Receiving-key registry metadata is independent of token selection. Ordinary proof capacity is selected automatically from combined input/output counts: 14/56/112/168. Credit creation/top-up supports up to ten funds inputs. Input reservations are atomic and cover both local roles of the same Credit leaf.

`Preview` exposes recipients, amounts, change, Credit roles, expiry and required permissions without exporting note secrets or private signatures. `Authorize` signs the complete frozen operation statement. `Prove` rejects stale circuit/VK identities and different public inputs, then independently verifies the returned bundle. Proof generation and private authorization do not consume the public payer's funding nonce.

ERC20 allowance, token permits, public funding grants and private wallet authorization are independent. Applications approve ERC20 allowance through standard bindings. Signed public funding uses a `FundingSigner` capability with `Address` and `SignTypedData`. `GethFundingVerifier` verifies EOAs and ERC-1271 contract wallets, including variable-length contract signatures. A custom `FundingAuthorizationVerifier` may be injected. The grant is requested after the actual proof exists. `Signer` is an ordinary `*bind.TransactOpts` and only pays gas/submits transactions. Caller-funded operations require that sender to equal the payer.

`ExportTransaction` checks deployment, deadlines, accepted roots, spent state and funding nonces before marking the managed call externalized. `TrackTransaction` records a broadcast association; `Reconcile` authenticates destination, exact calldata and canonical receipt. Both return a new operation handle; inspect `Status()` or `Preview()` on that result. Missing or reverted transactions still reconcile spent state and expiry, releasing inputs only when the proof has become unusable. `WalletSession.SendTransaction(ctx, ready)` thinly delegates to the configured geth signer/backend, returns the original `*types.Transaction`, and tracks its hash. Passing a plain `TransactionCall` does not infer an operation association. Use ordinary `bind.WaitMined` for receipt waiting; the SDK has no `Submit` or `Wait` API. A tracking failure after broadcast returns the sent transaction and an error containing its hash. Recursive members reconcile their exact request inside the aggregate. Refresh indexed history after confirmation to discover and authenticate resulting outputs. A timeout never releases potentially externalized inputs. Only unexternalized prepared operations may be abandoned; expired proofs require fresh canonical spent checks before reservations are released.

## Application-owned notes without a Store

Applications with their own database can call `Invisi.PrepareTransfer`, `PrepareWithdraw`, `PrepareSweep`, `PrepareShield`, `PrepareCreateCredit`, `PrepareDrawCredit`, `PrepareTopUpCredit` and `PrepareCancelCredit` directly. Use the corresponding `Explicit...Preparation` type. This path never calls a configured `WalletStore` and does not create a hidden memory store or managed session.

```go
prepared, err := invisi.PrepareTransfer(ctx, sdk.ExplicitTransferPreparation{
    TransactionIdentity: sdk.TransactionIdentity{Wallet: wallet, Protocol: protocol},
    TransferPreparation: sdk.TransferPreparation{
        TokenAddress: tokenAddress,
        Outputs: []sdk.PaymentOutput{{To: recipient, Amount: big.NewInt(30)}},
    },
    ExplicitFunds: sdk.ExplicitFunds{Inputs: selectedNotes},
})
if err != nil { return err }
fmt.Println(prepared.Preview()) // Review locally before authorization.
authorized, err := prepared.Authorize(ctx)
if err != nil { return err }
proved, err := authorized.Prove(ctx)
if err != nil { return err }
call, err := proved.ExportTransaction(ctx)
if err != nil { return err }
// Submit call.To/call.Data through the application's ordinary geth transaction pipeline.
```

Each `NoteInput` supplies a private `core.Note`, commitment, tree ID and leaf index. `CreditInput` supplies the complete delivery state and its `owner` or `spender` role. The SDK checks deployment, token, owner/role, commitment preimages, canonical membership and spent state. All supplied funds inputs are consumed, including when public funding covers the requested amount. `ChangeAddress` chooses the ordinary transfer/withdrawal change recipient; Credit creation/top-up change must remain bound to the Credit owner. Sweep can explicitly select a subset by commitments. Inputs from different owners cannot be combined into one spend. A shared `Invisi` can prepare independent transactions for multiple wallets.

`PreparedTransaction` exposes local `Preview`, `Authorize`, `Prove`, `Ready`, `AuthorizeFunding`, `ExportTransaction`, `ExportProof` and `ExportAggregateMember`. Successful stages return a new handle and fence the old one; errors and context cancellation allow retry on the previous frozen handle. Concurrent execution is rejected. `Authorize` can receive the matching wallet authorization explicitly for a public-only wallet. A `WalletAuthorizationProvider` can delegate signing without importing the private key. Proof backends and funding signers may be supplied through `InvisiOptions` or per call: `prepared.Prove(ctx, backend, options)` and `proved.AuthorizeFunding(ctx, signer)`. The optional proving arguments are checked at runtime because Go does not have method overloads. Default formatting and JSON redact private plans.

The application owns reservations, concurrency across processes, operation persistence, broadcasting, canonical receipt reconciliation and inventory updates. Save the exact signed transaction hash before broadcasting, retain reservations across uncertain outcomes, and never interpret a timeout as permission to spend the inputs again. Detached transactions have no persistence or recovery API. Managed `WalletSession` adds atomic reservations, immutable operation stages, external transaction tracking and recovery. A failed local proof does not externalize the operation; it remains retryable. `Resume` validates saved scope, proof identity and recomputed calldata before fencing the former worker. See `examples/03-funds/transfer.go` and `examples/08-wallet-session`.

## Local proving and aggregation

`prover.Prover` matches the TS SDK's `prove`/`verify` boundary. Progress phases are `witness`, `proving`, `verifying` and `complete`; cancellation uses the Go context. Proofs are bound to the circuit, artifact hash, VK hash, target, exact public inputs, BB version and proof digest. Proof bundle JSON uses the same lower-camel-case fields and `proofHex` encoding as TS. Circuit resources and verification keys are authenticated against the shared TS release.

`native.NewNativeProver(native.NativeOptions{})` uses the official `noir-execute` 1.0.0-rc.6 and BB 6.0.0 executables from PATH. `Prove` runs witness execution, proving and independent verification locally; applications do not need to implement a witness executor. `Verify` requires only BB. There is no Node, TS SDK or source/compiler dependency at runtime.

Install the matching executables from the official [Noir release](https://github.com/noir-lang/noir/releases/tag/v1.0.0-rc.6) and [Barretenberg release](https://github.com/AztecProtocol/aztec-packages/releases/tag/v6.0.0). These are local tool dependencies, similar to the TS native prover's BB requirement. The SDK checks tool versions and authenticated release assets before execution; it never downloads or upgrades tools automatically. Existing tool installations are not modified.

```go
import "github.com/invisi-lab/invisi-sdk-go/prover/native"

localProver, err := native.NewNativeProver(native.NativeOptions{
    // Optional: omit both fields when the pinned tools are on PATH.
    BBBinary: "/path/to/bb",
    NoirBinary: "/path/to/noir-execute",
})
```

The default executor uses private temporary directories and permission-restricted JSON/witness files, removes them on completion or failure, honors context cancellation/timeouts, and suppresses executor diagnostics that may contain private values. Stripped release artifacts receive empty source/debug metadata required by the official CLI; their bytecode, ABI and authenticated circuit identity remain unchanged. Advanced applications may explicitly inject another local `WitnessExecutor`. Private inputs are never sent to Protocol RPC or a remote proving service.

`prover.NewAggregator` verifies 4/8/12 recursive slots14 members before constructing the outer witness. Members must share the current circuit context, have no duplicate nullifiers/commitments and use the authenticated recursive key. `ExportAggregateMember` contains only the public proof, statement and request. Exported inputs remain reserved. `Protocol.Aggregate` constructs the public atomic call; each original operation independently reconciles its exact member after submission.

## Registry and disclosure

`NewAddressRegistry(protocol, provider)` selects the deployment's registry and validates its immutable `Protocol.addressRegistry()` binding. `Resolve` leaves full addresses offline and validates the chosen registry before resolving an EVM short address. `PrepareRegistration` uses registry nonces and ordinary EIP-712 owner signatures. Register and update share review/authorize/export/track/reconcile stages; authorization makes the operation ready without a prover. Geth owns sending and receipt waiting, and a separate EVM account may submit signed registrations.

A local private `Wallet` supplies registry signing automatically. For external signers and ERC-1271 owners, pass `RegistrationPreparation.Authorization` implementing `RegistrationAuthorizationProvider`, or obtain `operation.TypedData()` and pass collected signature bytes to `operation.Authorize(ctx, signature)`. ERC-1271 signatures remain opaque; EOA signatures use 65-byte `r || s || v` with EVM `v` values 27/28 (geth 0/1 values are normalized). The operation verifies authorization against the frozen registry message before exporting calldata. `Wallet.ExportRegistrationAuthorization()` grants registry signing independently of private circuit authorization.

Low-level `protocol.BuildRegistrationCalldata` preserves arbitrary signatures and permits an empty signature for direct owner transactions. The sender must equal the nonzero EVM owner encoded in the zero-padded `bytes32` identity. High-level operations require signatures. `protocol.RegistrationTypedData` and `protocol.ComputeRegistrationDigest` use `InvisiAddressRegistry` version `1`, the actual chain ID, and selected registry address; register and update have separate flat message types.

`DisclosureFromNote` opens one note. `DisclosureFromCredit` opens the Credit commitment using only its shared-key hash. `GetViewer().DiscloseNote` and `DiscloseCredit` release a selected envelope's symmetric payload key, never the wallet-wide viewing key. Credit payload disclosure reveals that envelope's shared recovery key and receiving snapshots. `DisclosureFromExport` and `Verify` interoperate with TS disclosure JSON and require an independently supplied matching output. `DisclosureVerifier` additionally authenticates canonical receipt inclusion and confirmations through Chain RPC; it establishes contents/inclusion, not payer identity, current spent state or complete history.

## Examples and validation

Read numbered topics in order:

| Directory | Content |
| --- | --- |
| `examples/00-wallet` | Executable offline account/wallet/permission example |
| `examples/01-connect` | Executable Chain RPC + Protocol RPC deployment verification |
| `examples/02-shield` | Caller-funded deposit walkthrough |
| `examples/03-funds` | Explicit manual note inputs for transfer, withdraw, mixed funding and view-only sweep |
| `examples/04-credit` | Explicit Credit inputs and receiving-key snapshots for create, draw, cancel and top-up |
| `examples/05-registry` | Register/update independent receiving keys |
| `examples/06-aggregation` | Explicit disjoint note inputs, recursive members and public aggregation |
| `examples/07-disclosure` | Executable offline scoped disclosure |
| `examples/08-wallet-session` | Encrypted SQLite, refresh, managed selection, geth sending and recovery |

The connection example reads only `CHAIN_RPC_URL` and `PROTOCOL_RPC_URL`. The managed-session example additionally reads `WALLET_ACCOUNT_PRIVATE_KEY` and `WALLET_STORAGE_KEY`; its default database is `./wallet.sqlite`. Tokens are editable local-deployment constants. Business walkthrough functions take explicit connected application objects so they do not hide signer, storage or prover configuration. Go `defer` closes caller-owned resources. All comments and output are English.

```sh
go test ./...
go test -race ./...
go vet ./...
go run ./examples/00-wallet
go run ./examples/07-disclosure

# Genuine local witness/proof tests, including recursive aggregation:
NOIR_EXECUTE_PATH=/path/to/noir-execute BB_BINARY_PATH=/path/to/bb \
    go test ./prover/native ./sdk -run TestReal -v
```

Tests use independent current KDF, encryption and FMD vectors, TS protocol statement/calldata vectors, and TS disclosure exports. Dummy proof bytes in formatting/backend tests do not establish proof soundness. Chain/RPC tests use controlled fixtures. The opt-in `TestReal` suite executes the official Noir/BB binaries for funds and Credit proofs, rejects modified signatures/public inputs, and proves a four-member recursive aggregate. These tests establish local proof generation and verification, not live deployment transaction acceptance.

Maintainers authenticate the selected protocol checkout before importing its proving assets and TS reference vectors:

```sh
python3 scripts/sync-reference.py /absolute/path/to/invisi-sdk-ts /absolute/path/to/invisi-protocol
INVISI_PROTOCOL_ROOT=/absolute/path/to/invisi-protocol go generate ./protocol/contracts
node scripts/generate-disclosure-vectors.mjs /absolute/path/to/invisi-sdk-ts /absolute/path/to/protocol-docs
python3 scripts/test-sync-reference.py
```

Published packages contain their own bindings, public test vectors and circuit assets. They do not depend on the TS SDK or Protocol RPC server at runtime.

## Gas reference before proving

Both `PreparedTransaction` and managed `Operation` support the same synchronous pre-proof reference as TypeScript:

```go
reference, err := prepared.EstimateGas()
if err != nil { return err }
fmt.Println("Reference gas units:", reference.GasUnits)
```

The public result contains only `GasUnits` (`uint64`). Estimation performs no proving, Store access, network requests or chain gas estimation. All eight direct preparation methods are supported. Recursive members have no standalone transaction cost and return an error; estimate complete aggregates separately.

The bundled receipt catalog matches TypeScript's catalog: 13 current Credit shapes measured with three real-verifier receipts each on isolated Anvil (Solidity 0.8.38, BB 6.0.0, Prague), alongside nine preserved historical Funds shapes. Credit provenance records protocol HEAD base `42f2546`, dirty source SHA-256 `621cbef5f84a133ca96674a2dba39b21b5f9fbc022977d4d2ab296d9d6e685e0`, and deployed runtime code hashes. Those Credit transactions used a 500,000,000 local block gas limit; network transaction and block limits can make large draws unmineable. The 120-input partial-draw mean is 17,458,420 gas, above the 16,777,216 transaction gas cap on chains enabling [EIP-7825](https://eips.ethereum.org/EIPS/eip-7825); that measured shape cannot execute there even with a larger block gas limit. Credit records require the same circuit artifact and verification key; historical Funds receipts remain available for Funds references. Exact shapes use the rounded-up arithmetic mean, taking the higher mean when policy variants share a public shape. References apply across deployment addresses and tokens. Unmeasured combinations first combine the highest observed gas across variants with the same public shape, then use the same upward-only workload envelope as TypeScript, which can substantially overestimate complex operations. This is not a calibrated prediction, a guaranteed upper bound or a transaction gas limit. Token behavior, address screening, tree/storage state and chain execution rules affect actual gas. The application's ordinary geth pipeline remains responsible for final estimation, fees, nonces and sending.

For trusted measurements of a private deployment, call `sdk.DefineGasBenchmarks(jsonBytes)` and set `InvisiOptions.GasBenchmarks`. `NewInvisi` snapshots those records. Exact deployment, circuit and shape matches override the bundled reference; unmatched records fall back to the standard model. Catalog validation checks format, not publisher authenticity. `scripts/sync-reference.py` checks Credit receipt circuit identities against the selected protocol release and copies the catalog with its SHA-256-bound TypeScript comparison vectors. Current receipt provenance includes source fingerprints, compiler and prover versions, block gas limit, and deployed runtime code hashes. The Go tests include TS-generated comparison vectors for measured and modeled combinations.
