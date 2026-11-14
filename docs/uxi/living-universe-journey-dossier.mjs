// Evidence-only journey dossier. No network, storage, status inference, or fabricated edges.
import {STAGES,canopyRoutes,canopyReadiness} from './living-universe-canopy.mjs';
export function journeyDossier(atlas,layout,id){
 if(!atlas?.missions||!layout?.nodes)throw Error('Imported atlas and layout required');
 const mission=atlas.missions.find(m=>m.id===id);
 if(!mission)return null;
 const routes=canopyRoutes(layout,id,{maxRoutes:12}),readiness=canopyReadiness(layout,id);
 const byId=new Map(atlas.missions.map(m=>[m.id,m]));
 const stages=STAGES.map(stage=>({stage,records:routes.routes.flatMap(r=>r.ids).filter((v,i,a)=>a.indexOf(v)===i).filter(v=>byId.get(v)?.stage===stage)}));
 const evidenceGaps=[...(mission.issues||[])];
 if(routes.unreachable)evidenceGaps.push('No declared directed path from a repository root');
 if(readiness.anomalies)evidenceGaps.push(readiness.anomalies+' non-adjacent or nonprogressive stage edge(s)');
 if(readiness.missingReferences)evidenceGaps.push(readiness.missingReferences+' unresolved reference(s)');
 if(mission.stage==='fruit'&&!mission.destination)evidenceGaps.push('No declared user-facing destination');
 return {id,stage:mission.stage,label:mission.label,routes:routes.routes.map(r=>({rootId:r.rootId,ids:[...r.ids],stageGaps:r.stageGaps})),truncated:routes.truncated,stages,readiness,evidenceGaps:[...new Set(evidenceGaps)],verified:false,claim:'Declared graph lineage only; independent verification required'};
}
