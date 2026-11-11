import test from 'node:test';
import assert from 'node:assert/strict';
import { adaptGenesisManifest } from './genesis-adapter.mjs';
const sample={kind:'tree-manifest',status:'canonical-draft',nodes:[
{id:'a',label:'Root',metadata:{part:'roots'},epistemicClass:'empirical'},
{id:'b',label:'Conceptual fruit',metadata:{part:'fruit'},epistemicClass:'doctrinal'},
{id:'c',label:'New seed',metadata:{part:'newSeeds'}}],edges:[{from:'a',to:'b',relation:'emerges_from'}]};
test('Genesis canonical graph adapts with source and epistemic labels',()=>{
 const v=adaptGenesisManifest(sample);assert.equal(v.nodes.length,3);assert.equal(v.edges.length,1);
 assert.equal(v.nodes[0].stage,'root');assert.equal(v.nodes[1].stage,'flower');assert.equal(v.nodes[2].stage,'flower');
 assert.equal(v.nodes[1].epistemicClass,'doctrinal');assert.equal(v.sourceStatus,'canonical-draft');
 assert.ok(v.notes.includes('not a verified product'));
});
test('rejects unknown relationships, unknown stages, duplicate ids and noncanonical data',()=>{
 assert.throws(()=>adaptGenesisManifest({nodes:[],edges:[]}));
 assert.throws(()=>adaptGenesisManifest({...sample,edges:[{from:'ghost',to:'b'}]}));
 assert.throws(()=>adaptGenesisManifest({...sample,nodes:[...sample.nodes,sample.nodes[0]]}));
 assert.throws(()=>adaptGenesisManifest({...sample,nodes:[{id:'x',metadata:{part:'unknown'}}]}));
});
