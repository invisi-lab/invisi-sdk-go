# Payment Router

Canonical API (Periphery `ffd81f4`) supports DirectNative, DirectToken, RelayedToken and SponsoredNative. `PaymentRouterABI()` and `PaymentRouterAccountABI()` expose the complete contract interfaces. Existing `Quote`, `BuildPaymentTransaction`, `SimulatePayment`, `LegacyPaymentRouterABI` and `ZeroExPaymentRouterABI` remain explicit legacy native-only APIs.

Use `PreparePayment` to produce an unchanged Core proof with the Router as funding payer. Export its Core call, then use `QuoteRoute` with `QuoteParameters.TokenIn` (zero for native), `AmountOut` including the Core token fee, an input cap and deadline. V3 routes start at that input token; ERC20 0x routes accept only the configured AllowanceHolder and current registry Settler. For same-token input or native-to-WETH, set `Route{Channel: ChannelNone}` without requesting a swap.

`BuildPayment(ctx, PaymentOptions{Core: coreCall, Mode: DirectToken, Payer: player, TokenIn: token, AmountInMaximum: budget, RefundRecipient: player, Deadline: deadline, Route: route})` reads the current Router nonce and binds the original Core proof, request and route. The builder returns a detached unsigned transaction. Native modes supply value equal to the cap; token modes supply zero. `ValidatePayment` rechecks deployment, nonce, route, sender and value; `SimulateCanonicalPayment` calls and estimates the complete Router transaction.

For token relay, first build with `Mode: RelayedToken`, a fixed `Relayer` and an empty signature. Decode the call with `DecodeRouterCall`, obtain `IntentTypedData`, and have the wallet sign it. Rebuild with the signature; `ValidatePayment` and simulation require a valid EOA or ERC1271 authorization. `ApprovalTransaction` builds an ERC20 approval to the Router. For ERC-2612 tokens, use the explicitly supplied token name/version/nonce with `PermitTypedData`, then `PermitFromSignature`; permit support is never assumed for WETH. `SignTypedData` is available for applications that explicitly own a local key. No helper sends transactions automatically.

For native sponsorship, build the inner SponsoredNative payment, sign `NativePaymentTypedData`, and call `BuildSponsoredPayment` with the player, pinned account implementation/code hash, account action and optional signed EIP-7702 authorization. `SignDelegation` signs only an explicitly selected implementation and EOA authorization nonce. Router nonce, account action nonce and EOA authorization nonce are independent. The provider must expose `StorageAt` and `NonceAt` when installing delegation; `GethProvider` does so. Persistent account storage is checked even when delegation was previously removed. Existing incompatible wallet implementations are rejected. Call `SimulateEnvelope(ctx, envelope, relayer, sponsoredOptions)` to validate the policy and simulate/estimate the actual zero-value outer type-4 transaction, including authorizationList. Wallet/chain delegation support and user approval remain host responsibilities.

Relay RPC methods live in `rpc/broadcaster`: `EstimateRouterQuote`, `GetRouterQuote` and `SubmitRouterTransaction` send the original embedded Core data plus `router` envelope. The signed quote binds `payment-router-envelope-v1` and the entire outer target, calldata, zero value and authorization list. Only direct execution is supported; schedules, aggregation and provenance sidecars are rejected for this path. Requote after changing any signed field.

Golden fixtures cover all four canonical selectors, exact Core/request/route hashes, intent/account signatures and envelopes with and without signed delegation. Structural 0x validation does not certify embedded liquidity actions; full simulation is required. Live API and interactive wallet acceptance remain environment-dependent.

## Legacy native-only integration

The optional module `github.com/invisi-labs/invisi-sdk-go/paymentrouter` retains these historical native-only helpers for existing deployments. The core SDK has no dependency on this module.

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

`PaymentRouterABI()` and `ZeroExPaymentRouterABI()` return fresh public ABI snapshots. `EncodeV3ExactOutputPath`, `ValidateV3Path`, `ValidateZeroExPayload`, and `InspectCorePayment` are available independently.

## Verification

`python3 scripts/test-modules.py` from the repository root runs race tests, vet, builds and module dependency checks. The module consumes the exact TS shared vectors, including complete payable transaction bytes. Optional compiled-contract tests require an isolated loopback Anvil fixture supplied by the TS repository's `scripts/payment-router-test-chain.mjs`:

```sh
INVISI_PAYMENT_ROUTER_FIXTURE=/absolute/path/go-anvil.json go test -v -run TestCompiledContractsAnvil
```

That test submits through the unlocked local host RPC only, outside the SDK. It checks all three native builder/simulator channels, actual spend/refund and custody, events, residual balances/approvals and core-failure rollback. Its core/DEX/Settler doubles do not demonstrate production ZK proof or external live DEX/API compatibility.
