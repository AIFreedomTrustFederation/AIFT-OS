/** Shared evidence-only constellation theatre for Journey Atlas and Mission Control.
 * Declared ancestry is not proof of a completed product or passing CI.
 */
import {spatialSummary,MAP_STAGES} from './living-universe-spatial-minimap.mjs';
const COLORS=['#ff547b','#ffa86b','#f5cd83','#42e6aa','#69baff','#b49cff'];
export function constellationModel(missions=[],selected=null){
 const summary=spatialSummary(missions,selected);
 const byId=new Map(summary.records.map(r=>[r.id,r]));
 const route=summary.routes[0]||null;
 const ids=route?.ids||summary.lineage.map(r=>r.id);
 const nodes=ids.map((id,i)=>{const record=byId.get(id);return record?{id,stage:record.stage,label:record.label||id,x:75+(i%6)*135,y:80+Math.floor(i/6)*110,selected:id===selected}:null}).filter(Boolean);
 const positions=new Map(nodes.map(n=>[n.id,n]));
 const edges=[];
 for(const node of nodes){const record=byId.get(node.id);
  for(const parent of Array.isArray(record.directParents)?record.directParents:[]){
   if(positions.has(parent)){const a=positions.get(parent),from=MAP_STAGES.indexOf(a.stage),to=MAP_STAGES.indexOf(node.stage);
    edges.push({from:parent,to:node.id,anomaly:to!==from+1});}
  }
 }
 return {nodes,edges,summary,route,empty:!nodes.length};
}
export function mountConstellationTheatre(host,{missions=[],selected=null,onSelect=()=>{}}={}){
 if(!host)return;
 const doc=host.ownerDocument,m=constellationModel(missions,selected);
 const el=(tag,text)=>{const n=doc.createElement(tag);if(text!==undefined)n.textContent=text;return n};
 const svgEl=(tag,attrs={})=>{const n=doc.createElementNS('http://www.w3.org/2000/svg',tag);for(const [k,v] of Object.entries(attrs))n.setAttribute(k,String(v));return n};
 host.replaceChildren();host.setAttribute('aria-label','Evidence constellation theatre');
 const h=el('h3','Source-to-outcome constellation');host.append(h);
 const desc=el('p','Imported ancestry only. Select a record to inspect its declared lineage; no live CI or product verification is implied.');desc.className='muted';host.append(desc);
 if(m.empty){const p=el('p','Select an imported record with declared ancestry to explore its constellation.');p.setAttribute('role','status');host.append(p);return}
 const width=900,height=Math.max(190,Math.ceil(m.nodes.length/6)*110+70);
 const svg=svgEl('svg',{viewBox:`0 0 ${width} ${height}`,role:'img','aria-label':'Directed lineage diagram; accessible record controls follow below'});
 svg.style.width='100%';svg.style.height='auto';svg.style.maxHeight='420px';
 const byId=new Map(m.nodes.map(n=>[n.id,n]));
 for(const edge of m.edges){const a=byId.get(edge.from),b=byId.get(edge.to);
  svg.append(svgEl('path',{d:`M ${a.x} ${a.y} Q ${(a.x+b.x)/2} ${Math.min(a.y,b.y)-35} ${b.x} ${b.y}`,fill:'none',stroke:edge.anomaly?'#f5cd83':'#69baff','stroke-width':edge.anomaly?3:2,'stroke-dasharray':edge.anomaly?'5 5':'none'}));
 }
 for(const n of m.nodes){const group=svgEl('g');
  const color=COLORS[MAP_STAGES.indexOf(n.stage)];group.append(svgEl('circle',{cx:n.x,cy:n.y,r:n.selected?25:19,fill:'#101b32',stroke:color,'stroke-width':n.selected?5:3}));
  const label=svgEl('text',{x:n.x,y:n.y+36,'text-anchor':'middle',fill:'currentColor','font-size':12});label.textContent=n.label.slice(0,22);group.append(label);svg.append(group);
 }
 host.append(svg);
 const nav=el('ol');nav.className='aift-constellation-directory';nav.setAttribute('aria-label','Accessible constellation records');
 for(const n of m.nodes){const li=el('li'),b=el('button',n.stage+' · '+n.label);b.type='button';b.dataset.recordId=n.id;b.setAttribute('aria-current',n.selected?'step':'false');b.addEventListener('click',()=>onSelect(n.id));li.append(b);nav.append(li)}
 host.append(nav);
 if(m.edges.some(e=>e.anomaly)){const warning=el('p','Dashed gold connections indicate declared nonadjacent or nonprogressive stages. No intermediate records were invented.');warning.setAttribute('role','status');host.append(warning)}
 if(!m.summary.hasRoot){const warning=el('p','Evidence gap: no complete repository-root lineage.');warning.setAttribute('role','status');host.append(warning)}
}
