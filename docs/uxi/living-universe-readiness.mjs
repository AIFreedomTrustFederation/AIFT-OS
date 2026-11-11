// Living Universe 2.0 — evidence-first operator readiness model.
// Pure, deterministic, local-only. A declared claim is never an independently verified check.
export const STAGES=Object.freeze(['root','trunk','branch','leaf','flower','fruit']);
const url=v=>{try{const u=new URL(v);return ['https:','http:'].includes(u.protocol)?u.href:null}catch{return null}};
const stage=n=>String(n?._stage||n?.stage||n?.kind||n?.type||'').toLowerCase().replace(/s$/,'');
export function buildReadinessAtlas(nodes,edges){
 if(!Array.isArray(nodes)||!Array.isArray(edges)||nodes.length>350||edges.length>1200)throw Error('Bounded graph required');
 const byId=new Map();
 for(const n of nodes){if(typeof n?.id!=='string'||!n.id||byId.has(n.id)||!STAGES.includes(stage(n)))throw Error('Invalid node identity or stage');byId.set(n.id,n)}
 const parents=new Map([...byId.keys()].map(id=>[id,[]]));
 const children=new Map([...byId.keys()].map(id=>[id,[]]));
 for(const e of edges){const a=e.from??e.source,b=e.to??e.target;if(!byId.has(a)||!byId.has(b))throw Error('Unknown edge endpoint');parents.get(b).push(a);children.get(a).push(b)}
 const ancestry=id=>{const seen=new Set(),todo=[id];while(todo.length){const current=todo.pop();if(seen.has(current))continue;seen.add(current);todo.push(...parents.get(current))}return [...seen]};
 const rows=nodes.map(n=>{
  const chain=ancestry(n.id),roots=chain.filter(id=>stage(byId.get(id))==='root');
  const destination=url(n.product_url||n.outcome_url||n.url||n.href);
  const issues=[];
  if(stage(n)==='fruit'){
   if(!destination)issues.push('No safe declared outcome destination');
   if(!parents.get(n.id).length)issues.push('No incoming contribution relationship');
   if(!roots.length)issues.push('No directed repository-root ancestry');
  }
  if(stage(n)==='flower'&&!parents.get(n.id).length)issues.push('Emerging work has no declared incoming relationship');
  return {id:n.id,label:String(n.label||n.name||n.title||n.id).slice(0,160),stage:stage(n),rootIds:roots,ancestorCount:chain.length-1,childCount:children.get(n.id).length,issues,destination,classification:issues.length?'evidence-gap':'declared-only'};
 });
 const fruits=rows.filter(r=>r.stage==='fruit'),gaps=rows.filter(r=>r.issues.length);
 return {rows,summary:{nodes:rows.length,roots:rows.filter(r=>r.stage==='root').length,declaredFruits:fruits.length,fruitEvidenceGaps:fruits.filter(r=>r.issues.length).length,reviewItems:gaps.length,verifiedProducts:0},note:'Declared graph evidence only; no independent CI, product verification or earned achievements.'};
}
export function readinessMissions(atlas){
 return atlas.rows.filter(r=>r.issues.length).map(r=>({id:r.id,title:'Review '+r.label,description:r.issues.join(' · '),state:'needs-evidence'}));
}
