# Router output recognition and later private spending

A Router payment publicly funds new private Notes. Source recognizes the principal path asynchronously from canonical execution; the transaction sender paying gas does not determine the principal's origin. There is no Router enrollment step and no Provenance input proof for this public funding. Later spending an old private Note requires a real Core proof and a separate Provenance proof.

## Query actual recognition

The guarded command verifies the Protocol deployment against Chain RPC and independently requires a successful finalized transaction before querying Source. It prints the exact service results, including pending/reasons. It does not infer recognition from a receipt or assume that the supplied commitment belongs to the supplied transaction; authenticate that association from canonical output evidence.

```sh
cd examples
go run ./09-provenance/01-router-recognition # offline help, no network
# Supply actual URLs, transaction hash and commitment:
go run ./09-provenance/01-router-recognition --execute \
  --chain-rpc "$CHAIN_RPC" --protocol-rpc "$PROTOCOL_RPC" \
  --transaction-hash "$ROUTER_TX_HASH" --commitment "$NOTE_COMMITMENT"
```

The command uses the existing `shared.SourceClient`: official authenticated HTTPS bootstrap by default, or independently authenticated operator pins for private endpoints. See [shared/source.go](../shared/source.go) for the endpoint, generation/provider IDs, Ed25519 key ID/key and pinned manifest configuration. Never promote unverified service responses or test signing roots to trust anchors. Querying a public commitment reveals interest in that Note; no viewing key or opening is sent.

## Spend a recognized private Note with both real proofs

`private_spend.go` is a callable host-integration example. It accepts the real SDK, canonical chain reader, authenticated Source client, App decryptor, wallet signer, durable artifact persistence and external transaction sender. This is intentionally not a stand-alone CLI: `shared.AppDeliveryDecryptor` is an explicit integration placeholder and returns an error until the App implements secure local decryption. No wallet keys, note openings or mock proof/quote bytes are bundled.

Configure `sdk.NewInvisi` with a real native prover plus `shared.SourceCapability(pool)`; pass that capability's client to `Spend`. The existing native backend pins its proof resources and manifest identity. Barretenberg 6.0.0 and compatible Noir 1.0.0 `noir-execute` must be installed. The host must authenticate the expected scope/manifest and use the same provider/deployment across capabilities.

`Spend` performs canonical inventory discovery through `shared.DiscoverNotes`, authenticates receipt ciphertext and membership through the session refresh path, independently requires finalized native execution, checks unspent inventory and actual Source `recognized` state, and chooses a Note with at least two base units. It transfers half to the recipient and returns change. Unknown/unrecognized Notes are never promoted to usable inputs; unavailable chain/Source evidence stops the operation. Recognition can change after selection: the real proof pipeline still rechecks trusted roots/certificates, member paths and anchors locally.

It then prepares the private-only transfer, obtains explicit wallet authorization, calls `Prove` for Core and `ProveProvenance` for Source, verifies identical exported Core calldata, calls `EvaluateTransaction`, and validates the candidate tracking identity. With `submit=true`, it sends the public packet using `SubmitProvenanceProof`, persists its tracking acknowledgement, sends the same ordinary Core call through the host sender, independently waits for native finality, then reads actual tracking. With `submit=false`, it proves and evaluates but sends neither the Source proof nor the main transaction. The host persistence callback runs before each external submission and after the transaction hash is returned; retain the frozen intent and reconcile uncertain sends instead of creating another payment. Production hosts should use an encrypted operation journal as in [07-broadcaster recovery](../07-broadcaster/03-recover-and-submit-provenance/main.go).

`EvaluateTransaction`/`SubmitProvenanceProof` are offchain JSON-RPC calls carrying only main calldata, public proof sidecars and certificates. Openings, amount/nonce witnesses, member paths and viewing keys remain local. Source proof submission is distinct from onchain execution. Source may remain `pending_confirmation` or `pending_source` after main transaction finality; preserve its reason and track the existing ID with `GetTracking` or bounded `WaitTracking`.

This example is one direct Core `transact`; its SDK packet uses operation index `0`. An operation index identifies a protocol operation inside a transaction, not a private input, output, token or log index. Do not reuse index zero or a Router array position for nested/multiple operations: obtain a reliable Source locator through the deployment-supported path. Core contracts and the existing **11 Provenance public inputs** remain unchanged.

```sh
go test ./09-provenance/...
go vet ./09-provenance/...
```

Compilation validates API integration only. It does not demonstrate live swap liquidity, wallet EIP-7702 compatibility, deployed proving assets or actual Source recognition.
