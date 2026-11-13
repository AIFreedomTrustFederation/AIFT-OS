/** Evidence-bound exploration lens, shared by Journey Atlas and Mission Control.
 * No external requests, storage, telemetry, inferred relationships or completion claims.
 */
export const LENS_STAGES=Object.freeze(['root','trunk','branch','leaf','flower','fruit']);
export function buildEvidenceLens(missions=[],selected=null){
 const records=Array.isArray(missions)?missions.filter(m=>m&&typeof m.id==='string'&&LENS_STAGES.includes(m.stage)).slice(0,350):[];
 const byId=new Map(records.map(m=>[m.id,m]));
 const current=byId.get(selected)||null;
 const parents=m=>Array.isArray(m?.directParents)?m.directParents.filter(id=>typeof id==='string'):[];
 const lineage=[],missing=new Set(),seen=new Set(),stack=current?[current.id]:[];
 while(stack.length&&seen.size<350){const id=stack.pop();if(seen.has(id))continue;seen.add(id);const m=byId.get(id);if(!m){missing.add(id);continue}lineage.push(m);for(const p of parents(m))stack.push(p)}
 const roots=lineage.filter(m=>m.stage==='root');
 const counts=Object.fromEntries(LENS_STAGES.map(s=>[s,records.filter(m=>m.stage===s).length]));
 const fruit=current?.stage==='fruit';
 const destination=fruit&&typeof current.destination==='string'?current.destination:null;
 let outcomeUrl=null;
 if(destination){try{const u=new URL(destination);if(u.protocol==='https:'&&!['github.com','api.github.com','raw.githubusercontent.com'].includes(u.hostname)&&!u.hostname.endsWith('.github.com'))outcomeUrl=u.href}catch{}}
 return {records,current,lineage,roots,missing:[...missing],counts,outcomeUrl,fruit,claim:'Imported declarations are not independently verified.'};
}
export function mountEvidenceLens(host,{missions=[],selected=null,onSelect=()=>{}}={}){
 if(!host)return;
 const data=buildEvidenceLens(missions,selected);
 host.replaceChildren();
 const doc=host.ownerDocument;
 const node=(tag,cls,content)=>{const e=doc.createElement(tag);if(cls)e.className=cls;if(content!==undefined)e.textContent=content;return e};
 host.append(node('h2','','Source-to-outcome evidence lens'));
 host.append(node('p','muted','Explore declared lineage without inventing intermediate systems or verified achievements.'));
 const stages=node('div','aift-lens-stages');
 for(const stage of LENS_STAGES){const cell=node('div','aift-lens-stage');cell.dataset.stage=stage;cell.append(node('strong','',stage.toUpperCase()),node('span','',String(data.counts[stage])+' imported'));stages.append(cell)}
 host.append(stages);
 if(!data.current){host.append(node('p','notice',missions.length?'Select an imported record to inspect its source lineage.':'Import a local evidence packet to begin.'));return}
 host.append(node('h3','',data.current.label||data.current.id));
 const path=node('ol','aift-lens-lineage');
 for(const m of [...data.lineage].reverse()){const li=node('li');const b=node('button','',m.stage+' · '+(m.label||m.id));b.type='button';b.setAttribute('aria-current',String(m.id===data.current.id));b.addEventListener('click',()=>onSelect(m.id));li.append(b);path.append(li)}
 host.append(path);
 if(!data.roots.length)host.append(node('p','notice','Evidence gap: no declared repository root is reachable from this record.'));
 if(data.missing.length)host.append(node('p','notice','Missing parent references: '+data.missing.length+'. No missing relationships were inferred.'));
 if(data.fruit){if(data.outcomeUrl){const a=node('a','','Open declared outcome');a.href=data.outcomeUrl;a.target='_blank';a.rel='noopener noreferrer';host.append(a)}else host.append(node('p','notice','Evidence gap: no eligible HTTPS user-facing outcome destination.'))}
 host.append(node('p','muted',data.claim));
}
