#!/usr/bin/env python3
"""Export pinned provenance resources into an offline native ResourceOptions.Directory.

Build the independent Periphery catalog with Periphery scripts/prover-assets.mjs.
This copies only public provenance objects and the SDK-pinned Provenance index; unrelated
circuit objects may be added to the same directory later from that snapshot.
"""
import argparse, hashlib, json, re, os, tempfile
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
p=argparse.ArgumentParser(description=__doc__)
p.add_argument('periphery_root',type=Path)
p.add_argument('--legacy-combined',action='store_true')
p.add_argument('--resource-dir',required=True,type=Path)
p.add_argument('--output',type=Path,default=ROOT/'.cache'/'provenance')
a=p.parse_args()
def require(ok,message):
    if not ok:raise ValueError(message)
def digest(data):return hashlib.sha256(data).hexdigest()
pin_name='LegacyResourceIndexSHA256' if a.legacy_combined else 'ProvenanceResourceIndexSHA256'
pin=re.search(r'const '+pin_name+r'\s*=\s*"([0-9a-f]{64})"',(ROOT/'prover/native/resource_pin.go').read_text())[1]
raw=(a.resource_dir/'index.json').read_bytes();require(digest(raw)==pin,'untrusted resource index')
index=json.loads(raw);require(index['format']=='invisi.prover-assets.v1' and index['noirVersion']=='1.0.0' and index['bbVersion']=='6.0.0','unsupported resource version')
payloads={pin+'.index.json':raw}
def load(name):
    row=index['files'][name];require(re.fullmatch('[0-9a-f]{64}',row['sha256']),'invalid object hash')
    filename=row['sha256']+('.vk' if name.endswith('/vk') else '.json')
    data=(a.resource_dir/filename).read_bytes();require(len(data)==row['size'] and digest(data)==row['sha256'],'corrupt resource')
    payloads[filename]=data;return data
manifest=load('provenance/manifest.json')
require(digest(manifest)==digest((a.periphery_root/'circuits/provenance/manifest.json').read_bytes())=='6107789c5c0d24c4130bff05da6d389336feb5ccb247c4030eae5198ab05d762','untrusted provenance manifest')
for entry in json.loads(manifest)['circuits']:
    for field in ('artifact','vk'):
        path=entry[field];require(path.startswith('circuits/provenance/') and '..' not in Path(path).parts and '\\' not in path,'invalid provenance path')
        require(digest(load('provenance/'+path))==entry[field+'Sha256'],'untrusted provenance identity')
a.output.mkdir(parents=True,exist_ok=True)
# Publish the verified index after all of its exported objects.
for name,data in sorted(payloads.items(),key=lambda item:item[0].endswith('.index.json')):
    temporary=None
    try:
        with tempfile.NamedTemporaryFile(dir=a.output,prefix='.',delete=False) as handle:
            temporary=Path(handle.name)
            handle.write(data)
        temporary.replace(a.output/name)
    finally:
        if temporary is not None:temporary.unlink(missing_ok=True)
try:os.link(a.output/(pin+'.index.json'),a.output/'index.json')
except FileExistsError:pass
print(f'Authenticated pinned offline provenance resources to {a.output}')
