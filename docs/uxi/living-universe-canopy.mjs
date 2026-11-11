// Directed, evidence-first canopy. Geometry never manufactures source relationships.
export const SPECTRUM=Object.freeze(['#ff547b','#f5cd83','#42e6aa','#69baff','#b49cff','#f5cd83']);
export const STAGES=Object.freeze(['root','trunk','branch','leaf','flower','fruit']);
export function canopyLayout(atlas,{width=960,height=480}={}){
 if(!atlas||!Array.isArray(atlas.missions))throw Error('Atlas required');
 if(!Number.isFinite(width)||!Number.isFinite(height)||width<320||height<240)throw Error('Invalid viewport');
 const ids=new Set(),groups=STAGES.map(s=>[]);
 for(const m of atlas.missions){
  if(!m||typeof m.id!=='string'||!m.id||ids.has(m.id))throw Error('Invalid or duplicate mission');
  ids.add(m.id);
  const i=STAGES.indexOf(m.stage);
  if(i>=0)groups[i].push(m);
 }
 const positions=new Map();
 groups.forEach((group,i)=>{
  group.sort((a,b)=>a.id.localeCompare(b.id));
  group.forEach((m,j)=>positions.set(m.id,{id:m.id,x:75+i*(width-150)/5,y:60+(j+1)*(height-120)/(group.length+1),stage:m.stage,label:m.label,issues:(m.issues||[]).length,roots:(m.roots||[]).length,status:m.status}));
 });
 const connections=[],seen=new Set(),omissions=[];
 for(const m of atlas.missions){
  for(const parent of m.directParents||[]){
   const key=JSON.stringify([parent,m.id]);
   if(seen.has(key))continue;
   seen.add(key);
   const a=positions.get(parent),b=positions.get(m.id);
   if(!a||!b){omissions.push({from:parent,to:m.id,reason:'unrenderable-stage-or-unknown-parent'});continue}
   const distance=STAGES.indexOf(b.stage)-STAGES.indexOf(a.stage);
   // Every visible connection is a declared direct edge. A gap is not a fabricated intermediate node.
   if(distance===1)connections.push({from:parent,to:m.id});
   else connections.push({from:parent,to:m.id,stageDistance:distance,kind:distance>1?'stage-gap':distance<=0?'nonprogressive':'adjacent'});
  }
 }
 return {nodes:[...positions.values()],connections,omissions};
}
export function canopyFocus(layout,id){
 const node=layout.nodes.find(n=>n.id===id);
 if(!node)return {node:null,neighbors:[]};
 return {node,neighbors:layout.connections.filter(e=>e.from===id||e.to===id)};
}
// Reversible source-to-outcome inspection: no transitive edges are ever drawn.
export function canopyLineage(layout,id){
 const index=new Map(layout.nodes.map(n=>[n.id,n]));
 if(!index.has(id))return {nodes:[],connections:[],rootIds:[],gaps:[]};
 const parents=new Map();
 for(const edge of layout.connections){if(!parents.has(edge.to))parents.set(edge.to,[]);parents.get(edge.to).push(edge)}
 const visited=new Set(),used=new Map(),stack=[id];
 while(stack.length){
  const current=stack.pop();if(visited.has(current))continue;visited.add(current);
  for(const edge of parents.get(current)||[]){used.set(JSON.stringify([edge.from,edge.to]),edge);stack.push(edge.from)}
 }
 return {nodes:[...visited].map(key=>index.get(key)),connections:[...used.values()],
  rootIds:[...visited].filter(key=>index.get(key)?.stage==='root'),
  gaps:[...used.values()].filter(edge=>edge.kind==='stage-gap'||edge.kind==='nonprogressive')};
}

// Evidence-only shortest declared routes from repository roots to a selected record.
// Never invents missing nodes, implied links, or verification receipts.
export function canopyRoutes(layout,id,{maxRoutes=12}={}){
 if(!Number.isInteger(maxRoutes)||maxRoutes<1||maxRoutes>50)throw Error('Invalid route limit');
 const index=new Map(layout.nodes.map(n=>[n.id,n]));
 if(!index.has(id))return {routes:[],truncated:false,unreachable:true};
 const outgoing=new Map();
 for(const edge of layout.connections){
  if(!index.has(edge.from)||!index.has(edge.to))continue;
  if(!outgoing.has(edge.from))outgoing.set(edge.from,[]);
  outgoing.get(edge.from).push(edge.to);
 }
 for(const values of outgoing.values())values.sort();
 const routes=[];
 const roots=layout.nodes.filter(n=>n.stage==='root').map(n=>n.id).sort();
 for(const root of roots){
  const queue=[[root]],visited=new Set([root]);let found=null;
  while(queue.length){
   const path=queue.shift(),last=path[path.length-1];
   if(last===id){found=path;break}
   for(const next of outgoing.get(last)||[]){
    if(!visited.has(next)){visited.add(next);queue.push([...path,next])}
   }
  }
  if(found)routes.push({rootId:root,ids:found,stageGaps:found.slice(1).filter((to,i)=>{
   const from=found[i];
   return layout.connections.some(e=>e.from===from&&e.to===to&&e.kind&&e.kind!=='adjacent');
  }).length});
 }
 return {routes:routes.slice(0,maxRoutes),truncated:routes.length>maxRoutes,unreachable:routes.length===0};
}
