import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
const html=readFileSync(new URL('./living-universe-2.html',import.meta.url),'utf8');
test('immersive panorama retains six biological stages and accessible alternative',()=>{
 for(const stage of ['root','trunk','branch','leaf','flower','fruit'])assert.match(html,new RegExp('<option value="'+stage+'">'));
 for(const id of ['living-panorama','panorama-sky','panorama-directory','panorama-detail','panorama-status'])assert.ok(html.includes('id="'+id+'"'));
 assert.match(html,/aria-live="polite"/);
 assert.match(html,/prefers-reduced-motion:reduce/);
});
test('panorama only renders locally imported evidence and canonical selection',()=>{
 assert.match(html,/const records=\(current\?\.missions\|\|\[\]\)/);
 assert.match(html,/canopyLayout\(current\)/);
 assert.match(html,/aift:explore-node/);
 assert.match(html,/No evidence imported\. Repository roots and outcomes are not fabricated/);
 assert.match(html,/\.slice\(0,180\)/);
 assert.match(html,/renderPanorama\(\);/);
});
test('panorama does not add independent verification, telemetry, or remote import',()=>{
 const section=html.slice(html.indexOf('\/\* Panorama uses'),html.indexOf('const render=()=>{'));
 assert.ok(!/fetch\s*\(|XMLHttpRequest|localStorage|sendBeacon/.test(section));
 assert.match(section,/not independently verified/);
});
