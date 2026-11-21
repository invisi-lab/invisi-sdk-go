// Reuses executed uint128 prototype receipts after verifying every Core event against current compiler ABI.
// This fixture does not claim production execution or refreshed gas measurements.
import fs from 'node:fs';
import {createRequire} from 'node:module';
const require=createRequire(new URL('../../../../invisi-core/package.json',import.meta.url));
const {Interface}=require('ethers');
const sdk=new URL('../../../',import.meta.url);
const core=new URL('../invisi-core/',sdk);
const base=new URL('.tools/credit-grant-uint128/grant-id-uint128/chain/',core);
const path=new URL('unified-receipts.json',import.meta.url);
const f=JSON.parse(fs.readFileSync(path));
const abi=JSON.parse(fs.readFileSync(new URL('../../contracts/abi/InvisiCore.abi.json',import.meta.url)));
const iface=new Interface(abi);
for(const [name,source] of [['credit-initial-payment','create-initial-sample-0.receipt.json'],['credit-draw','draw-1-sample-0.receipt.json']]){
 const c=f.cases.find(x=>x.name===name),r=JSON.parse(fs.readFileSync(new URL(source,base)));
 for(const log of r.logs.filter(l=>l.address.toLowerCase()===r.to.toLowerCase())) {
  const parsed=iface.parseLog(log); if(!parsed)throw Error('Unrecognized current Core log');console.log(source,parsed.name);
 }
 c.receipt=r;c.context.protocolAddress=r.to;c.context.timestamp=String(BigInt(r.blockTimestamp));
 c.provenance={source:'Core isolated uint128 grantId prototype Anvil experiment (before production adoption)',file:'.tools/credit-grant-uint128/grant-id-uint128/chain/'+source,eventABI:'Each Core event decoded against the current compiler ABI; receipt is not evidence of current production deployment or gas measurement'};
 if(name==='credit-draw')c.expected={ordinaryNotes:1,creditStates:1,nullifiers:1};
}
f.description='Executed local Anvil receipts. Note cases retain recorded historical provenance. Credit cases use the isolated uint128 grantId prototype, with exact current event ABI verification; no current production execution or gas measurement is claimed.';
fs.writeFileSync(path,JSON.stringify(f,null,2)+'\n');
