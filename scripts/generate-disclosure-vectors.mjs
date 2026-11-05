/** Cross-language disclosures from current TS APIs and authenticated envelope v1 delivery vectors. */
import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import { readFileSync,writeFileSync } from 'node:fs';
import {resolve,join} from 'node:path';
import {pathToFileURL,fileURLToPath} from 'node:url';
const root=resolve(fileURLToPath(new URL('..',import.meta.url)));
const ts=resolve(process.argv[2]??join(root,'../invisi-sdk-ts'));
const protocol=resolve(process.argv[3]??join(root,'../invisi-protocol'));
const {Disclosure}=await import(pathToFileURL(join(ts,'packages/core/dist/index.js')).href);
const v=JSON.parse(readFileSync(join(ts,'packages/core/src/testdata/wallet-note-vectors.json'),'utf8'));
for(const [kind,header,client] of [['note',1,'private-transfer-client.mjs'],['credit',129,'create-credit-client.mjs']]) {
 const value=v[kind], provenance=value.deliverySource;
 assert.equal(provenance.envelopeVersion,1,'current authenticated envelope v1 vectors required');
 assert.equal(provenance.header,header);
 assert.equal(provenance.protocolClientSha256,createHash('sha256').update(readFileSync(join(protocol,'scripts',client))).digest('hex'));
 for(const ciphertext of kind==='note'?[value.ciphertext]:[value.ownerCiphertext,value.spenderCiphertext]) {
  assert.equal(Number.parseInt(ciphertext.slice(2,4),16),header);
  assert.equal((ciphertext.length-2)/2,provenance.ciphertextBytes);
 }
}
const address=x=>'0x'+BigInt(x).toString(16).padStart(40,'0');
const context={chainId:BigInt(v.context.chainId),protocolAddress:BigInt(v.context.protocolAddress),tokenAddress:BigInt(v.context.tokenAddress)};
const note={...context,walletIdentity:address(v.note.walletIdentity),amount:BigInt(v.note.amount),nonce:BigInt(v.note.nonce)};
const credit={...context};
for(const key of ['spenderWalletIdentity','ownerWalletIdentity'])credit[key]=address(v.credit[key]);
for(const key of ['spenderPayout','ownerPayout','amount','nonce','expiry','creditSharedKey','singleLimit','dailyLimit','monthlyLimit','lastDay','dailySpent','lastMonth','monthlySpent'])credit[key]=BigInt(v.credit[key]);
const noteContext={...context,outputCommitment:v.note.commitment,walletIdentity:address(v.owner.walletIdentity)};
const creditContext={...context,outputCommitment:v.credit.commitment,walletIdentity:address(v.owner.walletIdentity),role:'owner'};
const noteOpening=(await Disclosure.fromNote(note)).export();
const creditOpening=(await Disclosure.fromCredit(credit)).export();
assert.equal(creditOpening.data.singleLimit,v.credit.singleLimit,'build the current TS core first');
const notePayload=(await Disclosure.fromEncryptedNote(v.owner.viewingPrivateKey,v.note.ciphertext,noteContext)).export();
const creditPayload=(await Disclosure.fromEncryptedCredit(v.owner.viewingPrivateKey,v.credit.ownerCiphertext,creditContext)).export();
writeFileSync(join(root,'internal/transaction/testdata/disclosure-vectors.json'),JSON.stringify({noteOpening,creditOpening,notePayload,creditPayload,noteCiphertext:v.note.ciphertext,creditCiphertext:v.credit.ownerCiphertext},null,2)+'\n');
console.log('Generated authenticated current TS envelope v1 disclosure vectors');
