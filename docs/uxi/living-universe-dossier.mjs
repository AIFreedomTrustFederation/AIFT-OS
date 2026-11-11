// Portable evidence dossier shared by the public portal and AIFT-OS.
// No remote access, persistence, inferred links, or verification claims.
export const DOSSIER_KIND = 'aift-living-universe-dossier';
const FIELDS = ['id','label','name','title','stage','kind','type','description','epistemicClass','status','genesisPart','evidenceBoundary','url','href','source_url','repository_url'];
const asId = value => typeof value === 'string' ? value : null;
export function createDossier(nodes, edges, focusId) {
  if (!Array.isArray(nodes) || !Array.isArray(edges) || nodes.length > 350 || edges.length > 1200)
    throw new Error('Unsupported or oversized graph');
  const byId = new Map(nodes.map(n => [asId(n.id), n]));
  if (!byId.has(focusId) || !focusId) throw new Error('Select a declared node before exporting');
  const incoming = new Map();
  for (const edge of edges) {
    const from = asId(edge.from ?? edge.source), to = asId(edge.to ?? edge.target);
    if (!byId.has(from) || !byId.has(to)) throw new Error('Graph contains an undeclared edge');
    if (!incoming.has(to)) incoming.set(to, []);
    incoming.get(to).push(from);
  }
  // Traverse every declared predecessor, not only a shortest path.
  const included = new Set([focusId]), queue = [focusId];
  while (queue.length) {
    const id = queue.shift();
    for (const parent of incoming.get(id) || []) {
      if (!included.has(parent)) { included.add(parent); queue.push(parent); }
    }
  }
  const subset = nodes.filter(n => included.has(n.id)).map(n => {
    const output = {};
    for (const field of FIELDS) if (typeof n[field] === 'string') output[field] = n[field];
    // _stage is a renderer field; stage must be canonical on export.
    output.stage = n._stage || n.stage || n.kind || n.type;
    return output;
  });
  const declared = edges.filter(e => included.has(e.from ?? e.source) && included.has(e.to ?? e.target))
    .map(e => ({from:e.from ?? e.source,to:e.to ?? e.target}));
  return {
    kind:DOSSIER_KIND, schemaVersion:'1.0.0',
    provenance:{classification:'user-exported-declaration',verification:'not-independently-verified'},
    focusId, nodes:subset, edges:declared
  };
}
export function validateDossier(value) {
  if (!value || value.kind !== DOSSIER_KIND || value.schemaVersion !== '1.0.0')
    throw new Error('Unsupported dossier format');
  if (!Array.isArray(value.nodes) || !Array.isArray(value.edges) || value.nodes.length > 350 || value.edges.length > 1200)
    throw new Error('Invalid dossier graph');
  const ids = new Set(value.nodes.map(n => n?.id));
  if (ids.size !== value.nodes.length || !ids.has(value.focusId) || [...ids].some(id => typeof id !== 'string' || !id))
    throw new Error('Dossier has invalid node identities');
  if (value.edges.some(e => !ids.has(e.from) || !ids.has(e.to)))
    throw new Error('Dossier has undeclared relationships');
  return value;
}
