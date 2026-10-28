#!/bin/sh
set -eu
ts_root=${INVISI_TS_SDK_ROOT:-../../../invisi-sdk-ts}
abigen_bin=${ABIGEN:-abigen}
if ! command -v "$abigen_bin" >/dev/null 2>&1; then
  abigen_bin="$HOME/go/bin/abigen"
fi
python3 - "$ts_root" <<'PY'
import json, pathlib, re, sys
root = pathlib.Path(sys.argv[1]) / 'packages/protocol/src/contracts/abis'
for name in ('InvisiProtocol', 'InvisiFactory', 'VerifierRouter', 'CreditVerifierRouter', 'InvisiAddressRegistry'):
    source = (root / (name + '.ts')).read_text()
    payload = re.split(r'export const \w+\s*=\s*', source, maxsplit=1)[1].rsplit(' as const', 1)[0]
    abi = json.loads(payload)
    pathlib.Path('abi', name + '.abi.json').write_text(json.dumps(abi, indent=2) + '\n')
PY
for entry in InvisiProtocol:invisi_protocol InvisiFactory:invisi_factory VerifierRouter:verifier_router CreditVerifierRouter:credit_verifier_router IERC20:ierc20 InvisiAddressRegistry:invisi_address_registry; do
  name=${entry%%:*}
  output=${entry#*:}
  "$abigen_bin" --abi "abi/$name.abi.json" --pkg contracts --type "$name" --out "$output.go"
done
