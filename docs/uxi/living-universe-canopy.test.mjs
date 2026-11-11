import test from 'node:test';import assert from 'node:assert/strict';import {canopyLayout,canopyFocus,STAGES} from './living-universe-canopy.mjs';
const missions=[{id:'repo',label:'Source repository',stage:'root',issues:[],roots:[],status:'unverified',ancestors:[]},{id:'system',label:'System',stage:'branch',issues:[],roots:['repo'],status:'unverified',ancestors:['repo']},{id:'outcome',label:'Product',stage:'fruit',issues:['No destination'],roots:['repo'],status:'evidence-gap',ancestors:['repo','system']}];
test('six biological layers are stable',()=>assert.deepEqual(STAGES,['root','trunk','branch','leaf','flower','fruit']));
test('roots stay roots, not fruits',()=>{const l=canopyLayout({missions});assert.equal(l.nodes.find(n=>n.id==='repo').stage,'root');assert.equal(l.nodes.find(n=>n.id==='outcome').stage,'fruit')});
test('nonadjacent ancestry never fabricates direct edges',()=>assert.equal(canopyLayout({missions}).connections.length,0));
test('deterministic bounded layout and missing selection',()=>{const l=canopyLayout({missions});assert.deepEqual(l,canopyLayout({missions}));assert.equal(canopyFocus(l,'missing').node,null)});
test('evidence gaps remain visible',()=>assert.equal(canopyLayout({missions}).nodes.find(n=>n.id==='outcome').issues,1));

test('transitive adjacent-stage ancestry does not masquerade as a declared edge',()=>{
 const records=[{id:'r',label:'Root',stage:'root',issues:[],roots:[],status:'unverified',ancestors:[],directParents:[]},
 {id:'t',label:'Trunk',stage:'trunk',issues:[],roots:['r'],status:'unverified',ancestors:['r'],directParents:[]}];
 assert.deepEqual(canopyLayout({missions:records}).connections,[]);
 records[1].directParents=['r'];
 assert.deepEqual(canopyLayout({missions:records}).connections,[{from:'r',to:'t'}]);
});

test('both atlas and canopy delegate selection to canonical observatory navigation',async()=>{
 const {readFileSync}=await import('node:fs');
 const {fileURLToPath}=await import('node:url');
 const {dirname,join}=await import('node:path');
 const html=readFileSync(join(dirname(fileURLToPath(import.meta.url)),'living-universe-2.html'),'utf8');
 assert.match(html,/addEventListener\('aift:explore-node'/);
 assert.match(html,/select\(id\);[\s\S]*?scrollIntoView/);
 assert.match(html,/CustomEvent\('aift:explore-node',\{detail:\{id:m\.id\}\}\)/);
 assert.doesNotMatch(html,/textContent\.includes\((?:m\.)?id\)/);
});
