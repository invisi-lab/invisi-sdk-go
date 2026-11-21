# Payment Router: public principal, private Note outputs

These are callable Go host-integration examples. `funding.go` and `relay.go` use the actual SDK builders, real quote adapters, proof callbacks and wallet callbacks. Importing them performs no network calls. They do not include a wallet UI, usable deployment constants, fake proofs or fixture DEX quotes. Run from this repository's Go workspace; the optional canonical Router API is unreleased, and module versions have not been bumped.

```sh
cd examples
go test ./08-payment-router
go vet ./08-payment-router
```

The host supplies an independently trusted `paymentrouter.Deployment`, `paymentrouter.GethProvider{Client: eth}`, `sdk.ProtocolFromRPC`, the public receiving wallet, a native Core prover, and external signing/submission capabilities. Use the actual token decimals and live routes. Token symbols such as USDT are illustrative; do not substitute an experimental token address while describing it as mainnet USDT.

`Prepare100PlusFee` creates two Notes of 50 Core tokens and a 0.2 Core-token fee, so the Router must acquire **100.2 Core tokens**, in integer base units. The fee recipient must be the authenticated selected Broadcaster for relay. If the current fee estimate does not fit 0.2, stop and rebuild the preparation with the actual approved fee; the recipe never changes a frozen request silently. Direct calls may use `client.PreparePayment` with a zero fee when no service fee applies. At least two positive Note outputs are required.

| Scenario | Route and workflow | Principal / gas |
| --- | --- | --- |
| Self ERC20 → Core token | `RouteV3`, `SelfPayment` with `DirectToken` | Player ERC20 / player gas |
| Self native → Core token | `RouteV3`, zero `TokenIn`, `SelfPayment` with `DirectNative` | Player native / player gas |
| Relayed ERC20 → Core token | `RouteV3`, `RelayToken` | Player ERC20 / fixed Broadcaster gas |
| Sponsored EIP-7702 native → 100.2 Core tokens | `RouteV3`, zero `TokenIn`, `SponsorNative` | Player account native / fixed Broadcaster gas |
| Self ERC20 → Core token through 0x | `RouteZeroEx`, `SelfPayment` with `DirectToken` | Player ERC20 / player gas |

For V3, supply the trusted Quoter and executable `V3Route` token/fee pairs. Native routes start with deployment WETH; the Router wraps native input. Set `QuoteParameters.AmountOut` to the funding amount returned by `Prepare100PlusFee`. For 0x, inject a real backend transport or `NewZeroExQuoteFetcher` with an explicitly supplied API key; the deployment pins AllowanceHolder and current registry Settler. Opaque action validation cannot establish liquidity execution: full Router simulation is mandatory. Same-token ERC20 and native→WETH use `Route{Channel: ChannelNone}` without a swap quote.

The host follows this order:

1. Prepare outputs, Core-token fee and funding using `Prepare100PlusFee` or `client.PreparePayment`.
2. Obtain the real route quote. Copy its `MaxInput`, `Deadline` and returned `Route` into `PaymentOptions`, alongside the selected payer, refund recipient and fixed relayer.
3. For relay, export `prepared.ExportRouterDraft(ctx)` into `PaymentOptions.Core`, then invoke `RelayToken` or `SponsorNative`. Pass `CoreProofCallback(prepared)` as the `prove` callback; it calls `prepared.Prove(ctx)` and exports the actual Core transaction. The recipe obtains and verifies the formal Broadcaster quote **before** invoking this callback or wallet signing. Keep all business fields unchanged. Fee convergence requires a fresh preparation/draft before the formal quote.
4. For self submission, prove and export Core before calling `SelfPayment`. It builds and simulates the complete Router call. For token input, its approval callback must submit the Router allowance and wait for a canonical successful receipt; this callback can spend approval gas even with `submit=false`, so hosts must explicitly authorize that capability.
5. Persist the final Core/envelope/quote before submission. `submit=false` still performs real quotation/proving/signing/simulation, but sends no payment. `submit=true` invokes the injected self sender or `SubmitRouterTransaction`. Persist the returned Broadcaster status ID and query `service.Status(ctx, id)`; successful submission is not finality. Reconcile actual transaction receipts independently through Chain RPC.

`sign` receives actual EIP-712 typed data after proof generation. Router/account EOA signatures require low-s `r||s||v` with `v=27/28`; geth parity `0/1` must be normalized for those calls. ERC-1271 uses its wallet format. EIP-7702 authorization parity remains `0/1`.

For `SponsorNative`, supply the pinned `SponsoredPaymentOptions` implementation/runtime policy, account action nonce and optional unsigned authorization metadata. The wallet `authorize` callback signs that exact delegation only after formal quotation and proof. An existing approved delegation may omit the authorization list. The outer envelope goes to the player, carries value zero, and is submitted by the Broadcaster; the player account's native balance funds the inner payment. EIP-7702 grants persistent delegation, including if execution reverts. Installation needs chain/RPC/type-4 wallet support. The Router nonce, account action nonce, EOA authorization nonce and Broadcaster transaction nonce are independent.

ERC20 allowance to the **Router** authorizes input principal. EIP-7702 authorization authorizes account delegation. Neither is Source enrollment. Public Router funding consumes no historical private Notes and needs no Provenance input proof. See [09-provenance](../09-provenance/README.md) for output recognition and later private spending. The Core/circuit/contract interfaces and Provenance's existing 11 public inputs remain unchanged.

For exact persistence/recovery wire details and deployment policy see [paymentrouter/README.md](../../paymentrouter/README.md). Preserve the original serialized submission after uncertain sends; do not rebuild a payment to locate its outcome. `SignedQuote.Context` is not serialized, so reauthenticate a restored quote with trusted configuration before using SDK submission helpers.
