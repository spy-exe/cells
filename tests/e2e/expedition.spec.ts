import {test,expect} from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";
import {readFileSync} from "node:fs";
import type {ExpeditionRun} from "../../apps/web/src/Expedition";
const offers=[{id:"cilia",name:"Cílios",detail:"+10% de velocidade de movimento."},{id:"digestion",name:"Metabolismo",detail:"+20% de massa recebida dos alimentos."},{id:"cohesion",name:"Coesão",detail:"Recombinação 20% mais cedo após dividir."}];
test("expedição real é solo, move e retoma a mesma semente",async({page},info)=>{
 const errors:string[]=[];page.on("pageerror",e=>errors.push(e.message));
 await page.goto("/?room=ignored-in-solo");
 await page.getByLabel("Modo",{exact:true}).selectOption("rogue");
 await expect(page.getByText("Jogar com amigos",{exact:true})).not.toBeVisible();
 await expect(page.getByRole("button",{name:"Só quero assistir"})).not.toBeVisible();
 await page.getByLabel("Como você quer aparecer?").fill("Explorador");
 await page.getByRole("button",{name:"Entrar na arena"}).click();
 await expect(page.getByLabel("Objetivo da expedição")).toContainText("Setor 1 de 3");
 await expect(page.getByRole("img",{name:/Arena Cells/})).toBeVisible();
 const first=await (await page.request.get("/api/expedition")).json();
 const position=await page.getByTestId("position").textContent();
 await page.keyboard.down("ArrowRight");await page.waitForTimeout(800);await page.keyboard.up("ArrowRight");
 await expect(page.getByTestId("position")).not.toHaveText(position!);
 await page.reload();
 await expect(page.getByLabel("Objetivo da expedição")).toBeVisible();
 const resumed=await (await page.request.get("/api/expedition")).json();
 expect(resumed.seed_label).toBe(first.seed_label);expect(resumed.revision).toBe(first.revision);
 expect(resumed.seconds_left).toBeLessThanOrEqual(first.seconds_left);
 await page.screenshot({path:`artifacts/${process.env.E2E_BASE_URL?.startsWith("https://cells.malha.app")?"public":"qa"}-expedition-${info.project.name}.png`});
 expect(errors).toEqual([]);
});
test("cenário visual: mutações, pausa, resultado e nova expedição",async({page},info)=>{
 const errors:string[]=[];page.on("pageerror",e=>errors.push(e.message));
 let state:ExpeditionRun={stage:1,phase:"choice",revision:2,seconds_left:0,target_mass:60,mass:90,peak:95,kills:1,seed_label:"ABCDEF",ranks:{cilia:0,digestion:0,cohesion:0},offers};
 let controls=0,commands=0;const picks:unknown[]=[];
 await page.route("**/api/expedition",r=>r.fulfill({json:state}));
 await page.route("**/api/expedition/mutation",async r=>{
  picks.push(r.request().postDataJSON());state={...state,stage:2,phase:"running",revision:3,seconds_left:25,target_mass:100,offers:[],ranks:{cilia:0,digestion:1,cohesion:0}};
  await r.fulfill({status:204});
 });
 await page.routeWebSocket("**/ws",ws=>{
  ws.send(Buffer.from(readFileSync("packages/protocol/fixtures/snapshot-v1.hex","utf8").trim(),"hex"));
  ws.onMessage(raw=>{const b=Buffer.from(raw);if(b[2]===1)commands++;if(b[2]===3){controls++;state={...state,stage:1,phase:"running",seconds_left:25,ranks:{},seed_label:"NEW123",revision:1};}});
 });
 await page.goto("/");await page.getByLabel("Modo",{exact:true}).selectOption("rogue");
 await page.getByLabel("Como você quer aparecer?").fill("Mutação QA");await page.getByRole("button",{name:"Entrar na arena"}).click();
 await expect(page.getByLabel("Escolher mutação")).toBeVisible();
 await page.keyboard.press(" ");expect(commands).toBe(0);
 const original=page.viewportSize()!;
 for(const [width,height] of [[360,800],[390,844],[412,915],[768,1024],[1440,900]]){
  await page.setViewportSize({width,height});
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
  const box=await page.getByLabel("Escolher mutação").boundingBox();
  expect(box!.y).toBeGreaterThanOrEqual(0);expect(box!.y+box!.height).toBeLessThanOrEqual(height);
 }
 await page.setViewportSize(original);
 expect((await new AxeBuilder({page}).analyze()).violations).toEqual([]);
 await page.screenshot({path:`artifacts/fixture-mutations-${info.project.name}.png`});
 await page.getByRole("button",{name:/Metabolismo/}).click();
 await expect(page.getByLabel("Objetivo da expedição")).toContainText("Setor 2 de 3");
 expect(picks).toEqual([{mutation:"digestion",revision:2}]);
 state={...state,phase:"failed",revision:4,seconds_left:0};
 await expect(page.getByLabel("Resultado da expedição")).toBeVisible();
 await page.keyboard.press("r");
 await expect(page.getByLabel("Resultado da expedição")).not.toBeVisible();expect(controls).toBe(1);
 expect(errors).toEqual([]);
});
