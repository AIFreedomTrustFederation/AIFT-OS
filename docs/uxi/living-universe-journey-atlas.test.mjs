import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
const html=readFileSync(new URL('./living-universe-journey-atlas.html',import.meta.url),'utf8');
test('six biological stages, accessible navigation, and spatial minimap',()=>{
 for(const stage of ['root','trunk','branch','leaf','flower','fruit'])assert.match(html,new RegExp('\\b'+stage+'\\b'));
 for(const id of ['graph','directory','minimap','inspector','quests','back','forward','file'])assert.match(html,new RegExp('id="'+id+'"'));
 assert.match(html,/aria-live="polite"/);
 assert.match(html,/prefers-reduced-motion:reduce/);
});
test('bounded local evidence import and no remote evidence acquisition',()=>{
 assert.match(html,/3\*1024\*1024/);
 assert.match(html,/nodes\.length>350/);
 assert.match(html,/edges\.length>1200/);
 assert.doesNotMatch(html,/\bfetch\s*\(/);
 assert.doesNotMatch(html,/\blocalStorage\b/);
});
test('canonical directed graph modules, no repository-as-fruit shortcut',()=>{
 assert.match(html,/from '\.\/living-universe-skill-atlas\.mjs'/);
 assert.match(html,/from '\.\/living-universe-canopy\.mjs'/);
 assert.match(html,/canopyLineage\(layout,m\.id\)/);
 assert.match(html,/evidenceQuestPlan\(atlas/);
 assert.match(html,/Source repositories are roots, never fruits/);
 assert.match(html,/Not independently verified/);
});
test('safe external navigation and keyboard interaction',()=>{
 assert.match(html,/noopener noreferrer/);
 assert.match(html,/e\.key==='Enter'/);
 assert.match(html,/e\.key===' '/);
 assert.match(html,/role:'button'/);
});
