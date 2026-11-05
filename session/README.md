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
    Wallet: wallet,
    Protocol: pool,
    Store: store,
})
```

`Connect` validates the pinned Protocol and RPC deployment before returning. Wallet, viewing, authorization, geth signer and optional public-funding signer are local session capabilities. Wallet and signer can represent different roles. Shared infrastructure contains no wallet state or default signing authority.

`session.NewMemoryWalletStore` is explicitly ephemeral. For persistence, import `session/sqlite`, provide an application-owned database, and construct `session.NewEncryptedWalletCodec` with an application-owned encryption key. Back up the database and key separately. Failed local work never advances the store revision.

Managed `Operation` uses the same preparation, statements, signatures, witnesses, proof validation and calldata as `sdk.PreparedTransaction`, and additionally tracks persistent worker leases, reservations, submission and canonical reconciliation. `SendTransaction` delegates to an optional geth signer and does not wait. Both paths can export calldata for external submission.

All packages belong to the existing Go module and share its version. No compatibility exports remain under `sdk` or `sdk/sqlite`.
