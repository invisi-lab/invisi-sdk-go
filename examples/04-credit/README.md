# Managed Credit examples

Create, draw, update and revoke retain encrypted inventory and managed cross-role grant reservations. Refresh uses the App delivery decryptor placeholder in [shared/decryptor.go](../shared/decryptor.go); replace it with your secure App integration. Draw/update reuse the validated refresh inventory.

Create/update/draw export the frozen EIP-712 statement and sign with ordinary geth using `OWNER_PRIVATE_KEY`. The key must match the selected transition's owner/spender. Revoke exports a direct-owner call. Set `INVISI_EXAMPLE_STORAGE_KEY` to the retained application-owned 32-byte storage key; use a fresh `--operation-id` for each new transition.

Business quota transitions and delivery encryption are explicit in these walkthroughs. `managed.Credit().Prepare` validates chain state and reserves grants; local proving independently verifies the result. The examples log unsigned calldata and call `Externalize` before exposing it. Retain the encrypted journal until canonical reconciliation resolves the original operation.

Relay delivery is consolidated under [07-broadcaster](../07-broadcaster/README.md). Source evidence applies to private Note ancestry; the ordinary Credit transitions shown here do not require a Source proof.
