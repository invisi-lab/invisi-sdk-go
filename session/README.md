# Managed wallet sessions

Import `github.com/invisi-lab/invisi-sdk-go/session` when the SDK should manage inventory, canonical refresh, reservations, private history and durable operations. Stateless callers use `sdk` without this package or a database driver.

```go
invisi := sdk.NewInvisi(sdk.InvisiOptions{
    Provider: provider,
    ProtocolRPC: protocolRPC,
    Prover: localProver,
})
managed, err := session.Connect(ctx, session.Options{
    Invisi: invisi,
    Wallet: wallet, // Public material only.
    Protocol: pool,
    Store: store,
})
```

`Connect` validates the pinned Protocol and RPC deployment before returning. Wallet contains only public material. Connect takes no wallet executor, viewing provider or authorization callback. Shared infrastructure contains no default signing authority. The geth signer and optional public-funding signer remain separate from private wallet authorization.

`pending, err := managed.Refresh(ctx, options)` captures canonical candidate pages without changing the Store. `pending.Decrypt(ctx, appDecryptor)` invokes the explicitly supplied `sdk.DeliveryDecryptor`. Only nil plaintext with no error is a nonrecipient; all operational failures abort. SDK validation binds plaintext to owner, token, deployment, commitment and Credit receiving keys. Apply rechecks receipts, roots, membership and spent/revoked state, then atomically commits inventory and checkpoints with the captured store revision. Concurrent application, stale revisions and reuse after success reject.

Authorization exports `operation.ExportWalletAuthorization(ctx)`, accepts an ordinary geth EIP-712 signature through `WithWalletSignature`, then proves. Public funding and revocation have their own frozen typed-data requests. Refresh unseen deliveries before Credit reconciliation; already authenticated stored openings can be reused.

`session.NewMemoryWalletStore` is explicitly ephemeral. For persistence, import `session/sqlite`, provide an application-owned database, and construct `session.NewEncryptedWalletCodec` with an application-owned encryption key. Back up the database and key separately. Failed local work never advances the store revision.

Managed `Operation` uses the same preparation, statements, signatures, witnesses, proof validation and calldata as `sdk.PreparedTransaction`, and additionally tracks persistent worker leases, reservations, submission and canonical reconciliation. `SendTransaction` delegates to an optional geth signer and does not wait. Both paths can export calldata for external submission.

All packages belong to the existing Go module and share its version. No compatibility exports remain under `sdk` or `sdk/sqlite`.
