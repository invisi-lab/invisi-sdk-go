# Payment Router

The unreleased canonical API (planned v1.3.3; published module versions remain v1.3.2) supports arbitrary native/ERC20 input through exact-output Uniswap V3 or validated 0x routes. `PaymentRouterABI()` and `PaymentRouterAccountABI()` expose the canonical contract interfaces. The optional module never signs or sends automatically; the host owns deployment trust, proof generation, wallet decisions and submission.

| Mode | Builder / entry point | Principal and gas |
| --- | --- | --- |
| `DirectNative` | `BuildPayment` / `payNative` | Player sends native input cap as transaction value and pays gas. |
| `DirectToken` | `BuildPayment` / `payToken` | Player supplies ERC20 and pays gas; value is zero. |
| `RelayedToken` | `BuildPayment` / `payTokenFor` | Player supplies ERC20 and signs the intent; fixed relayer pays gas; value is zero. |
| `SponsoredNative` | `BuildPayment`, then `BuildSponsoredPayment` / account `executeNativePayment` wrapping `payNative` | Player account supplies native principal; fixed relayer pays gas; outer value is zero. |

## Prepare, route and authorize

Create a `Client` with an explicitly trusted `Deployment` and a `Provider` (`GethProvider` adapts an existing geth client). Deployment checks pin chain, Router dependencies, code and registry state to a canonical block. `PreparePayment` prepares a Core `transact` with the Router as caller-authorized funding payer, empty private inputs, no withdrawal and at least two positive Note outputs. Use `prepared.ExportRouterDraft(ctx)` before proving to export the empty-proof Core draft. Finalize the Core fee, outputs, route and deadlines before requesting a signed quote. After quotation, prove and export the final Core call; retain every business field.

The exact output is the Core funding amount: output Notes **plus the Core-token fee**. For example, 100 USDT in Notes plus a 0.2 USDT Broadcaster fee requires 100.2 USDT from the Router. There is no separate input-token fee. Broadcaster token support and pricing concern this Core output token; the input token needs no separate price listing.

For a swap, call `QuoteRoute` with `QuoteParameters{TokenIn: input, TokenOut: coreToken, AmountOut: fundingAmount, MaxInput: budget, Deadline: deadline}`. Zero `TokenIn` means native currency. V3 routes use that ERC20 input (WETH for native) and encode the exact-output path in reverse. ERC20 0x routes require the configured AllowanceHolder wrapping the current registry Settler; native 0x routes use the validated native envelope. Liquidity action bytes remain opaque, so structural validation requires full execution simulation before sending. For same-token ERC20 or native-to-WETH wrapping, use `Route{Channel: ChannelNone}` without a swap quote. Canonical builders support WETH output; legacy native builders do not.

`BuildPayment(ctx, PaymentOptions{Core: coreCall, Mode: DirectToken, Payer: player, TokenIn: input, AmountInMaximum: budget, RefundRecipient: player, Deadline: deadline, Route: route})` reads the Router nonce and binds the exact Core proof/request and route. Use the returned quote's `MaxInput` and `Deadline` when building a quoted route. Native modes produce an inner call with value equal to that cap; token modes produce zero value. `ValidatePayment` rechecks deployment, nonce, route, sender and value. `SimulateCanonicalPayment` simulates and estimates the complete direct Router call. The host signs and sends a self-submitted call only after successful preflight.

Approve the **Router** to pull ERC20 input with `ApprovalTransaction`, and wait for the host-submitted approval to be mined before simulation. For ERC-2612 tokens, explicitly supply the token's name, version and current permit nonce to `PermitTypedData`, sign with the wallet, then attach `PermitFromSignature`. Permit support is optional and never assumed for WETH. A permit nonce is separate from payment nonces.

For `RelayedToken`, fix `Relayer` to the selected Broadcaster's EVM sender. Build with an empty signature, decode with `DecodeRouterCall`, obtain `IntentTypedData`, ask the wallet to sign, and rebuild with `PaymentOptions.Signature`. Empty signatures are for signing preparation only; validation and executable relay estimates require valid EOA or ERC-1271 authorization. `SignTypedData` is available when the host explicitly owns a local key.

For `SponsoredNative`, build the inner payment, sign `NativePaymentTypedData`, then use `BuildSponsoredPayment` with the player, pinned account implementation/code hash, signed `AccountPayment` and optional signed EIP-7702 authorization. The player EOA balance supplies principal. `SignDelegation` requires an explicit wallet decision and the current EOA authorization nonce. The Router nonce, account action nonce, payer EOA nonce used by EIP-7702 authorization, and Broadcaster sender transaction nonce are independent counters. Coordinate the authorization with other pending payer EOA transactions because they use the same EOA nonce. Delegation persists even if execution reverts; removal/reinstallation does not reset account action storage. Foreign wallet implementations are rejected. The provider must implement `StorageAt` and `NonceAt` when installing delegation; `GethProvider` does. Wallet, chain and RPC EIP-7702 support remain host responsibilities. `SimulateEnvelope(ctx, envelope, relayer, sponsoredOptions)` validates the pinned policy and simulates/estimates the complete outer call, including the signed authorization list. An existing delegation can omit the list; installing it requires the outer type-4 transaction. Outer value is always `"0"`.

