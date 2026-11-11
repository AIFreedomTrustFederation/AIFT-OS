import test from 'node:test';
import assert from 'node:assert/strict';
import {auditLineage,auditGraph} from './living-universe-lineage-audit.mjs';
const nodes=[{id:'repo',stage:'root',repository_url:'https://github.com/example/repo'},{id:'trunk',stage:'trunk'},{id:'service',stage:'branch'},{id:'app',stage:'fruit',url:'https://example.org/app'}];
const edges=[{from:'repo',to:'trunk'},{from:'trunk',to:'service'},{from:'service',to:'app'}];
test('directed root-to-outcome audit does not turn repository into fruit',()=>{
 const a=auditLineage(nodes,edges,'app');assert.deepEqual(a.roots,['repo']);assert.equal(a.classification,'declared-lineage');assert.equal(a.outcomeUrl,'https://example.org/app');assert.equal(a.ancestry.length,4);
 assert.equal(auditGraph(nodes,edges).outcomes,1);
});
test('unlinked outcomes remain evidence gaps',()=>{
 const a=auditLineage(nodes,[], 'app');assert.equal(a.classification,'evidence-gap');assert.equal(a.gaps.length,2);
});
test('cycle traversal is bounded and does not fabricate roots',()=>{
 const a=auditLineage([{id:'a',stage:'branch'},{id:'b',stage:'leaf'}],[{from:'a',to:'b'},{from:'b',to:'a'}],'a');assert.equal(a.ancestry.length,2);assert.equal(a.roots.length,0);
});
test('unsafe outcome destinations are rejected',()=>{
 const a=auditLineage([...nodes.slice(0,3),{id:'app',stage:'fruit',url:'javascript:alert(1)'}],edges,'app');assert.equal(a.outcomeUrl,null);assert.match(a.gaps.join(' '),/destination/);
});
test('invalid edges, duplicate nodes and unknown stages fail closed',()=>{
 assert.throws(()=>auditLineage(nodes,[{from:'ghost',to:'app'}],'app'));
 assert.throws(()=>auditLineage([...nodes,nodes[0]],edges,'app'));
 assert.throws(()=>auditLineage([{id:'x',stage:'repository-fruit'}],[],'x'));
});
