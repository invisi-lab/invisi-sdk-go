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
require(published['bbVersion'] == '6.0.0' and published['noirVersion'] == '1.0.0', 'unauthenticated protocol release')
release_entries = {entry['name']: entry for entry in published['entries']}
require(len(release_entries)==77,'current release requires 77 circuits')
files, layouts, expected = ({}, {}, set())
credit_manifest_sha = ''
for family, version, fmt in (('transact-user', 4, 7), ('credit', 7, 8)):
    manifest_path = protocol / 'circuits' / ('transact' if family=='transact-user' else family) / 'manifest.json'
    raw = manifest_path.read_bytes()
    if family == 'credit':
        credit_manifest_sha = digest(raw)
    manifest = json.loads(raw)
    require(manifest['complete'] and manifest['version'] == version and (manifest['formatVersion'] == fmt), 'retired/incomplete manifest')
    if family=='transact-user':
        require(manifest['aggregateFormatVersion']==10 and len(manifest['aggregates'])==34,'retired aggregate release')
    require(manifest['bbVersion'] == '6.0.0' and manifest.get('noirVersion','1.0.0') == '1.0.0', 'unauthenticated protocol release')
    require(next((row['sha256'] for row in published['manifests'] if row['name'] == family)) == digest(raw), 'SDK release does not match selected protocol')
    for path, sha in manifest['dependencies'].items():
        require(digest((protocol / path).read_bytes()) == sha, 'stale source: ' + path)
    for entry in manifest['entries'] + manifest.get('aggregates', []):
        name = entry['name']
        expected.add(name)
        artifact = json.loads((protocol / entry['artifact']).read_text())
        require(artifact['noir_version'].startswith('1.0.0+'),'wrong Noir artifact version')
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
recursive={}
for capacity in range(1,11):
    inner=release_entries['transact_user_slots'+str(capacity)]['keys']['noir-recursive']
    recursive[str(capacity)]={'verificationKeyFields':inner['vkFields'],'keyHash':inner['vkHash'],'innerVkCommitment':inner['innerVkCommitment']}
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
    release_dir=root/'protocol/release'
    release_dir.mkdir(parents=True,exist_ok=True)
    (release_dir/'current.go').write_text('// Code generated by scripts/sync-reference.py; DO NOT EDIT.\npackage release\n\nconst CreditManifestVersion = 7\nconst CreditFormatVersion = 8\nconst UserManifestVersion = 4\nconst UserFormatVersion = 7\nconst AggregateFormatVersion = 10\nconst CreditManifestSHA256 = ' + json.dumps(credit_manifest_sha) + '\n')

if proving_only:
    publish_proving()
    print("Synced authenticated current protocol proving assets and layouts; gas and SDK vectors unchanged")
    sys.exit(0)

# Gas data is published as documented medians, never relabeled as raw deployment receipts.
gas_files={name:(source/'packages/protocol/src/gas/data'/name).read_bytes() for name in ('documented-references.json','documented-reference-vectors.json')}
catalog=json.loads(gas_files['documented-references.json'])
require(catalog.get('kind')=='documented-reference' and catalog.get('formatVersion')==1,'unsupported gas reference schema')
require(json.loads(gas_files['documented-reference-vectors.json']).get('catalogSha256')==digest(compact(catalog)),'stale documented gas vectors')
publish_proving()
for name,data in gas_files.items():
    (root/'protocol/gas'/('testdata' if 'vectors' in name else 'data')/name).write_bytes(data)
print('Synced authenticated current protocol assets, layouts and documented gas references')
