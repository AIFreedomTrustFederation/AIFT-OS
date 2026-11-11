// Deterministic, local-only lineage audit. Relationships are declarations, not verified attestations.
const stages = new Set(['root','trunk','branch','leaf','flower','fruit']);
const role = n => String(n?._stage || n?.stage || n?.kind || n?.type || '').toLowerCase().replace(/s$/, '');
const url = value => { try { const u = new URL(value); return ['https:','http:'].includes(u.protocol) ? u.href : null; } catch { return null; } };
export function auditLineage(nodes, edges, focusId) {
  if (!Array.isArray(nodes) || !Array.isArray(edges) || nodes.length > 350 || edges.length > 1200) throw Error('Unsupported graph size');
  const byId = new Map();
  for (const n of nodes) {
    if (!n || typeof n.id !== 'string' || !n.id || byId.has(n.id) || !stages.has(role(n))) throw Error('Invalid node identity or biological stage');
    byId.set(n.id,n);
  }
  const parents = new Map([...byId.keys()].map(id => [id,[]]));
  for (const e of edges) {
    const from = e?.from ?? e?.source, to = e?.to ?? e?.target;
    if (!byId.has(from) || !byId.has(to)) throw Error('Undeclared directed relationship');
    parents.get(to).push(from);
  }
  const focus = byId.get(focusId);
  if (!focus) return {focusId:null,ancestry:[],roots:[],gaps:[],classification:'unselected',outcomeUrl:null};
  const seen = new Set([focusId]), queue=[focusId], ancestry=[], gaps=[];
  while (queue.length) {
    const id=queue.shift(); ancestry.push(id);
    for (const p of parents.get(id)) if (!seen.has(p)) {seen.add(p);queue.push(p);}
  }
  const roots=ancestry.filter(id=>role(byId.get(id))==='root');
  if (!roots.length) gaps.push('No declared repository-root ancestry');
  if (role(focus)==='fruit') {
    const productUrl=url(focus.url || focus.href);
    if (!productUrl) gaps.push('Outcome has no valid user-facing destination');
    if (!parents.get(focusId).length) gaps.push('Outcome has no incoming directed contribution');
  }
  const outcomeUrl=role(focus)==='fruit'?url(focus.url || focus.href):null;
  return {focusId,ancestry,roots,gaps,classification:gaps.length?'evidence-gap':'declared-lineage',outcomeUrl};
}
export function auditGraph(nodes,edges) {
  const rows=nodes.map(n=>({node:n,report:auditLineage(nodes,edges,n.id)}));
  return {roots:rows.filter(r=>role(r.node)==='root').length,
    outcomes:rows.filter(r=>role(r.node)==='fruit').length,
    gaps:rows.filter(r=>r.report.gaps.length).length,
    rows};
}
