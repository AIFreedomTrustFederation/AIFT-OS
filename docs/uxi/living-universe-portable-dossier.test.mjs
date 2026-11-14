import test from 'node:test';
import assert from 'node:assert/strict';
import {analyzeAtlas} from './living-universe-skill-atlas.mjs';
import {canopyLayout} from './living-universe-canopy.mjs';
import {portableDossier} from './living-universe-portable-dossier.mjs';
const nodes=[
 {id:'repo',stage:'root',label:'Repository',repository_url:'https://github.com/AIFreedomTrustFederation/AIFT-OS'},
 {id:'trunk',stage:'trunk'}, {id:'branch',stage:'branch'},
 {id:'leaf',stage:'leaf'}, {id:'flower',stage:'flower'},
 {id:'product',stage:'fruit',product_url:'https://example.org/product'}];
const edges=nodes.slice(1).map((n,i)=>({from:nodes[i].id,to:n.id}));
test('portable dossier includes directed root-to-product lineage and explicit unverified status',()=>{
 const atlas=analyzeAtlas(nodes,edges),d=portableDossier(atlas,canopyLayout(atlas),'product');
 assert.equal(d.schema,'aift.living-universe.lineage-dossier.v1');
 assert.equal(d.independentlyVerified,false);
 assert.deepEqual(d.routes[0].ids,nodes.map(n=>n.id));
 assert.equal(d.records.length,6);
 assert.equal(d.records[0].stage,'fruit');
 assert.ok(d.records.some(r=>r.id==='repo'&&r.stage==='root'));
});
test('unreachable fruit cannot acquire a fabricated root path',()=>{
 const atlas=analyzeAtlas(nodes,[]),d=portableDossier(atlas,canopyLayout(atlas),'product');
 assert.equal(d.routes.length,0);
 assert.ok(d.evidenceGaps.some(g=>g.includes('repository root')));
 assert.equal(d.independentlyVerified,false);
});
test('unknown record cannot create a dossier',()=>{
 const atlas=analyzeAtlas(nodes,edges);
 assert.equal(portableDossier(atlas,canopyLayout(atlas),'unknown'),null);
});
