# Portable wallet execution

`Request` and `Response` contain only public requests, encrypted envelopes and operation-specific results. Version, ID, owner and method must match. Use a fresh ID per request. The caller retains the original request and authenticates its transport; the SDK verifies identity, context, commitment and spending signatures independently.

| Method | Params | Result |
| --- | --- | --- |
| `wallet.public` | `{}` | `viewingPublicKey` (33 bytes), `spendingPublicKey` (65 bytes), `clueKey` (32 bytes) |
| `wallet.detect` | `clues: hex[]` | `matches: boolean[]` (advisory candidates) |
| `wallet.decryptNote` | `envelope: hex`, `context` | `plaintext: hex` |
| `wallet.decryptCredit` | `envelope: hex`, `context` including `role: owner\|spender` | `plaintext: hex` |
| `wallet.authorize` | `operation`, integer `capacity`, `context`, `words: decimal[]` | `spendingPublicKey`, raw low-s `signature` (64 bytes), `digest` (32 bytes) |

All context integers (`chainId`, `protocolAddress`, `tokenAddress`) are canonical nonnegative decimal strings, including addresses. `outputCommitment` is 32-byte hex. Owner identity at the top level is a 20-byte hex EVM address. Bytes are `0x` hex. No request accepts private keys or a detection key.

`Prepare` creates a request. `native.Executor.Execute` runs it with host-owned key material. `Accept` verifies the response and returns `PublicResult`, `[]bool`, `core.Note`, `credit.DeliveryState` or `core.WalletAuthorizationWitness` according to method. Supply the resolved public viewing key when accepting Credit plaintext.

For a transaction, `sdk.PreparedTransaction`, `session.Operation`, `credit.Prepared`, and `credit/session.Operation` expose `ExportWalletRequest(ctx, id)` and `WithWalletResponse(ctx, originalRequest, response)`. Attachment also checks the response against the operation's current frozen statement before recording authorization. The older EIP-712 signature attachment helpers remain available for external EVM signers.

`native/testdata/wallet-wire-vectors.json` is identical to the TypeScript/Rust portable fixture and covers all five methods with a public test key. It is not account storage.

Credit create, draw and update use the same wire authorization. Credit revocation retains separate public EVM owner authorization. `wallet/protocol` holds the dependency-free wire types so Credit and the generic wallet validator share one contract without import cycles.
