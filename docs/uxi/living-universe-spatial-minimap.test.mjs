import test from 'node:test';import assert from 'node:assert/strict';
import {spatialSummary,spatialRoutes,MAP_STAGES} from './living-universe-spatial-minimap.mjs';
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

test('directed routes choose shortest declared path and preserve stage gaps',()=>{
 const rows=[
  {id:'root',stage:'root'},
  {id:'slow',stage:'branch',directParents:['root']},
  {id:'leaf',stage:'leaf',directParents:['slow']},
  {id:'fast',stage:'leaf',directParents:['root']},
  {id:'fruit',stage:'fruit',directParents:['leaf','fast']}
 ];
 const routes=spatialRoutes(rows,'fruit');
 assert.deepEqual(routes.map(r=>r.ids),[['root','fast','fruit']]);
 assert.deepEqual(routes[0].anomalies.map(a=>a.kind),['stage-gap','stage-gap']);
});
test('routes are bounded, cycle safe, deterministic and do not invent roots',()=>{
 const rows=[{id:'fruit',stage:'fruit',directParents:['b','missing']},
 {id:'b',stage:'branch',directParents:['a','fruit']},
 {id:'a',stage:'root'},
 {id:'z',stage:'root',directParents:['fruit']}];
 // z is a descendant of fruit, not an ancestor; treating it as a contributing root would fabricate provenance.
 assert.deepEqual(spatialRoutes(rows,'fruit').map(r=>r.rootId),['a']);
 assert.equal(spatialRoutes(rows,'missing').length,0);
 assert.equal(spatialRoutes(rows,'fruit',{maxRoutes:1}).length,1);
 assert.equal(spatialRoutes([{id:'fruit',stage:'fruit',directParents:['unknown']}],'fruit').length,0);
});
test('shared spatial summary carries routes without asserting verified outcomes',()=>{
 const m=spatialSummary([{id:'r',stage:'root'},{id:'f',stage:'fruit',directParents:['r']}],'f');
 assert.deepEqual(m.routes[0].ids,['r','f']);
 assert.match(m.claim,/Imported declarations only/);
});

test('a repository root downstream of an outcome cannot be credited as its contributor',()=>{
 const records=[{id:'product',stage:'fruit'},{id:'downstream-repo',stage:'root',directParents:['product']}];
 assert.deepEqual(spatialRoutes(records,'product'),[]);
 assert.equal(spatialSummary(records,'product').hasRoot,false);
});
