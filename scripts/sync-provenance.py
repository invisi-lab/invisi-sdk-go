#!/usr/bin/env python3
"""从经过固定摘要认证的本地协议工作区同步独立来源证明产物。"""
import argparse
import hashlib
import json
from pathlib import Path

PIN = "f90f2ce43218de158c6392fa19e6517dda703c35823a4aa5bfc20b50788ad282"
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("protocol_root", type=Path)
parser.add_argument("--output", type=Path, default=Path(__file__).resolve().parents[1] / ".cache" / "provenance")
args = parser.parse_args()
source = args.protocol_root.resolve()
manifest_bytes = (source / "circuits/provenance/manifest.json").read_bytes()
if hashlib.sha256(manifest_bytes).hexdigest() != PIN:
    raise SystemExit("Untrusted provenance manifest; review and explicitly update the release pin")
manifest = json.loads(manifest_bytes)
if manifest["family"] != "provenance" or manifest["target"] != "evm" or len(manifest["circuits"]) != 28:
    raise SystemExit("Unsupported provenance release")
payloads = [("manifest.json", manifest_bytes)]
for entry in manifest["circuits"]:
    for key in ("artifact", "vk"):
        relative = Path(entry[key])
        if relative.is_absolute() or ".." in relative.parts or relative.parts[:2] != ("circuits", "provenance"):
            raise SystemExit("Invalid provenance artifact path")
        path = (source / relative).resolve()
        if not path.is_relative_to(source):
            raise SystemExit("Provenance artifact escaped checkout")
        data = path.read_bytes()
        if hashlib.sha256(data).hexdigest() != entry[key + "Sha256"]:
            raise SystemExit(f"Provenance {key} integrity mismatch for K{entry['capacity']}")
        payloads.append((str(relative), data))
for relative, data in payloads:
    destination = args.output / relative
    destination.parent.mkdir(parents=True, exist_ok=True)
    destination.write_bytes(data)
print(f"Authenticated {len(manifest['circuits'])} provenance artifacts/VKs to {args.output}")
