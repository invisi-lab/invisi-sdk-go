package protocol

type BuildCancelCreditCalldataParams struct {
	Context TransactionContext
	Proof   []byte
	Params  CancelCreditParams
}
