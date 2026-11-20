"""Optional real-catalog regression: no implicit Periphery sibling during Go metadata export."""
import hashlib
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]

@unittest.skipUnless(os.environ.get('INVISI_TEST_CORE_RESOURCES'), 'authenticated local Core resource set required')
class SeparateResourceRoots(unittest.TestCase):
    def test_catalog_is_sufficient_and_checkout_crosscheck_is_explicit(self):
        assets = Path(os.environ['INVISI_TEST_CORE_RESOURCES']).resolve()
        core = Path(os.environ['INVISI_TEST_CORE_ROOT']).resolve()
        pin = hashlib.sha256((assets / 'index.json').read_bytes()).hexdigest()
        with tempfile.TemporaryDirectory(prefix='sdk-reference-without-sibling-') as directory:
            root = Path(directory) / 'sdk'
            (root / 'scripts').mkdir(parents=True)
            (root / 'prover/native').mkdir(parents=True)
            (root / 'protocol/release').mkdir(parents=True)
            script = root / 'scripts/sync-reference.py'
            shutil.copyfile(ROOT / 'scripts/sync-reference.py', script)
            args = [sys.executable, str(script), '--proving-only', directory, str(core), '--resource-dir', str(assets), '--index-sha256', pin]
            result = subprocess.run(args, capture_output=True, text=True)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertTrue((root / 'prover/witness-layouts.json').is_file())
            result = subprocess.run(args + ['--periphery-root', str(Path(directory) / 'absent')], capture_output=True, text=True)
            self.assertNotEqual(result.returncode, 0)


@unittest.skipUnless(os.environ.get('INVISI_TEST_PROVENANCE_RESOURCES'), 'authenticated Provenance catalog required')
class SharedCacheExport(unittest.TestCase):
    def test_export_preserves_core_index_and_writes_pin_named_provenance_index(self):
        assets=Path(os.environ['INVISI_TEST_PROVENANCE_RESOURCES']).resolve()
        with tempfile.TemporaryDirectory() as temporary:
            output=Path(temporary);old=b'{"existingCoreCatalog":true}'
            (output/'index.json').write_bytes(old)
            result=subprocess.run([sys.executable,str(ROOT/'scripts/sync-provenance.py'),os.environ['INVISI_TEST_PERIPHERY_ROOT'],'--resource-dir',str(assets),'--output',str(output)],capture_output=True,text=True)
            self.assertEqual(result.returncode,0,result.stderr)
            self.assertEqual((output/'index.json').read_bytes(),old)
            pin=hashlib.sha256((assets/'index.json').read_bytes()).hexdigest()
            self.assertEqual((output/(pin+'.index.json')).read_bytes(),(assets/'index.json').read_bytes())

if __name__ == '__main__':
    unittest.main()
