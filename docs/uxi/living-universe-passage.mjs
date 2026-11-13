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
export function mountJourneyRibbon(host,{missions=[],selected=null,onSelect=()=>{}}={}){
  if(!host||typeof document==='undefined')return;
  const model=journeyStages(missions,selected);
  host.replaceChildren();
  const heading=document.createElement('h3');heading.textContent='Source-to-outcome passage';host.append(heading);
  const expl=document.createElement('p');expl.className='muted';
  expl.textContent='Explore documented stages. Empty stages are evidence gaps, not completed milestones. Repository sources remain roots; a fruit requires a declared user-facing outcome.';
  host.append(expl);
  const track=document.createElement('div');track.className='aift-passage';track.setAttribute('role','group');track.setAttribute('aria-label','Six biological evidence stages');
  for(const {stage,count,entries,active} of model.stages){
    const section=document.createElement('section');section.className='aift-passage-stage';
    section.dataset.stage=stage;
    const title=document.createElement('h4');title.textContent=stage+' · '+count;section.append(title);
    if(!count){const gap=document.createElement('p');gap.textContent='No imported evidence';section.append(gap)}
    else {
      const list=document.createElement('div');list.className='aift-passage-records';
      for(const record of entries.slice(0,12)){
        const button=document.createElement('button');button.type='button';button.textContent=record.label||record.id;
        button.setAttribute('aria-pressed',String(record.id===model.selected));
        button.title=stage+': '+(record.label||record.id);
        button.onclick=()=>onSelect(record.id);list.append(button);
      }
      section.append(list);
      if(entries.length>12){const note=document.createElement('p');note.textContent=(entries.length-12)+' additional records in the directory';section.append(note)}
    }
    if(active)section.setAttribute('data-selected','true');
    track.append(section);
  }
  host.append(track);
  const footer=document.createElement('p');footer.className='muted';
  footer.textContent='Declared evidence only · '+model.missing.length+' unrepresented stages · no independent outcome verification';
  host.append(footer);
}
