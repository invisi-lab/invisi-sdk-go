# Go examples

Each `main.go` is an ordinary Go walkthrough with flat public settings and console output. Create public wallets with `sdk.WalletFromAddress(receivingAddress)`. Wallets require a complete `0i` address with receiving keys; `0x` input is rejected. Use `OwnerAddress()` for owner identity and `GetAddress()` for complete receiving material. Identity-only requests can use `wallet.Prepare(id, ownerAddress, method, params)` directly. Construction performs no registry lookup; standalone receiving-address parsing still supports `0x`. Account signing uses geth EIP-712 signatures from environment keys, with an explicit required owner/payer address check. `prover/native` handles local proving.

Install Barretenberg 6.0.0 and Noir 1.0.0 compatible `noir-execute`. Live commands accept `--chain-rpc`, `--protocol-rpc`, `--wallet-address`, and where applicable `--token` and `--recipient`. Amounts are integer token base units. Deployment and canonical chain checks run before financial preparation.

```sh
go run ./examples/02-transparent-amount/00-note-output
export OWNER_PRIVATE_KEY='<owner EVM private key>'
go run ./examples/03-hidden-amount/03-notes-with-change \
  --wallet-address '<public receiving address>' \
  --recipient '<recipient receiving address>' --token '<token address>'
```

`OWNER_PRIVATE_KEY` must match the selected Note/Credit owner or spender; `PAYER_PRIVATE_KEY` must match the public ERC20 payer. Aggregate uses `SECOND_OWNER_PRIVATE_KEY` for its second owner. The SDK independently verifies attached signatures against each frozen statement. Direct-owner registration, withdrawal and revocation examples export calldata for a transaction sent by that owner.

## App delivery decryption

Replace `shared.AppDeliveryDecryptor` in [shared/decryptor.go](shared/decryptor.go) with your App's secure delivery decryption implementation. It is an explicit placeholder that returns an error until integrated. The App implementation is outside the SDK.

`DecryptNote` and `DecryptCredit` return plaintext bytes. Return `nil, nil` only when the delivery is not for this recipient; cancellation, unavailable keys and decryption failures return errors. The SDK decodes plaintext and checks its owner/role, commitment, context and canonical membership before persisting inventory. A failed refresh leaves prior inventory intact. No stdin wallet request/response bridge is used.

`shared.DiscoverNotes` and `DiscoverCredits` accept `sdk.DeliveryDecryptor` and use the same managed refresh validation in ephemeral inventory. Durable examples call `pending.Decrypt(ctx, decryptor)` and retain validated inventory in encrypted SQLite.

| Directory | Scenario |
| --- | --- |
| `01-address-registry` | Register/update public receiving keys and export a direct-owner call. |
| `02-transparent-amount` | Offline Note output; Shield one/nine outputs, memo and a separate payer signature. |
| `03-hidden-amount` | Public/private funding, change, withdrawals, owner-bound sweep and recursive aggregation. |
| `04-credit` | Managed postpaid create/draw/update/revoke, role-bound discovery and reservations. |
| `05-wallet-session` | Encrypted inventory restore, persistent shared public cache, scan/funds status, paginated history, transfer and Note/Credit operation recovery. |
| `06-disclosure` | Authenticated Note/Credit opening and memo disclosure; output is sensitive application data. |
| `07-broadcaster` | Standard relay delivery, independent Source submission, encrypted recovery and scheduled Shield cancellation. |

Most examples log unsigned calldata. Submit through an ordinary geth transaction signer and retain the operation ID/hash for canonical reconciliation. The examples in [07-broadcaster](07-broadcaster/README.md) demonstrate actual sending; its scheduled Shield cancellation example runs its submit-and-cancel flow directly when invoked. Public ERC20 allowance and transaction gas remain separate from private-spending authorization.

Session/Credit examples use `examples/wallet.sqlite`. Set `INVISI_EXAMPLE_STORAGE_KEY` to your application-owned 32-byte hex storage encryption key and retain it across runs. It is independent of the EVM signer and App viewing/decryption keys. Keep the encrypted inventory and journals: they contain sensitive openings and preserve reservations after uncertain submissions. Select a new Credit `--operation-id` only for a new intent; resume the original ID to recover an uncertain operation.

Credit business transition construction remains explicit through `credit.Prepare*`; `managed.Credit().Prepare` validates current allocation, roots/nullifiers and cross-role reservations. `05-wallet-session/04-recover-credit` resumes the frozen Credit journal and does not apply a quota change again. Call `Externalize` before handing direct Credit calldata to a signer, then reconcile the actual hash before releasing reservations.

The restore-session example composes public history and encrypted account records in separate tables of the same SQLite database. Reuse its PublicHistory instance for another public wallet and an account-specific private Store/codec. Default Refresh stages a recent page and Decrypt resumes head-first/backward recovery per committed batch. Return wallet.ErrAuthenticationFailed for ambiguous authentication; native detection keys remain in the App. Earlier committed batches survive interruption. See session/README.md for offline scans, retained coverage and bounded paging.
