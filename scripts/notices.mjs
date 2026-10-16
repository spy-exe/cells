import {execFileSync} from 'node:child_process';
import {readFileSync,readdirSync,existsSync,writeFileSync} from 'node:fs';
import {join} from 'node:path';
const packages=execFileSync('npm',['ls','--parseable','--omit=dev','--all'],{encoding:'utf8'}).trim().split('\n').slice(1);
const notices=['Cells — third-party license notices. Cells source is AGPL-3.0-only.'];
for(const directory of packages){
 const file=join(directory,'package.json');if(!existsSync(file))continue;
 const pkg=JSON.parse(readFileSync(file,'utf8'));
 const files=readdirSync(directory).filter(name=>/^(license|copying|notice)(\.|$)/i.test(name));
 for(const name of files)notices.push(`\n${pkg.name} ${pkg.version}\n${readFileSync(join(directory,name),'utf8')}`);
}
writeFileSync('apps/web/public/licenses.txt',notices.join('\n\n'));
