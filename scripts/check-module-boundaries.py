#!/usr/bin/env python3
"""Check the Go modules against their actual production package dependencies."""
import json, os, pathlib, subprocess
ROOT = pathlib.Path(__file__).resolve().parents[1]
PREFIX = 'github.com/invisi-labs/invisi-sdk-go'
MODULES = {'sdk': ROOT, **{n: ROOT / n for n in ('primitives', 'protocol', 'prover', 'rpc', 'session', 'ipp', 'paymentrouter')}}
ALLOWED = {'paymentrouter': {'sdk', 'primitives', 'protocol', 'rpc', 'prover'}, 'primitives': set(), 'protocol': {'primitives'}, 'prover': {'primitives', 'protocol'}, 'rpc': {'primitives', 'protocol'}, 'sdk': {'primitives', 'protocol', 'prover', 'rpc'}, 'session': {'sdk', 'primitives', 'protocol', 'rpc', 'prover'}, 'ipp': {'sdk', 'primitives', 'protocol', 'rpc'}}

def owner(path):
    for name in ('primitives', 'protocol', 'prover', 'rpc', 'session', 'ipp', 'paymentrouter'):
        if path == PREFIX + '/' + name or path.startswith(PREFIX + '/' + name + '/'):
            return name
    return 'sdk'
for module, root in MODULES.items():
    output = subprocess.check_output(['go', 'list', '-json', './...'], cwd=root, text=True)
    decoder = json.JSONDecoder()
    at = 0
    count = 0
    while at < len(output):
        while at < len(output) and output[at].isspace():
            at += 1
        if at == len(output):
            break
        package, n = decoder.raw_decode(output[at:])
        at += n
        count += 1
        for imported in package.get('Imports', []):
            if imported == PREFIX or imported.startswith(PREFIX + '/'):
                dependency = owner(imported)
                if module == 'ipp' and '/ipp/invisi' not in package['ImportPath'] and dependency not in ('ipp', 'primitives'):
                    raise SystemExit(f"ipp core imports forbidden {imported}")
                if dependency != module and dependency not in ALLOWED[module]:
                    raise SystemExit(f"{module}: {package['ImportPath']} imports forbidden {imported}")
            if module in ('primitives', 'protocol') and (imported in ('net/http', 'github.com/ethereum/go-ethereum/ethclient', 'github.com/ethereum/go-ethereum/accounts/abi/bind') or imported.startswith('modernc.org/sqlite')):
                raise SystemExit(f'{module}: runtime/network dependency {imported}')
    print(f'{module}: {count} production packages satisfy boundary')
