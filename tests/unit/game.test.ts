import {readFileSync} from 'node:fs';
import {describe,it,expect} from 'vitest';
import {inputFrame,decodeSnapshot} from '../../packages/protocol/game';
describe('binary v1',()=>{
 it('encodes target coordinates, sequence and actions without client state',()=>{
  const b=inputFrame(7,123,456,1);const v=new DataView(b);expect(b.byteLength).toBe(16);expect(Array.from(new Uint8Array(b).slice(0,3))).toEqual([206,1,1]);expect(v.getUint32(3,true)).toBe(7);expect(v.getFloat32(7,true)).toBe(123);expect(v.getUint8(15)).toBe(1);
 });
 it('rejects truncated, unsupported and oversized snapshots',()=>{
  for(const b of [new ArrayBuffer(0),new ArrayBuffer(53)])expect(()=>decodeSnapshot(b)).toThrow();
  const b=new ArrayBuffer(53);const v=new DataView(b);v.setUint8(0,206);v.setUint8(1,1);v.setUint8(2,2);v.setUint16(50,65000,true);expect(()=>decodeSnapshot(b)).toThrow();
 });
});

it('decodes the exact Go golden packet including UTF-8 and cooldowns',()=>{const hex=readFileSync(new URL('../../packages/protocol/fixtures/snapshot-v1.hex',import.meta.url),'utf8').trim();const bytes=Uint8Array.from(Buffer.from(hex,'hex'));const s=decodeSnapshot(bytes.buffer);expect(s.self).toBe(7);expect(s.tick).toBe(30);expect(s.splitIn).toBe(24);expect(s.entities[0]).toMatchObject({name:'João',bot:true,owner:7,mass:36});expect(s.leaderboard[0].name).toBe('João');expect(s.teamMass).toEqual([36,0,0]);});
