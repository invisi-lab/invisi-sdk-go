# Hidden-amount Note examples

These examples discover Notes through authenticated Protocol RPC receipts, decrypt deliveries through `shared.AppDeliveryDecryptor`, and independently validate openings and canonical membership. Replace the App placeholder before running live discovery.

Private transfers sign the exact frozen EIP-712 statement using geth and `OWNER_PRIVATE_KEY`. Public-funding signature variants use `PAYER_PRIVATE_KEY`; these keys must match the required owner/payer. `06-aggregate` signs each owner's members separately, using `SECOND_OWNER_PRIVATE_KEY` for the second public wallet. Full owner-bound sweeps require no private-spending signature and can pay only the original owner.

The commands log unsigned calldata. Preparation/proving does not send a payment. Ordinary broadcaster delivery and independent Source pre-/post-payment submission are consolidated under [07-broadcaster](../07-broadcaster/README.md).
