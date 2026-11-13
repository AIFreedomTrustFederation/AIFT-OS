/** Evidence-scoped biological journey ribbon. No inferred products, telemetry, or storage. */
export const BIO_STAGES=Object.freeze(['root','trunk','branch','leaf','flower','fruit']);
export function journeyStages(missions=[],selected=null){
  const records=Array.isArray(missions)?missions:[];
  const stages=BIO_STAGES.map(stage=>{
    const entries=records.filter(r=>r?.stage===stage&&typeof r.id==='string');
    return {stage,count:entries.length,entries,active:entries.some(r=>r.id===selected)};
  });
  return {stages,selected:records.some(r=>r.id===selected)?selected:null,
    missing:stages.filter(s=>s.count===0).map(s=>s.stage)};
}

/** Pure, bounded semantic-zoom model. Stage focus never changes evidence or selection. */
export function passageExplorer(missions=[],{stage='all',query='',limit=48}={}){
  const records=Array.isArray(missions)?missions:[];
  const focus=BIO_STAGES.includes(stage)?stage:'all';
  const needle=String(query||'').trim().slice(0,120).toLocaleLowerCase();
  const matches=records.filter(r=>r&&BIO_STAGES.includes(r.stage)&&typeof r.id==='string'&&
    (focus==='all'||r.stage===focus)&&
    (!needle||String(r.label||r.id).toLocaleLowerCase().includes(needle)||r.id.toLocaleLowerCase().includes(needle)));
  return {focus,query:needle,total:matches.length,visible:matches.slice(0,Math.max(1,Math.min(100,Number(limit)||48))),
    hidden:Math.max(0,matches.length-Math.max(1,Math.min(100,Number(limit)||48)))};
}
const passageViews=new WeakMap();

export function mountJourneyRibbon(host,{missions=[],selected=null,onSelect=()=>{}}={}){
  if(!host||typeof document==='undefined')return;
  const model=journeyStages(missions,selected);
  const previous=passageViews.get(host)||{stage:'all',query:''};
  const state={...previous};
  passageViews.set(host,state);
  const draw=()=>{
    const view=passageExplorer(missions,state);
    host.replaceChildren();
    const heading=document.createElement('h3');heading.textContent='Source-to-outcome passage';host.append(heading);
    const expl=document.createElement('p');expl.className='muted';
    expl.textContent='Repository sources remain roots. Fruit records are declared outcomes, not independently verified products. Empty stages are evidence gaps.';
    host.append(expl);
    const nav=document.createElement('nav');nav.className='aift-passage-navigation';nav.setAttribute('aria-label','Biological semantic zoom');
    const all=document.createElement('button');all.type='button';all.textContent='Whole tree';all.setAttribute('aria-pressed',String(view.focus==='all'));all.onclick=()=>{state.stage='all';draw()};nav.append(all);
    for(const {stage,count} of model.stages){
      const button=document.createElement('button');button.type='button';button.textContent=stage+' · '+count;
      button.setAttribute('aria-pressed',String(view.focus===stage));
      button.onclick=()=>{state.stage=stage;draw()};nav.append(button);
    }
    host.append(nav);
    const breadcrumb=document.createElement('p');breadcrumb.className='muted';breadcrumb.setAttribute('aria-live','polite');
    breadcrumb.textContent='Federation / '+(view.focus==='all'?'whole tree':view.focus)+' / '+view.total+' matching records';host.append(breadcrumb);
    const searchLabel=document.createElement('label');searchLabel.textContent='Search evidence records ';
    const search=document.createElement('input');search.type='search';search.value=state.query;search.placeholder='Record label or ID';search.setAttribute('aria-label','Search biological passage');
    search.oninput=()=>{state.query=search.value;const next=passageExplorer(missions,state);summary.textContent=next.total+' matching records; '+next.hidden+' not shown';populate(next)};
    searchLabel.append(search);host.append(searchLabel);
    const track=document.createElement('div');track.className='aift-passage';track.setAttribute('role','group');track.setAttribute('aria-label','Six biological evidence stages');host.append(track);
    const summary=document.createElement('p');summary.className='muted';summary.setAttribute('aria-live','polite');host.append(summary);
    function populate(current){
      track.replaceChildren();
      for(const {stage,count,active} of model.stages){
        if(current.focus!=='all'&&current.focus!==stage)continue;
        const section=document.createElement('section');section.className='aift-passage-stage';section.dataset.stage=stage;
        const title=document.createElement('h4');title.textContent=stage+' · '+count;section.append(title);
        const entries=current.visible.filter(r=>r.stage===stage);
        if(!count){const gap=document.createElement('p');gap.textContent='No imported evidence';section.append(gap)}
        else if(!entries.length){const gap=document.createElement('p');gap.textContent='No matching records in this view';section.append(gap)}
        else{
          const list=document.createElement('div');list.className='aift-passage-records';
          for(const record of entries){
            const button=document.createElement('button');button.type='button';button.textContent=record.label||record.id;
            button.setAttribute('aria-pressed',String(record.id===model.selected));
            button.title=stage+': '+(record.label||record.id);button.onclick=()=>onSelect(record.id);list.append(button);
          }
          section.append(list);
        }
        if(active)section.setAttribute('data-selected','true');
        track.append(section);
      }
    }
    populate(view);
    summary.textContent=view.total+' matching records; '+view.hidden+' not shown. '+model.missing.length+' unrepresented stages. No independent outcome verification.';
  };
  draw();
}
