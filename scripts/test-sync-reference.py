#!/usr/bin/env python3
"""Regression: retired manifests fail before publication, even with Python optimizations."""
import hashlib
import json
import pathlib
import os
import subprocess
import sys
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]
PROTOCOL = pathlib.Path(os.environ.get("INVISI_PROTOCOL_ROOT", str(ROOT.parent / "invisi-protocol")))
TS = pathlib.Path(os.environ.get("INVISI_TS_ROOT", str(ROOT.parent / "invisi-sdk-ts")))

class SynchronizationTests(unittest.TestCase):
    def test_proving_only_import_does_not_publish_gas_catalog(self):
        before = (ROOT / "internal/gas/data/gas-benchmarks.json").read_bytes()
        result = subprocess.run([sys.executable, str(ROOT / "scripts/sync-reference.py"), "--proving-only", str(TS), str(PROTOCOL)], capture_output=True)
        self.assertEqual(result.returncode, 0, result.stderr.decode())
        self.assertEqual(before, (ROOT / "internal/gas/data/gas-benchmarks.json").read_bytes())
        layouts = json.loads((ROOT / "prover/witness-layouts.json").read_text())
        self.assertIn("input_owner_address", layouts["transact_user_slots14"]["private"])
        self.assertIn("output_identities", layouts["transact_user_slots14"]["private"])

    def test_wallet_vectors_require_current_authenticated_envelopes(self):
        before = (ROOT / "prover/native/circuits/manifest.json").read_bytes()
        with tempfile.TemporaryDirectory(prefix="invisi-envelope-sync-test-") as directory:
            checkout = pathlib.Path(directory)
            for relative in ("packages/prover/circuits", "packages/sdk/src/gas/data", "packages/protocol/src", "packages/core/src/fmd/testdata"):
                destination = checkout / relative
                destination.parent.mkdir(parents=True, exist_ok=True)
                destination.symlink_to(TS / relative, target_is_directory=True)
            destination = checkout / "packages/core/src/wallet/testdata"
            destination.mkdir(parents=True)
            for path in (TS / "packages/core/src/wallet/testdata").iterdir():
                if path.name != "wallet-note-vectors.json":
                    (destination / path.name).symlink_to(path)
            vectors = json.loads((TS / "packages/core/src/wallet/testdata/wallet-note-vectors.json").read_text())
            vectors["credit"]["deliverySource"]["envelopeVersion"] = 5
            (destination / "wallet-note-vectors.json").write_text(json.dumps(vectors))
            result = subprocess.run([sys.executable, "-O", str(ROOT / "scripts/sync-reference.py"), str(checkout), str(PROTOCOL)], capture_output=True)
            self.assertNotEqual(result.returncode, 0)
            self.assertIn(b"retired/unauthenticated delivery vector", result.stderr)
        self.assertEqual(before, (ROOT / "prover/native/circuits/manifest.json").read_bytes())

    def test_stale_credit_gas_fingerprint_rejected_before_writes(self):
        before = (ROOT / "internal/gas/data/gas-benchmarks.json").read_bytes()
        with tempfile.TemporaryDirectory(prefix="invisi-gas-sync-test-") as directory:
            checkout = pathlib.Path(directory)
            for relative in ("packages/prover/circuits", "packages/core/src/wallet/testdata", "packages/core/src/fmd/testdata", "packages/protocol/src"):
                destination = checkout / relative
                destination.parent.mkdir(parents=True, exist_ok=True)
                destination.symlink_to(TS / relative, target_is_directory=True)
            destination = checkout / "packages/sdk/src/gas/data"
            destination.mkdir(parents=True)
            profiles = json.loads((TS / "packages/sdk/src/gas/data/gas-benchmarks.json").read_text())
            credit = next(measurement for profile in profiles for measurement in profile["measurements"] if "credit" in measurement["shape"]["kind"])
            credit["circuit"]["artifactSha256"] = "0" * 64
            (destination / "gas-benchmarks.json").write_text(json.dumps(profiles))
            result = subprocess.run([sys.executable, "-O", str(ROOT / "scripts/sync-reference.py"), str(checkout), str(PROTOCOL)], capture_output=True)
            self.assertNotEqual(result.returncode, 0, "stale Credit gas catalog was published")
            self.assertIn(b"stale Credit gas circuit", result.stderr)
        self.assertEqual(before, (ROOT / "internal/gas/data/gas-benchmarks.json").read_bytes())

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
