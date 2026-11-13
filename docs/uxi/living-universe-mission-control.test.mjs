import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
const html=readFileSync(new URL('./living-universe-mission-control.html',import.meta.url),'utf8');
test('integrated mission control exposes graph, lineage, routes, quests and accessible directory',()=>{
 for(const id of ['graph','directory','inspector','paths','quests','back','forward','file','search','stage','clear','orbit'])assert.match(html,new RegExp('id="'+id+'"'));
 for(const stage of ['root','trunk','branch','leaf','flower','fruit'])assert.match(html,new RegExp('value="'+stage+'"'));
 for(const name of ['canopyLayout','canopyLineage','canopyRoutes','journeyDossier','evidenceQuestPlan'])assert.match(html,new RegExp(name+'\\('));
 assert.match(html,/aria-live="polite"/);assert.match(html,/prefers-reduced-motion:reduce/);
});
test('canonical graph contracts and local evidence boundaries',()=>{
 assert.match(html,/filterAtlas\(atlas,\{search:/);
 assert.match(html,/layout\.connections/);
 assert.match(html,/new Set\(l\.nodes\.map\(n=>n\.id\)\)/);
 assert.match(html,/3\*1024\*1024/);assert.match(html,/nodes\.length>350/);assert.match(html,/edges\.length>1200/);
 assert.doesNotMatch(html,/\bfetch\s*\(/);assert.doesNotMatch(html,/\blocalStorage\b/);
});
test('source repositories remain roots and fruits are never automatically verified',()=>{
 assert.match(html,/Repository roots nourish the trunk/);
 assert.match(html,/not independently verified/i);
 assert.match(html,/noopener noreferrer/);
 assert.match(html,/u\.protocol==='https:'/);
 assert.match(html,/No remote fetch, telemetry, storage, fabricated records or automatic achievements/);
});
test('navigation preserves bounded history and keyboard access',()=>{
 assert.match(html,/history\.length>25/);
 assert.match(html,/cursor--/);assert.match(html,/cursor\+\+/);
 assert.match(html,/e\.key==='Enter'/);assert.match(html,/e\.key===' '/);
});
