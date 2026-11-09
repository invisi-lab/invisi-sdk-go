# Independent Provenance client

The Provenance family is separate from the existing transaction circuits and Note/Credit spending APIs. It has depth 32, exactly 13 public inputs, and capacities 1–10, then 20–190 in steps of 10. The trusted release manifest SHA-256 is `f90f2ce43218de158c6392fa19e6517dda703c35823a4aa5bfc20b50788ad282`.

Sync authenticated public artifacts from a local protocol checkout:

```sh
python3 scripts/sync-provenance.py ../invisi-protocol
```

This checks the pinned manifest and every artifact/VK digest before copying 28 independent artifacts and keys into `.cache/provenance/`. Existing main circuit manifests, assets and version numbers remain separate. The cache is generated and ignored by Git. The reviewed manifest is tracked with this SDK; updating its pin requires an explicit release review.

Configure scope (chain, protocol, policy), current chain generation, provider, trusted Ed25519 public keys indexed by signingKeyId, and the reviewed manifest digest locally. Responses never establish trust. Certificate payloads use the fixed 301-byte encoding and raw 64-byte Ed25519 signatures. Empty-tree certificates are supported, but proof sources must contain at least one leaf.

The v1 metadata methods `provenance_getVersion` and `provenance_getStatus` are policy-scoped. SDK clients always send `params: [{policyId: "0x<32-byte pinned policy>"}]` from their local trust context, including for a non-default policy. They still reject a response for another policy; the service default is never used as a substitute.

The recommended flow is: fetch a signed root, verify its trust and generation, locally replay paginated public leaves at that fixed root, check contiguous indexes and scoped commitment hashes, verify the final certified root, obtain local membership paths, build the private witness locally, and prove locally. No private opening, owner, nonce, amount, witness, IPN or private key belongs in a service request. The clients reject private-material field names in public submissions. Default membership lookup is a local tree replay; no per-input commitment query is necessary.

Export produces `{operationIndex, proofBundle}`. The bundle has only `format`, `circuitId`, `manifestSha256`, `target`, `publicInputs`, `proof`, and `rootCertificates`; format is `invisi-provenance-proof-v1`, circuit IDs are `provenance_k<K>`, and target is `evm`. Bundle validation checks scope, generation, configured keys, source ordering, sourcesDigest, capacity, operation digest limb bounds, and binding. A host-supplied prover must also verify its output against the authenticated local VK.

Use evaluateTransaction before broadcast, submitProof with a mined operation locator and stable idempotency key, then getProofJob/waitProofJob. Evaluate, submit and job responses must include matching scope, generation, providerId, signingKeyId and manifestSha256. A generation change requires refreshing canonical roots and local paths; it never authorizes changing already signed calldata. Recognition status is independent of whether a Note has been spent. Existing transaction journal schemas are unchanged; the independent recovery coordinator below persists source evidence in its own encrypted namespace.

Policy-enabled broadcaster quotes bind the original public request hash and required policy with `keccak256(abi.encode("InvisiBroadcasterProvenanceIntentV1", originalRequestHash, policyId))`. The signed Quote's six EIP-712 fields stay unchanged; outer metadata must contain `provenancePolicyId` and `requestHashFormat: "invisi-broadcaster-provenance-intent-v1"`. Missing, unexpected or different policy metadata fails closed. Preliminary fee discovery needs no proof; final quote/submission accepts proof sidecars without changing the signed public intent.

## Go packages

Import `github.com/invisi-lab/invisi-sdk-go/provenance`. `ScopeHash`, `NoteOpening`, `RecognizedLeaf`, digests and certificate codecs share the protocol golden vectors. `RecognizedTree`, `BuildWitness`, `AuthenticateManifest`, `ExportSidecar` and `ValidateProofBundle` are independent of main proof types.

`NativeProver` requires explicit trusted `NoirExecutePath` and `BBPath` plus a `LoadAssets` callback returning local manifest/artifact/VK bytes. It verifies all public artifact digests, uses a private temporary directory for local witness execution, checks exact expected public inputs, independently verifies with the pinned VK, and removes temporary material. Command errors omit private command output.

`Client.SyncRecognizedTree(ctx, rootID, certificate, maxLeaves)` requires an explicit replay budget. `Client.WaitProofJob` uses the caller context for cancellation and timeout.

`broadcaster.NewProvenanceClient(configuration, requiredPolicy, options...)` is an explicit policy-bound workflow. Its Quote/Submit methods accept typed proof sidecars. `VerifyProvenanceQuote` validates outer metadata and the bound signature. Policy-disabled SignedQuote decoding rejects provenance metadata rather than dropping it.

Genuine local K1 integration: `INVISI_PROTOCOL_ROOT=../invisi-protocol HARDWARE_CONCURRENCY=4 BB_MSM_LEGACY=1 go test -v ./provenance`.

## Signing-key rotation

Pin the provider and genesis signing public key externally. `getPolicy` may return `initialSigningPublicKey` and public `keyRotations`; these fields never establish trust themselves. Verify the complete history (or an explicitly pinned checkpoint including validity intervals) with `verifyRotationChain` / `VerifyRotationChain` / `verify_rotation_chain`. The shared 169-byte rotation payload uses version 1, provider, old/new key IDs, new public key, activation time and previous rotation hash. Verify the old trusted key signature, SHA-256 new key ID, provider, previous hash, strictly increasing activation times and a caller-supplied current time.

