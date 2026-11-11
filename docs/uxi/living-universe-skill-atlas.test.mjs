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
