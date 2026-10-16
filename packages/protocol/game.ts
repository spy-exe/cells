export interface Entity {id:number;owner:number;kind:number;x:number;y:number;mass:number;team:number;bot:boolean;name:string}
export interface Rank {id:number;mass:number;bot:boolean;team:number;name:string}
export interface Snapshot {tick:number;self:number;alive:boolean;teams:boolean;size:number;x:number;y:number;view:number;mass:number;splitIn:number;ejectIn:number;respawnIn:number;teamMass:number[];entities:Entity[];leaderboard:Rank[];received:number}
export function inputFrame(seq:number,x:number,y:number,actions=0):ArrayBuffer {
 const b=new ArrayBuffer(16),v=new DataView(b);v.setUint8(0,206);v.setUint8(1,1);v.setUint8(2,1);v.setUint32(3,seq,true);v.setFloat32(7,x,true);v.setFloat32(11,y,true);v.setUint8(15,actions);return b;
}
export function controlFrame(kind:3|4|5):ArrayBuffer {
 const b=new ArrayBuffer(kind===5?11:3),v=new DataView(b);v.setUint8(0,206);v.setUint8(1,1);v.setUint8(2,kind);if(kind===5)v.setFloat64(3,performance.now(),true);return b;
}
const decoder=new TextDecoder('utf-8',{fatal:true});
export function decodeSnapshot(b:ArrayBuffer):Snapshot {
 if(b.byteLength<53||b.byteLength>200000)throw Error('Tamanho de atualização inválido');
 const v=new DataView(b);if(v.getUint8(0)!==206||v.getUint8(1)!==1||v.getUint8(2)!==2)throw Error('Protocolo incompatível');
 const f=(offset:number)=>{const n=v.getFloat32(offset,true);if(!Number.isFinite(n))throw Error('Número inválido');return n};
 const count=v.getUint16(50,true),ranks=v.getUint8(52);if(count>2048||ranks>10)throw Error('Limite de entidades');
 const s:Snapshot={tick:v.getUint32(3,true),self:v.getUint32(7,true),alive:!!(v.getUint8(11)&1),teams:!!(v.getUint8(11)&2),size:f(12),x:f(16),y:f(20),view:f(24),mass:f(28),splitIn:v.getUint16(32,true),ejectIn:v.getUint16(34,true),respawnIn:v.getUint16(36,true),teamMass:[f(38),f(42),f(46)],entities:[],leaderboard:[],received:performance.now()};
 if(s.size<800||s.size>20000||s.view<=0||s.mass<0)throw Error('Mundo inválido');
 let p=53;
 const text=(size:number)=>{if(p+size>b.byteLength)throw Error('Nome truncado');const name=decoder.decode(new Uint8Array(b,p,size));p+=size;return name};
 for(let i=0;i<count;i++){
  if(p+24>b.byteLength)throw Error('Entidade truncada');
  const e:Entity={id:v.getUint32(p,true),owner:v.getUint32(p+4,true),x:f(p+8),y:f(p+12),mass:f(p+16),kind:v.getUint8(p+20),team:v.getUint8(p+21),bot:!!v.getUint8(p+22),name:''};const size=v.getUint8(p+23);p+=24;e.name=text(size);if(e.mass<=0||e.kind<1||e.kind>4)throw Error('Entidade inválida');s.entities.push(e);
 }
 for(let i=0;i<ranks;i++){
  if(p+11>b.byteLength)throw Error('Ranking truncado');const rank:Rank={id:v.getUint32(p,true),mass:f(p+4),bot:!!v.getUint8(p+8),team:v.getUint8(p+9),name:''};const size=v.getUint8(p+10);p+=11;rank.name=text(size);s.leaderboard.push(rank);
 }
 if(p!==b.byteLength)throw Error('Dados excedentes');return s;
}
