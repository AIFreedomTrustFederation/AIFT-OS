import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
const html=readFileSync(new URL('./living-universe-2.html',import.meta.url),'utf8');
test('biological ontology has six stages and repositories are not fruits by default',()=>{
 assert.match(html,/\['root','trunk','branch','leaf','flower','fruit'\]/);
 assert.match(html,/repositories are roots/i);
 assert.match(html,/Evidence gap: this declared outcome has no incoming source lineage/);
});
test('local-first bounded import and untrusted text handling',()=>{
 assert.match(html,/file.size>2\*1024\*1024/);
 assert.match(html,/list.length>350\|\|links.length>1200/);
 assert.match(html,/textContent=/);
 assert.doesNotMatch(html,/innerHTML\s*=/);
 assert.doesNotMatch(html,/localStorage|sessionStorage|sendBeacon|fetch\(/);
});
test('accessible spatial and directory views with reduced-motion fallback',()=>{
 assert.match(html,/aria-label="Evidence constellation/);
 assert.match(html,/id="directory"/);
 assert.match(html,/prefers-reduced-motion:reduce/);
 assert.match(html,/e.key==='Enter'\|\|e.key===' '/);
});
test('explicit source URL protocol allowlist and no unverified CI claims',()=>{
 assert.match(html,/\['https:','http:'\]/);
 assert.match(html,/No achievements, progress percentages or product readiness/);
});