import test from 'node:test';
import assert from 'node:assert/strict';
import {createDossier,validateDossier,DOSSIER_KIND} from './living-universe-dossier.mjs';
const nodes=[
 {id:'repo',stage:'root',label:'Source repository',secret:'must not export'},
 {id:'trunk',stage:'trunk',label:'Federation'},
 {id:'system',stage:'branch',label:'System'},
 {id:'leaf',stage:'leaf',label:'Capability'},
 {id:'outcome',stage:'fruit',label:'Declared outcome',url:'https://example.org/product'},
 {id:'other',stage:'root',label:'Unrelated repository'}
];
const edges=[{from:'repo',to:'trunk'},{from:'trunk',to:'system'},{from:'system',to:'leaf'},{from:'leaf',to:'outcome'}];
test('portable dossier preserves all declared ancestors and source-to-outcome edges',()=>{
 const d=createDossier(nodes,edges,'outcome');assert.equal(d.kind,DOSSIER_KIND);
 assert.deepEqual(d.nodes.map(n=>n.id),['repo','trunk','system','leaf','outcome']);
 assert.equal(d.edges.length,4);assert.equal(validateDossier(JSON.parse(JSON.stringify(d))).focusId,'outcome');
 assert.equal(d.provenance.verification,'not-independently-verified');
});
test('does not export unrelated nodes or unknown private metadata',()=>{
 const d=createDossier(nodes,edges,'outcome');assert.equal(d.nodes.some(n=>n.id==='other'),false);
 assert.equal(d.nodes[0].secret,undefined);
});
test('missing lineage stays missing rather than being inferred',()=>{
 const d=createDossier(nodes,[], 'outcome');assert.deepEqual(d.nodes.map(n=>n.id),['outcome']);assert.equal(d.edges.length,0);
});
test('cycles terminate and preserve only declared relationships',()=>{
 const d=createDossier(nodes,[...edges,{from:'outcome',to:'system'}],'outcome');assert.equal(d.nodes.length,5);assert.equal(d.edges.length,5);
});
test('rejects unknown focus, malformed edges and corrupted import',()=>{
 assert.throws(()=>createDossier(nodes,edges,'unknown'));
 assert.throws(()=>createDossier(nodes,[{from:'unknown',to:'outcome'}],'outcome'));
 const d=createDossier(nodes,edges,'outcome');d.edges.push({from:'unknown',to:'outcome'});assert.throws(()=>validateDossier(d));
});
