/** Current ordinary Note disclosure fixtures; Credit audit vectors live in credit/testdata. */
import assert from 'node:assert/strict';
import {readFileSync,writeFileSync} from 'node:fs';
import {resolve,join} from 'node:path';
import {pathToFileURL,fileURLToPath} from 'node:url';
const root=resolve(fileURLToPath(new URL('..',import.meta.url)));
const ts=resolve(process.argv[2]??join(root,'../invisi-sdk-ts'));
const {Disclosure}=await import(pathToFileURL(join(ts,'packages/core/dist/index.js')).href);
const v=JSON.parse(readFileSync(join(ts,'packages/core/src/wallet/testdata/wallet-note-vectors.json'),'utf8'));
assert.equal(v.note.ciphertext.slice(0,4),'0x01');
assert.equal((v.note.ciphertext.length-2)/2,v.note.deliverySource.ciphertextBytes);
const address=x=>'0x'+BigInt(x).toString(16).padStart(40,'0');
const context={chainId:BigInt(v.context.chainId),protocolAddress:BigInt(v.context.protocolAddress),tokenAddress:BigInt(v.context.tokenAddress)};
const note={...context,ownerAddress:address(v.note.ownerAddress),amount:BigInt(v.note.amount),nonce:BigInt(v.note.nonce)};
const noteContext={...context,outputCommitment:v.note.commitment,ownerAddress:address(v.owner.ownerAddress)};
const noteOpening=(await Disclosure.fromNote(note)).export();
const notePayload=(await Disclosure.fromEncryptedNote(v.owner.viewingPrivateKey,v.note.ciphertext,noteContext)).export();
writeFileSync(join(root,'internal/transaction/testdata/disclosure-vectors.json'),JSON.stringify({noteOpening,notePayload,noteCiphertext:v.note.ciphertext},null,2)+'\n');
console.log('Generated current ordinary Note disclosure vectors');
