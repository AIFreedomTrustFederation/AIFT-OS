import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {journeyStages,BIO_STAGES,passageExplorer} from './living-universe-passage.mjs';
const root=new URL('.',import.meta.url);
test('six biological stages always preserve source order',()=>{assert.deepEqual(BIO_STAGES,['root','trunk','branch','leaf','flower','fruit']);assert.equal(journeyStages().missing.length,6)});
test('repositories remain roots and missing fruits are explicit',()=>{const m=journeyStages([{id:'repo',stage:'root',label:'Repository'}],'repo');assert.equal(m.stages[0].count,1);assert.equal(m.stages[5].count,0);assert.equal(m.selected,'repo');assert.ok(m.missing.includes('fruit'))});
test('unknown selection fails closed',()=>{assert.equal(journeyStages([{id:'r',stage:'root'}],'fabricated').selected,null)});
for(const file of ['living-universe-journey-atlas.html','living-universe-mission-control.html']){
 test(file+' integrates passage without removing local evidence exploration',()=>{
  const html=readFileSync(new URL(file,root),'utf8');
  assert.match(html,/id="passage"/);assert.match(html,/mountJourneyRibbon/);
  assert.match(html,/living-universe-passage\.mjs/);
  assert.match(html,/type="file"/);assert.match(html,/createExplorationController/);
  assert.match(html,/prefers-reduced-motion/);
 });
}

test('semantic zoom filters biological stages without inventing fruit',()=>{
 const records=[{id:'repo',stage:'root',label:'Source repository'},{id:'leaf',stage:'leaf',label:'Search capability'}];
 assert.deepEqual(passageExplorer(records,{stage:'fruit'}).visible,[]);
 assert.deepEqual(passageExplorer(records,{stage:'root'}).visible.map(r=>r.id),['repo']);
 assert.equal(passageExplorer(records,{stage:'unknown'}).focus,'all');
});
test('search and bounded record lists are deterministic',()=>{
 const records=Array.from({length:110},(_,i)=>({id:'r'+i,stage:'root',label:'Source '+i}));
 assert.equal(passageExplorer(records,{query:'Source 10'}).total,11);
 const all=passageExplorer(records,{limit:200});
 assert.equal(all.visible.length,100);assert.equal(all.hidden,10);
 assert.equal(passageExplorer(null).total,0);
});
test('passage uses safe DOM APIs and accessible stage controls',()=>{
 const js=readFileSync(new URL('living-universe-passage.mjs',root),'utf8');
 assert.match(js,/aria-label','Biological semantic zoom/);
 assert.match(js,/aria-live','polite/);
 assert.doesNotMatch(js,/innerHTML\s*=/);
});
