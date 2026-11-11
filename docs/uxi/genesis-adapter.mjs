/**
 * Living Universe 2.0 canonical Genesis adapter.
 * Never infer a completed product from Genesis's conceptual 'fruit' part.
 * All data stays local; no fetch, telemetry, storage or network access.
 */
export function adaptGenesisManifest(input) {
 if (!input || input.kind !== 'tree-manifest' || !Array.isArray(input.nodes) || !Array.isArray(input.edges)) throw Error('Not a canonical Genesis tree-manifest');
 const stages = {seed:'root',roots:'root',trunk:'trunk',branches:'branch',leaves:'leaf',fruit:'flower',newSeeds:'flower'};
 const ids = new Set();
 const nodes = input.nodes.map((n,i)=>{
  if (!n || typeof n.id !== 'string' || !n.id || ids.has(n.id)) throw Error('Invalid or duplicate Genesis node '+i);
  ids.add(n.id);
  const part = n.metadata?.part;
  const stage = stages[part];
  if (!stage) throw Error('Unmapped Genesis part for '+n.id);
  // Genesis's fruit describes emergence; it is not evidence of a delivered application.
  return {id:n.id,label:n.label||n.id,description:n.description||'',stage,
   epistemicClass:n.epistemicClass||'unspecified',status:n.status||'unspecified',
   genesisPart:part,sourceRefs:Array.isArray(n.sourceRefs)?n.sourceRefs:[],
   canonicalUri:n.canonicalUri||'', evidenceBoundary:n.metadata?.evidenceBoundary||'',
   sourceRepository:'https://github.com/AIFreedomTrustFederation/AIFT-Genesis'};
 });
 const edges = input.edges.map((e,i)=>{
  if (!e || !ids.has(e.from) || !ids.has(e.to)) throw Error('Unresolved Genesis edge '+i);
  return {from:e.from,to:e.to,relation:e.relation||'unspecified',epistemicClass:e.epistemicClass||'unspecified'};
 });
 return {kind:'living-universe-evidence',origin:'AIFT-Genesis tree-manifest',sourceStatus:input.status||'unspecified',
  nodes,edges,notes:'Genesis conceptual fruit is displayed as an emerging flower, not a verified product. No root repository is a fruit.'};
}
