import test from 'node:test';
import assert from 'node:assert/strict';
import {receiptTimeline} from './living-universe-receipt-timeline.mjs';
const head='a'.repeat(40),repo='AIFreedomTrustFederation/AIFT-OS';
const good={repository:repo,headSha:head,workflow:'CI',runUrl:'https://github.com/'+repo+'/actions/runs/123',commitUrl:'https://github.com/'+repo+'/commit/'+head,observedAt:'2026-11-11T00:00:00Z',conclusion:'success'};
test('canonical receipt is linked but never independently verified',()=>{const x=receiptTimeline([good],{expectedRepo:repo,expectedHead:head,now:Date.parse('2026-11-12')});assert.equal(x.entries[0].state,'declared');assert.equal(x.entries[0].runUrl,good.runUrl);assert.equal(x.verified,0)});
test('malformed and mismatched receipts never become links',()=>{const x=receiptTimeline([{...good,runUrl:'javascript:alert(1)'},{...good,headSha:'b'.repeat(40)}],{expectedRepo:repo,expectedHead:head,now:Date.parse('2026-11-12')});assert.ok(x.entries.every(e=>e.state==='invalid'&&!e.runUrl));});
test('bounded display and absent data remain unverified',()=>{const x=receiptTimeline(Array(20).fill(good),{limit:3,now:Date.parse('2026-11-12')});assert.equal(x.entries.length,3);assert.equal(x.omitted,17);assert.equal(receiptTimeline(null).invalidCollection,true);assert.throws(()=>receiptTimeline([],{limit:0}));});
