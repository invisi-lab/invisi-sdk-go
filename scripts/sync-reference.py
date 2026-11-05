#!/usr/bin/env python3
"""Import SDK reference vectors and authenticated current protocol proving assets.

The TS release supplies independently generated SDK vectors/recursive key metadata;
all circuit ABI, bytecode, VKs and layouts are authenticated against the selected
protocol checkout before any destination is written. No proofs or witnesses ship.
"""
import hashlib
import json
import pathlib
import shutil
import sys
import tempfile
root = pathlib.Path(__file__).resolve().parents[1]

def require(condition, message):
    if not condition:
        raise ValueError(message)
arguments = sys.argv[1:]
proving_only = '--proving-only' in arguments
arguments = [value for value in arguments if value != '--proving-only']
source = pathlib.Path(arguments[0]).resolve() if arguments else root.parent / 'invisi-sdk-ts'
protocol = pathlib.Path(arguments[1]).resolve() if len(arguments) > 1 else root.parent / 'invisi-protocol'

def digest(data):
    return hashlib.sha256(data).hexdigest()

def compact(value):
    return json.dumps(value, separators=(',', ':'), ensure_ascii=False).encode()

def artifact_hash(value):
    return digest(compact({'noirVersion': value['noir_version'], 'abi': value['abi'], 'bytecode': value['bytecode']}))
published = json.loads((source / 'packages/prover/circuits/manifest.json').read_text())
require(published['bbVersion'] == '6.0.0' and published['noirVersion'] == '1.0.0-rc.6', 'unauthenticated protocol release')
release_entries = {entry['name']: entry for entry in published['entries']}
files, layouts, expected = ({}, {}, set())
for family, version, fmt in (('transact-user', 3, 7), ('credit', 4, 5)):
    manifest_path = protocol / 'circuits' / (family + '-manifest.json')
    raw = manifest_path.read_bytes()
    manifest = json.loads(raw)
    require(manifest['complete'] and manifest['version'] == version and (manifest['formatVersion'] == fmt), 'retired/incomplete manifest')
    require(manifest['bbVersion'] == '6.0.0' and manifest.get('noirVersion', '1.0.0-rc.6') == '1.0.0-rc.6', 'unauthenticated protocol release')
    require(next((row['sha256'] for row in published['manifests'] if row['name'] == family)) == digest(raw), 'SDK release does not match selected protocol')
    for path, sha in manifest['dependencies'].items():
        require(digest((protocol / path).read_bytes()) == sha, 'stale source: ' + path)
    for entry in manifest['entries'] + manifest.get('aggregates', []):
        name = entry['name']
        expected.add(name)
        artifact = json.loads((protocol / entry['artifact']).read_text())
        require(artifact_hash(artifact) == entry['artifactSha256'], 'stale artifact: ' + name)
        release = release_entries[name]
        require(release['artifactSha256'] == entry['artifactSha256'] and release['publicInputCount'] == entry['publicInputCount'], 'unauthenticated protocol release')
        imported = json.loads((source / 'packages/prover/circuits' / (name + '.json')).read_text())
        require(artifact_hash(imported) == entry['artifactSha256'], 'stale SDK artifact')
        files[name + '.json'] = compact({k: v for k, v in artifact.items() if k not in ('debug_symbols', 'file_map')}) + b'\n'
        keys = {}
        for key in (entry.get('direct', entry), entry.get('recursive')):
            if not key:
                continue
            vk = (protocol / key['vk']).read_bytes()
            require(digest(vk) == key['vkSha256'], 'stale VK')
            vk_json = (protocol / key['vkJson']).read_bytes()
            require(digest(vk_json) == key['vkJsonSha256'], 'stale VK JSON')
            imported_key = release['keys'][key['target']]
            require(imported_key['vkSha256'] == key['vkSha256'] and bytes.fromhex(imported_key['vkHex'][2:]) == vk, 'unauthenticated protocol release')
            if 'innerVkCommitment' in key:
                require(imported_key['innerVkCommitment'] == key['innerVkCommitment'], 'unauthenticated protocol release')
            files[name + '.' + key['target'] + '.vk'] = vk
            keys[key['target']] = key['vkSha256']
        parameters = artifact['abi']['parameters']
        layouts[name] = {'artifactSha256': entry['artifactSha256'], 'keys': keys, 'publicInputCount': entry['publicInputCount'], 'public': [{'name': p['name'], 'count': p['type']['length'] if p['type']['kind'] == 'array' else 1, 'array': p['type']['kind'] == 'array'} for p in parameters if p['visibility'] == 'public'], 'private': [p['name'] for p in parameters if p['visibility'] != 'public']}
