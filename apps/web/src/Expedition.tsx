import { useCallback, useEffect, useRef, useState } from "react";
import { ArrowRight, Layers, Leaf, Wind } from "lucide-react";
export interface ExpeditionRun {
 stage: number; phase: "running" | "choice" | "fallen" | "failed" | "victory";
 revision: number; seconds_left: number; target_mass: number; mass: number;
 peak: number; kills: number; seed_label: string;
 ranks: Record<string, number>;
 offers: {id: string; name: string; detail: string}[];
}
export function useExpedition(enabled: boolean) {
 const [run,setRun]=useState<ExpeditionRun|null>(null),[error,setError]=useState("");
 const alive=useRef(false),busy=useRef(false);
 const refresh=useCallback(async()=>{
  if(!enabled || busy.current)return;busy.current=true;
  try {
   const response=await fetch("/api/expedition");
   if(response.status===404)return;
   if(!response.ok)throw Error("Não foi possível atualizar a expedição.");
   const value=await response.json() as ExpeditionRun;
   if(!["running","choice","fallen","failed","victory"].includes(value.phase)||![value.stage,value.revision,value.seconds_left,value.target_mass,value.mass].every(Number.isFinite)||!Array.isArray(value.offers))throw Error("Atualização incompatível. Recarregue a página.");
   if(alive.current){setRun(value);setError("");}
  }catch(e){if(alive.current)setError(e instanceof Error?e.message:"Falha de conexão.");}
  finally{busy.current=false;}
 },[enabled]);
 useEffect(()=>{
  alive.current=true;
  if(!enabled)return ()=>{alive.current=false;};
  void refresh();const timer=setInterval(()=>void refresh(),1000);
  return ()=>{alive.current=false;clearInterval(timer);};
 },[enabled,refresh]);
 const choose=async(id:string)=>{
  if(!run)return;
  try{
   const response=await fetch("/api/expedition/mutation",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({mutation:id,revision:run.revision})});
   if(!response.ok && response.status!==409)throw Error("Não foi possível escolher. Tente novamente.");
   await refresh();
  }catch(e){if(alive.current)setError(e instanceof Error?e.message:"Falha de conexão.");}
 };
 return {run,error,choose};
}
export function Expedition({run,error,choose,onRestart,onLeave}:{run:ExpeditionRun|null;error:string;choose:(id:string)=>Promise<void>;onRestart:()=>void;onLeave:()=>void}) {
 const [pending,setPending]=useState(false);const focus=useRef<HTMLHeadingElement>(null);
 useEffect(()=>{if(run && run.phase!=="running")focus.current?.focus();},[run?.phase]);
 if(!run)return <div className="expedition-loading" role="status">Preparando expedição…{error && <p role="alert">{error}</p>}</div>;
 const ended=run.phase!=="running"&&run.phase!=="choice";
 return <>
  <aside className="expedition-hud" aria-label="Objetivo da expedição">
   <div><span>Setor {run.stage} de 3</span><b>{run.seconds_left}s</b></div>
   <p>Alcance {run.target_mass} de massa e sobreviva.</p>
   <progress aria-label="Massa para concluir o setor" max={run.target_mass} value={Math.min(run.mass,run.target_mass)}/>
   <small>{Math.floor(run.mass)} / {run.target_mass} · {Object.values(run.ranks).reduce((a,b)=>a+b,0)} mutações</small>
  </aside>
  {run.phase!=="running" && <section className="arena-overlay expedition-overlay" aria-label={ended?"Resultado da expedição":"Escolher mutação"}>
   <p className="eyebrow">Expedição · {run.seed_label}</p>
   <h1 ref={focus} tabIndex={-1}>{run.phase==="choice"?"Evolua à sua maneira.":run.phase==="victory"?"Você atravessou o bioma.":run.phase==="fallen"?"A cadeia continua.":"Faltou um pouco de massa."}</h1>
   <p>{run.phase==="choice"?`Setor ${run.stage} concluído. O mundo está pausado. Escolha uma mutação para os próximos setores.`:run.phase==="victory"?"Três setores, duas mutações e uma célula que encontrou seu caminho.":"Uma nova semente traz outro mapa e novas oportunidades."}</p>
   {run.phase==="choice" && <div className="mutation-options">{run.offers.map((offer,i)=>{
    const Icon=[Wind,Leaf,Layers][i%3];return <button className="mutation-option" key={offer.id} disabled={pending} onClick={()=>{
     setPending(true);void choose(offer.id).finally(()=>setPending(false));
    }}><Icon size={23} aria-hidden="true"/><span><strong>{offer.name}</strong><small>{offer.detail}</small></span><ArrowRight size={16} aria-hidden="true"/></button>;
   })}</div>}
   {ended && <><dl className="run-summary"><div><dt>Pico de massa</dt><dd>{Math.floor(run.peak)}</dd></div><div><dt>Eliminações</dt><dd>{run.kills}</dd></div></dl><button className="play-button" onClick={onRestart}>Nova expedição <ArrowRight size={18}/></button><small className="respawn-hint">R para recomeçar</small><button className="text-button" onClick={onLeave}>Voltar ao menu</button></>}
   {error && <p className="form-error" role="alert">{error}</p>}
  </section>}
 </>;
}
