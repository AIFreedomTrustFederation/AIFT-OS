/** Evidence-only spatial lineage navigator shared by Journey Atlas and Mission Control.
 * No network calls, persistence, synthetic achievements, or repository-as-fruit mapping.
 */
export const MAP_STAGES=Object.freeze(['root','trunk','branch','leaf','flower','fruit']);
const LIMIT=350;
export function spatialSummary(missions=[],selected=null){
 const records=Array.isArray(missions)?missions.filter(m=>m&&typeof m.id==='string'&&MAP_STAGES.includes(m.stage)).slice(0,LIMIT):[];
 const index=new Map(records.map(m=>[m.id,m])),active=index.get(selected)||null;
 const ancestors=new Set(),missing=new Set(),pending=active?[active.id]:[];
 while(pending.length&&ancestors.size<=LIMIT){
  const id=pending.pop();if(ancestors.has(id))continue;ancestors.add(id);
  const m=index.get(id);if(!m){missing.add(id);continue}
  for(const p of Array.isArray(m.directParents)?m.directParents:[])if(typeof p==='string')pending.push(p);
 }
 const stages=MAP_STAGES.map(stage=>({stage,count:records.filter(m=>m.stage===stage).length,
  lineage:records.filter(m=>m.stage===stage&&ancestors.has(m.id)).length}));
 const routes=active?spatialRoutes(records,active.id):[];
 const lineage=records.filter(m=>ancestors.has(m.id)).sort((a,b)=>MAP_STAGES.indexOf(a.stage)-MAP_STAGES.indexOf(b.stage));
 return {stages,active,lineage,routes,missing:[...missing],hasRoot:lineage.some(m=>m.stage==='root'),
  records, truncated:Array.isArray(missions)&&missions.length>LIMIT,
  claim:'Imported declarations only; no independent product or CI verification.'};
}

/** Deterministic, evidence-only directed paths. Never synthesize missing stages. */
export function spatialRoutes(records=[],targetId=null,{maxRoutes=12}={}){
 const rows=Array.isArray(records)?records.slice(0,LIMIT):[];
 const byId=new Map(rows.filter(r=>r&&typeof r.id==='string'&&MAP_STAGES.includes(r.stage)).map(r=>[r.id,r]));
 if(!byId.has(targetId))return [];
 const roots=[...byId.values()].filter(r=>r.stage==='root').sort((a,b)=>a.id.localeCompare(b.id));
 const children=new Map([...byId.keys()].map(id=>[id,[]]));
 for(const record of byId.values())for(const p of Array.isArray(record.directParents)?record.directParents:[]){
  if(children.has(p))children.get(p).push(record.id);
 }
 for(const ids of children.values())ids.sort();
 const routes=[];
 for(const root of roots){
  const queue=[[root.id]],seen=new Set([root.id]);let found=null;
  while(queue.length){
   const path=queue.shift(),last=path.at(-1);
   if(last===targetId){found=path;break}
   for(const child of children.get(last)||[])if(!seen.has(child)){seen.add(child);queue.push([...path,child])}
  }
  if(found){
   const steps=found.map(id=>byId.get(id));
   const anomalies=[];
   for(let i=1;i<steps.length;i++){
    const from=MAP_STAGES.indexOf(steps[i-1].stage),to=MAP_STAGES.indexOf(steps[i].stage);
    if(to!==from+1)anomalies.push({from:steps[i-1].id,to:steps[i].id,kind:to<=from?'nonprogressive':'stage-gap'});
   }
   routes.push({rootId:root.id,ids:found,steps,anomalies});
   if(routes.length>=Math.max(1,Math.min(12,Number.isInteger(maxRoutes)?maxRoutes:12)))break;
  }
 }
 return routes;
}

