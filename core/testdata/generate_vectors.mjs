// Run from any directory: node core/testdata/generate_vectors.mjs [protocol-root].
// Expected hashes come directly from the current protocol's Barretenberg client.
import {writeFile} from 'node:fs/promises';
import {fileURLToPath,pathToFileURL} from 'node:url';
import {resolve,dirname} from 'node:path';
const here=dirname(fileURLToPath(import.meta.url));
const root=resolve(process.argv[2]??resolve(here,'../../../invisi-protocol'));
const {poseidon2,identityCommitment,noteCommitment,noteNullifier,FIELD_MODULUS,viewingPublicKey,encryptPrivateTransferNote}=await import(pathToFileURL(resolve(root,'scripts/private-transfer-client.mjs')));
const {creditCommitment,encryptCreditNote}=await import(pathToFileURL(resolve(root,'scripts/create-credit-client.mjs')));
const {creditNullifier,deriveCreditRemainder}=await import(pathToFileURL(resolve(root,'scripts/credit-lifecycle-client.mjs')));
const address=n=>'0x'+n.toString(16).padStart(40,'0');
const hex=n=>'0x'+n.toString(16).padStart(64,'0');
const vectors=[];
const envelopes=[];
const viewingPrivateKey=Buffer.from(hex(123n).slice(2),'hex');
const viewingKey=viewingPublicKey(viewingPrivateKey);
const entropy={ephemeralPrivateKey:Buffer.from(hex(456n).slice(2),'hex'),salt:Buffer.alloc(16,0x33),iv:Buffer.alloc(12,0x44)};
for(const seed of [1n,19n,FIELD_MODULUS-1000n]) {
 const context={chainId:31337n,protocolAddress:address(0x1234n),tokenAddress:address(0x5678n)};
 const note={amount:500n,nonce:seed+1n,noteSecret:seed,receivingAddress:address(0xabcden)};
 const credit={spenderIdentityCommitment:await identityCommitment(seed,address(0xabcden)),ownerIdentityCommitment:await identityCommitment(seed+3n,address(0x7777n)),amount:1000n,nonce:seed+2n,expiry:2000000000n,creditSharedKey:seed+4n};
 const consumed=BigInt(await creditNullifier(context,credit));
 const recoveryArgs=[context.chainId,BigInt(context.protocolAddress),BigInt(context.tokenAddress),consumed,credit.creditSharedKey];
 const add=(name,args,value)=>vectors.push({name,inputs:args.map(String),expected:typeof value==='string'?value:hex(value)});
 add('DeriveNoteSecret',[seed,42n],await poseidon2([1n,seed,42n]));
 add('ComputeIdentityCommitment',[seed,BigInt(note.receivingAddress)],credit.spenderIdentityCommitment);
 add('ComputeNoteCommitment',[context.chainId,credit.spenderIdentityCommitment,BigInt(context.tokenAddress),note.amount,note.nonce],await noteCommitment(context,note));
 add('ComputeChangeCommitment',[context.chainId,seed,BigInt(note.receivingAddress),BigInt(context.tokenAddress),note.amount,note.nonce],await noteCommitment(context,note));
 add('ComputeNoteNullifier',[context.chainId,seed,BigInt(note.receivingAddress),BigInt(context.tokenAddress),note.amount,note.nonce],await noteNullifier(context,note));
 add('ComputeCreditSharedKeyHash',[credit.creditSharedKey],await poseidon2([7n,credit.creditSharedKey]));
 add('ComputeCreditCommitment',[context.chainId,credit.spenderIdentityCommitment,credit.ownerIdentityCommitment,BigInt(context.tokenAddress),credit.amount,credit.nonce,credit.expiry,await poseidon2([7n,credit.creditSharedKey])],await creditCommitment(context,credit));
 add('ComputeCreditNullifier',[context.chainId,credit.creditSharedKey,BigInt(context.tokenAddress),credit.amount,credit.nonce,credit.expiry],consumed);
 add('DeriveRemainderNonce',recoveryArgs,await poseidon2([10n,...recoveryArgs]));
 add('ComputeCreditRecoveryPad',recoveryArgs,await poseidon2([9n,...recoveryArgs]));
 for(const [kind,value,commit,encrypt] of [['ordinary',note,await noteCommitment(context,note),encryptPrivateTransferNote],['credit',credit,await creditCommitment(context,credit),encryptCreditNote]]) {
 const envelope=await encrypt(viewingKey,value,{...context,outputCommitment:commit},entropy);
 envelopes.push({kind,privateKey:'0x'+viewingPrivateKey.toString('hex'),publicKey:'0x'+viewingKey.toString('hex'),entropy:'0x'+Buffer.concat([entropy.ephemeralPrivateKey,entropy.salt,entropy.iv]).toString('hex'),context:{chainId:context.chainId,protocolAddress:BigInt(context.protocolAddress),tokenAddress:BigInt(context.tokenAddress),outputCommitment:commit},note:kind==='ordinary'?{...value,receivingAddress:BigInt(value.receivingAddress),tokenAddress:BigInt(context.tokenAddress)}:{...value,tokenAddress:BigInt(context.tokenAddress)},envelope:'0x'+envelope.toString('hex')});
 }
 const result=await deriveCreditRemainder(context,credit,400n);
 vectors.push({name:'RecoverCreditRemainder',context:{chainId:context.chainId,protocolAddress:BigInt(context.protocolAddress)},previous:{...credit,tokenAddress:BigInt(context.tokenAddress)},recovery:{creditNullifier:result.nullifier,remainderCommitment:result.remainderCommitment,recoveryAmount:result.recoveryAmount},expected:{...result.remainder,tokenAddress:BigInt(context.tokenAddress)}});
}
await writeFile(resolve(here,'envelopes.json'),JSON.stringify({source:'invisi-protocol ede4bc6 deterministic client envelopes',envelopes},(_,v)=>typeof v==='bigint'?v.toString():v,2)+'\n');
await writeFile(resolve(here,'vectors.json'),JSON.stringify({source:'invisi-protocol ede4bc6 scripts; Barretenberg Poseidon2',vectors},(_,v)=>typeof v==='bigint'?v.toString():v,2)+'\n');
process.exit(0);
