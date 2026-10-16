import {test,expect,type Page} from '@playwright/test';
import {decodeSnapshot,type Snapshot} from '../../packages/protocol/game';

function observe(page:Page){let current:Snapshot|null=null;page.on('websocket',ws=>ws.on('framereceived',event=>{const p=event.payload;if(typeof p==='string'||p[2]!==2)return;const buffer=p.buffer.slice(p.byteOffset,p.byteOffset+p.byteLength) as ArrayBuffer;current=decodeSnapshot(buffer)}));return()=>current}

test('dois convidados, renderização, movimento, split e reconexão',async({page,browser},info)=>{
 const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message));const a=observe(page);
 const room=`qa-${Date.now().toString(36)}-${info.project.name}`;
 await page.goto(`/?room=${room}`);await page.getByLabel('Como você quer aparecer?').fill('Célula A');await page.getByRole('button',{name:'Entrar na arena'}).click();
 await expect(page.getByRole('main',{name:'Partida Cells'})).toBeVisible();await expect.poll(()=>a()?.alive).toBe(true);await expect(page.getByRole('img',{name:/Arena Cells/})).toBeVisible();await expect.poll(async()=>{const text=await page.locator('.connection-hud').textContent();return Number(text?.match(/(\d+) FPS/)?.[1]??0)>0}).toBe(true);
 const other=await browser.newContext({viewport:{width:1280,height:800}});const bpage=await other.newPage(),b=observe(bpage);
 try {
  await bpage.goto(`/?room=${room}`);await bpage.getByLabel('Como você quer aparecer?').fill('Célula B');await bpage.getByRole('button',{name:'Entrar na arena'}).click();await expect.poll(()=>b()?.alive).toBe(true);
  await expect.poll(()=>a()?.leaderboard.some(p=>p.name==='Célula B')).toBe(true);expect(a()!.self).not.toBe(b()!.self);
  const start=a()!.x;
  const direction=start>a()!.size/2?'ArrowLeft':'ArrowRight';await page.keyboard.down(direction);await page.waitForTimeout(1000);await page.keyboard.up(direction);expect(Math.abs(a()!.x-start)).toBeGreaterThan(15);
  await page.getByRole('button',{name:'Dividir célula'}).click();await expect.poll(()=>a()?.entities.filter(e=>e.owner===a()?.self&&e.kind===2).length).toBeGreaterThan(1);
  // A second player remains independent while A acts.
  expect(b()!.alive).toBe(true);
  await page.screenshot({path:`artifacts/classic-${info.project.name}.png`});
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
  const oldTick=b()!.tick,oldID=b()!.self;await other.setOffline(true);await bpage.waitForTimeout(500);await other.setOffline(false);await expect.poll(()=>b()?.tick,{timeout:15000}).toBeGreaterThan(oldTick+30);expect(b()!.self).toBe(oldID);
  const identity=a()!.self;await page.reload();await expect(page.getByRole('main',{name:'Partida Cells'})).toBeVisible();await expect.poll(()=>a()?.self).toBe(identity);expect(errors).toEqual([]);
 } finally {await other.close()}
});

test('espectador pode entrar e erros gráficos permanecem visíveis',async({page},info)=>{
 await page.goto(`/?room=viewer-${Date.now().toString(36)}-${info.project.name}`);
 await page.getByRole('button',{name:'Só quero assistir'}).click();
 await expect(page.getByLabel('Fim da vida')).toBeVisible();await expect(page.getByText('Modo espectador',{exact:true})).toBeVisible();
 await expect(page.getByRole('button',{name:'Voltar à arena'})).toBeEnabled();await page.getByRole('button',{name:'Voltar à arena'}).click();await expect(page.getByLabel('Fim da vida')).not.toBeVisible();
 await page.getByRole('button',{name:'Sair da partida'}).click();await expect(page.getByRole('heading',{level:1})).toContainText('Comece pequeno');
 await page.addInitScript(()=>{const original=HTMLCanvasElement.prototype.getContext;HTMLCanvasElement.prototype.getContext=function(this:HTMLCanvasElement,...args:Parameters<typeof original>){if(String(args[0]).includes('webgl'))return null;return original.apply(this,args)} as typeof original});
 await page.reload();await page.getByLabel('Como você quer aparecer?').fill('Sem WebGL');await page.getByRole('button',{name:'Entrar na arena'}).click();
 await expect(page.getByRole('alert')).toContainText('Não foi possível iniciar WebGL');await page.waitForTimeout(1200);await expect(page.getByRole('alert')).toContainText('Não foi possível iniciar WebGL');
});
