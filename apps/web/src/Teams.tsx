import {useRef} from "react";
import type {Snapshot} from "../../../packages/protocol/game";
const teams=[{name:"Musgo",color:"#73983b"},{name:"Coral",color:"#c46f57"},{name:"Maré",color:"#4f8fbc"}];
export function Teams({snapshot}:{snapshot:Snapshot|null}) {
 const mass=(snapshot?.teamMass??[0,0,0]).map(v=>Number.isFinite(v)?Math.max(0,v):0);
 const total=mass.reduce((sum,value)=>sum+value,0),share=mass.map(value=>total?value/total:0);
 const own=snapshot?.entities.find(e=>e.kind===2&&e.owner===snapshot.self)?.team;
 const assigned=useRef<number|undefined>(undefined);
 if(assigned.current===undefined && own!==undefined)assigned.current=own;
 return <aside className="leaderboard team-standings" aria-label="Disputa entre equipes">
  <h2>O espaço é de todos<span>TEAMS</span></h2>
  <div className="team-shares" role="img" aria-label={teams.map((t,i)=>`${t.name}: ${Math.round(share[i]*100)}%`).join(", ")}>{teams.map((t,i)=><i key={t.name} style={{width:`${share[i]*100}%`,background:t.color}}/>)}</div>
  <ol>{teams.map((t,i)=><li key={t.name} className={i===assigned.current?"is-you":""}>
   <i className="team-dot" style={{background:t.color}} aria-hidden="true"/><span>{t.name}{i===assigned.current&&<small>você</small>}</span><b>{Math.round(share[i]*100)}%</b>
  </li>)}</ol>
  <p>Some massa. Proteja sua equipe.</p>
 </aside>;
}
