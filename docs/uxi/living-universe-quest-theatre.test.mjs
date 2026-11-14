import test from 'node:test';import assert from 'node:assert/strict';
import {questTheatreModel} from './living-universe-quest-theatre.mjs';
test('empty import never manufactures quests',()=>{const m=questTheatreModel(null);assert.deepEqual(m.quests,[]);assert.match(m.claim,/Import evidence/)});
test('empty atlas never awards achievements',()=>{const m=questTheatreModel({missions:[]});assert.deepEqual(m.quests,[]);assert.ok(!('xp' in m))});
test('unknown selected record is not promoted to verified',()=>{const m=questTheatreModel({missions:[]},{selected:'missing'});assert.equal(m.selected,null);assert.match(m.claim,/No XP/)});
