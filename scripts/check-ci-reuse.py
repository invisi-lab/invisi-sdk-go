#!/usr/bin/env python3
"""Reuse only completed push CI for the identical commit on trusted branches."""
import json
import os
import subprocess
import sys


def verified(runs, commit, branches):
    return any(run.get('head_sha') == commit
               and run.get('head_branch') in branches
               and run.get('event') == 'push'
               and run.get('status') == 'completed'
               and run.get('conclusion') == 'success'
               for run in runs)


def main():
    commit, branches = sys.argv[1:]
    repository = os.environ['GITHUB_REPOSITORY']
    response = subprocess.check_output([
        'gh', 'api', f'repos/{repository}/actions/workflows/ci.yml/runs'
        f'?head_sha={commit}&status=success&event=push&per_page=100',
    ], text=True)
    matches = verified(json.loads(response)['workflow_runs'], commit, branches.split(','))
    print('verified=' + str(matches).lower())


if __name__ == '__main__':
    main()
