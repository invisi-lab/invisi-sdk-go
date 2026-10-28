#!/usr/bin/env python3
"""Synchronize public cryptographic/protocol fixtures and circuit assets from a TS SDK release."""
import pathlib
import shutil
import sys

root = pathlib.Path(__file__).resolve().parents[1]
source = pathlib.Path(sys.argv[1]) if len(sys.argv) > 1 else root.parent / "invisi-sdk-ts"
for name in (
    "wallet-note-vectors.json", "wallet-authorization-vectors.json", "wallet-kdf-vectors.json",
    "fmd-rust-vectors.json", "decaf377-clue-keys.json",
):
    shutil.copyfile(source / "packages/core/src/testdata" / name, root / "core/testdata" / name)
(root / "sdk/data").mkdir(parents=True, exist_ok=True)
shutil.copyfile(source / "packages/sdk/src/data/gas-benchmarks.json", root / "sdk/data/gas-benchmarks.json")
shutil.copyfile(source / "packages/protocol/src/protocol-vectors.json", root / "protocol/protocol-vectors.json")
(root / "protocol/testdata").mkdir(parents=True, exist_ok=True)
shutil.copyfile(source / "packages/protocol/src/testdata/address-registration-vectors.json", root / "protocol/testdata/address-registration-vectors.json")
shutil.copyfile(source / "packages/protocol/src/witness-layouts.json", root / "prover/witness-layouts.json")
# Mirror the released asset set so removed circuits cannot remain embedded.
circuit_source = source / "packages/prover/circuits"
circuit_target = root / "prover/native/circuits"
shutil.copytree(circuit_source, circuit_target, dirs_exist_ok=True)
for target in circuit_target.iterdir():
    if target.is_file() and not (circuit_source / target.name).is_file():
        target.unlink()

import json
entries = json.loads((source / "packages/prover/circuits/manifest.json").read_text())["entries"]
inner = next(entry for entry in entries if entry["name"] == "transact_user_slots14")["keys"]["noir-recursive"]
(root / "prover/recursive-key.json").write_text(json.dumps({
    "verificationKeyFields": inner["vkFields"], "keyHash": inner["vkHash"],
    "innerVkCommitment": inner["innerVkCommitment"],
}, indent=2) + "\n")
