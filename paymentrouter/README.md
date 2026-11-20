# Payment Router

Optional module `github.com/invisi-labs/invisi-sdk-go/paymentrouter` composes the SDK's detached proof lifecycle with native ETH funding. It supports exact-output Uniswap V3 and registry-verified native 0x routes. It exposes no signing or broadcasting operation. The core SDK has no dependency on this module.

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
