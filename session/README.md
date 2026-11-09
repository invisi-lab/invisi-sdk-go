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

`pending, err := managed.Refresh(ctx, options)` captures public candidate pages and returns `pending.WalletRequests` without changing the real Store. Execute each request outside the session. Pass `[]session.WalletRefreshResponse{{RequestID: request.ID, Response: &response}}` to `pending.ApplyWalletResponses`; an explicit nil response marks a nonmatch. Every ID must appear exactly once. A pending handle applies at most once and rejects concurrent application or a changed store revision. Apply revalidates chain receipts, roots, membership and spent/revoked state against captured candidates, then atomically replaces inventory and checkpoints.

Authorization follows the same explicit boundary: `operation.ExportWalletRequest(ctx, id)`, external execution, then `operation.WithWalletResponse(ctx, request, response)`. No local executor is retained in the session. Refresh unseen Note/Credit deliveries before `CreditReconciler`; reconciliation can reuse already authenticated stored openings.

`session.NewMemoryWalletStore` is explicitly ephemeral. For persistence, import `session/sqlite`, provide an application-owned database, and construct `session.NewEncryptedWalletCodec` with an application-owned encryption key. Back up the database and key separately. Failed local work never advances the store revision.

Managed `Operation` uses the same preparation, statements, signatures, witnesses, proof validation and calldata as `sdk.PreparedTransaction`, and additionally tracks persistent worker leases, reservations, submission and canonical reconciliation. `SendTransaction` delegates to an optional geth signer and does not wait. Both paths can export calldata for external submission.

All packages belong to the existing Go module and share its version. No compatibility exports remain under `sdk` or `sdk/sqlite`.