export function mountSpatialMinimap(host,{missions=[],selected=null,onStage=()=>{},onSelect=()=>{}}={}){
 if(!host)return;const doc=host.ownerDocument,model=spatialSummary(missions,selected);
 const el=(tag,value,cls)=>{const n=doc.createElement(tag);if(value!==undefined)n.textContent=value;if(cls)n.className=cls;return n};
 host.replaceChildren();host.setAttribute('aria-label','Source-to-outcome spatial minimap');
 host.append(el('h3','Federation lineage minimap'));
 host.append(el('p','Explore six biological stages, inspect the selected record’s ancestry, and return to its source roots.','muted'));
 const nav=el('div',undefined,'aift-minimap-track');nav.setAttribute('role','group');nav.setAttribute('aria-label','Biological stage navigation');
 for(const s of model.stages){
  const b=el('button',s.stage+' · '+s.count,'aift-minimap-stage');b.type='button';b.dataset.stage=s.stage;
  b.dataset.lineage=String(s.lineage>0);b.setAttribute('aria-label',s.stage+': '+s.count+' imported records; '+s.lineage+' in selected lineage');
  b.title=s.lineage+' selected-lineage records';b.addEventListener('click',()=>onStage(s.stage));nav.append(b);
 }
 host.append(nav);
 const summary=el('p',model.active?'Selected '+model.active.stage+': '+(model.active.label||model.active.id):'No record selected. Import evidence to explore.','muted');
 summary.setAttribute('aria-live','polite');host.append(summary);
 if(model.active){
  const actions=el('div',undefined,'aift-minimap-actions');
  const focus=el('button','Focus selected record');focus.type='button';focus.addEventListener('click',()=>onSelect(model.active.id));actions.append(focus);
  const roots=model.lineage.filter(m=>m.stage==='root');
  if(roots.length){const root=el('button','Trace to root ('+roots.length+')');root.type='button';root.addEventListener('click',()=>onSelect(roots[0].id));actions.append(root)}
  host.append(actions);
  const trail=el('ol',undefined,'aift-minimap-lineage');trail.setAttribute('aria-label','Selected source-to-outcome lineage');
  for(const record of model.lineage){
   const li=el('li');const b=el('button',record.stage+' · '+(record.label||record.id));b.type='button';
   b.dataset.recordId=record.id;b.setAttribute('aria-current',record.id===model.active.id?'step':'false');
   b.addEventListener('click',()=>onSelect(record.id));li.append(b);trail.append(li);
  }
  host.append(trail);
  const routes=el('section',undefined,'aift-minimap-routes');routes.setAttribute('aria-label','Declared root-to-record routes');
  routes.append(el('h4','Directed provenance journeys'));
  if(!model.routes.length)routes.append(el('p','No complete declared path from a repository root. The lineage remains unverified.','notice'));
  for(const [i,route] of model.routes.entries()){
   const details=el('details',undefined,'aift-minimap-route');const head=el('summary','Route '+(i+1)+' · '+(route.steps[0].label||route.rootId)+' · '+route.steps.length+' declared steps');
   details.append(head);
   const path=el('ol',undefined,'aift-minimap-route-steps');
   for(const step of route.steps){
    const li=el('li'),button=el('button',step.stage+' · '+(step.label||step.id));button.type='button';
    button.dataset.recordId=step.id;button.addEventListener('click',()=>onSelect(step.id));li.append(button);path.append(li);
   }
   details.append(path);
   if(route.anomalies.length)details.append(el('p',route.anomalies.length+' direct stage transition anomalies; no missing intermediate stages inferred.','notice'));
   routes.append(details);
  }
  host.append(routes);
 }
 if(model.active&&!model.hasRoot)host.append(el('p','Evidence gap: no declared repository-root lineage.','notice'));
 if(model.missing.length)host.append(el('p','Evidence gap: '+model.missing.length+' missing parent references.','notice'));
 if(model.truncated)host.append(el('p','Evidence limit: showing the first '+LIMIT+' imported records.','notice'));
 host.append(el('p',model.claim,'muted'));
}
