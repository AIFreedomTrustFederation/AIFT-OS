/**
 * Shared, local-only Living Universe exploration state.
 * IDs and routes are always resolved against the currently imported evidence.
 * No storage, telemetry, URL-sourced claims, or cross-origin communication.
 */
export const EXPLORATION_MODES=Object.freeze(['all','lineage','routes']);
export function createExplorationController({limit=25}={}){
  let valid=new Set(),selected=null,mode='all',history=[],cursor=-1,route=0;
  const snapshot=()=>Object.freeze({selected,mode,history:[...history],cursor,route,canBack:cursor>0,canForward:cursor<history.length-1});
  const select=(id,{remember=true}={})=>{
    if(typeof id!=='string'||!valid.has(id))return false;
    selected=id;route=0;
    if(remember){history=history.slice(0,cursor+1);if(history.at(-1)!==id)history.push(id);if(history.length>limit)history.shift();cursor=history.length-1}
    return true;
  };
  return Object.freeze({
    snapshot,
    importRecords(records){
      valid=new Set((Array.isArray(records)?records:[]).map(r=>r?.id).filter(id=>typeof id==='string'&&id.length>0));
      selected=null;mode='all';history=[];cursor=-1;route=0;
      return snapshot();
    },
    clear(){valid.clear();selected=null;mode='all';history=[];cursor=-1;route=0;return snapshot()},
    select,
    setMode(next){if(!EXPLORATION_MODES.includes(next)||next!=='all'&&!selected)return false;mode=next;route=0;return true},
    setRoute(index,count){if(!selected||!Number.isInteger(index)||index<0||index>=count)return false;mode='routes';route=index;return true},
    back(){if(cursor<=0)return false;cursor--;return select(history[cursor],{remember:false})},
    forward(){if(cursor>=history.length-1)return false;cursor++;return select(history[cursor],{remember:false})},
    contains(id){return valid.has(id)}
  });
}
