// Evidence-first spatial canopy. Pure geometry: no network, no implied achievements.
export const SPECTRUM=Object.freeze(['#ff547b','#f5cd83','#42e6aa','#69baff','#b49cff','#f5cd83']);
export const STAGES=Object.freeze(['root','trunk','branch','leaf','flower','fruit']);
export function canopyLayout(atlas,{width=960,height=480}={}){
 if(!atlas||!Array.isArray(atlas.missions))throw Error('Atlas required');
 const groups=STAGES.map(s=>atlas.missions.filter(m=>m.stage===s).sort((a,b)=>a.id.localeCompare(b.id)));
 const positions=new Map();
 groups.forEach((group,i)=>group.forEach((m,j)=>positions.set(m.id,{id:m.id,x:75+i*(width-150)/5,y:60+(j+1)*(height-120)/(group.length+1),stage:m.stage,label:m.label,issues:m.issues.length,roots:m.roots.length,status:m.status})));
 const connections=[];
 for(const m of atlas.missions)for(const ancestor of (m.directParents||[])){if(!positions.has(ancestor)||!positions.has(m.id))continue;
 // Draw only explicitly declared direct edges between adjacent biological stages. Transitive ancestry is not an edge.
 const a=positions.get(ancestor),b=positions.get(m.id);
 if(STAGES.indexOf(b.stage)-STAGES.indexOf(a.stage)===1)connections.push({from:ancestor,to:m.id});
 }
 return {nodes:[...positions.values()],connections};
}
export function canopyFocus(layout,id){
 const node=layout.nodes.find(n=>n.id===id);
 if(!node)return {node:null,neighbors:[]};
 return {node,neighbors:layout.connections.filter(e=>e.from===id||e.to===id)};
}
