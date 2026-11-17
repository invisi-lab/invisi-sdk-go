#!/usr/bin/env python3
"""Render complete release notes from versioned local changelogs."""
import argparse
import json
from pathlib import Path
import re


def extract_version(markdown, version):
    headings = list(re.finditer(r'^##[ \t]+([^\n]+)$', markdown, re.MULTILINE))
    matches = [i for i, heading in enumerate(headings) if heading.group(1).strip() == version]
    if len(matches) != 1:
        raise ValueError(f'Expected exactly one CHANGELOG entry for {version}')
    index = matches[0]
    end = headings[index + 1].start() if index + 1 < len(headings) else len(markdown)
    content = markdown[headings[index].end():end].strip()
    definitions = re.compile(r'^\s*\[([^\]]+)\]:\s*\S+.*$', re.MULTILINE)
    labels = {match.group(1).casefold() for match in definitions.finditer(content)}
    substantive = definitions.sub('', content)
    substantive = re.sub(r'<a\b[^>]*>.*?</a>', '', substantive, flags=re.IGNORECASE | re.DOTALL)
    substantive = re.sub(r'\[[^\]]*\]\([^)]*\)', '', substantive)
    substantive = re.sub(r'\[[^\]]+\]\s*\[[^\]]*\]', '', substantive)
    substantive = re.sub(r'\[([^\]]+)\]', lambda match: '' if match.group(1).casefold() in labels else match.group(), substantive)
    substantive = re.sub(r'https?://\S+', '', substantive)
    if not re.search(r'[A-Za-z\u4e00-\u9fff]', substantive):
        raise ValueError(f'CHANGELOG entry for {version} must contain release details, not only links')
    return content


def release_notes(root, tag):
    if not re.fullmatch(r'v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)', tag):
        raise ValueError('Release tag must be canonical vMAJOR.MINOR.PATCH')
    root = Path(root)
    version = tag[1:]
    if (root / 'CHANGELOG.md').is_file():
        sections = [extract_version((root / 'CHANGELOG.md').read_text(), version)]
    else:
        sections = []
        for manifest in sorted((root / 'packages').glob('*/package.json')):
            package = json.loads(manifest.read_text())
            if package.get('private'):
                continue
            if package['version'] != version:
                raise ValueError(f"{package['name']}: version differs from release tag")
            content = extract_version((manifest.parent / 'CHANGELOG.md').read_text(), version)
            sections.append(f"## {package['name']}\n\n{content}")
        if not sections:
            raise ValueError('No public package changelogs found')
    return f'# SDK {tag}\n\n' + '\n\n'.join(sections) + '\n'


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('tag')
    parser.add_argument('--output', type=Path)
    args = parser.parse_args()
    rendered = release_notes(Path(__file__).resolve().parents[1], args.tag)
    if args.output:
        args.output.write_text(rendered)
    else:
        print(rendered, end='')


if __name__ == '__main__':
    main()
