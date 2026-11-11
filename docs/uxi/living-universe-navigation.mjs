// Pure, bounded navigation state for both AIFT observatories.
// No persistence, network, or claim of verified evidence.
export const BIOLOGICAL_STAGES=Object.freeze(['root','trunk','branch','leaf','flower','fruit']);
export function createNavigation(nodes=[],edges=[]){
 const byId=new Map(nodes.map(n=>[n.id,n]));
 const adjacent=new Map(nodes.map(n=>[n.id,[]]));
 for(const e of edges){const from=e.from??e.source,to=e.to??e.target;if(adjacent.has(from)&&byId.has(to))adjacent.get(from).push(to)}
 const visited=[];let cursor=-1;
 function inspect(id,{record=true}={}){
  if(!byId.has(id))return null;
  if(record){visited.splice(cursor+1);if(visited[cursor]!==id)visited.push(id);if(visited.length>100)visited.shift();cursor=visited.length-1}
  return byId.get(id);
 }
 function step(direction){const next=cursor+direction;if(next<0||next>=visited.length)return null;cursor=next;return byId.get(visited[cursor])}
 function neighbors(id){return [...new Set(adjacent.get(id)||[])].filter(id=>byId.has(id))}
 function snapshot(){return {selected:visited[cursor]||null,back:cursor>0,forward:cursor<visited.length-1,visited:visited.slice(),cursor}}
 function reset(){visited.length=0;cursor=-1}
 return {inspect,back:()=>step(-1),forward:()=>step(1),neighbors,snapshot,reset}
}
export function stageFor(node){const stage=String(node?._stage||node?.stage||node?.kind||'').toLowerCase().replace(/s$/,'');return BIOLOGICAL_STAGES.includes(stage)?stage:null}
