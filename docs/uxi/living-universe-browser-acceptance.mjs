import { chromium } from 'playwright';
import assert from 'node:assert/strict';
import { mkdirSync } from 'node:fs';
const base = process.env.AIFT_BROWSER_BASE || 'http://127.0.0.1:8765';
const prefix = process.env.AIFT_UXI_PREFIX || '';
const out = process.env.AIFT_SCREENSHOT_DIR || 'uxi-browser-evidence';
mkdirSync(out,{recursive:true});
const browser = await chromium.launch({headless:true});
let failures=[];
try {
 for(const width of [375,1280]){
  const context=await browser.newContext({viewport:{width,height:900},reducedMotion:'reduce'});
  const page=await context.newPage();
  const errors=[];
  page.on('pageerror',e=>errors.push(e.message));
  for(const [name,path] of [['observatory','living-universe-2.html'],['expedition','living-universe-expedition.html']]){
   const response=await page.goto(base+'/'+prefix+path,{waitUntil:'load',timeout:30000});
   assert.equal(response?.status(),200,name+' HTTP status');
   await page.screenshot({path:out+'/'+name+'-'+width+'.png',fullPage:true,animations:'disabled'});
   const body=await page.locator('body').innerText();
   assert.ok(body.length>300,name+' rendered content');
   assert.ok(/root/i.test(body),name+' root vocabulary');
   assert.ok(/evidence|provenance/i.test(body),name+' evidence vocabulary');
   const links=await page.locator('a[href]').evaluateAll(nodes=>nodes.filter(a=>a.href.startsWith('https:')&&a.target==='_blank'&&!a.rel.includes('noopener')).map(a=>a.href));
   assert.deepEqual(links,[],name+' external target protections');
   const buttons=page.locator('button:visible');
   if(await buttons.count()){
    await buttons.first().focus();
    assert.equal(await buttons.first().evaluate(e=>document.activeElement===e),true,name+' keyboard focus');
   }
  }
  if(errors.length)failures.push({width,errors});
  await context.close();
 }
 assert.deepEqual(failures,[],'uncaught browser exceptions');
 console.log('Browser acceptance passed: Observatory and Expedition at 375px/1280px, reduced motion, keyboard focus, links, screenshots');
} finally {await browser.close()}
