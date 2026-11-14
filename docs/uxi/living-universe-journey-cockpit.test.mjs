import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
const html=readFileSync(new URL('./living-universe-2.html',import.meta.url),'utf8');
test('Journey Cockpit is integrated with imported evidence and Tree',()=>{
 for(const marker of ['id="journey-cockpit"','id="journey-query"','id="journey-target"','id="journey-routes"','id="journey-stages"','id="journey-detail"','function renderJourney()','renderJourney();','canopyRoutes(layout,mission.id','canopyReadiness(layout,mission.id','aift:explore-node'])assert.ok(html.includes(marker),marker);
});
test('Journey Cockpit does not assert unverified achievements',()=>{
 assert.match(html,/independent verification unavailable/);
 assert.match(html,/Missing relationships are not synthesized/);
 assert.match(html,/Open declared product \(not independently verified\)/);
 assert.match(html,/prefers-reduced-motion:reduce/);
 assert.match(html,/aria-live="polite"/);
});
test('Journey Cockpit limits imported record options and routes',()=>{
 assert.match(html,/slice\(0,250\)/);
 assert.match(html,/maxRoutes:12/);
 assert.match(html,/new URL\(selected.destination\)/);
 assert.match(html,/u.protocol==='https:'/);
});

test('Journey Cockpit exposes reversible breadcrumb, stage minimap and traversal controls',()=>{
 for(const marker of ['id="journey-trail"','id="journey-overview"','id="journey-actions"','function renderJourneyNavigation(path)','renderJourneyNavigation(path);','Previous stage','Next stage','Return to source root','Return to selected outcome','aria-current','Evidence gap'])assert.ok(html.includes(marker),marker);
});
test('Journey semantic minimap is accessible and honors reduced motion',()=>{
 for(const marker of ['aria-label="Reversible lineage breadcrumbs"','aria-label="Biological stage minimap"','aria-label="Journey traversal controls"','@media(prefers-reduced-motion:reduce){.journey-trail'])assert.ok(html.includes(marker),marker);
});
