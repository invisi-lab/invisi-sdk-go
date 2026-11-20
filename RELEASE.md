

## CI result reuse

Push checks run on `develop` and `main`; pull requests targeting either branch remain checked. `main` reuses successful `develop` push CI only for the identical commit SHA. Tag releases reuse successful `develop` or `main` push CI for that exact SHA; otherwise they run verification before release. Running, cancelled, skipped and failed results are never accepted. API failures block release instead of granting reuse. New merge commits require their own checks. Publishing remains tag-only, with package/archive validation retained. Branch pushes do not create releases.

Tagged releases validate the canonical version, every internal dependency pin, and all eight actual remote Git tags (including annotated tags) against the checked-out commit before checking archives. A missing dependency cache is treated as a cold start; external dependencies can be downloaded into the isolated cache.

Release jobs require `!cancelled()` as well as successful verification. Cancelling a workflow therefore also cancels publication work; completed registry writes remain immutable.

The current release cohort is **v1.3.2**: root SDK plus primitives, protocol, prover, rpc, session, ipp, and paymentrouter. Every internal module requirement uses the same version; optional modules publish only after their SDK and protocol dependencies.
