import test from 'node:test';import assert from 'node:assert/strict';import {analyzeAtlas,filterAtlas} from './living-universe-skill-atlas.mjs';
const nodes=[{id:'repo',stage:'root',repository_url:'https://github.com/AIFreedomTrustFederation/AIFT-Genesis'},{id:'trunk',stage:'trunk'},{id:'branch',stage:'branch'},{id:'fruit',stage:'fruit',url:'https://example.org/app',evidence:'claimed'}];
const edges=[{from:'repo',to:'trunk'},{from:'trunk',to:'branch'},{from:'branch',to:'fruit'}];
test('directed ancestry gives roots to fruits, never turns repositories into fruit',()=>{const a=analyzeAtlas(nodes,edges);assert.equal(a.summary.roots,1);assert.equal(a.summary.declaredOutcomes,1);assert.deepEqual(a.missions[3].roots,['repo']);assert.equal(a.summary.verifiedAchievements,0)});
test('claims do not become achievements',()=>{const a=analyzeAtlas(nodes,edges);assert.equal(a.missions[3].status,'evidence-declared')});
test('unlinked fruit surfaces gaps',()=>{const a=analyzeAtlas(nodes,[]);assert.match(a.missions[3].issues.join(' '),/root ancestry/);assert.equal(a.missions[3].status,'evidence-gap')});
test('cyclic lineage bounded and rejects unknown edges',()=>{assert.equal(analyzeAtlas(nodes,[...edges,{from:'fruit',to:'trunk'}]).missions.length,4);assert.throws(()=>analyzeAtlas(nodes,[{from:'missing',to:'fruit'}]),/Undeclared/)});
test('filters use real declared evidence only',()=>{const a=analyzeAtlas(nodes,edges);assert.equal(filterAtlas(a,{stage:'root'}).length,1);assert.equal(filterAtlas(a,{search:'fruit'}).length,1)});
test('unsafe URL is not a product destination',()=>{const a=analyzeAtlas([{id:'x',stage:'fruit',url:'javascript:alert(1)'}],[]);assert.equal(a.missions[0].destination,null);assert.ok(a.missions[0].issues.length)});

import {evidenceQuestPlan} from './living-universe-skill-atlas.mjs';
test('quests prioritize outcomes without fabricated achievements',()=>{const p=evidenceQuestPlan(analyzeAtlas(nodes,edges));assert.equal(p.quests[0].id,'fruit');assert.equal(p.verified,0);assert.equal(p.quests[0].receipt,'not-verified')});
test('quests reveal missing provenance',()=>{const p=evidenceQuestPlan(analyzeAtlas(nodes,[]));assert.ok(p.quests.find(x=>x.id==='fruit').checks.some(x=>/root/.test(x)))});
test('quest limit bounded',()=>{assert.equal(evidenceQuestPlan(analyzeAtlas(nodes,edges),{limit:2}).quests.length,2);assert.throws(()=>evidenceQuestPlan(analyzeAtlas(nodes,edges),{limit:0}),/Invalid quest/)});

import {evidenceReadinessGates} from './living-universe-skill-atlas.mjs';
test('fruit with declared source and destination still has independent pending gates',()=>{const m=analyzeAtlas(nodes,edges).missions.find(x=>x.id==='fruit');const r=evidenceReadinessGates(m);assert.equal(r.verified,false);assert.equal(r.ready,false);assert.equal(r.blocking,0);assert.equal(r.pending,2);assert.deepEqual(r.gates.map(g=>g.id),['provenance','destination','evidence','accessibility']);assert.equal(r.gates[1].state,'declared')});
test('unlinked fruit blocks source provenance and outcome verification',()=>{const m=analyzeAtlas([{id:'orphan',stage:'fruit'}],[]).missions[0];const r=evidenceReadinessGates(m);assert.equal(r.gates.find(g=>g.id==='provenance').state,'blocked');assert.equal(r.gates.find(g=>g.id==='destination').state,'blocked');assert.ok(r.blocking>=2);assert.equal(r.verified,false)});
test('root is never a product fruit or verified achievement',()=>{const q=evidenceQuestPlan(analyzeAtlas(nodes,edges)).quests.find(x=>x.id==='repo');assert.equal(q.stage,'root');assert.equal(q.destination!==null,true);assert.equal(q.receipt,'not-verified');assert.equal(q.readiness.gates.some(g=>g.id==='destination'),false)});
test('invalid readiness input is rejected',()=>{assert.throws(()=>evidenceReadinessGates({id:'bad',roots:null,issues:[]}),/Invalid atlas mission/)});

test('GitHub source repositories cannot masquerade as finished fruits',()=>{
 const a=analyzeAtlas([{id:'repo',stage:'root',repository_url:'https://github.com/AIFreedomTrustFederation/AIFT-OS'},{id:'fake-fruit',stage:'fruit',url:'https://github.com/AIFreedomTrustFederation/AIFT-OS'}],[{from:'repo',to:'fake-fruit'}]);
 const fruit=a.missions.find(m=>m.id==='fake-fruit');assert.equal(fruit.destination,null);assert.equal(fruit.status,'evidence-gap');assert.match(fruit.issues.join(' '),/destination/);
 assert.equal(evidenceQuestPlan(a).quests[0].readiness.gates.find(g=>g.id==='destination').state,'blocked');
});
test('GitHub pull requests and Actions URLs are not finished products',()=>{
 for(const url of ['https://github.com/AIFreedomTrustFederation/AIFT-OS/pull/65','https://github.com/AIFreedomTrustFederation/AIFT-OS/actions/runs/123','https://api.github.com/repos/x/y']){
  const a=analyzeAtlas([{id:'fruit',stage:'fruit',url}],[]);assert.equal(a.missions[0].destination,null);
 }
});
test('declared application destination can be a fruit with source ancestry but not verified',()=>{
 const a=analyzeAtlas([{id:'source',stage:'root',repository_url:'https://github.com/AIFreedomTrustFederation/AIFT-OS'},{id:'outcome',stage:'fruit',product_url:'https://example.org/app'}],[{from:'source',to:'outcome'}]);
 const outcome=a.missions.find(m=>m.id==='outcome');assert.equal(outcome.destination,'https://example.org/app');assert.deepEqual(outcome.roots,['source']);assert.equal(a.summary.verifiedAchievements,0);
});

test('outcome dossier provides reversible provenance and accessible evidence gaps',async()=>{
 const {readFileSync}=await import('node:fs');
 const html=readFileSync(new URL('./living-universe-2.html',import.meta.url),'utf8');
 assert.match(html,/fruit-lineage-dossier/);
 assert.match(html,/Explore outcome → contributing roots/);
 assert.match(html,/Return to outcome in Tree/);
 assert.match(html,/No qualified product destination/);
 assert.match(html,/aift:explore-node/);
 assert.match(html,/prefers-reduced-motion:reduce/);
});

test('outcome dossier renders complete directed paths with intermediate systems and stage anomalies',async()=>{
 const {readFileSync}=await import('node:fs');
 const html=readFileSync(new URL('./living-universe-2.html',import.meta.url),'utf8');
 assert.match(html,/canopyRoutes\(graph,q\.id,\{maxRoutes:12\}\)/);
 assert.match(html,/for\(const id of route\.ids\)/);
 assert.match(html,/No missing links are inferred/i);
 assert.match(html,/route\.stageGaps/);
 assert.match(html,/aift:explore-node/);
});
