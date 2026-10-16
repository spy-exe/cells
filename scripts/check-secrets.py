"""Pattern gate for tracked source, complementing review and dependency audits."""
import re
import subprocess
from pathlib import Path

patterns = [re.compile(r'gh[pousr]_[A-Za-z0-9]{30,}'), re.compile(r'-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----'), re.compile(r'AKIA[A-Z0-9]{16}')]
blocked = []
for name in subprocess.check_output(['git', 'ls-files'], text=True).splitlines():
    path = Path(name)
    if path.name in {'.env', 'tunnel.json', 'cert.pem'} or '.runtime' in path.parts:
        blocked.append(name)
    if path.is_file() and path.stat().st_size < 2_000_000:
        data = path.read_text(errors='replace')
        if any(pattern.search(data) for pattern in patterns):
            blocked.append(name)
if blocked:
    raise SystemExit('Sensitive patterns in: ' + ', '.join(sorted(set(blocked))))
print('Tracked source: no configured secret patterns found.')
