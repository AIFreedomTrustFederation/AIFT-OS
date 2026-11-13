/** Shared, evidence-scoped spatial minimap for Atlas and Mission Control.
 * Repository roots are sources, never product fruits. No network or persistence.
 */
export const MAP_STAGES=Object.freeze(['root','trunk','branch','leaf','flower','fruit']);
export function spatialSummary(missions=[],selected=null){
 const records=Array.isArray(missions)?missions.filter(m=>m&&typeof m.id==='string'&&MAP_STAGES.includes(m.stage)).slice(0,350):[];
 const index=new Map(records.map(m=>[m.id,m]));
 const active=index.get(selected)||null,ancestors=new Set(),missing=new Set(),pending=active?[active.id]:[];
 while(pending.length&&ancestors.size<350){
  const id=pending.pop();if(ancestors.has(id))continue;ancestors.add(id);
  const m=index.get(id);if(!m){missing.add(id);continue}
  for(const p of Array.isArray(m.directParents)?m.directParents:[])if(typeof p==='string')pending.push(p);
 }
 const stages=MAP_STAGES.map(stage=>({stage,count:records.filter(m=>m.stage===stage).length,
  lineage:records.filter(m=>m.stage===stage&&ancestors.has(m.id)).length}));
 return {stages,active,missing:[...missing],hasRoot:records.some(m=>m.stage==='root'&&ancestors.has(m.id)),
  claim:'Imported declarations only; no independent product or CI verification.'};
}
export function mountSpatialMinimap(host,{missions=[],selected=null,onStage=()=>{},onSelect=()=>{}}={}){
 if(!host)return;const doc=host.ownerDocument,model=spatialSummary(missions,selected);
 const el=(tag,text,cls)=>{const n=doc.createElement(tag);if(text!==undefined)n.textContent=text;if(cls)n.className=cls;return n};
 host.replaceChildren();host.setAttribute('aria-label','Source-to-outcome spatial minimap');
 const heading=el('h3','Federation lineage minimap');host.append(heading);
 const help=el('p','Six evidence stages. Choose a stage to filter the directory, or follow the selected record.','muted');host.append(help);
 const nav=el('div',undefined,'aift-minimap-track');nav.setAttribute('role','group');nav.setAttribute('aria-label','Biological stage navigation');
 for(const s of model.stages){
  const b=el('button',s.stage+' · '+s.count,'aift-minimap-stage');b.type='button';b.dataset.stage=s.stage;
  b.dataset.lineage=String(s.lineage>0);b.setAttribute('aria-label',s.stage+': '+s.count+' imported records; '+s.lineage+' in selected lineage');
  b.title=s.lineage+' selected-lineage records';b.addEventListener('click',()=>onStage(s.stage));nav.append(b);
 }
 host.append(nav);
 const summary=el('p',model.active?'Selected '+model.active.stage+': '+(model.active.label||model.active.id):'No record selected. Import evidence to explore.','muted');summary.setAttribute('aria-live','polite');host.append(summary);
 if(model.active){const focus=el('button','Focus selected record');focus.type='button';focus.addEventListener('click',()=>onSelect(model.active.id));host.append(focus)}
 if(model.active&&!model.hasRoot)host.append(el('p','Evidence gap: no declared repository-root lineage.','notice'));
 if(model.missing.length)host.append(el('p','Evidence gap: '+model.missing.length+' missing parent references.','notice'));
 host.append(el('p',model.claim,'muted'));
}
