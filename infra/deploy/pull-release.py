#!/usr/bin/env python3
"""Deploy only a successful main/push build from this public repository.

Install this reviewed file as root; application releases never replace the helper.
"""
import fcntl
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tarfile
import tempfile
import time
import urllib.error
import urllib.request

REPO = 'spy-exe/cells'
ROOT = Path('/opt/cells')
STATE = Path('/var/lib/cells-deploy')
SHA = re.compile(r'^[a-f0-9]{40}$')

def event(kind, **fields):
    print(json.dumps({'event': kind, **fields}), flush=True)

def get(url, binary=False):
    request = urllib.request.Request(url, headers={'User-Agent': 'cells-deployer/1', 'Accept': 'application/vnd.github+json'})
    with urllib.request.urlopen(request, timeout=40) as response:
        data = response.read(100 * 1024 * 1024 + 1)
    if len(data) > 100 * 1024 * 1024:
        raise ValueError('Download exceeds deployment budget')
    return data if binary else json.loads(data)

def run(*args, **kwargs):
    return subprocess.run(args, check=True, capture_output=True, **kwargs)

def current_sha():
    try:
        value = json.loads((ROOT / 'current/manifest.json').read_text())['revision']
        return value if SHA.fullmatch(value) else None
    except (OSError, ValueError, KeyError):
        return None

def backup(sha):
    folder = Path('/var/backups/cells')
    folder.mkdir(mode=0o700, parents=True, exist_ok=True)
    target = folder / f'{time.time_ns()}-{sha[:7]}.dump'
    with target.open('xb') as output:
        os.chmod(target, 0o600)
        subprocess.run(['runuser', '-u', 'postgres', '--', 'pg_dump', '-Fc', 'cells'], stdout=output, check=True, stderr=subprocess.PIPE)
    run('pg_restore', '--list', str(target))
    event('backup_created', file=target.name)

def healthy(url, sha):
    try:
        result = get(url)
        return result.get('ok') is True and result.get('revision') == sha
    except (OSError, ValueError, urllib.error.URLError):
        return False

def switch(sha):
    if not SHA.fullmatch(sha):
        raise ValueError('Invalid revision')
    link = ROOT / '.next'
    if link.is_symlink():
        link.unlink()
    link.symlink_to(ROOT / 'releases' / sha)
    link.replace(ROOT / 'current')
    run('systemctl', 'restart', 'cells')
    for _ in range(20):
        if healthy('http://127.0.0.1:8080/readyz', sha):
            return
        time.sleep(.5)
    raise RuntimeError('Activated service failed readiness')

def deploy():
    if (STATE / 'hold').exists():
        event('deployment_held')
        return
    runs = get(f'https://api.github.com/repos/{REPO}/actions/workflows/ci.yml/runs?branch=main&event=push&status=success&per_page=1')['workflow_runs']
    if not runs:
        event('no_verified_build')
        return
    workflow = runs[0]
    sha = workflow['head_sha']
    if not SHA.fullmatch(sha) or workflow['head_branch'] != 'main' or workflow['event'] != 'push' or workflow['conclusion'] != 'success':
        raise ValueError('Build does not satisfy deployment policy')
    previous = current_sha()
    if previous == sha:
        event('already_current', revision=sha)
        return
    release = get(f'https://api.github.com/repos/{REPO}/releases/tags/build-{sha}')
    if release.get('draft') or release['tag_name'] != f'build-{sha}':
        raise ValueError('Release metadata rejected')
    prefix = f'https://github.com/{REPO}/releases/download/build-{sha}/'
    checksum = get(prefix + 'SHA256SUMS', True).decode().split()
    if len(checksum) != 2 or not re.fullmatch(r'[a-f0-9]{64}', checksum[0]) or checksum[1] != 'cells-linux-amd64.tar.gz':
        raise ValueError('Checksum manifest rejected')
    payload = get(prefix + 'cells-linux-amd64.tar.gz', True)
    if hashlib.sha256(payload).hexdigest() != checksum[0]:
        raise ValueError('Artifact checksum mismatch')
    with tempfile.TemporaryDirectory(prefix='cells-release-') as scratch:
        archive = Path(scratch) / 'release.tar.gz'
        archive.write_bytes(payload)
        unpack = Path(scratch) / 'unpack'
        unpack.mkdir()
        with tarfile.open(archive) as bundle:
            members = bundle.getmembers()
            total = 0
            if len(members) > 10000:
                raise ValueError('Too many archive entries')
            for member in members:
                name = Path(member.name)
                if name.is_absolute() or '..' in name.parts or not (member.isfile() or member.isdir()):
                    raise ValueError('Unsafe archive entry')
                allowed = str(name) in {'.', 'cells', 'manifest.json', 'LICENSE', 'THIRD_PARTY_NOTICES.txt'} or (name.parts and name.parts[0] == 'web')
                total += member.size
                if not allowed or total > 100 * 1024 * 1024 or member.size > 50 * 1024 * 1024:
                    raise ValueError('Archive outside release contract')
            bundle.extractall(unpack, filter='data')
        manifest = json.loads((unpack / 'manifest.json').read_text())
        if manifest.get('revision') != sha or (unpack / 'cells').read_bytes()[:4] != b'\x7fELF':
            raise ValueError('Release identity or executable rejected')
        destination = ROOT / 'releases' / sha
        if not destination.exists():
            shutil.copytree(unpack, destination)
        for path in destination.rglob('*'):
            os.chmod(path, 0o755 if path.is_dir() or path.name == 'cells' else 0o644)
    backup(sha)
    unit = f'cells-candidate-{sha[:7]}'
    try:
        run('systemd-run', '--unit=' + unit, '--property=User=cells', '--property=CPUQuota=50%', '--property=MemoryMax=512M', '--setenv=CELLS_ADDR=127.0.0.1:8081', '--setenv=CELLS_WEB_DIR=' + str(destination / 'web'), str(destination / 'cells'))
        for _ in range(20):
            if healthy('http://127.0.0.1:8081/readyz', sha):
                break
            time.sleep(.5)
        else:
            raise RuntimeError('Candidate failed readiness')
    finally:
        subprocess.run(['systemctl', 'stop', unit], check=False, capture_output=True)
    try:
        switch(sha)
        for _ in range(12):
            if healthy('https://cells.malha.app/readyz', sha):
                break
            time.sleep(1)
        else:
            raise RuntimeError('Public TLS readiness failed')
    except Exception:
        if previous:
            switch(previous)
            event('rolled_back', revision=previous)
        else:
            subprocess.run(['systemctl', 'stop', 'cells'], check=False, capture_output=True)
        raise
    if previous:
        (STATE / 'previous').write_text(previous)
    event('published', revision=sha, version=manifest['version'])

def main():
    if os.geteuid() != 0:
        raise SystemExit('Install and run the reviewed helper as root inside the dedicated Cells CT.')
    STATE.mkdir(mode=0o700, parents=True, exist_ok=True)
    with (STATE / 'deploy.lock').open('w') as lock:
        fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        if '--rollback' in sys.argv:
            (STATE / 'hold').write_text('Manual rollback; resume explicitly.')
            sha = (STATE / 'previous').read_text().strip()
            switch(sha)
            event('manual_rollback', revision=sha)
        else:
            if '--resume' in sys.argv:
                (STATE / 'hold').unlink(missing_ok=True)
            deploy()

if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        event('deployment_failed', error=type(error).__name__, detail=str(error)[:200])
        raise SystemExit(1)
