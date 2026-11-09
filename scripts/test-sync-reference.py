#!/usr/bin/env python3
"""Authenticated reference imports fail before publication, including with Python assertions disabled."""
import hashlib,json,pathlib,os,subprocess,sys,tempfile,unittest
ROOT=pathlib.Path(__file__).resolve().parents[1]
PROTOCOL=pathlib.Path(os.environ.get('INVISI_PROTOCOL_ROOT',str(ROOT.parent/'invisi-protocol')))
TS=pathlib.Path(os.environ.get('INVISI_TS_ROOT',str(ROOT.parent/'invisi-sdk-ts')))
class SynchronizationTests(unittest.TestCase):
 def command(self,source=TS):
  return subprocess.run([sys.executable,'-O',str(ROOT/'scripts/sync-reference.py'),'--proving-only',str(source),str(PROTOCOL)],capture_output=True)
 def snapshot(self):
  return [hashlib.sha256((ROOT/p).read_bytes()).hexdigest() for p in ['prover/native/circuits/manifest.json','prover/witness-layouts.json','prover/recursive-key.json','internal/release/current.go']]
 def test_current_proving_only_preserves_gas_and_imports_all_native_routes(self):
  before=(ROOT/'gas/data/documented-references.json').read_bytes();result=self.command();self.assertEqual(result.returncode,0,result.stderr.decode());self.assertEqual(before,(ROOT/'gas/data/documented-references.json').read_bytes())
  layouts=json.loads((ROOT/'prover/witness-layouts.json').read_text());self.assertEqual(len(layouts),77);self.assertEqual(sum(x.startswith('transact_aggregate_') for x in layouts),34)
  self.assertIn('input_owner_address',layouts['transact_user_slots1']['private']);self.assertNotIn('transact_user_slots14',layouts)
  keys=json.loads((ROOT/'prover/recursive-key.json').read_text());self.assertEqual(set(keys),set(map(str,range(1,11))))
 def test_retired_identity_or_substituted_key_rejected_before_writes(self):
  for kind in ['retired-noir','missing-route','foreign-key-bytes','foreign-inner-pin','foreign-artifact']:
   with self.subTest(kind=kind),tempfile.TemporaryDirectory(prefix='invisi-sync-negative-') as directory:
    source=pathlib.Path(directory);circuits=source/'packages/prover/circuits';circuits.mkdir(parents=True)
    for p in (TS/'packages/prover/circuits').iterdir():
     if p.name!='manifest.json':(circuits/p.name).symlink_to(p)
    m=json.loads((TS/'packages/prover/circuits/manifest.json').read_text())
    if kind=='retired-noir':m['noirVersion']='1.0.0-rc.6'
    elif kind=='missing-route':m['entries'].pop()
    elif kind=='foreign-key-bytes':m['entries'][0]['keys']['evm']['vkHex']='0x00'
    elif kind=='foreign-inner-pin':m['entries'][0]['keys']['noir-recursive']['innerVkCommitment']='0x01'
    else:m['entries'][0]['artifactSha256']='0'*64
    (circuits/'manifest.json').write_text(json.dumps(m));before=self.snapshot();result=self.command(source);self.assertNotEqual(result.returncode,0);self.assertEqual(before,self.snapshot())
if __name__=='__main__':unittest.main()
