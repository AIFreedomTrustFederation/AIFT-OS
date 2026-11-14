/**
 * Portable, evidence-limited lineage dossier for Journey Atlas and Mission Control.
 * Explicitly not an attestation or a substitute for the imported source packet.
 */
import {journeyDossier} from './living-universe-journey-dossier.mjs';
const MAX_RECORDS=350;
export function portableDossier(atlas,layout,id){
  const d=journeyDossier(atlas,layout,id);
  if(!d)return null;
  const byId=new Map(atlas.missions.map(m=>[m.id,m]));
  const ids=[...new Set([id,...d.routes.flatMap(r=>r.ids)])].slice(0,MAX_RECORDS);
  const records=ids.map(key=>{
    const m=byId.get(key);
    return m?{id:m.id,label:m.label,stage:m.stage,status:m.status??null,
      destination:typeof m.destination==='string'?m.destination:null}:null;
  }).filter(Boolean);
  return {schema:'aift.living-universe.lineage-dossier.v1',subject:id,
    evidenceClass:'imported-declarations-unverified',independentlyVerified:false,
    records,routes:d.routes.map(r=>({rootId:r.rootId,ids:r.ids,stageGaps:r.stageGaps})),
    evidenceGaps:d.evidenceGaps,routeDisplayTruncated:d.truncated,
    notice:'This export is a selected view of locally imported declarations, not proof of product readiness or a complete source graph.'};
}
export function mountPortableDossier(host,{atlas,layout,selected}={}){
  if(!host)return;
  host.replaceChildren();
  const title=document.createElement('h3');title.textContent='Portable lineage dossier';host.append(title);
  const note=document.createElement('p');note.textContent='Export the selected source-to-outcome trail for review. No verification, signing, live status, or source completeness is implied.';host.append(note);
  const packet=atlas&&layout&&selected?portableDossier(atlas,layout,selected):null;
  const button=document.createElement('button');button.type='button';button.textContent='Export selected lineage JSON';
  button.disabled=!packet;
  button.addEventListener('click',()=>{
    if(!packet)return;
    const blob=new Blob([JSON.stringify(packet,null,2)+'\n'],{type:'application/json'});
    const url=URL.createObjectURL(blob);
    const a=document.createElement('a');a.href=url;a.download='aift-lineage-dossier.json';
    document.body.append(a);a.click();a.remove();
    setTimeout(()=>URL.revokeObjectURL(url),1000);
  });
  host.append(button);
  const state=document.createElement('p');
  state.textContent=packet?packet.records.length+' declared records in selected view; '+packet.evidenceGaps.length+' evidence gap(s).':'Import evidence and select a record to export.';
  host.append(state);
}
