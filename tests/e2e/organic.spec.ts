import { test, expect } from "@playwright/test";
function frame(cells: {id:number;owner:number;mass:number;name?:string;x?:number;y?:number}[],tick=60,mass=36) {
 const header=Buffer.alloc(53);header[0]=206;header[1]=1;header[2]=2;
 header.writeUInt32LE(tick,3);header.writeUInt32LE(1,7);header[11]=1;
 for(const [offset,value] of [[12,6000],[16,3000],[20,3000],[24,1500],[28,mass]]) header.writeFloatLE(value,offset);
 header.writeUInt16LE(cells.length,50);
 return Buffer.concat([header,...cells.map(c=>{
  const name=Buffer.from(c.name??'');const b=Buffer.alloc(24);
  b.writeUInt32LE(c.id,0);b.writeUInt32LE(c.owner,4);b.writeFloatLE(c.x??3000,8);b.writeFloatLE(c.y??3000,12);b.writeFloatLE(c.mass,16);b[20]=2;b[23]=name.length;
  return Buffer.concat([b,name]);
 })]);
}
for(const kind of ['depth','caption'] as const) test(`renderização real: ${kind==='depth'?'célula maior cobre a menor':'nome pequeno permanece no centro'}`,async({page},info)=>{
 const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message));
 await page.routeWebSocket('**/ws',ws=>ws.send(frame(kind==='depth'?[{id:2,owner:8,mass:225},{id:1,owner:1,mass:36}]:[{id:1,owner:1,mass:36,name:'I'}])));
 await page.goto(`/?room=organic-${Date.now().toString(36)}`);
 await page.getByLabel('Como você quer aparecer?').fill('Organic QA');
 await page.getByRole('button',{name:'Entrar na arena'}).click();
 await expect(page.getByTestId('mass')).toHaveText('36');
 await expect.poll(async()=>Number((await page.locator('.connection-hud').textContent())?.match(/(\d+) FPS/)?.[1]??0)).toBeGreaterThan(0);
 const png=await page.screenshot({path:`artifacts/fixture-${kind}-${info.project.name}.png`});
 const pixels=await page.evaluate(async bytes=>{
  const bitmap=await createImageBitmap(new Blob([new Uint8Array(bytes)],{type:'image/png'}));
  const canvas=new OffscreenCanvas(bitmap.width,bitmap.height),ctx=canvas.getContext('2d')!;ctx.drawImage(bitmap,0,0);
  const x=Math.floor(bitmap.width/2),y=Math.floor(bitmap.height/2),dpr=devicePixelRatio;
  const pixel=Array.from(ctx.getImageData(x,y,1,1).data);
  const span=Math.ceil(5*dpr),region=ctx.getImageData(x-span,y-span,span*2,span*2).data;
  // Count antialiased caption pixels, distinct from the orange body, at any DPR.
  let ink=0;for(let i=0;i<region.length;i+=4)if(region[i]<140&&region[i+1]<125&&region[i+2]<110)ink++;
  return {pixel,ink};
 },Array.from(png));
 if(kind==='depth') expect(pixels.pixel.slice(0,3)).toEqual([144,188,75]);
 else expect(pixels.ink).toBeGreaterThan(2);
 expect(errors).toEqual([]);
});

test("recombinação mantém o nome no corpo e sob a célula maior durante a absorção",async({page},info)=>{
 const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message));
 await page.clock.install({time:new Date('2026-10-16T22:00:00Z')});
 let publish:((merged:boolean)=>void)|undefined;
 await page.routeWebSocket('**/ws',ws=>{
  publish=merged=>ws.send(frame(merged?[{id:1,owner:1,mass:72,name:'A'}]:[{id:1,owner:1,mass:36,name:'A'},{id:2,owner:1,mass:36,name:'B',x:3048}],merged?62:60,72));
  publish(false);
 });
 await page.goto(`/?room=merge-${Date.now().toString(36)}`);
 await page.getByLabel('Como você quer aparecer?').fill('Merge QA');await page.getByRole('button',{name:'Entrar na arena'}).click();
 await expect(page.getByTestId('mass')).toHaveText('72');
 await expect.poll(async()=>Number((await page.locator('.connection-hud').textContent())?.match(/(\d+) FPS/)?.[1]??0)).toBeGreaterThan(0);
 await page.clock.pauseAt(new Date('2026-10-16T22:01:00Z'));
 await page.clock.runFor(1200);
 const inspect=async()=>{
  const png=await page.screenshot();const viewport=page.viewportSize()!;
  return page.evaluate(async({bytes,viewport})=>{
   const bitmap=await createImageBitmap(new Blob([new Uint8Array(bytes)],{type:'image/png'}));const canvas=new OffscreenCanvas(bitmap.width,bitmap.height),ctx=canvas.getContext('2d')!;ctx.drawImage(bitmap,0,0);
   const ratio=bitmap.width/viewport.width,scale=Math.max(viewport.width/1.9,viewport.height/1.4)/1500;
   const x=Math.round((viewport.width/2+48*scale)*ratio),y=Math.round(viewport.height/2*ratio),span=Math.ceil(4*ratio);
   const region=ctx.getImageData(x-span,y-span,span*2,span*2).data;let ink=0;
   for(let i=0;i<region.length;i+=4)if(region[i]<160&&region[i+1]<160&&region[i+2]<160)ink++;
   return ink;
  },{bytes:Array.from(png),viewport});
 };
 expect(await inspect()).toBeGreaterThan(2);
 publish!(true);await page.clock.runFor(48);
 expect(await inspect()).toBe(0);
 await page.screenshot({path:`artifacts/fixture-merge-${info.project.name}.png`});
 await page.clock.runFor(400);
 expect(await inspect()).toBe(0);expect(errors).toEqual([]);
});
