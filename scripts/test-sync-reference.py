#!/usr/bin/env python3
"""Exercise pinned resource synchronization without vendored circuit fixtures."""
import hashlib,os,pathlib,subprocess,sys,tempfile,unittest
ROOT=pathlib.Path(__file__).resolve().parents[1]
RESOURCE=os.environ.get('INVISI_PROVER_RESOURCE_DIR')
class SynchronizationTests(unittest.TestCase):
 def snapshot(self):return [hashlib.sha256((ROOT/p).read_bytes()).hexdigest() for p in ['prover/native/resource_pin.go','prover/witness-layouts.json','prover/recursive-key.json','protocol/release/current.go']]
 def command(self,directory):return subprocess.run([sys.executable,'-O',str(ROOT/'scripts/sync-reference.py'),'--proving-only','--resource-dir',str(directory)],capture_output=True)
 def test_wrong_index_rejected_before_writes(self):
  with tempfile.TemporaryDirectory() as d:
   pathlib.Path(d,'index.json').write_text('{}');before=self.snapshot();r=self.command(d);self.assertNotEqual(r.returncode,0);self.assertEqual(before,self.snapshot())
 @unittest.skipUnless(RESOURCE,'set INVISI_PROVER_RESOURCE_DIR for authenticated release integration')
 def test_current_release_is_reproducible(self):
  before=self.snapshot();r=self.command(RESOURCE);self.assertEqual(r.returncode,0,r.stderr.decode());self.assertEqual(before,self.snapshot())
if __name__=='__main__':unittest.main()
