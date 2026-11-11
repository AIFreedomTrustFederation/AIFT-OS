// Living Universe 2.0: evidence-backed skill atlas (pure, no network or storage).
export const STAGES=Object.freeze(['root','trunk','branch','leaf','flower','fruit']);
export function analyzeAtlas(nodes,edges){
 if(!Array.isArray(nodes)||!Array.isArray(edges))throw Error('Expected nodes and directed edges');
 const index=new Map();
 for(const n of nodes){if(!n||typeof n.id!=='string'||!n.id||index.has(n.id))throw Error('Invalid or duplicate node');index.set(n.id,n)}
 const inbound=new Map(nodes.map(n=>[n.id,[]]));
 for(const e of edges){const a=e.from??e.source,b=e.to??e.target;if(!index.has(a)||!index.has(b))throw Error('Undeclared edge');inbound.get(b).push(a)}
 const stage=n=>String(n._stage||n.stage||n.kind||n.type||'').toLowerCase().replace(/s$/,'');
 const isRoot=n=>stage(n)==='root' && (n.repository_url||n.source_url||n.url||n.href);
 const sourceUrl=n=>[n.url,n.href,n.source_url,n.repository_url].find(u=>{try{return new URL(u).protocol==='https:'}catch{return false}})||null;
 const ancestors=id=>{const visited=new Set(),stack=[id];while(stack.length){const next=stack.pop();if(visited.has(next))continue;visited.add(next);stack.push(...(inbound.get(next)||[]))}visited.delete(id);return [...visited]};
 const missions=nodes.map(n=>{
 const s=stage(n),parents=ancestors(n.id),roots=parents.filter(id=>isRoot(index.get(id))),destination=sourceUrl(n);
 const issues=[];
 if(!STAGES.includes(s))issues.push('Unrecognized biological stage');
 if(s==='fruit'&&!destination)issues.push('No declared user-facing destination');
 if(s==='fruit'&&roots.length===0)issues.push('No directed repository-root ancestry');
 if(s==='fruit'&&inbound.get(n.id).length===0)issues.push('No declared incoming contribution');
 const evidence=n.evidence||n.verification||n.ci;
 // Never infer a verified achievement from an arbitrary evidence string or URL.
 return {id:n.id,label:String(n.label||n.name||n.title||n.id),stage:s,roots,destination,issues,
  status:issues.length?'evidence-gap':evidence?'evidence-declared':'unverified',ancestors:parents,directParents:[...new Set(inbound.get(n.id))]};
 });
 return {missions,summary:{nodes:nodes.length,roots:nodes.filter(isRoot).length,declaredOutcomes:missions.filter(m=>m.stage==='fruit').length,gaps:missions.filter(m=>m.issues.length).length,verifiedAchievements:0}};
}
export function filterAtlas(atlas,{search='',stage='',status=''}={}){
 const q=search.toLocaleLowerCase().trim();
 return atlas.missions.filter(m=>(!stage||m.stage===stage)&&(!status||m.status===status)&&(!q||[m.label,m.id,m.stage,...m.roots].join(' ').toLocaleLowerCase().includes(q)));
}

export function evidenceQuestPlan(atlas,{limit=24}={}){
 if(!atlas||!Array.isArray(atlas.missions))throw Error('Atlas required');
 if(!Number.isInteger(limit)||limit<1||limit>100)throw Error('Invalid quest limit');
 const priority={fruit:0,flower:1,leaf:2,branch:3,trunk:4,root:5};
 const quests=atlas.missions.map(m=>{
 const checks=[...(m.issues||[])];
 if(m.stage==='fruit'&&!m.roots.length)checks.push('Document a directed repository-root contribution');
 if(m.status==='evidence-declared')checks.push('Independently verify declared evidence against exact-head CI or source');
 if(m.status==='unverified')checks.push('Attach source-backed evidence before claiming readiness');
 if(!checks.length)checks.push('Review source provenance; no completion receipt exists');
 return {id:m.id,label:m.label,stage:m.stage,rootIds:[...m.roots],destination:m.destination,status:m.status,checks,receipt:'not-verified',priority:priority[m.stage]??6};
 }).sort((a,b)=>a.priority-b.priority||b.checks.length-a.checks.length||a.id.localeCompare(b.id));
 return {quests:quests.slice(0,limit),total:quests.length,truncated:quests.length>limit,verified:0};
}
