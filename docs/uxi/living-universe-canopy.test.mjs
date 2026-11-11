import test from 'node:test';
import assert from 'node:assert/strict';
import {canopyLayout,canopyFocus,canopyLineage,STAGES} from './living-universe-canopy.mjs';
const mission=(id,stage,directParents=[])=>({id,label:id,stage,directParents,issues:[],roots:[],status:'unverified'});
test('six stages and repository roots remain distinct from fruits',()=>{
 assert.deepEqual(STAGES,['root','trunk','branch','leaf','flower','fruit']);
 const l=canopyLayout({missions:[mission('repo','root'),mission('app','fruit',['repo'])]});
 assert.equal(l.nodes.find(n=>n.id==='repo').stage,'root');
 assert.equal(l.nodes.find(n=>n.id==='app').stage,'fruit');
});
test('transitive ancestry cannot manufacture an edge',()=>{
 const l=canopyLayout({missions:[{...mission('r','root')},{...mission('t','trunk'),ancestors:['r']}]});
 assert.deepEqual(l.connections,[]);
});
test('direct adjacent relationships retain stable contract',()=>{
 assert.deepEqual(canopyLayout({missions:[mission('r','root'),mission('t','trunk',['r'])]}).connections,[{from:'r',to:'t'}]);
});
test('direct stage-skipping edges are visible and explicitly marked, never invented intermediates',()=>{
 const l=canopyLayout({missions:[mission('r','root'),mission('f','fruit',['r'])]});
 assert.deepEqual(l.connections,[{from:'r',to:'f',stageDistance:5,kind:'stage-gap'}]);
 assert.equal(l.nodes.length,2);
 assert.deepEqual(canopyLineage(l,'f').rootIds,['r']);
 assert.equal(canopyLineage(l,'f').gaps.length,1);
});
test('backward or cyclic relationships terminate and expose anomalies',()=>{
 const l=canopyLayout({missions:[mission('r','root',['f']),mission('f','fruit',['r'])]});
 const lineage=canopyLineage(l,'f');
 assert.equal(lineage.nodes.length,2);
 assert.equal(lineage.connections.length,2);
 assert.equal(lineage.gaps.length,2);
});
test('unknown parents are omitted and recorded, never visualized as roots',()=>{
 const l=canopyLayout({missions:[mission('f','fruit',['missing'])]});
 assert.equal(l.connections.length,0);
 assert.equal(l.omissions[0].reason,'unrenderable-stage-or-unknown-parent');
 assert.deepEqual(canopyLineage(l,'f').rootIds,[]);
});
test('duplicate nodes and invalid viewport are rejected',()=>{
 assert.throws(()=>canopyLayout({missions:[mission('x','root'),mission('x','root')]}));
 assert.throws(()=>canopyLayout({missions:[]},{width:0}));
});
test('deterministic geometry, focus and missing lineage',()=>{
 const a={missions:[mission('b','trunk',['a']),mission('a','root')]};
 const l=canopyLayout(a);assert.deepEqual(l,canopyLayout(a));
 assert.equal(canopyFocus(l,'b').neighbors.length,1);
 assert.deepEqual(canopyLineage(l,'absent').nodes,[]);
});
test('canonical observatory selection remains wired',async()=>{
 const {readFileSync}=await import('node:fs');
 const {fileURLToPath}=await import('node:url');
 const {dirname,join}=await import('node:path');
 const html=readFileSync(join(dirname(fileURLToPath(import.meta.url)),'living-universe-2.html'),'utf8');
 assert.match(html,/addEventListener\('aift:explore-node'/);
 assert.match(html,/select\(id\);[\s\S]*?scrollIntoView/);
});
