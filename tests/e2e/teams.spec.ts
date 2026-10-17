import {test,expect,type Page} from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";
import {decodeSnapshot,type Snapshot} from "../../packages/protocol/game";
function observe(page:Page){let state:Snapshot|null=null,team:number|undefined;page.on('websocket',ws=>ws.on('framereceived',({payload})=>{if(typeof payload==='string'||payload[2]!==2)return;state=decodeSnapshot(payload.buffer.slice(payload.byteOffset,payload.byteOffset+payload.byteLength) as ArrayBuffer);team=state.entities.find(e=>e.kind===2&&e.owner===state!.self)?.team??team;}));return ()=>({state,team});}
test("Teams real tem duas sessões, cores e disputa de massa compartilhada",async({page,browser},info)=>{
 const a=observe(page),room=`teams-${Date.now().toString(36)}-${info.project.name}`;const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message));
 await page.goto(`/?room=${room}`);await expect(page.locator('#mode option[value="teams"]')).toHaveCount(1);
 await page.getByLabel('Modo',{exact:true}).selectOption('teams');await page.getByLabel('Como você quer aparecer?').fill('Musgo QA');await page.getByRole('button',{name:'Entrar na arena'}).click();
 await expect(page.getByLabel('Disputa entre equipes')).toBeVisible();await expect.poll(()=>a().state?.teams).toBe(true);await expect.poll(()=>a().team??-1).toBeGreaterThanOrEqual(0);
 const context=await browser.newContext({viewport:{width:800,height:600},deviceScaleFactor:1}),other=await context.newPage(),b=observe(other);
 try{
  await other.goto(`/?room=${room}`);await other.getByLabel('Modo',{exact:true}).selectOption('teams');await other.getByLabel('Como você quer aparecer?').fill('Coral QA');await other.getByRole('button',{name:'Entrar na arena'}).click();
  await expect(other.getByLabel('Disputa entre equipes')).toBeVisible();await expect.poll(()=>b().state?.teams).toBe(true);
  expect(a().state!.self).not.toBe(b().state!.self);expect(a().state!.teamMass.every(m=>m>0)).toBe(true);expect(b().state!.teamMass.every(m=>m>0)).toBe(true);
  await expect(page.locator('.team-standings li')).toHaveCount(3);await expect(page.locator('.team-standings .is-you')).toHaveCount(1);
  await page.screenshot({path:`artifacts/${process.env.E2E_BASE_URL?.startsWith('https://cells.malha.app')?'public':'qa'}-teams-${info.project.name}.png`});expect(errors).toEqual([]);
 }finally{await context.close();}
});
test("cenário visual: placar de equipes é proporcional, legível e acessível",async({page},info)=>{
 await page.route('**/api/status',r=>r.fulfill({json:{ok:true,ready:true,version:'fixture',revision:'fixture',modes:['classic','arena','teams','rogue'],players:1,bots:30}}));
 await page.route('**/api/guest',r=>r.fulfill({status:201,json:{nickname:'Time QA',mode:'teams',room:''}}));
 let eliminate=()=>{};
 await page.routeWebSocket('**/ws',ws=>{
  const h=Buffer.alloc(53);h[0]=206;h[1]=1;h[2]=2;h.writeUInt32LE(60,3);h.writeUInt32LE(1,7);h[11]=3;
  for(const [offset,value] of [[12,6000],[16,3000],[20,3000],[24,1500],[28,36],[38,100],[42,200],[46,300]])h.writeFloatLE(value,offset);h.writeUInt16LE(2,50);
  const e=Buffer.alloc(24);e.writeUInt32LE(1,0);e.writeUInt32LE(1,4);e.writeFloatLE(3000,8);e.writeFloatLE(3000,12);e.writeFloatLE(36,16);e[20]=2;e[21]=1;const pellet=Buffer.from(e);pellet.writeUInt32LE(2,0);pellet.writeUInt32LE(88,4);pellet.writeFloatLE(3100,8);pellet.writeFloatLE(12,16);pellet[20]=3;pellet[21]=2;ws.send(Buffer.concat([h,e,pellet]));
  eliminate=()=>{h[11]=2;h.writeUInt32LE(63,3);h.writeFloatLE(0,28);h.writeFloatLE(164,42);h.writeUInt16LE(0,50);ws.send(h);};
 });
 await page.goto('/');await page.getByLabel('Modo',{exact:true}).selectOption('teams');await page.getByLabel('Como você quer aparecer?').fill('Time QA');await page.getByRole('button',{name:'Entrar na arena'}).click();
 await expect(page.getByRole('img',{name:/Arena Cells/})).toBeVisible();
 const captureColor=async()=>{const shot=await page.screenshot(),viewport=page.viewportSize()!;
 return page.evaluate(async({bytes,viewport})=>{
  const bitmap=await createImageBitmap(new Blob([new Uint8Array(bytes)],{type:'image/png'})),canvas=new OffscreenCanvas(bitmap.width,bitmap.height),ctx=canvas.getContext('2d')!;ctx.drawImage(bitmap,0,0);
  const ratio=bitmap.width/viewport.width,scale=Math.max(viewport.width/1.9,viewport.height/1.4)/1500;
  return Array.from(ctx.getImageData(Math.round((viewport.width/2+100*scale)*ratio),Math.round(viewport.height/2*ratio),1,1).data).slice(0,3);
 },{bytes:Array.from(shot),viewport});};await expect.poll(captureColor).toEqual([136,182,215]);
 await expect(page.getByLabel('Disputa entre equipes')).toBeVisible();await expect(page.locator('.team-standings .is-you')).toContainText('Coral');await expect(page.getByRole('img',{name:'Musgo: 17%, Coral: 33%, Maré: 50%'})).toBeVisible();
 await expect.poll(async()=>{const widths=await page.locator('.team-shares i').evaluateAll(items=>items.map(e=>e.getBoundingClientRect().width));return widths[2]/widths[0];}).toBeCloseTo(3,1);
 expect((await new AxeBuilder({page}).analyze()).violations).toEqual([]);expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
 await page.screenshot({path:`artifacts/fixture-teams-${info.project.name}.png`});
 eliminate();await expect(page.getByTestId("mass")).toHaveText("0");await expect(page.locator(".team-standings .is-you")).toContainText("Coral");
});