## Broadcaster Router workflow

Use the dedicated `rpc/broadcaster` methods `EstimateRouterQuote`, `GetRouterQuote` and `SubmitRouterTransaction`. Ordinary `Quote` / `EstimateQuote` / `GetQuote` / `SubmitTransaction` use the ordinary Core lifecycle and request-hash format; they cannot authenticate a Router envelope.

The wire methods remain `broadcaster_estimateQuote`, `broadcaster_getQuote` and `broadcaster_submitTransaction`, with one object in the JSON-RPC `params` array:

```json
{
  "data": "0x<exact embedded Core transact calldata>",
  "router": {
    "to": "0x<Router for token relay; player for native relay>",
    "data": "0x<complete signed Router/account calldata>",
    "value": "0",
    "authorizationList": []
  }
}
```

Abbreviated hex values are placeholders. Submission additionally sends the returned `quote` and its `signature` as separate top-level fields. Authorization entries use standard Ethereum RPC `chainId`, `address`, `nonce`, `yParity`, `r`, `s` hex quantities. `data` must exactly match the Core call embedded in `router.data`.

Router estimation and formal quotation happen **before proving and wallet signing**. Prepare public funding and outputs, preview with `prepared.EstimateRouterGas(RouterGasInput{...})` (or `EstimateRouterGas` using an existing Core preview), export `prepared.ExportRouterDraft(ctx)`, and use `BuildPaymentDraft`. For sponsored native, wrap with `BuildSponsoredDraft`; unsigned account signatures and EIP-7702 tuples are accepted only for this draft stage. Estimate a fee, enforce the host cap, finalize the Core-token fee/funding and route, then call `GetRouterQuote` with that exact unsigned draft. If fee convergence changes outputs or route, rebuild the draft before requesting the formal quote.

After the quote, call `Prove`, rebuild with `BuildPayment`, sign the actual proof-dependent Router intent (or account payment and delegation), and rebuild the final envelope. `VerifyRouterQuote` accepts both stages with their own exact embedded Core calldata. Proof bytes, Router/account/permit signatures and EIP-7702 signature fields can change; fee, outputs, route, permit deadline, nonces, outer target and authorization metadata cannot. `SubmitRouterTransaction` rejects empty proof or missing wallet authorization, and the service validates signatures, current deployment/nonces and simulates the full final call within signed gas/fee caps.

Quotes use `execution: "direct"` and `requestHashFormat: "payment-router-intent-v1"`. `RouterIntentHash` strictly decodes and normalizes proof/signatures, recomputes the empty-proof Core hash, then hashes `abi.encode(keccak256("InvisiPaymentRouterIntentV1"), EnvelopeHash(normalized))`. `EnvelopeHash` remains the raw envelope identity helper. Scheduling and aggregation are unavailable. Pure public Router funding needs no Provenance source sidecars.

The `payment-router-gas-v1` preview has source `model`: its conservative engineering allowances are not receipt-calibrated swap measurements. It supports all four modes, no-swap/V3 (1–8 hops), permits, sponsored delegation and route bytes. Opaque 0x returns `zeroex-route-gas-unavailable` unless an independent trusted route budget is supplied. The Broadcaster uses operator policy, never a client gas hint. Final simulation remains mandatory.

See the [compiling host-integration examples](workflow_example_test.go) for self-submitted ERC20, relayed ERC20 and sponsored native workflows. The relay recipe accepts an empty-proof Core draft, a proving callback and wallet callbacks; `go test` compiles them without executing RPC calls or broadcasting fixture proofs. Golden fixtures separately cover canonical calldata, hashes, signatures and authorization envelopes; they do not certify live DEX or wallet compatibility.

EOA signatures attached to Router intents or account payments use low-s
`r || s || v` bytes with on-chain `v` equal to 27 or 28. A host signer such as
geth `crypto.Sign` returns recovery parity 0 or 1; normalize that final byte
before attaching it. ERC-1271 signatures follow the contract wallet's format.
EIP-7702 authorization `yParity` remains 0 or 1 and must not be changed to 27/28.

For recovery, persist the literal submission body (`data`, `router`, `quote`,
`signature`) before sending and retain the returned opaque ID. Use
`service.Status(ctx, id)` to query a known submission. `SubmitRouterTransaction`
requires a currently valid quote with its verified `Context`; that context is
not serialized into JSON. Restoring a still-valid quote requires re-verifying it
with trusted configuration before restoring its context. If the acknowledgement
was lost and the frozen quote has expired, the host can resend the exact saved
JSON-RPC submission body to retrieve the server's existing record. The SDK helper
does not bypass expiry checks for that recovery path. Do not rebuild or sign a
new payment merely to discover the original outcome; preserve the original
body and reconcile uncertain sends.

