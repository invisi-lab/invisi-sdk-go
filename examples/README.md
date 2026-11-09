# Go examples

Each main is a flat walkthrough. No example loads account keys, creates a local wallet executor, signs EVM transactions or broadcasts a transaction. Wallet execution belongs to an external native host. `prover/native` runs proof computation and is separate from wallet execution.

Install Barretenberg 6.0.0 and a Noir 1.0.0 compatible `noir-execute` for proof examples. Live examples accept `--chain-rpc`, `--protocol-rpc`, `--wallet-address` (public long receiving address), and where applicable `--token` and `--recipient`. Chain and deployment checks run before preparation. Amounts are integer token base units.

```sh
go run ./examples/00-wallet/01-wallet --owner 0x7e5f4552091a69125d5dfcb7b8c2659029395bdf
go run ./examples/00-wallet/02-note-output
go run ./examples/03-hidden-amount/03-notes-with-change \
  --wallet-address '<public receiving address>' --recipient '<public receiving address>' \
  --token 0x2222222222222222222222222222222222222222
```

The first command prints a `wallet.public` request array and waits for the native response array on stdin. The second runs offline using public fixture material only. Live commands use your selected deployment and token.

For an external wallet exchange, read the printed request array, execute it in your native host and write the response array to stdin. Responses may arrive in any order; their exact IDs and identity fields bind them to the original requests. Use explicit per-item errors such as `{"version":1,"id":"...","ownerAddress":"...","method":"wallet.decryptNote","error":{"code":"WALLET_NOT_RECIPIENT"}}`. Only this decrypt nonmatch is skipped during sync; cancellation and unlock failures leave inventory unchanged. Batches have at most 1024 items and 16 MiB serialized JSON, unique IDs and one owner. Do not put keys in requests, responses or command-line settings.

| Directory | Scenario and stdin order |
| --- | --- |
| `00-wallet` | Public request/response bootstrap; offline encrypted Note output from public receiving material. |
| `01-address-registry` | Prepare register/update and export an unsigned direct-owner call. The external transaction caller must be the owner. |
| `02-transparent-amount` | Shield one/nine outputs, memo output, external payer authorization, or authenticated broadcaster pricing. The signature-funded Shield uses `--payer` for Bob while `--wallet-address` selects the recipient-side public wallet. Signature variants print a public EIP-712 funding request and read a JSON hex signature; they export the final unsigned call. |
| `03-hidden-amount` | Public funding, Note inputs, change, withdrawal, owner-bound sweep and recursive aggregation. Note-funded examples automatically discover indexed candidates through canonical receipts and membership checks, print native decrypt request arrays and read response arrays, then print an authorization request array and read its response array. Full sweeps have no private-spending signature and pay only the original owner. Aggregate selects two discovered Notes for each of two owners, obtains one authorization batch per owner, proves four recursive members and exports one aggregate call. Use `--second-wallet-address` for the second owner. |
| `04-credit` | Separate postpaid create/draw/update/revoke plans. The examples discover current role-bound grants automatically, build the business state transition and obtain current chain allocation/time; no grant input JSON is required. Create/update/draw print a wallet request array and read the native response array. Prepare through the managed Credit session for current roots, nullifiers, allocation, balances and cross-role reservations. Revoke exports a direct-owner call without a wallet-spending proof/signature. |
| `05-wallet-session` | Encrypted SQLite restore/refresh, managed transfer and operation recovery. Every refresh prints its request array and reads the response array. Transfer then does the explicit authorization exchange and proof. Recovery uses `--operation-id`, reauthorizes if the saved plan has no proof, and exports calldata. |
| `06-disclosure` | Automatically discover owned Note/Credit candidates using `--wallet-address` and `--token`, exchange native decrypt request/response arrays and independently verify the opening before exporting the scoped disclosure. The payload variant decodes the application memo. A scoped disclosure contains private application data; direct its output to its intended auditor. |

Session and Credit examples use `examples/wallet.sqlite`. Set `INVISI_EXAMPLE_STORAGE_KEY` to an application-owned 32-byte lowercase hex **storage encryption key**. This is independent of wallet cryptographic keys. Keep this key across runs to reopen the journal. A reused Credit operation ID rejects; select a fresh `--operation-id` for each new plan. Selected Note/Credit openings and recovery intents are sensitive inventory data, despite containing no wallet private keys.

`shared.DiscoverNotes` and `DiscoverCredits` use the same canonical managed refresh checks in an ephemeral inventory, with an external JSON native transport; they obtain tree positions, membership and spent/revoked status automatically. Preparation independently revalidates current financial state. Large historical scans exceeding the batch bound must narrow their refresh range. Credit creation takes `--spender`; draw takes `--recipient`; revocation takes the original `--shard` and `--slot` reference. Application-owned detached reservations must remain held until the exported call is canonically resolved. Managed examples retain reservations in the encrypted journal when exporting calldata. Submit the unsigned call through your external EVM signer and track/reconcile its transaction hash through the SDK. Public ERC20 payer allowance and gas funding are separate from wallet authorization.
