# Invisi Prover

Local witness construction, authenticated proving assets and proof execution.

Module version: `1.2.0`. See the [SDK repository](https://github.com/invisi-labs/invisi-sdk-go) for the six-module dependency boundary.

## Contributing

See the [contribution guidelines](../README.md#contributing) for issues, pull requests, validation, and handling sensitive material.

Native proving resources are loaded on the first proof or verification. Constructing
`native.NewNativeProver` and inspecting provenance metadata performs no downloads.
The module ships only small statement/witness metadata; public circuit JSON and VKs
come from the pinned snapshot at `https://assets.invisipay.fi/prover/sets/<index-sha256>/`,
with the immutable Core GitHub release as a fallback. The index SHA-256 is fixed in this SDK.
Every index and object is authenticated before use, including cached files.
Existing artifact identities, verification keys, and provenance manifest checks still
apply. Witness generation and proof execution remain local.

`NativeOptions.Resources` optionally configures `native.ResourceOptions`:

```go
Resources: native.ResourceOptions{
    Directory: "/opt/invisi/prover-assets", // index.json and flat <sha256>.json/.vk objects
    CacheDir: "/var/cache/invisi/prover-assets", // default: system cache/invisi/prover-assets
    Offline: true, // fail on missing/corrupt local resources without HTTP
}
```

`Sources: []string{"https://mirror.example/assets/"}` selects ordered mirrors of
that same pinned snapshot. An empty non-nil slice disables remote sources. A
custom source cannot change the trusted index or protocol versions. Downloads have
a two-minute deadline. Canceling a caller returns promptly; shared downloads continue
for other active callers and are aborted when their final caller cancels. An optional `HTTPClient` allows
custom transport configuration. No private witness data enters resource requests.

For air-gapped use, copy an authenticated Core resource output directory to
`Directory`, then enable `Offline`. Generate new tiny metadata with
`scripts/sync-reference.py --proving-only --resource-dir CORE_OUTPUT` after building
Core's `scripts/prover-assets.mjs`. A reviewed new release can explicitly set
`--index-sha256`; synchronization validates all resources before writing metadata.
Set `INVISI_PROVER_RESOURCE_DIR` to run the full local release authentication tests.

`CacheDir` overrides `INVISI_PROVER_CACHE_DIR`, which overrides the system cache directory.

The default disk cache is the system cache directory plus `invisi/prover-assets`:
`~/Library/Caches/invisi/prover-assets` on macOS,
`${XDG_CACHE_HOME:-~/.cache}/invisi/prover-assets` on Linux, and
`%LOCALAPPDATA%/invisi/prover-assets` on Windows.
An explicit cache option overrides `INVISI_PROVER_CACHE_DIR`, which overrides this default.
Index files use `<index-sha256>.index.json`; objects use `<sha256>.json` or `<sha256>.vk`,
so Node, Go and Rust can share the same verified disk cache. Offline mode defaults to false.

Deployment builds can call `native.PrefetchResources(ctx, options, nil)` to prepare
the complete pinned snapshot without invoking Noir or BB. Pass logical resource
paths to prepare only a subset. `native.ResourceIndexSHA256` exposes the SDK's
fixed index identity for build reports. Existing directory/cache files are verified
and reused; missing resources download into `CacheDir` unless offline mode is set.
