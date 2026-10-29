#!/bin/sh
set -eu
protocol_root=${INVISI_PROTOCOL_ROOT:-../../../invisi-protocol}
abigen_bin=${ABIGEN:-abigen}
if ! command -v "$abigen_bin" >/dev/null 2>&1; then
  abigen_bin="$HOME/go/bin/abigen"
fi
python3 - "$protocol_root" <<'PY'
import json, pathlib, sys
root = pathlib.Path(sys.argv[1]) / 'contracts/out'
rows = []
for name in ('InvisiProtocol', 'InvisiFactory', 'VerifierRouter', 'CreditVerifierRouter', 'InvisiAddressRegistry'):
    artifact = json.loads((root / (name + '.sol') / (name + '.json')).read_text())
    assert isinstance(artifact['abi'], list), 'invalid protocol artifact ABI'
    rows.append((name, artifact['abi']))
for name, abi in rows:
    pathlib.Path('abi', name + '.abi.json').write_text(json.dumps(abi, indent=2) + '\n')
PY
for entry in InvisiProtocol:invisi_protocol InvisiFactory:invisi_factory VerifierRouter:verifier_router CreditVerifierRouter:credit_verifier_router IERC20:ierc20 InvisiAddressRegistry:invisi_address_registry; do
  name=${entry%%:*}
  output=${entry#*:}
  "$abigen_bin" --abi "abi/$name.abi.json" --pkg contracts --type "$name" --out "$output.go"
done
