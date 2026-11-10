// Public cross-language vectors from the generated Protocol ABI and independent ethers encoding.
import {readFileSync,writeFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import assert from 'node:assert/strict';
import {createRequire} from 'node:module';
const require=createRequire(new URL('../../../../invisi-protocol/package.json',import.meta.url));
const {Interface,Wallet,TypedDataEncoder,AbiCoder,keccak256}=require('ethers');
const artifactBytes=readFileSync(new URL('../../../../invisi-protocol/contracts/out/InvisiProtocol.sol/InvisiProtocol.json',import.meta.url)),artifact=JSON.parse(artifactBytes);
// Refuse artifacts compiled from a different local Solidity source closure.
const sourceRoot=new URL('../../../../invisi-protocol/',import.meta.url);
assert(Object.keys(artifact.metadata?.sources??{}).length>0,'artifact source metadata missing');
for(const[path,record]of Object.entries(artifact.metadata.sources))assert.equal(keccak256(readFileSync(new URL(path,sourceRoot))),record.keccak256,'artifact source mismatch: '+path);
const methods=['transact','shield','createCredit','creditDraw','updateCredit','revokeCredit'];
const abi=artifact.abi.filter(x=>x.type==='function'&&methods.includes(x.name)&&(x.name!=='transact'||x.inputs[1].type==='tuple'));
if(abi.length!==6)throw Error('Six current public ABIs required');
writeFileSync(new URL('../protocol.abi.json',import.meta.url),JSON.stringify(abi,null,2)+'\n');
writeFileSync(new URL('../fees.abi.json',import.meta.url),JSON.stringify(artifact.abi.filter(x=>x.type==='function'&&['broadcasterFeeBalance','withdrawBroadcasterFees'].includes(x.name)),null,2)+'\n');
const iface=new Interface(abi),word=n=>'0x'+BigInt(n).toString(16).padStart(64,'0'),zero='0x'+'00'.repeat(20),wallet=new Wallet(word(7));
const trust={chainId:56n,protocol:'0x1111111111111111111111111111111111111111',signer:wallet.address};
const token='0x2222222222222222222222222222222222222222',owner='0x3333333333333333333333333333333333333333',feeRecipient='0x4444444444444444444444444444444444444444';
const source={treeId:1n,merkleRoot:word(101)},deadline=2000000000n,proof='0xdeadbeef';
const delivery={ciphertext:'0x02'+'ab'.repeat(159),clue:'0x'+'cd'.repeat(64)+'18'+'ef'.repeat(3)};
const creditDelivery={ciphertext:'0x82'+'ab'.repeat(483),clue:delivery.clue};
const funding={payer:'0x5555555555555555555555555555555555555555',amount:67n,sigNonce:8n,deadline:1999999999n,signature:'0x123456',permit:{deadline:1999999998n,v:28,r:word(77),s:word(88)}};
const fee={deadline,feeRecipient,feeAmount:45n};
const request={params:{token,sources:[source,{treeId:9n,merkleRoot:word(109)}],nullifiers:[word(11),word(12)],noteCommitments:[word(21)],withdrawalAddress:owner,withdrawalAmount:123n,...fee},deliveries:[delivery],funding};
const shield={token,noteCommitments:[word(21)],...fee};
const creation={token,owner,revocationShard:1n,revocationSlot:0n,creditCommitment:word(31),...fee};
const draw={token,sources:[source],payers:[{payer:owner,amount:100n}],grantRefs:[{payerIndex:0n,shard:1n,slot:0n}],nullifiers:[word(41)],creditCommitments:[word(42)],recoveryAmounts:[word(43)],noteCommitment:word(44),totalAmount:100n,...fee};
const update={token,owner,source,nullifier:word(41),creditCommitment:word(45),revocationShard:1n,revocationSlot:0n,...fee};
const args={transact:[proof,request],shield:[shield,[{identity:word(51),amount:22n}],[delivery],funding],createCredit:[proof,creation,[creditDelivery,creditDelivery]],creditDraw:[proof,draw,[creditDelivery,creditDelivery,delivery]],updateCredit:[proof,update,[creditDelivery,creditDelivery]],revokeCredit:[token,owner,1n,0n,deadline,feeRecipient,45n,'0x1234']};
const quoteTypes={Quote:[{name:'token',type:'address'},{name:'feeRecipient',type:'address'},{name:'feeAmount',type:'uint256'},{name:'deadline',type:'uint256'},{name:'maxGas',type:'uint256'},{name:'requestHash',type:'bytes32'}]};
const domain={name:'InvisiBroadcaster',version:'1',chainId:trust.chainId,verifyingContract:trust.protocol};
function requestHash(method,values){const copy=structuredClone(values);if(method==='transact'){copy[0]='0x';copy[1].funding.signature='0x'}else if(method==='shield')copy[3].signature='0x';else if(method==='revokeCredit')copy[7]='0x';else copy[0]='0x';return keccak256(AbiCoder.defaultAbiCoder().encode(['string','string','bytes'],['InvisiBroadcasterIntentV1',method,iface.encodeFunctionData(method,copy)]));}
async function vector(method,values,publicRequest){const data=iface.encodeFunctionData(method,values),hash=requestHash(method,values),credit=method.endsWith('Credit')||method==='creditDraw';const quote={token,feeRecipient,feeAmount:'45',deadline:'1999999900',maxGas:'16777216',requestHash:hash};const variant=structuredClone(values);if(method==='revokeCredit')variant[7]='0xabcdef';else if(method==='shield')variant[3].signature='0xabcdef';else {variant[0]='0x01';if(method==='transact')variant[1].funding.signature='0xabcdef';}return {method,generatedBy:'independent ethers canonical Protocol ABI reference; SDK parity asserted separately',artifactSha256:createHash('sha256').update(artifactBytes).digest('hex'),privateKey:word(7),trust,request:publicRequest,proof:method==='shield'||method==='revokeCredit'?null:proof,data,requestHash:hash,quote,signature:await wallet.signTypedData(domain,quoteTypes,quote),quoteDigest:TypedDataEncoder.hash(domain,quoteTypes,quote),unsignedVariant:iface.encodeFunctionData(method,variant)};}
const json=x=>JSON.stringify(x,(_,v)=>typeof v==='bigint'?v.toString():v,2)+'\n';
const vectors=[];for(const method of methods)vectors.push(await vector(method,args[method],method==='transact'?request:args[method]));
writeFileSync(new URL('sdk-vectors.json',import.meta.url),json(vectors));writeFileSync(new URL('sdk-vector.json',import.meta.url),json(vectors[0]));
const publicRequest={...request,params:{...request.params,sources:[],nullifiers:[],noteCommitments:[word(21),word(22)],withdrawalAddress:zero,withdrawalAmount:0n},deliveries:[delivery,delivery]};writeFileSync(new URL('sdk-public-funding-vector.json',import.meta.url),json(await vector('transact',[proof,publicRequest],publicRequest)));
const secondRequest=structuredClone(request);secondRequest.params.nullifiers=[word(111),word(112)];secondRequest.params.noteCommitments=[word(121)];
const aggregate=new Interface([artifact.abi.find(x=>x.type==='function'&&x.name==='transact'&&x.inputs[1].type==='tuple[]')]);writeFileSync(new URL('unsupported.json',import.meta.url),json({aggregate:aggregate.encodeFunctionData('transact',[proof,[request,secondRequest]]),unknown:'0xdeadbeef'}));
