/** Shared evidence quest theatre. Read-only: imported declarations are not achievements. */
import {evidenceQuestPlan} from './living-universe-skill-atlas.mjs';
import {spatialSummary} from './living-universe-spatial-minimap.mjs';
const STAGES=['root','trunk','branch','leaf','flower','fruit'];
export function questTheatreModel(atlas,{selected=null,stage='',limit=24}={}){
 if(!atlas||!Array.isArray(atlas.missions))return {quests:[],stages:[],claim:'Import evidence to reveal review work; no achievements awarded.'};
 const records=atlas.missions.slice(0,350);
 const summary=spatialSummary(records,selected);
 const planned=evidenceQuestPlan(atlas,{limit:Math.max(1,Math.min(24,limit))}).quests;
 const byId=new Map(records.map(r=>[r.id,r]));
 const quests=planned.filter(q=>byId.has(q.id)&&(!stage||q.stage===stage)).map(q=>{
  const source=byId.get(q.id),trace=spatialSummary(records,q.id);
  return {id:q.id,label:q.label,stage:q.stage,receipt:q.receipt,checks:[...q.checks],
   roots:trace.lineage.filter(r=>r.stage==='root').map(r=>r.id),
   gaps:trace.missing.length,selected:q.id===selected,
   destinationDeclared:source.stage==='fruit'&&Boolean(source.destination),
   verified:false};
 });
 return {quests,stages:STAGES.map(s=>({stage:s,count:quests.filter(q=>q.stage===s).length})),
  truncated:atlas.missions.length>350,selected:summary.active?.id||null,
  claim:'Review quests are derived from imported evidence. No XP, completion, CI pass or product verification is inferred.'};
}
export function mountQuestTheatre(host,{atlas=null,selected=null,stage='',onSelect=()=>{}}={}){
 if(!host)return;const doc=host.ownerDocument,model=questTheatreModel(atlas,{selected,stage});
 const el=(tag,txt,cls)=>{const n=doc.createElement(tag);if(txt!==undefined)n.textContent=txt;if(cls)n.className=cls;return n};
 host.replaceChildren();host.classList.add('aift-quest-theatre');host.setAttribute('aria-label','Evidence-backed quest theatre');
 host.append(el('h3','Evidence expeditions'));
 host.append(el('p',model.claim,'muted'));
 if(!atlas){host.append(el('p','Import an evidence packet to reveal review quests.'));return}
 const ribbon=el('div',undefined,'aift-quest-spectrum');ribbon.setAttribute('aria-label','Quests by biological stage');
 for(const s of model.stages){const item=el('span',s.stage+' · '+s.count);item.dataset.stage=s.stage;ribbon.append(item)}host.append(ribbon);
 if(!model.quests.length){host.append(el('p','No evidence-derived quests for this view.'));return}
 const list=el('ol',undefined,'aift-quest-grid');
 for(const q of model.quests.slice(0,12)){
  const li=el('li',undefined,'aift-quest-card');li.dataset.stage=q.stage;
  const btn=el('button',q.label);btn.type='button';btn.setAttribute('aria-pressed',String(q.selected));btn.addEventListener('click',()=>onSelect(q.id));li.append(btn);
  li.append(el('p',q.stage+' · '+q.receipt,'muted'));
  const checks=el('ul');for(const check of q.checks.slice(0,4))checks.append(el('li',check));li.append(checks);
  li.append(el('p',q.roots.length?q.roots.length+' declared repository-root ancestor(s)':'Evidence gap: no declared repository-root ancestry','muted'));
  if(q.gaps)li.append(el('p',q.gaps+' unresolved parent reference(s)','notice'));
  if(q.stage==='fruit')li.append(el('p',q.destinationDeclared?'Destination declared; independent verification outstanding':'Evidence gap: no product destination','notice'));
  list.append(li);
 }
 host.append(list);
 if(model.quests.length>12)host.append(el('p','Showing 12 of '+model.quests.length+' review quests. Narrow the stage filter to explore more.','muted'));
 if(model.truncated)host.append(el('p','Imported record display limited to 350.','notice'));
}
