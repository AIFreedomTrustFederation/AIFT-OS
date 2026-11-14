import {readFileSync} from 'node:fs';
import {test} from 'node:test';
import assert from 'node:assert/strict';
const html=readFileSync(new URL('./living-universe-2.html',import.meta.url),'utf8');
test('spatial constellation integrates with journey render',()=>{
 assert.match(html,/function renderConstellation\(\)/);
 assert.match(html,/renderJourney\(\);\s*renderConstellation\(\)/);
 assert.match(html,/constellationSearch\.addEventListener\('input',renderConstellation\)/);
 assert.match(html,/constellationZoom\.addEventListener\('input',renderConstellation\)/);
});
test('visual edges are exclusively declared canopy connections',()=>{
 assert.match(html,/for\(const edge of layout\.connections\)/);
 assert.match(html,/canopyReadiness\(layout,n\.id\)/);
 assert.match(html,/independent verification unavailable/);
});
test('keyboard, text alternatives and motion safety',()=>{
 assert.match(html,/role:'button',tabindex:0/);
 assert.match(html,/e\.key==='Enter'\|\|e\.key===' '/);
 assert.match(html,/constellationList\.append\(b\)/);
 assert.match(html,/prefers-reduced-motion:reduce/);
});
test('no repository-to-fruit promotion or network/telemetry',()=>{
 assert.match(html,/Repositories are roots, never fruits/);
 assert.doesNotMatch(html.slice(html.indexOf('Federation constellation: declared graph only'),html.indexOf('const render=()=>{')),/\bfetch\s*\(|sendBeacon|localStorage|sessionStorage/);
});
