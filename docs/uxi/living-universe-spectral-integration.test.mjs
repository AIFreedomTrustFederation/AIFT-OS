import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
const root=new URL('./',import.meta.url);
const read=name=>readFileSync(new URL(name,root),'utf8');
const css=read('living-universe-spectral-theatre.css');
for(const page of ['living-universe-journey-atlas.html','living-universe-mission-control.html']){
 test(page+' loads the shared cinematic design layer',()=>{
  const html=read(page);
  assert.match(html,/href="\.\/living-universe-spectral-theatre\.css"/);
  assert.match(html,/id="graph"/);
  assert.match(html,/id="directory"/);
  assert.match(html,/id="quests"/);
  assert.match(html,/id="status"/);
 });
}
test('shared palette retains six biological spectrum colors and focus visibility',()=>{
 for(const token of ['--aift-red','--aift-green','--aift-blue','--aift-gold','--aift-violet','--aift-spectrum'])assert.ok(css.includes(token),token);
 assert.match(css,/:focus-visible/);
});
test('cinematic treatment honors reduced motion and forced color modes',()=>{
 assert.match(css,/@media\s*\(prefers-reduced-motion:reduce\)/);
 assert.match(css,/@media\s*\(forced-colors:active\)/);
 assert.match(css,/transition:none!important/);
});
test('shared decoration cannot insert synthetic graph records',()=>{
 assert.doesNotMatch(css,/@import\s|url\(\s*['"]?https?:/i);
 assert.match(css,/pointer-events:none/);
});
