import test from 'node:test';
import assert from 'node:assert/strict';
import {planJourneys,lineageFor} from './journey-planner.mjs';
test('only directed root paths are recognized',()=>{
 const nodes=[{id:'repo',stage:'root'},{id:'system',stage:'branch'},{id:'app',stage:'fruit',url:'https://example.org',productEvidence:{verified:true,sourceUrl:'https://example.org/evidence'}}];
 const edges=[{from:'repo',to:'system'},{from:'system',to:'app'}];
 const p=planJourneys(nodes,edges);
 assert.equal(p.routes[0].classification,'declared-evidence-route');
 assert.deepEqual(p.routes[0].path,['repo','system','app']);
 assert.deepEqual(lineageFor('app',nodes,edges),['repo','system','app']);
 assert.equal(planJourneys(nodes,[{from:'app',to:'repo'}]).routes[0].classification,'evidence-gap');
});
test('repository membership and product labels never grant verification',()=>{
 const p=planJourneys([{id:'repo',stage:'root'},{id:'concept',stage:'fruit',url:'https://example.org'}],[{from:'repo',to:'concept'}]);
 assert.equal(p.routes[0].classification,'evidence-gap');
 assert.match(p.routes[0].gap,/verified product evidence/);
});
test('cycles terminate and do not create paths',()=>{
 const p=planJourneys([{id:'a',stage:'branch'},{id:'b',stage:'fruit'}],[{from:'a',to:'b'},{from:'b',to:'a'}]);
 assert.equal(p.routes[0].hasDirectedRootPath,false);
});
test('no synthetic quests or XP when no evidence nodes exist',()=>{
 assert.deepEqual(planJourneys([],[]).tasks,[]);
});
test('tasks are review opportunities, not earned achievements',()=>{
 const p=planJourneys([{id:'r',stage:'root'}],[]);
 assert.equal(p.tasks[0].readiness,'available');
 assert.match(p.tasks[0].note,/not a verified achievement/);
});
