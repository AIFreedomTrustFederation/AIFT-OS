import {verifyReceiptShape} from './living-universe-receipts.mjs';
// Imported evidence is untrusted. This is a bounded, read-only review projection.
export function receiptTimeline(receipts,{expectedRepo,expectedHead,now=Date.now(),limit=12}={}){
 if(!Number.isInteger(limit)||limit<1||limit>50)throw Error('Invalid receipt limit');
 if(!Array.isArray(receipts))return {entries:[],total:0,omitted:0,invalidCollection:true,verified:0};
 const entries=receipts.slice(0,limit).map((r,i)=>{
  const check=verifyReceiptShape(r,{expectedRepo,expectedHead,now});
  const valid=check.valid;
  return {index:i,state:valid?'declared':'invalid',workflow:typeof r?.workflow==='string'?r.workflow.slice(0,120):'Unidentified workflow',
   repository:typeof r?.repository==='string'?r.repository.slice(0,120):'Unknown repository',
   head:typeof r?.headSha==='string'?r.headSha.slice(0,40):'',
   conclusion:typeof r?.conclusion==='string'?r.conclusion.slice(0,24):'unknown',
   observedAt:typeof r?.observedAt==='string'?r.observedAt.slice(0,40):'',
   runUrl:valid?r.runUrl:null,commitUrl:valid?r.commitUrl:null,
   issues:check.errors,independentlyVerified:false};
 });
 return {entries,total:receipts.length,omitted:Math.max(0,receipts.length-limit),invalidCollection:false,verified:0};
}
