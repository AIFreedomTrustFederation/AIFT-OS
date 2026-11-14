import { chromium } from 'playwright';
import assert from 'node:assert/strict';
import { mkdirSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

const base = process.env.AIFT_BROWSER_BASE || 'http://127.0.0.1:8765';
const prefix = process.env.AIFT_UXI_PREFIX || 'docs/uxi/';
const out = process.env.AIFT_SCREENSHOT_DIR || 'uxi-browser-evidence';
mkdirSync(out, {recursive:true});
const fixture = {
  nodes:[
    {id:'test-repo',stage:'root',label:'TEST ONLY source repository',repository_url:'https://github.com/AIFreedomTrustFederation/AIFT-Genesis'},
    {id:'test-trunk',stage:'trunk',label:'TEST ONLY federation trunk'},
    {id:'test-branch',stage:'branch',label:'TEST ONLY documented system'},
    {id:'test-leaf',stage:'leaf',label:'TEST ONLY capability'},
    {id:'test-flower',stage:'flower',label:'TEST ONLY emerging work'},
    {id:'test-fruit',stage:'fruit',label:'TEST ONLY outcome',product_url:'https://example.org/test-only'}
  ],
  edges:[
    {from:'test-repo',to:'test-trunk'},
    {from:'test-trunk',to:'test-branch'},
    {from:'test-branch',to:'test-leaf'},
    {from:'test-leaf',to:'test-flower'},
    {from:'test-flower',to:'test-fruit'}
  ]
};
const report=[];
const browser=await chromium.launch({headless:true});
try {
 for(const width of [375,1280]){
  const context=await browser.newContext({viewport:{width,height:900},reducedMotion:'reduce'});
  for(const [name,path] of [['journey-atlas','living-universe-journey-atlas.html'],['mission-control','living-universe-mission-control.html']]){
   const page=await context.newPage(),errors=[];
   page.on('pageerror',e=>errors.push(e.message));
   const response=await page.goto(new URL(prefix+path,base+'/').href,{waitUntil:'load'});
   assert.equal(response?.status(),200,name+' page response');
   assert.match(await page.locator('body').innerText(),/root/i);
   assert.match(await page.locator('#status').innerText(),/import|evidence|empty|no /i,'empty evidence state');
   await page.screenshot({path:join(out,name+'-'+width+'-before.png'),fullPage:true,animations:'disabled'});
   await page.locator('#file').setInputFiles({name:'TEST-ONLY-lineage.json',mimeType:'application/json',buffer:Buffer.from(JSON.stringify(fixture))});
   await assert.doesNotReject(async()=>{
    await page.locator('#directory').getByText('TEST ONLY source repository').first().waitFor({timeout:5000});
   },name+' root imported; status='+JSON.stringify(await page.locator('#status').innerText())+'; page errors='+JSON.stringify(errors));
   const directory=await page.locator('#directory').innerText();
   assert.match(directory,/TEST ONLY source repository/,'root retained');
   assert.match(directory,/TEST ONLY outcome/,'outcome retained');
   assert.equal(await page.locator('#directory').getByText('TEST ONLY source repository').count(),1,'repository appears once as a source record');
   assert.equal(await page.locator('#directory').getByText('TEST ONLY outcome').count(),1,'outcome appears once as a separate record');
   assert.equal(await page.evaluate(()=>matchMedia('(prefers-reduced-motion: reduce)').matches),true,'reduced motion enabled');
   const text=await page.locator('body').innerText();
   assert.match(text,/not independently verified|not verified|declared/i,'verification disclaimer');
   await page.locator('#search').fill('TEST ONLY outcome');
   assert.match(await page.locator('#directory').innerText(),/TEST ONLY outcome/,'searchable outcome');
   await page.locator('#search').fill('');
   const root=page.locator('#directory').getByText('TEST ONLY source repository').first();
   await root.click();
   assert.match(await page.locator('#inspector').innerText(),/TEST ONLY source repository|TEST ONLY source/i,'root inspector');
   const fruit=page.locator('#directory').getByText('TEST ONLY outcome').first();
   await fruit.click();
   assert.match(await page.locator('#inspector').innerText(),/TEST ONLY outcome/,'outcome inspector');
   const external=await page.locator('a[target="_blank"]').evaluateAll(nodes=>nodes.filter(a=>!a.rel.includes('noopener')||!a.rel.includes('noreferrer')).map(a=>a.href));
   assert.deepEqual(external,[],'external links isolated');
   const enabled=page.locator('button:visible:not([disabled])');
   if(await enabled.count()){
    await enabled.first().focus();
    assert.equal(await enabled.first().evaluate(el=>document.activeElement===el),true,'keyboard-focusable controls');
   }
   assert.deepEqual(errors,[],'no browser runtime exceptions');
   await page.screenshot({path:join(out,name+'-'+width+'-after.png'),fullPage:true,animations:'disabled'});
   await page.locator('#clear').click();
   assert.doesNotMatch(await page.locator('#directory').innerText(),/TEST ONLY outcome|TEST ONLY source repository/,'clear removes all imported records');
   assert.match(await page.locator('#status').innerText(),/clear|no evidence|empty/i,'clear remains fail-closed');
   await page.screenshot({path:join(out,name+'-'+width+'-cleared.png'),fullPage:true,animations:'disabled'});
   report.push({surface:name,width,importedNodes:fixture.nodes.length,synthetic:true,errors:errors.length});
   await page.close();
  }
  await context.close();
 }
} finally {await browser.close();writeFileSync(join(out,'multi-surface-results.json'),JSON.stringify(report,null,2));}
console.log('Synthetic multi-surface browser acceptance:',report);
