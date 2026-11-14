import {receiptGateState} from './living-universe-receipts.mjs';
// Living Universe 2.0: evidence-backed skill atlas (pure, no network or storage).
export const STAGES=Object.freeze(['root','trunk','branch','leaf','flower','fruit']);
export function analyzeAtlas(nodes,edges){
 if(!Array.isArray(nodes)||!Array.isArray(edges))throw Error('Expected nodes and directed edges');
 const index=new Map();
 for(const n of nodes){if(!n||typeof n.id!=='string'||!n.id||index.has(n.id))throw Error('Invalid or duplicate node');index.set(n.id,n)}
 const inbound=new Map(nodes.map(n=>[n.id,[]]));
 for(const e of edges){const a=e.from??e.source,b=e.to??e.target;if(!index.has(a)||!index.has(b))throw Error('Undeclared edge');inbound.get(b).push(a)}
 const stage=n=>String(n._stage||n.stage||n.kind||n.type||'').toLowerCase().replace(/s$/,'');
 const isRoot=n=>stage(n)==='root' && [n.repository_url,n.source_url,n.url,n.href].some(u=>typeof u==='string'&&/^https:\/\/github\.com\/[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+\/?$/.test(u));
 // A GitHub repository is source infrastructure, never a user-facing fruit.
 // Reject repository, commit, tree, blob, PR, Actions and API endpoints as outcomes.
 const productUrl=n=>[n.product_url,n.application_url,n.url,n.href].find(u=>{try{const p=new URL(u);return p.protocol==='https:'&&!!p.hostname&&!['github.com','www.github.com','api.github.com','raw.githubusercontent.com'].includes(p.hostname)&&!p.hostname.endsWith('.githubusercontent.com')}catch{return false}})||null;
 const sourceUrl=n=>[n.url,n.href,n.source_url,n.repository_url].find(u=>{try{return new URL(u).protocol==='https:'}catch{return false}})||null;
 const ancestors=id=>{const visited=new Set(),stack=[id];while(stack.length){const next=stack.pop();if(visited.has(next))continue;visited.add(next);stack.push(...(inbound.get(next)||[]))}visited.delete(id);return [...visited]};
 const missions=nodes.map(n=>{
 const s=stage(n),parents=ancestors(n.id),roots=parents.filter(id=>isRoot(index.get(id))),destination=s==='fruit'?productUrl(n):sourceUrl(n);
 const issues=[];
 if(!STAGES.includes(s))issues.push('Unrecognized biological stage');
 if(s==='fruit'&&!destination)issues.push('No declared user-facing destination distinct from source repository');
 if(s==='fruit'&&roots.length===0)issues.push('No directed repository-root ancestry');
 if(s==='fruit'&&inbound.get(n.id).length===0)issues.push('No declared incoming contribution');
 if(s==='fruit'&&[n.repository_url,n.source_url].some(Boolean))issues.push('Fruit contains repository-source fields; review biological classification');
 const evidence=n.evidence||n.verification||n.ci;
 // Never infer a verified achievement from an arbitrary evidence string or URL.
 return {id:n.id,label:String(n.label||n.name||n.title||n.id),stage:s,roots,destination,issues,
  status:issues.length?'evidence-gap':evidence?'evidence-declared':'unverified',receipts:Array.isArray(n.receipts)?n.receipts:[],ancestors:parents,directParents:[...new Set(inbound.get(n.id))]};
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
 return {id:m.id,label:m.label,stage:m.stage,rootIds:[...m.roots],destination:m.destination,status:m.status,checks,readiness:evidenceReadinessGates(m),receipts:m.receipts||[],receipt:'not-verified',priority:priority[m.stage]??6};
 }).sort((a,b)=>a.priority-b.priority||b.checks.length-a.checks.length||a.id.localeCompare(b.id));
 return {quests:quests.slice(0,limit),total:quests.length,truncated:quests.length>limit,verified:0};
}

/**
 * Evidence readiness is intentionally NOT an achievement or live CI assertion.
 * Each gate describes what an operator must verify outside the imported graph.
 * A declared evidence URL alone cannot satisfy a gate.
 */
export function evidenceReadinessGates(mission){
 if(!mission||typeof mission.id!=='string'||!Array.isArray(mission.roots)||!Array.isArray(mission.issues))throw Error('Invalid atlas mission');
 const gates=[
  {id:'provenance',label:'Directed source provenance',state:mission.roots.length?'declared':'blocked',detail:mission.roots.length?mission.roots.length+' contributing root(s) declared':'No contributing repository root declared'},
  {id:'evidence',label:'Independent evidence verification',state:receiptGateState(mission.receipts||[]).state,detail:receiptGateState(mission.receipts||[]).reason+'; independently confirm run conclusion and head before marking complete'},
  {id:'accessibility',label:'Accessible user journey',state:'pending',detail:'Verify keyboard, screen reader, mobile and reduced-motion behavior in a real browser'}
 ];
 if(mission.stage==='fruit')gates.splice(1,0,{id:'destination',label:'User-facing deliverable',state:mission.destination?'declared':'blocked',detail:mission.destination?'HTTPS destination declared; live availability unverified':'No HTTPS user-facing destination declared'});
 if(mission.issues.length)gates.push({id:'classification',label:'Biological classification review',state:'blocked',detail:mission.issues.join('; ')});
 return {id:mission.id,stage:mission.stage,gates,ready:false,verified:false,blocking:gates.filter(g=>g.state==='blocked').length,pending:gates.filter(g=>g.state==='pending').length};
}
