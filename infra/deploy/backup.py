#!/usr/bin/env python3
"""Back up only Cells; optionally verify a restore in a disposable isolated DB."""
import json
import os
from pathlib import Path
import subprocess
import sys
import tarfile
import time

ROOT=Path('/var/backups/cells')

def run(*args,**kwargs):
    return subprocess.run(args,check=True,stderr=subprocess.PIPE,**kwargs)

def main():
    if os.geteuid()!=0:raise SystemExit('Run inside the dedicated Cells LXC as root.')
    if Path('/etc/hostname').read_text().strip()!='cells-prd-01':raise SystemExit('Wrong host.')
    os.umask(0o077)
    ROOT.mkdir(mode=0o700,parents=True,exist_ok=True)
    stamp=str(time.time_ns())
    dump=ROOT/f'{stamp}-scheduled.dump'
    try:
        with dump.open('xb') as output:
            run('runuser','-u','postgres','--','pg_dump','-Fc','cells',stdout=output)
        run('pg_restore','--list',str(dump),stdout=subprocess.DEVNULL)
    except Exception:
        dump.unlink(missing_ok=True)
        raise
    configs=ROOT/f'{stamp}-configs.tar.gz'
    with tarfile.open(configs,'w:gz') as bundle:
        bundle.add('/etc/cells',arcname='cells-config')
        for name in ['cells.service','cells-tunnel.service','cells-deploy.service','cells-deploy.timer','cells-backup.service','cells-backup.timer']:
            path=Path('/etc/systemd/system')/name
            if path.is_file():bundle.add(path,arcname=f'units/{name}')
    if '--verify' in sys.argv:
        clone='cells_verify_'+stamp
        run('runuser','-u','postgres','--','createdb','-O','cells',clone)
        try:
            with dump.open('rb') as source:
                run('runuser','-u','postgres','--','pg_restore','--exit-on-error','-d',clone,stdin=source)
            run('runuser','-u','postgres','--','psql','-v','ON_ERROR_STOP=1','-d',clone,'-c','SELECT count(*) FROM cells_schema',stdout=subprocess.DEVNULL)
        finally:
            run('runuser','-u','postgres','--','dropdb',clone)
    cutoff=time.time()-14*86400
    for path in ROOT.iterdir():
        if path.is_file() and (path.name.endswith('.dump') or path.name.endswith('-configs.tar.gz')) and path.stat().st_mtime<cutoff:
            path.unlink()
    print(json.dumps({'event':'cells_backup_verified' if '--verify' in sys.argv else 'cells_backup_created','file':dump.name,'bytes':dump.stat().st_size}))

if __name__=='__main__':
    try:main()
    except Exception as error:
        print(json.dumps({'event':'cells_backup_failed','error':type(error).__name__}),file=sys.stderr)
        raise SystemExit(1)
