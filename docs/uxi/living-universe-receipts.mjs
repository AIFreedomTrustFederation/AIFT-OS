// Pure, fail-closed verification of external evidence receipts. No network, storage or telemetry.
const SHA=/^[0-9a-f]{40}$/i;
const RUN=/^https:\/\/github\.com\/([\w.-]+)\/([\w.-]+)\/actions\/runs\/([1-9]\d*)$/;
const COMMIT=/^https:\/\/github\.com\/([\w.-]+)\/([\w.-]+)\/commit\/([0-9a-f]{40})$/i;
const VALID=new Set(['success','failure','cancelled','pending']);
function safeDate(value){if(typeof value!=='string'||!/^\d{4}-\d\d-\d\dT/.test(value))return false;return Number.isFinite(Date.parse(value));}
export function verifyReceiptShape(receipt,{expectedRepo,expectedHead,now}={}){
 const errors=[];
 if(!receipt||typeof receipt!=='object'||Array.isArray(receipt))return {valid:false,errors:['Receipt must be an object'],state:'unverified'};
 const {repository,headSha,runUrl,commitUrl,conclusion,observedAt,workflow}=receipt;
 if(typeof repository!=='string'||!/^[-\w.]+\/[-\w.]+$/.test(repository))errors.push('Invalid repository');
 if(typeof headSha!=='string'||!SHA.test(headSha))errors.push('Invalid exact-head SHA');
 if(typeof workflow!=='string'||!workflow.trim()||workflow.length>120)errors.push('Missing workflow identity');
 if(!VALID.has(conclusion))errors.push('Invalid workflow conclusion');
 const run=typeof runUrl==='string'&&runUrl.match(RUN),commit=typeof commitUrl==='string'&&commitUrl.match(COMMIT);
 if(!run)errors.push('Missing canonical GitHub Actions run URL');
 if(!commit)errors.push('Missing canonical GitHub commit URL');
 if(run&&repository&&run[1]+'/'+run[2]!==repository)errors.push('Run repository mismatch');
 if(commit&&repository&&commit[1]+'/'+commit[2]!==repository)errors.push('Commit repository mismatch');
 if(commit&&headSha&&commit[3].toLowerCase()!==headSha.toLowerCase())errors.push('Commit SHA mismatch');
 if(expectedRepo&&repository!==expectedRepo)errors.push('Unexpected repository');
 if(expectedHead&&headSha!==expectedHead)errors.push('Stale head SHA');
 if(!safeDate(observedAt))errors.push('Missing ISO observation timestamp');
 const timestamp=Date.parse(observedAt);
 if(now!==undefined&&(!Number.isFinite(now)||timestamp>now+60000))errors.push('Future or invalid observation');
 // Shape validation is not independent verification. Never promote an imported receipt to verified.
 return {valid:errors.length===0,errors,state:errors.length?'invalid':'declared',independentlyVerified:false};
}
export function receiptGateState(receipts,options={}){
 if(!Array.isArray(receipts))return {state:'blocked',reason:'Receipt collection must be an array',receipts:[]};
 const checked=receipts.map(r=>verifyReceiptShape(r,options));
 const valid=checked.filter(r=>r.valid).length;
 return {state:valid?'declared':'pending',reason:valid?valid+' structurally valid receipt(s), independently unverified':'No structurally valid exact-head receipt',receipts:checked,independentlyVerified:false};
}
