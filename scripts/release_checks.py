"""Validate tagged Go module identity and seed optional dependency caches."""
import json
import re
import subprocess

PREFIX = 'github.com/invisi-labs/invisi-sdk-go'
MODULE_NAMES = ['primitives', 'protocol', 'prover', 'rpc', '', 'session', 'ipp', 'paymentrouter']


def cache_links(src, dst):
    dst.mkdir(parents=True, exist_ok=True)
    if not src.is_dir():
        return
    for path in src.iterdir():
        if path.name == 'cache' or path.name.startswith('invisi-sdk-go'):
            continue
        target = dst / path.name
        if '@' in path.name:
            target.symlink_to(path, target_is_directory=path.is_dir())
        elif path.is_dir():
            cache_links(path, target)


def validate_manifests(manifests, version):
    expected = {PREFIX + ('/' + name if name else '') for name in MODULE_NAMES}
    if set(manifests) != expected:
        raise ValueError('Release must include all eight Go modules')
    for name, manifest in manifests.items():
        if manifest['Module']['Path'] != name:
            raise ValueError(f'Module identity mismatch: {name}')
        for dependency in manifest.get('Require') or []:
            path = dependency['Path']
            if path == PREFIX or path.startswith(PREFIX + '/'):
                if path not in expected or dependency['Version'] != version:
                    raise ValueError(f'Internal dependency must match {version}: {name}/{path}')
        if manifest.get('Replace'):
            raise ValueError(f'Release modules must not depend on replacement directives: {name}')


def validate_tag_refs(output, version, commit):
    refs = {}
    for line in output.splitlines():
        sha, ref = line.split()
        refs[ref] = sha
    for name in MODULE_NAMES:
        tag = (name + '/' if name else '') + version
        ref = 'refs/tags/' + tag
        actual = refs.get(ref + '^{}', refs.get(ref))
        if actual != commit:
            raise ValueError(f'Missing tag or wrong commit: {tag}; expected {commit}, got {actual}')


def validate_tagged_release(root, version):
    if not re.fullmatch(r'v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)', version):
        raise ValueError('Release tag must be canonical stable vMAJOR.MINOR.PATCH')
    manifests = {}
    for name in MODULE_NAMES:
        directory = root / name if name else root
        content = subprocess.check_output(['go', 'mod', 'edit', '-json'], cwd=directory, text=True)
        manifests[PREFIX + ('/' + name if name else '')] = json.loads(content)
    validate_manifests(manifests, version)
    commit = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip()
    patterns = []
    for name in MODULE_NAMES:
        ref = 'refs/tags/' + (name + '/' if name else '') + version
        patterns.extend([ref, ref + '^{}'])
    output = subprocess.check_output(['git', 'ls-remote', '--tags', 'origin', *patterns], cwd=root, text=True)
    validate_tag_refs(output, version, commit)
    print(f'All eight remote module tags match {version} at {commit}', flush=True)
