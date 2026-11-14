import test from 'node:test';
import assert from 'node:assert/strict';
import {canopyLayout,canopyFocus,canopyLineage,canopyRoutes,canopyReadiness,STAGES} from './living-universe-canopy.mjs';
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

test('cinematic lineage journey exposes source roots, anomalies and missing references without invented nodes',async()=>{
 const {readFileSync}=await import('node:fs');const {fileURLToPath}=await import('node:url');const {dirname,join}=await import('node:path');
 const html=readFileSync(join(dirname(fileURLToPath(import.meta.url)),'living-universe-2.html'),'utf8');
 assert.match(html,/id="canopy-lineage-journey"/);assert.match(html,/Contributing repository roots/);assert.match(html,/Direct stage anomalies/);assert.match(html,/Missing graph references/);assert.match(html,/layout\.omissions\.filter/);assert.match(html,/document\.createElement\('button'\)/);
});

test('root-to-record journey resolves directed shortest paths without fabricated ancestry',()=>{
 const l=canopyLayout({missions:[mission('r','root'),mission('s','root'),mission('t','trunk',['r']),mission('f','fruit',['t','s'])]});
 const result=canopyRoutes(l,'f');
 assert.deepEqual(result.routes.map(r=>r.ids),[['r','t','f'],['s','f']]);
 assert.deepEqual(result.routes.map(r=>r.stageGaps),[1,1]);
 assert.equal(result.unreachable,false);
});
test('journeys are cycle-safe, bounded, deterministic and report absent provenance',()=>{
 const l=canopyLayout({missions:[mission('r','root',['f']),mission('f','fruit',['r']),mission('orphan','leaf')]});
 assert.deepEqual(canopyRoutes(l,'f').routes[0].ids,['r','f']);
 assert.equal(canopyRoutes(l,'orphan').unreachable,true);
 assert.equal(canopyRoutes(l,'unknown').unreachable,true);
 assert.throws(()=>canopyRoutes(l,'f',{maxRoutes:0}));
});

test('route theatre offers selectable full-path highlighting and accessible step navigation',async()=>{
 const {readFileSync}=await import('node:fs');const {fileURLToPath}=await import('node:url');const {dirname,join}=await import('node:path');
 const html=readFileSync(join(dirname(fileURLToPath(import.meta.url)),'living-universe-2.html'),'utf8');
 assert.match(html,/dataset\.routeIds=route\.ids\.join/);
 assert.match(html,/routeEdges\.has\(key\)/);
 assert.match(html,/routeIds\.has\(n\.id\)/);
 assert.match(html,/activate\.setAttribute\('aria-pressed'/);
 assert.match(html,/const steps=document\.createElement\('ol'\)/);
 assert.match(html,/prefers-reduced-motion:reduce/);
});

test('semantic zoom and minimap remain accessible and evidence-only',async()=>{
 const {readFileSync}=await import('node:fs');
 const {fileURLToPath}=await import('node:url');
 const {dirname,join}=await import('node:path');
 const html=readFileSync(join(dirname(fileURLToPath(import.meta.url)),'living-universe-2.html'),'utf8');
 for(const id of ['canopy-stage-nav','canopy-mini','canopy-zoom-reset','canopy-directory'])assert.match(html,new RegExp('id="'+id+'"'));
 assert.match(html,/aria-pressed/);
 assert.match(html,/sky\.setAttribute\('viewBox',view\.join\(' '\)\)/);
 assert.match(html,/prefers-reduced-motion:reduce/);
 assert.match(html,/paintNavigation\(layout,sky\)/);
 assert.match(html,/Viewport-only semantic zoom: never mutates the declared provenance graph/);
});

test('spatial keyboard navigation preserves canonical record identity and focus after rerender',async()=>{
 const {readFileSync}=await import('node:fs');const {fileURLToPath}=await import('node:url');const {dirname,join}=await import('node:path');
 const html=readFileSync(join(dirname(fileURLToPath(import.meta.url)),'living-universe-2.html'),'utf8');
 assert.match(html,/const canopyKeyboardMove=/);
 assert.match(html,/ArrowLeft/);assert.match(html,/ArrowRight/);assert.match(html,/ArrowUp/);assert.match(html,/ArrowDown/);
 assert.match(html,/focusCanopyRecord\(next/);
 assert.match(html,/data-canopy-id/);
 assert.match(html,/focus\(\{preventScroll:true\}\)/);
 assert.match(html,/source==='directory'/);
 assert.match(html,/activate\.textContent/);
});

test('provenance readiness is evidence-only and exposes source gaps without awarding verification',()=>{
 const l=canopyLayout({missions:[mission('repo','root'),mission('system','branch',['repo']),mission('product','fruit',['system','missing'])]});
 const r=canopyReadiness(l,'product');
 assert.equal(r.roots,1);assert.equal(r.intermediate,1);
 assert.equal(r.anomalies,2);assert.equal(r.missingReferences,1);
 assert.equal(r.hasDirectedRootPath,true);assert.equal(r.verified,false);
 assert.equal(canopyReadiness(l,'absent').found,false);
 const orphan=canopyLayout({missions:[mission('outcome','fruit',['unknown'])]});
 assert.equal(canopyReadiness(orphan,'outcome').hasDirectedRootPath,false);
});
test('readiness ribbon is present and never labels declared paths verified',async()=>{
 const {readFileSync}=await import('node:fs');const {fileURLToPath}=await import('node:url');const {dirname,join}=await import('node:path');
 const html=readFileSync(join(dirname(fileURLToPath(import.meta.url)),'living-universe-2.html'),'utf8');
 assert.match(html,/className='canopy-readiness'/);
 assert.match(html,/Independent verification is still required/);
 assert.match(html,/Missing references/);
 assert.match(html,/aria-label','Declared provenance journey readiness'/);
});
