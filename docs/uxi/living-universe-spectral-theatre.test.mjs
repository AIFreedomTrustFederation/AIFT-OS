import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
const here=new URL('.',import.meta.url);
const css=readFileSync(new URL('living-universe-spectral-theatre.css',here),'utf8');
const pages=['living-universe-journey-atlas.html','living-universe-mission-control.html'];
test('both real exploration surfaces load the same additive cinematic layer',()=>{
 for(const page of pages){
  const html=readFileSync(new URL(page,here),'utf8');
  assert.match(html,/<link rel="stylesheet" href="\.\/living-universe-spectral-theatre\.css">/);
  assert.match(html,/id="graph"/);
  assert.match(html,/id="directory"/);
  assert.match(html,/id="quests"/);
  assert.match(html,/type="module"/);
 }
});
test('the shared layer supports motion, contrast, keyboard, mobile and stage navigation',()=>{
 for(const contract of ['--aift-spectrum','--aift-red','--aift-green','--aift-blue','--aift-gold','prefers-reduced-motion:reduce','forced-colors:active',':focus-visible','max-width:700px','#graph','#directory','#quests','[aria-pressed="true"]']){
  assert.ok(css.includes(contract),contract);
 }
});
test('presentation layer never invents products, telemetry or provenance',()=>{
 assert.doesNotMatch(css,/fetch\s*\(|XMLHttpRequest|localStorage|sessionStorage|@import|url\s*\(\s*https?:/i);
 assert.doesNotMatch(css,/content:\s*["'][^"']*(?:verified|complete|achievement|fruit)[^"']*["']/i);
});
