import test from 'node:test';
import assert from 'node:assert/strict';
import {createNavigation,stageFor,BIOLOGICAL_STAGES} from './living-universe-navigation.mjs';
test('six biological stages remain ordered, and a repository is not a fruit by inference',()=>{
 assert.deepEqual(BIOLOGICAL_STAGES,['root','trunk','branch','leaf','flower','fruit']);
 assert.equal(stageFor({stage:'root'}),'root');assert.equal(stageFor({stage:'repository'}),null);
});
test('navigation is reversible and branching after back discards forward history',()=>{
 const nav=createNavigation([{id:'r'},{id:'t'},{id:'f'}],[{from:'r',to:'t'},{from:'t',to:'f'}]);
 nav.inspect('r');nav.inspect('t');nav.inspect('f');
 assert.equal(nav.back().id,'t');assert.equal(nav.snapshot().forward,true);
 nav.inspect('r');assert.equal(nav.snapshot().forward,false);
 assert.equal(nav.snapshot().selected,'r');
});
test('navigation refuses undeclared nodes and deduplicates directed neighbors',()=>{
 const nav=createNavigation([{id:'a'},{id:'b'}],[{from:'a',to:'b'},{from:'a',to:'b'},{from:'a',to:'missing'}]);
 assert.deepEqual(nav.neighbors('a'),['b']);assert.deepEqual(nav.neighbors('b'),[]);
 assert.equal(nav.inspect('missing'),null);assert.equal(nav.snapshot().selected,null);
});
test('history bounded and clearable without persistent state',()=>{
 const nav=createNavigation(Array.from({length:110},(_,i)=>({id:String(i)})));
 for(let i=0;i<110;i++)nav.inspect(String(i));
 assert.equal(nav.snapshot().visited.length,100);nav.reset();assert.equal(nav.snapshot().selected,null);
});
