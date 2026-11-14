import {test} from 'node:test';import assert from 'node:assert/strict';import {readFileSync} from 'node:fs';
const page=readFileSync(new URL('./living-universe-lineage-theatre.html',import.meta.url),'utf8');
const observatory=readFileSync(new URL('./living-universe-2.html',import.meta.url),'utf8');
test('comparison is reachable from existing observatory',()=>{assert.match(observatory,/href="\.\/living-universe-lineage-theatre\.html"/);assert.match(page,/living-universe-2\.html/);});
test('comparison preserves six-stage biology and bounded local imports',()=>{for(const s of ['root','trunk','branch','leaf','flower','fruit'])assert.match(page,new RegExp("'"+s+"'"));assert.match(page,/data\.nodes\.length>350/);assert.match(page,/data\.edges\.length>1200/);assert.match(page,/f\.size>3000000/);assert.match(page,/canopyRoutes\(layout,m\.id/);});
test('comparison does not invent verification or network access',()=>{assert.match(page,/Independent verification unavailable/);assert.match(page,/shared contributing repository roots/);assert.doesNotMatch(page,/\bfetch\s*\(|XMLHttpRequest|sendBeacon|localStorage/);assert.match(page,/textContent=text/);});
test('responsive accessible controls and reduced motion',()=>{assert.match(page,/aria-live="polite"/);assert.match(page,/max-width:720px/);assert.match(page,/prefers-reduced-motion:reduce/);assert.match(page,/focus-visible/);});
