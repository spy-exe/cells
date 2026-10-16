"""Deployment boundary tests use synthetic archives, never production services."""
import hashlib
import importlib.util
import io
import json
import os
from pathlib import Path
import tarfile
import tempfile
import unittest
from unittest.mock import patch

spec=importlib.util.spec_from_file_location('deploy',Path(__file__).parents[2]/'infra/deploy/pull-release.py')
deploy=importlib.util.module_from_spec(spec)
spec.loader.exec_module(deploy)
SHA='a'*40

def archive(name='cells',link=False,revision=SHA):
    data=io.BytesIO()
    with tarfile.open(fileobj=data,mode='w:gz') as bundle:
        item=tarfile.TarInfo(name)
        if link:
            item.type=tarfile.SYMTYPE;item.linkname='/etc/passwd';bundle.addfile(item)
        else:
            item.size=4;bundle.addfile(item,io.BytesIO(b'\x7fELF'))
        manifest=json.dumps({'revision':revision,'version':'0.0.1'}).encode()
        item=tarfile.TarInfo('manifest.json');item.size=len(manifest);bundle.addfile(item,io.BytesIO(manifest))
    return data.getvalue()

class DeployBoundary(unittest.TestCase):
    def exercise(self,payload):
        def get(url,binary=False):
            if '/actions/' in url:return {'workflow_runs':[{'head_sha':SHA,'head_branch':'main','event':'push','conclusion':'success'}]}
            if '/releases/tags/' in url:return {'tag_name':'build-'+SHA,'draft':False}
            if url.endswith('SHA256SUMS'):return (hashlib.sha256(payload).hexdigest()+'  cells-linux-amd64.tar.gz').encode()
            return payload
        with tempfile.TemporaryDirectory() as scratch:
            root=Path(scratch)
            with patch.object(deploy,'ROOT',root),patch.object(deploy,'STATE',root/'state'),patch.object(deploy,'get',get),patch.object(deploy,'backup') as backup,patch.object(deploy,'run'),patch.object(deploy,'healthy',return_value=False),patch.object(deploy.time,'sleep'):
                previous_mask=os.umask(0o077)
                try:
                    with self.assertRaises((ValueError,RuntimeError)):
                        deploy.deploy()
                finally:
                    os.umask(previous_mask)
                destination=root/'releases'/SHA
                if destination.exists():
                    self.assertEqual(destination.stat().st_mode & 0o777,0o755)
                self.assertFalse((root/'current').exists())
                self.assertFalse((root.parent/'escape').exists())
                return backup.call_count
    def test_rejects_path_escape_before_backup(self):
        self.assertEqual(self.exercise(archive('../../escape')),0)
    def test_rejects_symlinks_before_backup(self):
        self.assertEqual(self.exercise(archive('web/payload',link=True)),0)
    def test_rejects_manifest_from_another_commit(self):
        self.assertEqual(self.exercise(archive(revision='b'*40)),0)
    def test_unhealthy_candidate_does_not_activate(self):
        self.assertEqual(self.exercise(archive()),1)

if __name__=='__main__':unittest.main()
