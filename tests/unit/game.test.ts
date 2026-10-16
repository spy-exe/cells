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
