import { expect, it } from 'vitest';
import { parseStatus } from '../../packages/protocol/status';
it('aceita estado real sem anunciar modos indisponíveis',()=>{
 expect(parseStatus({ok:true,ready:true,version:'0.0.1',revision:'abc',modes:[],players:0,bots:0}).modes).toEqual([]);
});
it('recusa payload sem versão ou com modos não reconhecidos',()=>{
 expect(()=>parseStatus({ok:true})).toThrow();
 expect(()=>parseStatus({ok:true,ready:true,version:'1',revision:'x',modes:['poderes'],players:0,bots:0})).toThrow();
});
