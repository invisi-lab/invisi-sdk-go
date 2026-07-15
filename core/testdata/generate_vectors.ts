/**
 * Cross-language test vector generator.
 *
 * Runs against the TypeScript SDK (@invisi-labs/core) and outputs a JSON file
 * that the Go core package test suite loads to verify bit-for-bit compatibility.
 *
 * Usage:
 *   cd /path/to/invisi-sdk-ts
 *   npx tsx /path/to/generate_vectors.ts > /path/to/vectors.json
 */

import {
  poseidonHash,
  deriveNullifierKey,
  deriveCreditNullifierKey,
  deriveChangeNullifierKey,
  computeIdentityCommitment,
  computePaymentCommitment,
  computeChangeCommitment,
  computePaymentNullifier,
  computeNullifierKeyHash,
  computeCreditCommitment,
  computeCreditNullifier,
  deriveRemainderNonce,
  deriveRemainderNullifierKey,
  computeRefundLink,
} from "@invisi-labs/core";

function toBigintStr(v: bigint | `0x${string}`): string {
  if (typeof v === "string") {
    return BigInt(v).toString();
  }
  return v.toString();
}

function toHex(v: bigint | `0x${string}`): string {
  if (typeof v === "string" && v.startsWith("0x")) return v;
  return "0x" + v.toString(16).padStart(64, "0");
}

async function main() {
  const vectors: Record<string, any> = {};

  // ---- Poseidon2 raw hash ----
  vectors.poseidon2 = [
    { inputs: ["1", "2"], expected: await poseidonHash([1n, 2n]) },
    { inputs: ["1", "2", "3", "4"], expected: await poseidonHash([1n, 2n, 3n, 4n]) },
    { inputs: ["0", "0"], expected: await poseidonHash([0n, 0n]) },
    { inputs: ["42"], expected: await poseidonHash([42n]) },
    { inputs: ["1", "2", "3", "4", "5", "6", "7"], expected: await poseidonHash([1n, 2n, 3n, 4n, 5n, 6n, 7n]) },
  ];

  // ---- DeriveNullifierKey ----
  const nk = await deriveNullifierKey(100n, 200n);
  vectors.deriveNullifierKey = [
    { masterKey: "100", invoiceId: "200", expected: toHex(nk) },
  ];

  // Also derive with different inputs for extra coverage
  const nk2 = await deriveNullifierKey(0n, 0n);
  vectors.deriveNullifierKey.push({
    masterKey: "0", invoiceId: "0", expected: toHex(nk2),
  });

  // ---- DeriveCreditNullifierKey (D_CREDIT_NKEY = 10) ----
  const cnk1 = await deriveCreditNullifierKey(100n, 200n);
  const cnk2 = await deriveCreditNullifierKey(0n, 0n);
  vectors.deriveCreditNullifierKey = [
    { receiverMasterKey: "100", creditId: "200", expected: toHex(cnk1) },
    { receiverMasterKey: "0", creditId: "0", expected: toHex(cnk2) },
  ];

  // ---- DeriveChangeNullifierKey (D_CHANGE_NKEY = 11) ----
  const chg1 = await deriveChangeNullifierKey(100n, 200n);
  const chg2 = await deriveChangeNullifierKey(0n, 0n);
  vectors.deriveChangeNullifierKey = [
    { payerMasterKey: "100", paymentCommitment: "200", expected: toHex(chg1) },
    { payerMasterKey: "0", paymentCommitment: "0", expected: toHex(chg2) },
  ];

  // ---- ComputeIdentityCommitment ----
  const ic = await computeIdentityCommitment(nk, 0x999n);
  vectors.computeIdentityCommitment = [
    { nullifierKey: toBigintStr(nk), receivingAddress: "2457", expected: ic },
  ];

  // ---- ComputePaymentCommitment ----
  const pc = await computePaymentCommitment(1n, ic, 0xABCn, 1000n, 42n);
  vectors.computePaymentCommitment = [
    {
      chainId: "1",
      identityCommitment: ic,
      tokenAddress: "2748",
      amount: "1000",
      nonce: "42",
      expected: pc,
    },
  ];

  // ---- ComputeChangeCommitment ----
  // End-to-end change-commitment: wrap a change nullifier key + receiving
  // address into a change identity commitment, then a payment commitment.
  // Guards Go DeriveChange/ComputeChangeCommitment parity against the TS
  // reference (both = D_PAY over computeIdentityCommitment(nk, addr)).
  const chgCommit = await computeChangeCommitment(1n, chg1, 0x999n, 0xABCn, 250n, 7n);
  vectors.computeChangeCommitment = [
    {
      chainId: "1",
      changeNullifierKey: toBigintStr(chg1),
      changeReceivingAddress: "2457",
      tokenAddress: "2748",
      amount: "250",
      nonce: "7",
      expected: chgCommit,
    },
  ];

  // ---- ComputePaymentNullifier ----
  const pn = await computePaymentNullifier(1n, nk, 0xABCn, 1000n, 42n);
  vectors.computePaymentNullifier = [
    {
      chainId: "1",
      nullifierKey: toBigintStr(nk),
      tokenAddress: "2748",
      amount: "1000",
      nonce: "42",
      expected: pn,
    },
  ];

  // ---- ComputeNullifierKeyHash ----
  const nkh = await computeNullifierKeyHash(nk);
  vectors.computeNullifierKeyHash = [
    { nullifierKey: toBigintStr(nk), expected: nkh },
  ];

  // ---- ComputeCreditCommitment ----
  const cc = await computeCreditCommitment(1n, ic, ic, 0xABCn, 1000n, 42n, 9999n, nkh);
  vectors.computeCreditCommitment = [
    {
      chainId: "1",
      identityCommitment: ic,
      payerCommitment: ic,
      tokenAddress: "2748",
      amount: "1000",
      nonce: "42",
      expiry: "9999",
      nullifierKeyHash: nkh,
      expected: cc,
    },
  ];

  // ---- ComputeCreditNullifier ----
  const cn = await computeCreditNullifier(1n, nk, 0xABCn, 1000n, 42n, 9999n);
  vectors.computeCreditNullifier = [
    {
      chainId: "1",
      nullifierKey: toBigintStr(nk),
      tokenAddress: "2748",
      amount: "1000",
      nonce: "42",
      expiry: "9999",
      expected: cn,
    },
  ];

  // ---- DeriveRemainderNonce ----
  const rn = await deriveRemainderNonce(100n, 200n);
  vectors.deriveRemainderNonce = [
    { nonce: "100", newNonce: "200", expected: toHex(rn) },
  ];

  // ---- DeriveRemainderNullifierKey ----
  const rnk = await deriveRemainderNullifierKey(100n, 200n);
  vectors.deriveRemainderNullifierKey = [
    { nullifierKey: "100", newNonce: "200", expected: toHex(rnk) },
  ];

  // ---- ComputeRefundLink ----
  const rl = await computeRefundLink(12345n, 67890n);
  vectors.computeRefundLink = [
    { originalCommitment: "12345", refundNonce: "67890", expected: rl },
  ];

  console.log(JSON.stringify(vectors, null, 2));
}

main().catch(console.error);
