package protocol

type BuildCreditDrawCalldataParams struct {
	Context    TransactionContext
	Proof      []byte
	Params     CreditDrawParams
	Deliveries []Delivery
}
