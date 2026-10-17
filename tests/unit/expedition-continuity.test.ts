import {expect,it} from "vitest";
import type {ExpeditionRun} from "../../apps/web/src/Expedition";
import {continuesExpedition} from "../e2e/helpers/expedition";
const start:ExpeditionRun={stage:1,phase:"running",revision:1,seed_label:"ABCDEF",seconds_left:24,ranks:{cilia:0,digestion:0,cohesion:0},target_mass:60,mass:36,peak:36,kills:0,offers:[]};
it("accepts continuous simulation and a legitimate death or sector end during reload",()=>{
 expect(continuesExpedition(start,{...start,seconds_left:22})).toBe(true);
 for(const phase of ["fallen","failed","choice"] as const)expect(continuesExpedition(start,{...start,phase,revision:2,seconds_left:0})).toBe(true);
});
it("rejects lost saves, unexplained revisions, changed mutations and impossible victories",()=>{
 for(const changed of [{seed_label:"OTHER"},{revision:0},{revision:5},{phase:"choice"},{seconds_left:25},{ranks:{...start.ranks,cilia:1}},{phase:"victory",revision:2}] as Partial<ExpeditionRun>[])
  expect(continuesExpedition(start,{...start,...changed})).toBe(false);
 const dead={...start,phase:"fallen" as const,revision:2,seconds_left:0};
 expect(continuesExpedition(dead,{...dead,phase:"running"})).toBe(false);
 expect(continuesExpedition(dead,{...dead,revision:3})).toBe(false);
});
