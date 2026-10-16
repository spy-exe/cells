import {test,expect} from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';
test('entrada honesta, servidor real e preferência de movimento',async({page})=>{
 const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message));
 await page.goto('/');
 await expect(page.getByRole('heading',{level:1})).toContainText('Comece pequeno');
 await expect(page.getByRole('status')).toContainText('Servidor online');
 await expect(page.getByRole('button',{name:'Entrar na arena'})).toBeEnabled();
 await page.getByRole('switch',{name:/Reduzir movimento/}).click();
 await page.reload();
 await expect(page.getByRole('switch',{name:/Reduzir movimento/})).toHaveAttribute('aria-checked','true');
 for(const [width,height] of [[360,800],[390,844],[412,915],[768,1024],[1440,900]]){
  await page.setViewportSize({width,height});
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
 }
 expect((await new AxeBuilder({page}).analyze()).violations).toEqual([]);
 expect(errors).toEqual([]);
});
