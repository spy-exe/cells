import { expect, it } from "vitest";
import { entityDepth, membraneGeometry, updateMembrane, absorptionPose, detectAbsorptions, cellCaption } from "../../apps/web/src/organic";
import type { Snapshot } from "../../packages/protocol/game";
it("fits captions inside small cells while retaining the bot marker",()=>{
 const small=cellCaption('Alexandria',true,15),large=cellCaption('Alexandria',false,80);
 expect(small.text).toContain('IA');expect(small.text.length).toBeLessThan(8);
 expect(small.font).toBeGreaterThanOrEqual(10);
 expect(small.font).toBeLessThan(large.font);expect(large.text).toBe('Alexandria');
});

it("draws larger bodies over smaller bodies and their labels regardless of owner", () => {
  expect(entityDepth(2, 100)).toBeGreaterThan(entityDepth(2, 36, true));
  expect(entityDepth(2, 36, true)).toBeGreaterThan(entityDepth(2, 36));
  expect(entityDepth(2, 36)).toBeGreaterThan(entityDepth(4, 100));
});
it("keeps the membrane closed, bounded and equal-area while it moves", () => {
  const mesh=membraneGeometry();
  updateMembrane(mesh.vertices,17,1000,50,.4,false);
  const first=mesh.vertices.slice();
  let twiceArea=0;
  for(let i=2;i<mesh.vertices.length-2;i+=2){
    const r=Math.hypot(mesh.vertices[i],mesh.vertices[i+1]);
    expect(r).toBeGreaterThan(.95);expect(r).toBeLessThan(1.05);
    twiceArea+=mesh.vertices[i]*mesh.vertices[i+3]-mesh.vertices[i+2]*mesh.vertices[i+1];
  }
  expect(twiceArea/2).toBeCloseTo(Math.PI,4);
  expect(mesh.vertices[2]).toBe(mesh.vertices.at(-2));
  expect(mesh.vertices[3]).toBe(mesh.vertices.at(-1));
  updateMembrane(mesh.vertices,17,1300,350,.4,false);
  expect(mesh.vertices).not.toEqual(first);
  updateMembrane(mesh.vertices,17,1000,50,.4,true);
  const reduced=mesh.vertices.slice();
  updateMembrane(mesh.vertices,17,1300,350,.4,true);
  expect(mesh.vertices).toEqual(reduced);
});
it("animates confirmed visible food toward the eater with shrinking and no overshoot", () => {
  expect(absorptionPose(100,50,120,70,0)).toEqual({x:100,y:50,scale:1,alpha:1});
  const mid=absorptionPose(100,50,120,70,.5);
  expect(mid.x).toBeGreaterThan(100);expect(mid.x).toBeLessThan(120);
  expect(mid.scale).toBeGreaterThan(0);expect(mid.scale).toBeLessThan(1);
  expect(absorptionPose(100,50,120,70,1)).toEqual({x:120,y:70,scale:0,alpha:0});
});
it("requires disappearance, a nearby existing cell and enough observed growth", () => {
  const cell={id:2,owner:7,kind:2,x:100,y:100,mass:36,team:0,name:"A",bot:false};
  const food={...cell,id:3,owner:0,kind:1,x:120,mass:1,name:""};
  const before={entities:[cell,food]} as Snapshot;
  const after={entities:[{...cell,mass:37}]} as Snapshot;
  expect(detectAbsorptions(before,after)).toEqual([{item:3,eater:2}]);
  expect(detectAbsorptions(before,{entities:[cell]} as Snapshot)).toEqual([]);
  expect(detectAbsorptions(before,{entities:[{...cell,mass:37},food]} as Snapshot)).toEqual([]);
  expect(detectAbsorptions(before,{entities:[{...cell,mass:37,x:2000}]} as Snapshot)).toEqual([]);
});
it("keeps complete Unicode code points in tiny centered captions",()=>{
 expect(cellCaption("🧬Bioma",false,5).text).toBe("🧬");
});
it("tracks a captured visible cell, ejected mass and a same-owner recombination",()=>{
 const eater={id:10,owner:7,kind:2,x:100,y:100,mass:225,team:0,name:"A",bot:false};
 const item={...eater,id:11,owner:8,x:125,mass:36,name:"B"};
 const scene=(source:typeof item,prior=eater)=>[
  {tick:60,self:7,entities:[prior,source]},
  {tick:62,self:7,entities:[{...prior,mass:prior.mass+source.mass}]},
 ] as [Snapshot,Snapshot];
 expect(detectAbsorptions(...scene(item))).toEqual([{item:11,eater:10}]);
 expect(detectAbsorptions(...scene({...item,kind:3,mass:12}))).toEqual([{item:11,eater:10}]);
 expect(detectAbsorptions(...scene({...item,owner:7}, {...eater,mass:36}))).toEqual([{item:11,eater:10}]);
});
it("does not invent a capture by a smaller enemy, virus absorption or a reset-world merge",()=>{
 const a={id:10,owner:7,kind:2,x:100,y:100,mass:36,team:0,name:"A",bot:false};
 const b={...a,id:11,owner:8,x:110,mass:100};
 const before={tick:60,self:7,entities:[a,b]} as Snapshot;
 const after={tick:62,self:7,entities:[{...a,mass:136}]} as Snapshot;
 expect(detectAbsorptions(before,after)).toEqual([]);
 expect(detectAbsorptions({...before,entities:[{...a,mass:225},{...b,kind:4}]}, {...after,entities:[{...a,mass:325}]})).toEqual([]);
 expect(detectAbsorptions({...before,entities:[{...a,mass:225},b]}, {...after,tick:1,entities:[{...a,mass:325}]})).toEqual([]);
});
it("never shows an allied player being eaten in Teams, while allowing enemy captures",()=>{
 const a={id:10,owner:7,kind:2,x:100,y:100,mass:225,team:0,name:"A",bot:false};
 const b={...a,id:11,owner:8,mass:36,x:115};
 const before={tick:60,self:7,teams:true,entities:[a,b]} as Snapshot;
 const after={tick:62,self:7,teams:true,entities:[{...a,mass:261}]} as Snapshot;
 expect(detectAbsorptions(before,after)).toEqual([]);
 expect(detectAbsorptions({...before,entities:[a,{...b,team:1}]},after)).toEqual([{item:11,eater:10}]);
});
