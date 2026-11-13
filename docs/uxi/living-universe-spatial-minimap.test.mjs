import test from 'node:test';import assert from 'node:assert/strict';
import {spatialSummary,MAP_STAGES} from './living-universe-spatial-minimap.mjs';
test('six biological stages and repository roots stay distinct',()=>{const m=spatialSummary([{id:'repo',stage:'root'},{id:'product',stage:'fruit',directParents:['repo']}],'product');assert.deepEqual(m.stages.map(s=>s.stage),MAP_STAGES);assert.equal(m.stages[0].lineage,1);assert.equal(m.stages[5].count,1);assert.equal(m.hasRoot,true)});
test('missing provenance fails visibly without invented roots',()=>{const m=spatialSummary([{id:'fruit',stage:'fruit',directParents:['absent']}],'fruit');assert.deepEqual(m.missing,['absent']);assert.equal(m.hasRoot,false);assert.equal(m.stages[0].count,0)});
test('bounded imported data and unknown selections',()=>{const rows=Array.from({length:500},(_,i)=>({id:String(i),stage:'leaf'}));assert.equal(spatialSummary(rows).stages[3].count,350);assert.equal(spatialSummary(rows,'unknown').active,null)});

test('lineage is ordered from roots to selected fruit and does not invent products',()=>{
 const records=[{id:'outcome',stage:'fruit',directParents:['cap']},{id:'cap',stage:'leaf',directParents:['source']},{id:'source',stage:'root'}];
 const m=spatialSummary(records,'outcome');
 assert.deepEqual(m.lineage.map(x=>x.id),['source','cap','outcome']);
 assert.equal(m.hasRoot,true);
 assert.match(m.claim,/no independent product or CI verification/);
});
test('lineage traversal handles cycles and unknown parents without synthetic nodes',()=>{
 const m=spatialSummary([{id:'a',stage:'branch',directParents:['b','absent']},{id:'b',stage:'root',directParents:['a']}],'a');
 assert.deepEqual(m.lineage.map(x=>x.id),['b','a']);
 assert.deepEqual(m.missing,['absent']);
 assert.equal(m.records.length,2);
});
test('bounded imports disclose truncation and keep selected record evidence-scoped',()=>{
 const rows=Array.from({length:351},(_,i)=>({id:String(i),stage:'leaf'}));
 assert.equal(spatialSummary(rows,'350').active,null);
 assert.equal(spatialSummary(rows).truncated,true);
 assert.equal(spatialSummary(rows.slice(0,350)).truncated,false);
});
