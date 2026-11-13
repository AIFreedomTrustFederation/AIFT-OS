import test from 'node:test';import assert from 'node:assert/strict';
import {spatialSummary,MAP_STAGES} from './living-universe-spatial-minimap.mjs';
test('six biological stages and repository roots stay distinct',()=>{const m=spatialSummary([{id:'repo',stage:'root'},{id:'product',stage:'fruit',directParents:['repo']}],'product');assert.deepEqual(m.stages.map(s=>s.stage),MAP_STAGES);assert.equal(m.stages[0].lineage,1);assert.equal(m.stages[5].count,1);assert.equal(m.hasRoot,true)});
test('missing provenance fails visibly without invented roots',()=>{const m=spatialSummary([{id:'fruit',stage:'fruit',directParents:['absent']}],'fruit');assert.deepEqual(m.missing,['absent']);assert.equal(m.hasRoot,false);assert.equal(m.stages[0].count,0)});
test('bounded imported data and unknown selections',()=>{const rows=Array.from({length:500},(_,i)=>({id:String(i),stage:'leaf'}));assert.equal(spatialSummary(rows).stages[3].count,350);assert.equal(spatialSummary(rows,'unknown').active,null)});
