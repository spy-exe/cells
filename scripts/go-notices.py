"""Preserve upstream notices for modules linked into the distributed Go binary."""
from pathlib import Path
import subprocess

lines = subprocess.check_output(['go', 'list', '-m', '-f', '{{.Path}}|{{.Dir}}', 'all'], text=True).splitlines()
notices = []
for line in lines:
    name, _, directory = line.partition('|')
    if not directory or name == 'github.com/spy-exe/cells':
        continue
    root = Path(directory)
    for file in sorted(root.iterdir()):
        if file.is_file() and file.name.lower().split('.')[0] in {'license', 'copying', 'notice'}:
            notices.append(name + '\n' + file.read_text(errors='replace'))
Path('release/THIRD_PARTY_NOTICES.txt').write_text('\n\n'.join(notices))
