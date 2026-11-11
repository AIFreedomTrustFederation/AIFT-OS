/**
 * Living Universe 2.0 — evidence-first journey planner.
 * Pure local graph analysis: no inferred achievements, no remote checks.
 */
const stages=['root','trunk','branch','leaf','flower','fruit'];
const str=v=>typeof v==='string'?v:'';
export function planJourneys(nodes,edges) {
 if(!Array.isArray(nodes)||!Array.isArray(edges))throw Error('Expected nodes and edges arrays');
 const byId=new Map(nodes.map(n=>[n.id,n]));
 const incoming=new Map(nodes.map(n=>[n.id,[]]));
 const outgoing=new Map(nodes.map(n=>[n.id,[]]));
 for(const edge of edges){
  const from=str(edge.from||edge.source),to=str(edge.to||edge.target);
  if(!byId.has(from)||!byId.has(to))continue;
  incoming.get(to).push(from);outgoing.get(from).push(to);
 }
 const isRoot=n=>n.stage==='root'||n._stage==='root';
 const isFruit=n=>n.stage==='fruit'||n._stage==='fruit';
 const roots=nodes.filter(isRoot),fruits=nodes.filter(isFruit);
 const rootSet=new Set(roots.map(n=>n.id));
 const routes=[];
 // Traverse only actual directed edges; cycle guard prevents invented reachability.
 for(const fruit of fruits){
  const queue=[[fruit.id]],visited=new Set([fruit.id]);let path=null;
  while(queue.length){
   const backward=queue.shift(),head=backward[backward.length-1];
   if(rootSet.has(head)){path=[...backward].reverse();break}
   for(const parent of incoming.get(head)||[]){
    if(!visited.has(parent)){visited.add(parent);queue.push([...backward,parent])}
   }
  }
  const url=str(fruit.url||fruit.href);
  const verifiedProduct=fruit.productEvidence?.verified===true &&
    /^https?:\/\//i.test(url) && fruit.productEvidence?.sourceUrl &&
    /^https?:\/\//i.test(fruit.productEvidence.sourceUrl);
  routes.push({fruitId:fruit.id,fruitLabel:str(fruit.label||fruit.name||fruit.id),
   path:path||[],hasDirectedRootPath:!!path,hasProductEvidence:!!verifiedProduct,
   classification:verifiedProduct&&path?'evidence-linked-outcome':'evidence-gap',
   gap:!path?'No directed root-to-outcome path':!verifiedProduct?'No verified product evidence and destination':'none'});
 }
 const tasks=[];
 if(!nodes.length)return {routes,tasks,counts:{nodes:0,edges:0,roots:0,fruits:0}};
 for(const root of roots.slice(0,12))tasks.push({id:'root:'+root.id,type:'source-review',title:'Inspect source root: '+str(root.label||root.name||root.id),targetId:root.id,readiness:'available',note:'Exploration only; not a verified achievement.'});
 for(const route of routes.slice(0,24))tasks.push({id:'outcome:'+route.fruitId,type:'lineage-review',title:'Trace outcome: '+route.fruitLabel,targetId:route.fruitId,readiness:route.classification,note:route.gap});
 if(!roots.length)tasks.push({id:'gap:roots',type:'evidence-gap',title:'Identify a verifiable source root',targetId:nodes[0].id,readiness:'evidence-gap',note:'No source root declared'});
 return {routes,tasks,counts:{nodes:nodes.length,edges:edges.length,roots:roots.length,fruits:fruits.length}};
}
export function lineageFor(nodeId,nodes,edges,max=80){
 const byId=new Map(nodes.map(n=>[n.id,n]));if(!byId.has(nodeId))return [];
 const queue=[[nodeId]],visited=new Set([nodeId]);
 while(queue.length){const path=queue.shift(),current=path[0];const n=byId.get(current);
  if(n&&(n.stage==='root'||n._stage==='root'))return path;
  for(const e of edges){const to=e.to||e.target,from=e.from||e.source;
   if(to===current&&byId.has(from)&&!visited.has(from)){visited.add(from);queue.push([from,...path]);if(visited.size>max)return []}
  }
 }
 return [];
}