require(expected == set(release_entries), 'obsolete/missing SDK circuit')
files['manifest.json'] = (json.dumps(published, indent=2) + '\n').encode()
inner = release_entries['transact_user_slots14']['keys']['noir-recursive']
recursive = {'verificationKeyFields': inner['vkFields'], 'keyHash': inner['vkHash'], 'innerVkCommitment': inner['innerVkCommitment']}
def publish_proving():
    target = root / 'prover/native/circuits'
    with tempfile.TemporaryDirectory(prefix='invisi-prover-sync-', dir=target.parent) as staging:
        stage = pathlib.Path(staging)
        for name, data in files.items():
            (stage / name).write_bytes(data)
        for name in files:
            (stage / name).replace(target / name)
        for path in target.iterdir():
            if path.is_file() and path.name not in files:
                path.unlink()
    (root / 'prover/witness-layouts.json').write_text(json.dumps(layouts, indent=2) + '\n')
    (root / 'prover/recursive-key.json').write_text(json.dumps(recursive, indent=2) + '\n')

if proving_only:
    publish_proving()
    print("Synced authenticated current protocol proving assets and layouts; gas and SDK vectors unchanged")
    sys.exit(0)

gas_catalog = (source / 'packages/sdk/src/data/gas-benchmarks.json').read_bytes()
for profile in json.loads(gas_catalog):
    for measurement in profile['measurements']:
        if 'credit' not in measurement['shape']['kind']:
            continue
        circuit = measurement.get('circuit') or {}
        release = release_entries.get(circuit.get('circuitName'))
        key = release.get('keys', {}).get(circuit.get('target')) if release else None
        require(release and key and circuit.get('artifactSha256') == release['artifactSha256'] and circuit.get('vkSha256') == key['vkSha256'], 'stale Credit gas circuit')
gas_vectors = (source / 'packages/sdk/src/data/gas-reference-vectors.json').read_bytes()
require(json.loads(gas_vectors).get('catalogSha256') == digest(gas_catalog), 'stale gas reference vectors')
# Authenticate transport fixtures before publishing any files. Never fall back to retired docs vectors.
vector_root = source / 'packages/core/src/testdata'
wallet_vectors = json.loads((vector_root / 'wallet-note-vectors.json').read_text())
for name, header, client in (('note', 1, 'private-transfer-client.mjs'), ('credit', 129, 'create-credit-client.mjs')):
    value = wallet_vectors[name]
    provenance = value.get('deliverySource', {})
    ciphertexts = [value['ciphertext']] if name == 'note' else [value['ownerCiphertext'], value['spenderCiphertext']]
    require(provenance.get('envelopeVersion') == 1 and provenance.get('envelopeType') == header >> 7 and provenance.get('header') == header,
            'retired/unauthenticated delivery vector')
    require(provenance.get('protocolClientSha256') == digest((protocol / 'scripts' / client).read_bytes()),
            'retired/unauthenticated delivery vector source')
    for ciphertext in ciphertexts:
        raw = bytes.fromhex(ciphertext[2:])
        require(raw[0] == header and len(raw) == provenance.get('ciphertextBytes') and len(raw) <= 4096,
                'retired/unauthenticated delivery vector bytes')
publish_proving()
for name in ('wallet-note-vectors.json', 'wallet-authorization-vectors.json', 'wallet-kdf-vectors.json', 'fmd-rust-vectors.json', 'decaf377-clue-keys.json'):
    origin = vector_root / name
    shutil.copyfile(origin, root / 'core/testdata' / name)
shutil.copyfile(source / 'packages/protocol/src/testdata/address-registration-vectors.json', root / 'protocol/testdata/address-registration-vectors.json')
(root / 'sdk/data/gas-benchmarks.json').write_bytes(gas_catalog)
(root / 'sdk/testdata/gas-reference-vectors.json').write_bytes(gas_vectors)
shutil.copyfile(source / 'packages/protocol/src/protocol-vectors.json', root / 'protocol/protocol-vectors.json')
print('Synced authenticated current protocol assets, layouts and SDK reference vectors')
