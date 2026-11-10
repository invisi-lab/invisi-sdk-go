#!/usr/bin/env python3
"""Test unpublished module archives through an isolated local Go proxy/cache."""
import json, os, pathlib, shutil, subprocess, tempfile, zipfile
ROOT = pathlib.Path(__file__).resolve().parents[1]
PREFIX = 'github.com/invisi-lab/invisi-sdk-go'
VERSION = 'v1.2.0'
MODULES = [('primitives', ROOT / 'primitives'), ('protocol', ROOT / 'protocol'), ('prover', ROOT / 'prover'), ('rpc', ROOT / 'rpc'), ('', ROOT), ('session', ROOT / 'session')]
work = pathlib.Path(tempfile.mkdtemp(prefix='invisi-go-package-check-'))
proxy = work / 'proxy'
cache = work / 'modcache'
cache.mkdir()
print('Package evidence ' + str(work), flush=True)
existing = pathlib.Path(subprocess.check_output(['go', 'env', 'GOMODCACHE'], text=True).strip())

def cache_links(src, dst):
    dst.mkdir(parents=True, exist_ok=True)
    for p in src.iterdir():
        if p.name == 'cache' or p.name.startswith('invisi-sdk-go'):
            continue
        target = dst / p.name
        if '@' in p.name:
            target.symlink_to(p, target_is_directory=p.is_dir())
        elif p.is_dir():
            cache_links(p, target)
cache_links(existing, cache)

def download_links(src, dst, relative=''):
    dst.mkdir(parents=True, exist_ok=True)
    for p in src.iterdir():
        rel = relative + '/' + p.name
        if 'github.com/invisi-lab/invisi-sdk-go' in rel:
            continue
        target = dst / p.name
        if p.name == '@v':
            target.symlink_to(p, target_is_directory=True)
        elif p.is_dir():
            download_links(p, target, rel)
        elif p.name not in ('lock', 'sumdb'):
            target.symlink_to(p)
if (existing / 'cache/download').exists():
    download_links(existing / 'cache/download', cache / 'cache/download')
files = subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard'], cwd=ROOT, text=True).splitlines()
for name, base in MODULES:
    module = PREFIX + ('/' + name if name else '')
    dest = proxy / module / '@v'
    dest.mkdir(parents=True, exist_ok=True)
    (dest / (VERSION + '.mod')).write_bytes((base / 'go.mod').read_bytes())
    (dest / (VERSION + '.info')).write_text(json.dumps({'Version': VERSION, 'Time': '2026-11-10T00:00:00Z'}))
    (dest / 'list').write_text(VERSION + '\n')
    with zipfile.ZipFile(dest / (VERSION + '.zip'), 'w', zipfile.ZIP_DEFLATED) as archive:
        for relative in files:
            p = ROOT / relative
            if not p.is_file() or not p.is_relative_to(base):
                continue
            rel = p.relative_to(base)
            if rel.parts[0].startswith('.') or any((part in ('vendor', 'testdata-private') for part in rel.parts)):
                continue
            if any(((base / pathlib.Path(*rel.parts[:i]) / 'go.mod').exists() for i in range(1, len(rel.parts)))):
                continue
            target = module + '@' + VERSION + '/' + rel.as_posix()
            if rel.as_posix() == 'go.sum':
                # Checkout archives differ from immutable release tags; retain external
                # checksums but let the isolated proxy authenticate synthetic siblings.
                content = ''.join(line for line in p.read_text().splitlines(keepends=True)
                                  if not (line.split() and line.split()[0].split('@')[0].startswith(PREFIX)))
                archive.writestr(target, content)
            else:
                archive.write(p, target)
env = {**os.environ, 'GOWORK': 'off', 'GOPROXY': proxy.as_uri() + ',https://proxy.golang.org', 'GOMODCACHE': str(cache), 'GONOSUMDB': PREFIX + '*', 'GONOPROXY': 'none', 'GOFLAGS': '-mod=mod'}
for name, _ in MODULES:
    module = PREFIX + ('/' + name if name else '')
    print('ARCHIVE ' + module, flush=True)
    download = subprocess.run(['go', 'mod', 'download', '-json', module + '@' + VERSION], cwd=work, env=env, text=True, capture_output=True)
    if download.returncode:
        raise SystemExit(download.stdout + download.stderr)
    output = download.stdout
    record = json.loads(output)
    (work / ((name or 'sdk') + '-download.json')).write_text(output)
    consumer = pathlib.Path(tempfile.mkdtemp(prefix=(name or 'sdk') + '-consumer-', dir=work))
    shutil.copytree(record['Dir'], consumer, dirs_exist_ok=True)
    for entry in [consumer, *consumer.rglob('*')]:
        entry.chmod(493 if entry.is_dir() else 420)
    for command in [['go', 'test', './...'], ['go', 'vet', './...'], ['go', 'build', './...']]:
        subprocess.run(command, cwd=consumer, env=env, check=True)
print('PASS six independently extracted module archives: ' + str(work), flush=True)
