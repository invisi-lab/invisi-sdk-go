# App delivery decryption and optional wallet codecs

Normal SDK flows inject `DeliveryDecryptor`, implemented by the App. `DecryptNote` receives ciphertext plus the owner/deployment/token/commitment context. `DecryptCredit` also receives the grant role. Only `nil, nil` is a nonrecipient; any error aborts. `sdk.Wallet` methods validate returned plaintext and Credit receiving keys. The App owns authenticated decryption and key management.

Session stages canonical pages through `Refresh` then atomically validates/commits through `pending.Decrypt(ctx, decryptor)`. Signing uses exported EIP-712 requests and ordinary geth signatures. No native transport or wallet response matching is part of these normal flows.

The remainder documents optional transport-neutral codecs for App-defined integrations.

`Request` and `Response` contain only public requests, encrypted envelopes and operation-specific results. Version, ID, owner and method must match. Each response has exactly one `result` or `error: {"code":"..."}`; error codes match `^[A-Z][A-Z0-9_]{0,127}$`. Use a fresh ID per request. The caller retains the original request and authenticates its transport; the SDK verifies identity, context, commitment and spending signatures independently.

| Method | Params | Result |
| --- | --- | --- |
| `wallet.public` | `{}` | `viewingPublicKey` (33 bytes), `spendingPublicKey` (65 bytes), `clueKey` (32 bytes) |
| `wallet.detect` | `clues: hex[]` | `matches: boolean[]` (advisory candidates) |
| `wallet.decryptNote` | `envelope: hex`, `context` | `plaintext: hex` |
| `wallet.decryptCredit` | `envelope: hex`, `context` including `role: owner\|spender` | `plaintext: hex` |
| `wallet.authorize` | `operation`, integer `capacity`, `context`, `words: decimal[]` | `spendingPublicKey`, raw low-s `signature` (64 bytes), `digest` (32 bytes) |

All context integers (`chainId`, `protocolAddress`, `tokenAddress`) are canonical nonnegative decimal strings, including addresses. `outputCommitment` is 32-byte hex. Owner identity at the top level is a 20-byte hex EVM address. Bytes are lowercase `0x` hex. No request accepts private keys or a detection key.

`Prepare` creates a request. An App may implement `wallet.Client.Execute(ctx, []Request) ([]Response, error)` for its own transport; the SDK implements no bridge or executor. `ValidateRequests` checks unique printable ASCII IDs, same owner, at most 1024 items and 16 MiB serialized transport before execution. `ValidateResponses` checks every response binding and returns canonical request order. Empty batches are valid. `AcceptBatch` independently verifies every successful result and returns an error with no partial result if any response fails. `Accept` verifies one response and returns `PublicResult`, `[]bool`, `core.Note`, `credit.DeliveryState` or `core.WalletAuthorizationWitness` according to method. Supply the resolved public viewing key when accepting Credit plaintext.

Low-level operation wire requests remain optional; EIP-712 export/signature attachment is the normal transaction authorization interface. Both paths bind authorization to the current frozen statement.

`internal/testsupport/testdata/wallet-wire-vectors.json` is identical to the TypeScript/Rust portable fixture and covers all five methods. Its key and execution simulator exist only for tests.

Credit create, draw and update use the same wire authorization. Credit revocation retains separate public EVM owner authorization. `wallet/protocol` holds the dependency-free wire types so Credit and the generic wallet validator share one contract without import cycles.
