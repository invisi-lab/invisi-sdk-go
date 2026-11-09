# Portable wallet execution

`Request` and `Response` contain only public requests, encrypted envelopes and operation-specific results. Version, ID, owner and method must match. Each response has exactly one `result` or `error: {"code":"..."}`; error codes match `^[A-Z][A-Z0-9_]{0,127}$`. Use a fresh ID per request. The caller retains the original request and authenticates its transport; the SDK verifies identity, context, commitment and spending signatures independently.

| Method | Params | Result |
| --- | --- | --- |
| `wallet.public` | `{}` | `viewingPublicKey` (33 bytes), `spendingPublicKey` (65 bytes), `clueKey` (32 bytes) |
| `wallet.detect` | `clues: hex[]` | `matches: boolean[]` (advisory candidates) |
| `wallet.decryptNote` | `envelope: hex`, `context` | `plaintext: hex` |
| `wallet.decryptCredit` | `envelope: hex`, `context` including `role: owner\|spender` | `plaintext: hex` |
| `wallet.authorize` | `operation`, integer `capacity`, `context`, `words: decimal[]` | `spendingPublicKey`, raw low-s `signature` (64 bytes), `digest` (32 bytes) |

All context integers (`chainId`, `protocolAddress`, `tokenAddress`) are canonical nonnegative decimal strings, including addresses. `outputCommitment` is 32-byte hex. Owner identity at the top level is a 20-byte hex EVM address. Bytes are lowercase `0x` hex. No request accepts private keys or a detection key.

`Prepare` creates a request. A host-owned `wallet.Client.Execute(ctx, []Request) ([]Response, error)` exchanges a complete batch with the external native wallet; no key-holding executor is included in the SDK. `ValidateRequests` checks unique printable ASCII IDs, same owner, at most 1024 items and 16 MiB serialized transport before execution. `ValidateResponses` checks every response binding and returns canonical request order. Empty batches are valid. `AcceptBatch` independently verifies every successful result and returns an error with no partial result if any response fails. `Accept` verifies one response and returns `PublicResult`, `[]bool`, `core.Note`, `credit.DeliveryState` or `core.WalletAuthorizationWitness` according to method. Supply the resolved public viewing key when accepting Credit plaintext.

Managed refresh accepts `[]wallet.Response` directly. Only a matching decrypt response with code `WALLET_NOT_RECIPIENT` skips a candidate. Unlock/cancellation failures and every other error reject the entire apply without inventory or checkpoint writes. Null replies are not nonmatches. An oversized staged refresh fails before returning a handle; narrow its block range.

For a transaction, `sdk.PreparedTransaction`, `session.Operation`, `credit.Prepared`, and `credit/session.Operation` expose `ExportWalletRequest(ctx, id)` and `WithWalletResponse(ctx, originalRequest, response)`. Attachment also checks the response against the operation's current frozen statement before recording authorization. The older EIP-712 signature attachment helpers remain available for external EVM signers.

`internal/testsupport/testdata/wallet-wire-vectors.json` is identical to the TypeScript/Rust portable fixture and covers all five methods. Its key and execution simulator exist only for tests.

Credit create, draw and update use the same wire authorization. Credit revocation retains separate public EVM owner authorization. `wallet/protocol` holds the dependency-free wire types so Credit and the generic wallet validator share one contract without import cycles.
