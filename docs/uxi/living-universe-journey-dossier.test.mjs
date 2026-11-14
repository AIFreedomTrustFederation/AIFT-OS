import test from 'node:test';
import assert from 'node:assert/strict';
import {analyzeAtlas} from './living-universe-skill-atlas.mjs';
import {canopyLayout} from './living-universe-canopy.mjs';
import {journeyDossier} from './living-universe-journey-dossier.mjs';
const nodes=[
 {id:'repo',stage:'root',label:'Source',repository_url:'https://github.com/AIFreedomTrustFederation/AIFT-OS'},
 {id:'trunk',stage:'trunk'},{id:'branch',stage:'branch'},{id:'leaf',stage:'leaf'},
 {id:'flower',stage:'flower'},{id:'product',stage:'fruit',product_url:'https://example.org/product'}];
const edges=nodes.slice(1).map((n,i)=>({from:nodes[i].id,to:n.id}));
test('directed root-to-outcome dossier exposes declared lineage without verification',()=>{
 const atlas=analyzeAtlas(nodes,edges),layout=canopyLayout(atlas),d=journeyDossier(atlas,layout,'product');
 assert.equal(d.routes.length,1);assert.deepEqual(d.routes[0].ids,nodes.map(n=>n.id));
 assert.deepEqual(d.stages.map(s=>s.records.length),[1,1,1,1,1,1]);
 assert.equal(d.verified,false);assert.equal(d.readiness.verified,false);
});
test('missing ancestry remains explicit and cannot award an achievement',()=>{
 const a=analyzeAtlas(nodes,[]),d=journeyDossier(a,canopyLayout(a),'product');
 assert.equal(d.routes.length,0);assert.ok(d.evidenceGaps.some(x=>x.includes('repository root')));assert.equal(d.verified,false);
});
test('source repository can never be a fruit or destination',()=>{
 const bad=[{id:'repo-as-fruit',stage:'fruit',repository_url:'https://github.com/AIFreedomTrustFederation/AIFT-OS'}];
 const a=analyzeAtlas(bad,[]),d=journeyDossier(a,canopyLayout(a),'repo-as-fruit');
 assert.equal(a.missions[0].destination,null);assert.ok(d.evidenceGaps.some(x=>x.includes('destination')));
});
test('unknown selection returns null, never a fabricated record',()=>{
 const a=analyzeAtlas(nodes,edges);assert.equal(journeyDossier(a,canopyLayout(a),'missing'),null);
});
