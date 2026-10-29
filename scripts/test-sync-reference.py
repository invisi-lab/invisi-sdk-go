#!/usr/bin/env python3
"""Regression: retired manifests fail before publication, even with Python optimizations."""
import hashlib
import json
import pathlib
import subprocess
import sys
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]
PROTOCOL = ROOT.parent / "invisi-protocol"
TS = ROOT.parent / "invisi-sdk-ts"

class SynchronizationTests(unittest.TestCase):
    def test_stale_credit_gas_fingerprint_rejected_before_writes(self):
        before = (ROOT / "sdk/data/gas-benchmarks.json").read_bytes()
        with tempfile.TemporaryDirectory(prefix="invisi-gas-sync-test-") as directory:
            checkout = pathlib.Path(directory)
            for relative in ("packages/prover/circuits", "packages/core/src/testdata", "packages/protocol/src"):
                destination = checkout / relative
                destination.parent.mkdir(parents=True, exist_ok=True)
                destination.symlink_to(TS / relative, target_is_directory=True)
            destination = checkout / "packages/sdk/src/data"
            destination.mkdir(parents=True)
            profiles = json.loads((TS / "packages/sdk/src/data/gas-benchmarks.json").read_text())
            credit = next(measurement for profile in profiles for measurement in profile["measurements"] if "credit" in measurement["shape"]["kind"])
            credit["circuit"]["artifactSha256"] = "0" * 64
            (destination / "gas-benchmarks.json").write_text(json.dumps(profiles))
            result = subprocess.run([sys.executable, "-O", str(ROOT / "scripts/sync-reference.py"), str(checkout), str(PROTOCOL)], capture_output=True)
            self.assertNotEqual(result.returncode, 0, "stale Credit gas catalog was published")
            self.assertIn(b"stale Credit gas circuit", result.stderr)
        self.assertEqual(before, (ROOT / "sdk/data/gas-benchmarks.json").read_bytes())

    def test_retired_credit_manifest_rejected_even_when_optimized(self):
        before = (ROOT / "prover/native/circuits/manifest.json").read_bytes()
        with tempfile.TemporaryDirectory(prefix="invisi-sync-test-") as directory:
            checkout = pathlib.Path(directory)
            for path in PROTOCOL.iterdir():
                if path.name != "circuits":
                    (checkout / path.name).symlink_to(path, target_is_directory=path.is_dir())
            circuits = checkout / "circuits"
            circuits.mkdir()
            for path in (PROTOCOL / "circuits").iterdir():
                if path.name != "credit-manifest.json":
                    (circuits / path.name).symlink_to(path, target_is_directory=path.is_dir())
            manifest = json.loads((PROTOCOL / "circuits/credit-manifest.json").read_text())
            manifest["version"] = 3
            (circuits / "credit-manifest.json").write_text(json.dumps(manifest))
            result = subprocess.run([sys.executable, "-O", str(ROOT / "scripts/sync-reference.py"), str(TS), str(checkout)], capture_output=True)
            self.assertNotEqual(result.returncode, 0, "optimized Python bypassed release authentication")
            self.assertIn(b"retired/incomplete manifest", result.stderr)
        self.assertEqual(hashlib.sha256(before).digest(), hashlib.sha256((ROOT / "prover/native/circuits/manifest.json").read_bytes()).digest())

if __name__ == "__main__":
    unittest.main()