## Legacy native-only integration

The optional module `github.com/invisi-labs/invisi-sdk-go/paymentrouter` retains `Quote`, `BuildPaymentTransaction`, `SimulatePayment`, `LegacyPaymentRouterABI` and `ZeroExPaymentRouterABI` as historical native-only APIs for existing deployments. The core SDK has no dependency on this module.

Create a `Client` with an explicit trusted `Deployment` and a `Provider`. `GethProvider{Client: existingEthclient}` adapts geth RPC and pins both `eth_call` and `eth_estimateGas` to the selected block. RPC servers must support block-tagged gas estimation. Deployment checks authenticate chain ID, Router getters, dependency code and an optional Router runtime code hash; 0x additionally authenticates registry `ownerOf(2)` and Settler code. Code, getters, quotes and simulation use the same canonical block and reject observed reorgs.

```go
client, err := paymentrouter.NewClient(deployment, paymentrouter.GethProvider{Client: eth})
if err != nil { return err }
prepared, err := client.PreparePayment(ctx, paymentrouter.PreparePaymentOptions{
    Invisi: invisi, Wallet: wallet, Protocol: protocol,
    TokenAddress: token, Outputs: outputs, Fee: fee, Deadline: deadline,
})
if err != nil { return err }
ready, err := prepared.Prove(ctx) // use the SDK's explicitly configured prover
if err != nil { return err }
adapter, err := paymentrouter.NewUniswapV3Adapter(provider, quoter,
    []paymentrouter.V3Route{{Tokens: []common.Address{weth, token}, Fees: []uint32{3000}}})
if err != nil { return err }
quote, err := client.Quote(ctx, adapter, paymentrouter.QuoteParameters{
    TokenOut: token, AmountOut: fundingAmount, MaxInput: ethBudget,
    Deadline: swapDeadline, // *big.Int, in seconds
})
if err != nil { return err }
call, err := client.BuildPaymentTransaction(ctx, ready, quote, refundRecipient)
if err != nil { return err }
simulation, err := client.SimulatePayment(ctx, call, submittingAddress)
if err != nil { return err }
_ = simulation // host decides whether/how to submit call
```

Preparation requires at least two positive note outputs, empty private inputs, no withdrawal, and direct Router funding equal to output amounts plus the fee. WETH output is unsupported. Building validates the actual core tuple and nonempty proof, canonical funding, output count/commitments/deliveries, fee pairing and deadlines through the shared protocol module. Quotes cannot rewrite a proved statement. A quote may be refreshed while retaining the same proof.

`QuoteParameters.SlippageBps` defaults to 50 when nil. Set a pointer to zero for no slippage. Values must be integers in [0,10000]. V3 quotes round the authorized ETH input upward with integer arithmetic and select the least estimated input among configured executable routes. No floating-point financial arithmetic is used.

For 0x, inject `ZeroExQuoteTransport` into `NewZeroExAdapter`, or construct an explicit backend transport with `NewZeroExQuoteFetcher(ZeroExHTTPOptions{APIKey: suppliedKey})`. The HTTPS endpoint defaults to the 0x allowance-holder quote endpoint. Requests use `buyAmount`, native ETH, and the Router as taker and recipient. The transport never reads credentials from the environment and rejects redirects. A proxy transport is appropriate for browser-facing services.

0x metadata must agree with the exact-buy statement and budget. Direct Settler and AllowanceHolder-wrapped calls require the registered Settler, exact minimum output, native input and exactly one current `NATIVE_CHECK` binding the entire ETH budget. Other action bytes remain opaque. Payloads are decoded and re-encoded before forwarding. After transport, the client rechecks current time and registry identity. A successful preflight is a snapshot, not an execution guarantee.

Mutable financial inputs are copied before provider/transport callbacks. Returned Go values are detached copies and can be modified by the caller; subsequent builders revalidate them. Context cancellation is propagated throughout.

`LegacyPaymentRouterABI()` and `ZeroExPaymentRouterABI()` return fresh legacy public ABI snapshots. `PaymentRouterABI()` exposes the canonical four-mode interface. `EncodeV3ExactOutputPath`, `ValidateV3Path`, `ValidateZeroExPayload`, and `InspectCorePayment` are available independently.

## Verification

`python3 scripts/test-modules.py` from the repository root runs race tests, vet, builds and module dependency checks. The module consumes the exact TS shared vectors, including complete payable transaction bytes. Optional compiled-contract tests require an isolated loopback Anvil fixture supplied by the TS repository's `scripts/payment-router-test-chain.mjs`:

```sh
INVISI_PAYMENT_ROUTER_FIXTURE=/absolute/path/go-anvil.json go test -v -run TestCompiledContractsAnvil
```

That test submits through the unlocked local host RPC only, outside the SDK. It checks all three native builder/simulator channels, actual spend/refund and custody, events, residual balances/approvals and core-failure rollback. Its core/DEX/Settler doubles do not demonstrate production ZK proof or external live DEX/API compatibility.
