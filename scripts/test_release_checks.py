"""Guard cold-cache operation and actual Go release tag/version identity."""
import copy
import tempfile
from pathlib import Path
import unittest
from unittest.mock import patch
from release_checks import MODULE_NAMES, PREFIX, cache_links, validate_manifests, validate_tag_refs, validate_tagged_release


class ReleaseChecksTests(unittest.TestCase):
    def setUp(self):
        self.version = 'v1.2.6'
        self.commit = 'a' * 40
        self.manifests = {PREFIX + ('/' + name if name else ''): {
            'Module': {'Path': PREFIX + ('/' + name if name else '')},
            'Require': [{'Path': PREFIX + '/primitives', 'Version': self.version}],
        } for name in MODULE_NAMES}
        self.refs = '\n'.join(self.commit + ' refs/tags/' + (name + '/' if name else '') + self.version for name in MODULE_NAMES)

    def test_missing_dependency_cache_is_optional(self):
        with tempfile.TemporaryDirectory() as d:
            root = Path(d)
            cache_links(root / 'missing', root / 'seeded')
            self.assertTrue((root / 'seeded').is_dir())
            self.assertEqual(list((root / 'seeded').iterdir()), [])

    def test_warm_cache_keeps_external_dependencies_only(self):
        with tempfile.TemporaryDirectory() as d:
            root = Path(d); source = root / 'source'
            (source / 'example.org/pkg@v1.0.0').mkdir(parents=True)
            (source / 'github.com/invisi-labs/invisi-sdk-go@v1.2.6').mkdir(parents=True)
            (source / 'cache').mkdir()
            cache_links(source, root / 'seeded')
            self.assertTrue((root / 'seeded/example.org/pkg@v1.0.0').is_symlink())
            self.assertFalse((root / 'seeded/github.com/invisi-labs/invisi-sdk-go@v1.2.6').exists())
            self.assertFalse((root / 'seeded/cache').exists())

    def test_all_lightweight_and_annotated_tags_match_commit(self):
        validate_tag_refs(self.refs, self.version, self.commit)
        annotated = self.refs.replace(self.commit + ' refs/tags/primitives/', 'b' * 40 + ' refs/tags/primitives/')
        annotated += '\n' + self.commit + ' refs/tags/primitives/' + self.version + '^{}'
        validate_tag_refs(annotated, self.version, self.commit)

    def test_missing_or_moved_submodule_tag_is_rejected(self):
        for refs in [self.refs.split('\n', 1)[1], self.refs.replace(self.commit, 'b' * 40, 1)]:
            with self.assertRaisesRegex(ValueError, 'Missing tag or wrong commit'):
                validate_tag_refs(refs, self.version, self.commit)

    def test_exact_module_versions(self):
        validate_manifests(self.manifests, self.version)
        for change in ['version', 'identity', 'missing', 'replacement', 'unknown']:
            manifests = copy.deepcopy(self.manifests)
            item = manifests[PREFIX]
            if change == 'version': item['Require'][0]['Version'] = 'v1.2.5'
            if change == 'identity': item['Module']['Path'] = 'example.org/wrong'
            if change == 'missing': del manifests[PREFIX + '/ipp']
            if change == 'replacement': item['Replace'] = [{'Old': {'Path': PREFIX + '/primitives'}, 'New': {'Path': '../primitives'}}]
            if change == 'unknown': item['Require'][0]['Path'] = PREFIX + '/unknown'
            with self.subTest(change=change), self.assertRaises(ValueError):
                validate_manifests(manifests, self.version)

    def test_release_queries_actual_remote_refs_after_reading_all_manifests(self):
        import json
        output = [json.dumps(self.manifests[PREFIX + ('/' + name if name else '')]) for name in MODULE_NAMES]
        with patch('release_checks.subprocess.check_output', side_effect=output + [self.commit, self.refs]) as command:
            validate_tagged_release(Path('/checkout'), self.version)
            args = command.call_args.args[0]
            self.assertEqual(args[:4], ['git', 'ls-remote', '--tags', 'origin'])
            self.assertIn('refs/tags/ipp/' + self.version, args)
            self.assertIn('refs/tags/' + self.version + '^{}', args)
        with patch('release_checks.subprocess.check_output', side_effect=output + [self.commit, '']):
            with self.assertRaisesRegex(ValueError, 'Missing tag'):
                validate_tagged_release(Path('/checkout'), self.version)

    def test_invalid_release_tag_fails_before_commands(self):
        for tag in ['v01.2.6', '1.2.6', 'v1.2.6-rc.1', 'main']:
            with patch('release_checks.subprocess.check_output') as command:
                with self.assertRaises(ValueError): validate_tagged_release(Path('.'), tag)
                command.assert_not_called()


if __name__ == '__main__':
    unittest.main()
