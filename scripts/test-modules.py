#!/usr/bin/env python3
"""Run the six public modules and private examples in an isolated local workspace."""
import os, pathlib, subprocess, tempfile
ROOT = pathlib.Path(__file__).resolve().parents[1]
MODULES = ['primitives', 'protocol', 'prover', 'rpc', '.', 'session', 'examples']
with tempfile.TemporaryDirectory(prefix='invisi-six-workspace-') as tmp:
    work = pathlib.Path(tmp) / 'go.work'
    paths = [(ROOT / m).resolve() for m in MODULES]
    work.write_text('go 1.26.1\nuse (\n' + ''.join((str(p) + '\n' for p in paths)) + ')\nreplace (\n' + ''.join(('github.com/invisi-lab/invisi-sdk-go' + ('/' + m if m != '.' else '') + ' v1.2.3 => ' + str(p) + '\n' for m, p in zip(MODULES, paths))) + ')\n')
    env = {**os.environ, 'GOWORK': str(work)}
    subprocess.run(['python3', str(ROOT / 'scripts/check-module-boundaries.py')], env=env, check=True)
    for module, path in zip(MODULES, paths):
        print('VERIFY ' + module, flush=True)
        for command in [['go', 'test', '-race', './...'], ['go', 'vet', './...'], ['go', 'build', './...']]:
            subprocess.run(command, cwd=path, env=env, check=True)
