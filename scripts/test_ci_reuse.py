#!/usr/bin/env python3
"""Guard CI reuse against unrelated and incomplete validation results."""
import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location('ci_reuse', Path(__file__).with_name('check-ci-reuse.py'))
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


class ReuseTests(unittest.TestCase):
    def setUp(self):
        self.run_record = dict(head_sha='abc', head_branch='develop', event='push',
                               status='completed', conclusion='success')

    def test_exact_completed_push(self):
        self.assertTrue(module.verified([self.run_record], 'abc', ['develop']))

    def test_rejects_incomplete_or_untrusted_results(self):
        for key, value in [('head_sha', 'other'), ('head_branch', 'feature'),
                           ('event', 'pull_request'), ('status', 'in_progress'),
                           ('conclusion', 'failure'), ('conclusion', 'cancelled'),
                           ('conclusion', 'skipped')]:
            with self.subTest(key=key, value=value):
                self.assertFalse(module.verified([dict(self.run_record, **{key: value})], 'abc', ['develop']))

    def test_main_requires_develop(self):
        record = dict(self.run_record, head_branch='main')
        self.assertFalse(module.verified([record], 'abc', ['develop']))
        self.assertTrue(module.verified([record], 'abc', ['develop', 'main']))

    def test_missing_records_require_verification(self):
        self.assertFalse(module.verified([], 'abc', ['develop']))


if __name__ == '__main__':
    unittest.main()
