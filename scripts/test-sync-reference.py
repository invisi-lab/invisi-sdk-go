#!/usr/bin/env python3
"""Exercise pinned resource synchronization without vendored circuit fixtures."""
import ast,hashlib,json,os,pathlib,re,subprocess,sys,tempfile,unittest
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
class RecursiveMetadataTests(unittest.TestCase):
 def validate(self,binary,vk,pin):
  # Execute the pure validator without running the synchronization CLI or writing files.
  tree=ast.parse((ROOT/'scripts/sync-reference.py').read_text())
  fn=next((node for node in tree.body if isinstance(node,ast.FunctionDef) and node.name=='recursive_metadata'),None)
  self.assertIsNotNone(fn,'generator needs a pure authenticated recursive metadata validator')
  def require(ok,message):
   if not ok:raise ValueError(message)
  scope={'require':require,'re':re}
  exec(compile(ast.Module(body=[fn],type_ignores=[]),'<recursive_metadata>','exec'),scope)
  return scope['recursive_metadata'](binary,vk,pin)
 def fixture(self):
  binary=(ROOT/'prover/testdata/recursive/transact_user_slots1.vk').read_bytes()
  pin=json.loads((ROOT/'prover/recursive-key.json').read_text())['1']
  vk={'vk':['0x'+binary[i:i+32].hex() for i in range(0,len(binary),32)],'hash':pin['keyHash']}
  return binary,vk,pin
 def test_authenticated_binary_json_emits_only_metadata(self):
  binary,vk,pin=self.fixture()
  self.assertEqual(self.validate(binary,vk,pin['innerVkCommitment']),pin)
 def test_binary_json_correspondence_required(self):
  binary,vk,pin=self.fixture();vk['vk'][0]='0x'+'00'*32
  with self.assertRaisesRegex(ValueError,'binary/JSON mismatch'):self.validate(binary,vk,pin['innerVkCommitment'])
 def test_width_and_canonical_fields_required(self):
  binary,vk,pin=self.fixture()
  with self.assertRaisesRegex(ValueError,'width'):self.validate(binary[:-1],vk,pin['innerVkCommitment'])
  modulus=21888242871839275222246405745257275088548364400416034343698204186575808495617
  binary=modulus.to_bytes(32,'big')+binary[32:]
  with self.assertRaisesRegex(ValueError,'noncanonical'):self.validate(binary,vk,pin['innerVkCommitment'])
if __name__=='__main__':unittest.main()