Apply the returned keyring explicitly with `withRotationHistory` / `WithRotationHistory` / `TrustContext::with_rotation_history`; the clients do not accept unknown keys automatically. Historical certificates issued before a key's retirement remain valid; that old key cannot issue certificates at or after the successor's activation. `getPolicy` alone permits an unknown reported current key so its history can be fetched for explicit verification; proving, evaluation, submission and job responses still require trusted keys.

Optional service membership lookup requires an explicit interest-disclosure opt-in; returned paths are still checked locally against the trusted certificate. Local paginated replay remains the default.

## Independent recovery coordinator

Recovery is opt-in and separate from Note/Credit inventory, signing, reservations and the App. Construct an encrypted recovery store over caller-owned atomic ciphertext records, then inject the current trusted Provenance client and local host capabilities. There is no default plaintext file store. The AES-256-GCM record purpose/AAD binds `Invisi.Provenance.RecoveryRecord.v1`, the namespaced record key and CAS revision; wrong keys, keys from another record, and stale revisions fail closed. Local revisions are bounded to 2^53−1 in all three languages so JSON/CAS round trips remain exact for TypeScript. The host keeps encryption keys and raw openings separately; recovery records contain only public evidence and frozen public transaction identifiers.

`prepare` / `Prepare` persists the complete sidecar, certificates, scope/policy, manifest, generation, operation digest, ordered input nullifiers, operation index, original calldata/hash, signed request hash and base idempotency key with `broadcast_possible` **before returning permission for external broadcast**. A crash after receiving permission cannot make the record unfrozen. The host continues to own its existing transaction journal and input/nonce reservations. `markBroadcast` records a known transaction hash but is not the safety fence. Local cancellation stops proving/polling and never cancels an external broadcast or releases reservations.

`resume` / `Resume` / `run` first calls the injected host's authenticated reconciliation of the original frozen transaction. Confirmed canonical locators are persisted before automatic proof submission. An uncertain network response is retried with the same key, derived as `base:SHA256(canonical JSON proofBundle)` (base: 1–63 printable ASCII bytes, derived key: at most 128 bytes; checked before permission); a replacement bundle gets a different submission key. The original calldata, callee/scope, token, ordered nullifiers and operation digest remain frozen. The signed request hash also remains fixed, so the coordinator never silently changes nonce, fee or quote. The host must decode/authenticate this frozen transaction and independently verify local proofs; these are required capabilities, not trust inferred from a service response.

After proof submission, the coordinator stores the job ID and polls it. **A `valid` job only sets `proofValid`; completion requires `getOperationRecognition` for the exact canonical locator and matching operation digest/nullifiers to return `recognized` with its operation checkpoint.** Mixed funding or dependencies can remain `pending_source` or `source_rejected`; provider outages leave resumable state, and resuming a rejected source operation can observe a later decision without changing any old Note's status. No output Note is required, so a zero-output withdrawal still completes through operation recognition.

Current job wire states are `queued`, `verifying`, `retryable_error`, `valid`, `invalid` and `stale_generation`. `retryable_error` keeps the same job pending for the worker to retry; `invalid` becomes local `proof_rejected` and allows explicit local reproof with a different bundle-derived key. `stale_generation` requires fresh host reconciliation and reproof. Unknown spellings fail closed; none of these cryptographic-job states is a funding/source risk rejection. Public `waitProofJob` returns terminal `stale_generation` immediately and rejects unknown states.

Recovery completion is confirmed by the pinned source service's operation-recognition response under the configured trust model. It is not independent authentication of output Note membership. Receivers still separately authenticate root certificates, canonical anchors and membership when accepting a Note.

A generation change or reorg returns `needs_reproof` after host reconciliation. With explicit local reproof enabled, the host may supply a new sidecar only after the service trust context is explicitly updated to that generation; the coordinator rechecks certificates, binding, ordered input digest and the entire frozen statement. No remote proving URL or raw opening argument exists. On interruption/restart, instantiate another coordinator over the same encrypted records and call resume/run with the same ID. Local `filterRecognizedCommitments` / `FilterRecognizedCommitments` / `filter_recognized_commitments` helpers filter through a rebuilt `RecognizedTree`; protocol spendability and source recognition remain separate, with no per-note network interest query.

Recovery state-machine tests use explicitly marked synthetic proof bytes and real signed root certificates; they validate recovery/encryption behavior and do not claim another native-prover or App integration test.

Go: `NewSQLiteRecoveryRecords(ctx, callerOwnedDB)` creates only the independent `invisi_provenance_recovery` ciphertext table; pass it to `NewEncryptedRecoveryStore(records, key32)`. Alternatively implement `AtomicRecoveryRecords` for your KV. Construct `RecoveryCoordinator{Store: store, Trust: trust, Service: client, Host: capabilities}`; a successful `Prepare` is the broadcast permit. The caller owns the database/key lifecycle.

## Flat examples

See [`examples/07-provenance`](examples/07-provenance/README.md) for local proof generation, public sidecars, verified Broadcaster quote/explicit submission and confirmed source publication.
