/** Regenerate with: node internal/testvectors/generate.mjs /path/to/invisi-protocol
 * Source authority: supplied protocol checkout executable helpers and built contract ABI.
 * All proof bytes and secrets in these formatting fixtures are TEST values, not valid proofs.
 */
import { readFileSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";
import { pathToFileURL } from "node:url";
const root = process.argv[2];
if (!root) throw new Error("protocol checkout path required");
const from = (relative) => import(pathToFileURL(resolve(root, relative)).href);
const {
  Interface,
  TypedDataEncoder,
  toBeHex,
  AbiCoder,
  keccak256,
  toUtf8Bytes,
} = await from("node_modules/ethers/lib.esm/index.js");
const t = await from("scripts/transact-slots-client.mjs"),
  c = await from("scripts/create-credit-client.mjs"),
  l = await from("scripts/credit-lifecycle-client.mjs"),
  privateTransfer = await from("scripts/private-transfer-client.mjs");
const w = (x) => toBeHex(x, 32),
  ctx = {
    chainId: 31337n,
    protocolAddress: "0x1111111111111111111111111111111111111111",
    tokenAddress: "0x2222222222222222222222222222222222222222",
  },
  owner = "0x3333333333333333333333333333333333333333",
  payer = "0x4444444444444444444444444444444444444444",
  cipher = "0x01" + "11".repeat(177),
  cc = "0x02" + "22".repeat(253),
  f = { payer, amount: 100n, sigNonce: 7n, deadline: 1000n };
const fixtures = {};
for (const count of [1, 15, 57, 113]) {
  const p = {
      sources: [{ treeId: 1n, merkleRoot: w(1) }],
      nullifiers: [w(2)],
      outputCommitments: Array.from({ length: count }, (_, i) => w(i + 3)),
      withdrawalAddress: "0x0000000000000000000000000000000000000000",
      withdrawalAmount: 0n,
      deadline: 900n,
    },
    cs = Array(count).fill(cipher);
  fixtures["transact" + count] = {
    params: p,
    ciphertexts: cs,
    publicInputs: t.transactSlotsPublicInputs(ctx, p, cs),
    typedData: t.encodeTransactSlotsTypedData(ctx, p, cs),
    digest: TypedDataEncoder.hash(
      ...Object.values(t.encodeTransactSlotsTypedData(ctx, p, cs)).filter(
        (_, i) => i !== 1
      )
    ),
  };
}
const p = {
    sources: [],
    nullifiers: [],
    creditCommitment: w(5),
    changeCommitment: w(0),
    owner,
    expiry: 1000n,
    deadline: 900n,
  },
  cs = [cc, cc],
  a = { sigNonce: 4n };
fixtures.create = {
  params: p,
  ciphertexts: cs,
  funding: f,
  publicInputs: c.createCreditPublicInputs(ctx, p, cs, f),
  requestHash: c.createCreditRequestHash(p, cs, f),
  typedData: c.encodeCreateCreditTypedData(ctx, p, cs, a, f),
  fundingTypedData: c.encodeCreateCreditFundingTypedData(ctx, p, cs, f),
};
const cancel = {
  source: { treeId: 1n, merkleRoot: w(1) },
  nullifier: w(2),
  owner,
  amount: 50n,
  deadline: 900n,
};
fixtures.cancel = {
  params: cancel,
  publicInputs: l.cancelCreditPublicInputs(ctx, cancel),
  requestHash: l.cancelCreditRequestHash(cancel),
  typedData: l.encodeCancelCreditTypedData(ctx, cancel, a),
};
const top = {
  creation: p,
  creditSource: cancel.source,
  creditNullifier: w(2),
  previousExpiry: 950n,
};
fixtures.topUp = {
  params: top,
  ciphertexts: cs,
  funding: f,
  publicInputs: l.topUpCreditPublicInputs(ctx, top, cs, f),
  requestHash: l.topUpCreditRequestHash(top, cs, f),
  typedData: l.encodeTopUpCreditTypedData(ctx, top, cs, a, f),
  fundingTypedData: l.encodeTopUpCreditFundingTypedData(ctx, top, cs, f),
};
for (const count of [1, 2, 13, 51]) {
  const p = {
      sources: [{ treeId: 1n, merkleRoot: w(1) }],
      nullifiers: Array.from({ length: count }, (_, i) => w(i + 2)),
      remainderCommitments: Array(count).fill(w(0)),
      expiries: Array(count).fill(1000n),
      recoveryAmounts: Array(count).fill(w(0)),
      spender: owner,
      totalAmount: 50n,
      deadline: 900n,
    },
    cs = Array(count * 2).fill("0x");
  fixtures["draw" + count] = {
    params: p,
    ciphertexts: cs,
    publicInputs: l.creditDrawPublicInputs(ctx, p, cs),
  };
}
fixtures.transact1.memberRequestHash = keccak256(
  AbiCoder.defaultAbiCoder().encode(
    ["bytes32", "uint256", "bytes32[]"],
    [
      keccak256(toUtf8Bytes("InvisiAtomicBatchMemberV1")),
      14,
      fixtures.transact1.publicInputs,
    ]
  )
);
const ftuple =
    "tuple(address payer,uint256 amount,uint256 sigNonce,uint256 deadline,bytes signature,tuple(uint256 deadline,uint8 v,bytes32 r,bytes32 s) permit)",
  at = "tuple(uint256 sigNonce,bytes signature)",
  fund = {
    ...f,
    signature: "0xabcd",
    permit: { deadline: 0n, v: 0, r: w(0), s: w(0) },
  },
  auth = { sigNonce: 4n, signature: "0xabcd" },
  proof = "0x1234";
const shieldOutput = { identityCommitment: w(3), amount: 100n, nonce: w(42) },
  shieldP = {
    outputCommitments: [
      w(
        await privateTransfer.poseidon2([
          3n,
          ctx.chainId,
          3n,
          BigInt(ctx.tokenAddress),
          100n,
          42n,
        ])
      ),
    ],
    deadline: 900n,
  },
  shieldCs = [cipher],
  st = "tuple(bytes32[] outputCommitments,uint256 deadline)",
  sot = "tuple(bytes32 identityCommitment,uint256 amount,bytes32 nonce)[]";
const shieldHash = keccak256(
  AbiCoder.defaultAbiCoder().encode(
    ["bytes32", st, "bytes32", "bytes32"],
    [
      keccak256(toUtf8Bytes("InvisiShield")),
      shieldP,
      keccak256(AbiCoder.defaultAbiCoder().encode([sot], [[shieldOutput]])),
      keccak256(AbiCoder.defaultAbiCoder().encode(["bytes[]"], [shieldCs])),
    ]
  )
);
fixtures.shield = {
  params: shieldP,
  outputs: [shieldOutput],
  ciphertexts: shieldCs,
  funding: fund,
  requestHash: shieldHash,
  calldata: new Interface([
    `function shield(${st} params,${sot} outputs,bytes[] ciphertexts,${ftuple} funding)`,
  ]).encodeFunctionData("shield", [shieldP, [shieldOutput], shieldCs, fund]),
};
const framedProof = AbiCoder.defaultAbiCoder().encode(
  ["uint8", "bytes"],
  [0, proof]
);
fixtures.create.calldata = new Interface([
  `function createCredit(bytes proof,${c.CREATE_CREDIT_PARAMS_ABI} params,bytes[] ciphertexts,${ftuple} funding,${at} authorization)`,
]).encodeFunctionData("createCredit", [
  framedProof,
  fixtures.create.params,
  fixtures.create.ciphertexts,
  fund,
  auth,
]);
fixtures.cancel.calldata = new Interface([
  `function cancelCredit(bytes proof,${l.CANCEL_CREDIT_PARAMS_ABI} params,${at} authorization)`,
]).encodeFunctionData("cancelCredit", [
  framedProof,
  fixtures.cancel.params,
  auth,
]);
fixtures.topUp.calldata = new Interface([
  `function topUpCredit(bytes proof,${l.TOP_UP_CREDIT_PARAMS_ABI} params,bytes[] ciphertexts,${ftuple} funding,${at} authorization)`,
]).encodeFunctionData("topUpCredit", [
  framedProof,
  fixtures.topUp.params,
  fixtures.topUp.ciphertexts,
  fund,
  auth,
]);
fixtures.draw1.calldata = new Interface([
  `function creditDraw(bytes proof,${l.CREDIT_DRAW_PARAMS_ABI} params,bytes[] ciphertexts)`,
]).encodeFunctionData("creditDraw", [
  framedProof,
  fixtures.draw1.params,
  fixtures.draw1.ciphertexts,
]);
const sp = { ...fixtures.transact1.params, inputTreeId: 1n, merkleRoot: w(1) };
delete sp.sources;
const singleTuple =
  "tuple(bytes32 merkleRoot,bytes32[] nullifiers,bytes32[] outputCommitments,uint256 inputTreeId,address withdrawalAddress,uint256 withdrawalAmount,uint256 deadline)";
fixtures.single = {
  params: fixtures.transact1.params,
  ciphertexts: fixtures.transact1.ciphertexts,
  fundingRequestHash: keccak256(
    AbiCoder.defaultAbiCoder().encode(
      ["bytes32", singleTuple, "bytes32"],
      [
        keccak256(proof),
        sp,
        fixtures.transact1.typedData.message.ciphertextsHash,
      ]
    )
  ),

};
const crossTwoParams = {
  ...fixtures.transact1.params,
  sources: [
    { treeId: 1n, merkleRoot: w(1) },
    { treeId: 2n, merkleRoot: w(2) },
  ],
  nullifiers: [w(20), w(21)],
};
fixtures.crossTwo = {
  params: crossTwoParams,
  ciphertexts: fixtures.transact1.ciphertexts,

};
// Encode against the freshly built protocol ABI, independently from SDK builders.
const protocolAbi = JSON.parse(readFileSync(resolve(root, "contracts/out/InvisiProtocol.sol/InvisiProtocol.json"))).abi;
const iface = new Interface(protocolAbi);
const direct = iface.fragments.find(f => f.type === "function" && f.name === "transact" && f.inputs[1].type === "tuple");
const emptyFunding = {payer: "0x"+"00".repeat(20), amount:0n, sigNonce:0n, deadline:0n, signature:"0x", permit:{deadline:0n,v:0,r:w(0),s:w(0)}};
for (const name of ["transact1", "single", "crossTwo"]) {
  const fixture=fixtures[name];
  const params = fixture.params;
  fixture.calldata=iface.encodeFunctionData(direct,[proof,{params,ciphertexts:fixture.ciphertexts,funding:emptyFunding}]);
}
fixtures.transact1.fundingRequestHash=fixtures.single.fundingRequestHash;
const user = await from("scripts/transact-user-client.mjs");
for (const [count,capacity] of [[1,14],[15,56],[57,112],[113,168]]) {
  fixtures["transact"+count].userPublicInputs = user.userPublicInputs(fixtures["transact"+count].publicInputs,capacity);
}
const cross = await from("scripts/cross-tree-client.mjs");
fixtures.crossTwo.fundingRequestHash = cross.crossTreeFundingRequestHash(proof,fixtures.crossTwo.params,fixtures.crossTwo.ciphertexts);
const shield = await from("scripts/shield-client.mjs");
fixtures.shield.fundingTypedData = await shield.encodeShieldFundingTypedData(ctx,shieldP,[shieldOutput],shieldCs,fund);
const aggregateRequests = user.prepareAggregateRequests(ctx,Array.from({length:4},(_,i)=>({
 params:{...fixtures.transact1.params,nullifiers:[w(100+i)],outputCommitments:[w(200+i)]},
 ciphertexts:fixtures.transact1.ciphertexts,
})));
const aggregateMethod=iface.fragments.find(f=>f.type==='function'&&f.name==='transact'&&f.inputs[1].type==='tuple[]');
fixtures.aggregate4={requests:aggregateRequests,publicInputs:user.aggregatePublicInputs(ctx,aggregateRequests),calldata:iface.encodeFunctionData(aggregateMethod,[proof,aggregateRequests])};
writeFileSync(
  new URL("./current-protocol.json", import.meta.url),
  JSON.stringify(
    { source: "invisi-protocol@ede4bc6 executable client helpers and built ABI", context: ctx, fixtures },
    (_, v) => (typeof v === "bigint" ? v.toString() : v),
    2
  )
);
