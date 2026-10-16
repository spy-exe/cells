import 'pixi.js/browser';
// CSP-compatible uniform/shader dispatch: this shim avoids eval.
import 'pixi.js/unsafe-eval';
import {WebGLRenderer,Container,Graphics,Sprite,Text,Texture} from 'pixi.js';
import type {Snapshot,Entity} from '../../../packages/protocol/game';

const palette=[0x90bc4b,0xe19a76,0x8c91d6,0x68b5bd,0xd6b465,0xc584ae,0x709bd2,0xb2b557];
interface View {sprite:Sprite;label:Text|null}
export class ArenaRenderer {
 private renderer!:WebGLRenderer;
 private stage=new Container();
 private resize:ResizeObserver|null=null;
 private frame=0;
 private renderedAt=0;
 private measuredFPS=0;
 private world=new Container();
 private views=new Map<number,View>();
 private circle:Texture=Texture.EMPTY;
 private virus:Texture=Texture.EMPTY;
 private current:Snapshot|null=null;
 private previous=new Map<number,Entity>();
 private camera={x:0,y:0,scale:1};
 private lastReceived=0;
 private quality='high';
 private destroyed=false;
 private ready=false;
 private gridSize=0;
 private grid:Graphics|null=null;
 private fpsAt=0;
 onFPS:(fps:number)=>void=()=>{};
 async init(host:HTMLElement){
  this.renderer=new WebGLRenderer();

  await this.renderer.init({width:host.clientWidth,height:host.clientHeight,background:0xf4f7ef,antialias:true,resolution:Math.min(devicePixelRatio,2),autoDensity:true,failIfMajorPerformanceCaveat:false,skipExtensionImports:true,powerPreference:'low-power'});


  if(this.destroyed){this.renderer.destroy({removeView:true});return}
  host.appendChild(this.renderer.canvas);this.renderer.canvas.setAttribute('aria-label','Arena Cells. Use mouse, setas ou toque para mover. Espaço divide e W ejeta massa.');this.renderer.canvas.setAttribute('role','img');
  const circle=new Graphics().circle(64,64,61).fill(0xffffff).stroke({width:3,color:0xffffff,alpha:.6});
  this.circle=this.renderer.generateTexture(circle);circle.destroy();
  const points:number[]=[];for(let i=0;i<48;i++){const a=i*Math.PI/24,r=i%2?53:63;points.push(64+Math.cos(a)*r,64+Math.sin(a)*r)}
  const virus=new Graphics().poly(points).fill(0xffffff).stroke({width:2,color:0x456f35});this.virus=this.renderer.generateTexture(virus);virus.destroy();
  this.stage.addChild(this.world);this.ready=true;
  this.resize=new ResizeObserver(()=>this.renderer.resize(host.clientWidth,host.clientHeight));this.resize.observe(host);
  const tick=(now:number)=>{if(this.destroyed)return;this.frame=requestAnimationFrame(tick);const dt=now-this.renderedAt;if(dt<(this.quality==='low'?32:14))return;this.measuredFPS=1000/Math.max(1,dt);this.renderedAt=now;this.render();this.renderer.render({container:this.stage})};this.frame=requestAnimationFrame(tick);
 }
 setQuality(quality:string){this.quality=quality}
 update(snapshot:Snapshot){
  if(this.current){this.previous=new Map(this.current.entities.map(e=>[e.id,e]));this.lastReceived=this.current.received}
  const first=!this.current||this.current.alive!==snapshot.alive;this.current=snapshot;
  if(first){this.camera.x=snapshot.x;this.camera.y=snapshot.y}
 }
 target(clientX:number,clientY:number){
  if(!this.ready)return {x:this.current?.x??0,y:this.current?.y??0};
  const rect=this.renderer.canvas.getBoundingClientRect();return{x:this.camera.x+(clientX-rect.left-rect.width/2)/this.camera.scale,y:this.camera.y+(clientY-rect.top-rect.height/2)/this.camera.scale};
 }
 private render(){
  const s=this.current;if(!s||!this.ready)return;
  if(this.gridSize!==s.size){
   this.grid?.destroy();const g=new Graphics();
   for(let x=0;x<=s.size;x+=80){g.moveTo(x,0).lineTo(x,s.size);g.moveTo(0,x).lineTo(s.size,x)}
   g.stroke({width:1,color:0x9bac8e,alpha:.16});g.rect(0,0,s.size,s.size).stroke({width:8,color:0x9bac8e,alpha:.5});g.zIndex=-1;this.world.addChild(g);this.grid=g;this.gridSize=s.size;
  }
  const {width,height}=this.renderer.screen;const targetScale=Math.max(width/1.9,height/1.4)/s.view;
  this.camera.x+=(s.x-this.camera.x)*.2;this.camera.y+=(s.y-this.camera.y)*.2;this.camera.scale+=(targetScale-this.camera.scale)*.12;
  this.world.position.set(width/2-this.camera.x*this.camera.scale,height/2-this.camera.y*this.camera.scale);this.world.scale.set(this.camera.scale);
  const t=Math.min(1,Math.max(0,(performance.now()-100-this.lastReceived)/Math.max(1,s.received-this.lastReceived)));
  const seen=new Set<number>();
  for(const e of s.entities){
   seen.add(e.id);let v=this.views.get(e.id);
   if(!v){
    const sprite=new Sprite(e.kind===4?this.virus:this.circle);sprite.anchor.set(.5);
    sprite.zIndex=e.kind===1?0:e.kind===4?2:3;this.world.addChild(sprite);
    const label=e.kind===2?new Text({text:'',style:{fontFamily:'Inter Variable, sans-serif',fontWeight:'600',fontSize:14,fill:0x17251c,align:'center'}}):null;
    if(label){label.anchor.set(.5);label.zIndex=4;this.world.addChild(label)}v={sprite,label};this.views.set(e.id,v);
   }
   const prev=this.previous.get(e.id)??e,x=prev.x+(e.x-prev.x)*t,y=prev.y+(e.y-prev.y)*t,r=4*Math.sqrt(prev.mass+(e.mass-prev.mass)*t);
   v.sprite.position.set(x,y);v.sprite.width=r*2;v.sprite.height=r*2;
   v.sprite.tint=e.kind===4?0x8bae65:s.teams&&e.kind===2?[0xa6bd65,0xe89b87,0x88b6d7][e.team%3]:palette[(e.kind===2?e.owner:e.id)%palette.length];
   v.sprite.alpha=e.kind===1?.65:1;
   if(v.label){v.label.text=`${e.name}${e.bot?' · IA':''}`;v.label.style.fontSize=Math.max(13,Math.min(28,r*.34));v.label.position.set(x,y);v.label.visible=r*this.camera.scale>12&&this.quality!=='low';v.label.scale.set(Math.min(1,r*1.6/Math.max(1,v.label.width)))}
  }
  for(const[id,v]of this.views){if(!seen.has(id)){v.sprite.destroy();v.label?.destroy();this.views.delete(id)}}
  this.world.sortableChildren=true;
  if(performance.now()-this.fpsAt>500){this.onFPS(Math.round(this.measuredFPS));this.fpsAt=performance.now()}
 }
 destroy(){this.destroyed=true;cancelAnimationFrame(this.frame);this.resize?.disconnect();if(this.ready){this.stage.destroy({children:true});this.renderer.destroy({removeView:true});this.circle.destroy(true);this.virus.destroy(true)}this.views.clear()}
}
