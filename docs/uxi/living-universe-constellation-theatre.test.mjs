import test from 'node:test';
import assert from 'node:assert/strict';
import {constellationModel} from './living-universe-constellation-theatre.mjs';
const records=[
 {id:'repo',stage:'root',label:'Repository',directParents:[]},
 {id:'system',stage:'branch',label:'System',directParents:['repo']},
 {id:'product',stage:'fruit',label:'Declared outcome',directParents:['system']}
];
test('directed constellation follows imported parent references without invented intermediate stages',()=>{
 const m=constellationModel(records,'product');
 assert.deepEqual(m.nodes.map(n=>n.id),['repo','system','product']);
 assert.equal(m.edges.length,2);
 assert.equal(m.edges.every(e=>e.anomaly),true);
 assert.equal(m.nodes.filter(n=>n.selected).length,1);
 assert.equal(m.summary.hasRoot,true);
});
test('unlinked record does not invent root-to-outcome evidence',()=>{
 const m=constellationModel([{id:'orphan',stage:'fruit',directParents:['missing']}],'orphan');
 assert.equal(m.summary.hasRoot,false);
 assert.deepEqual(m.edges,[]);
 assert.equal(m.summary.missing.includes('missing'),true);
});
test('empty and invalid selection fail closed',()=>{
 assert.equal(constellationModel([],null).empty,true);
 assert.equal(constellationModel(records,'not-found').empty,true);
});
test('repository roots are never transformed into fruit nodes',()=>{
 const m=constellationModel(records,'product');
 assert.equal(m.nodes.find(n=>n.id==='repo').stage,'root');
});
