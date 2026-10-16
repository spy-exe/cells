"""Check Go statement coverage for authoritative rules and wire validation."""
from pathlib import Path
import sys
profile=Path(sys.argv[1] if len(sys.argv)>1 else 'coverage-go.out')
blocks={}
for line in profile.read_text().splitlines()[1:]:
    block,count,hits=line.split()
    statements,covered=blocks.get(block,(int(count),False))
    blocks[block]=(statements,covered or int(hits)>0)
failed=False
for package in ['engine','protocol','bot','training','expedition']:
    selected=[v for k,v in blocks.items() if f'/internal/{package}/' in k]
    total=sum(v[0] for v in selected)
    pct=100*sum(n for n,covered in selected if covered)/total if total else 0
    print(f'{package}: {pct:.1f}% statement coverage (minimum 85%)')
    failed|=pct<85
sys.exit(1 if failed else 0)
