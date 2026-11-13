import {readFileSync} from 'node:fs';
import {test} from 'node:test';
import assert from 'node:assert/strict';
const page=readFileSync(new URL('./living-universe-2.html',import.meta.url),'utf8');
test('shared lineage navigation is accessible and reversible',()=>{
 for(const token of ['aria-label="Living lineage history"','aiftVisit(m.id)','aiftVisit(n.id)','aiftMove(-1)','aiftMove(1)','Clear trail','aria-live="polite"'])assert.ok(page.includes(token),token);
});
test('lineage history is bounded, local and evidence-derived',()=>{
 assert.match(page,/aiftVisited\.length>20/);
 assert.match(page,/current\?\.missions\?\.some\(m=>m\.id===id\)/);
 assert.match(page,/declared evidence, not independently verified/);
 assert.match(page,/aift:observatory-evidence/);
});
test('mobile and reduced-motion equivalents exist',()=>{
 assert.match(page,/@media\(max-width:600px\)\{\.aift-trail/);
 assert.match(page,/@media\(prefers-reduced-motion:reduce\)/);
});
