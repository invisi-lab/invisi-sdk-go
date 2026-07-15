package core

const (
	DomainNKey       uint64 = 1
	DomainIdent      uint64 = 2
	DomainPay        uint64 = 3
	DomainNull       uint64 = 4
	DomainCredit     uint64 = 5
	DomainCNull      uint64 = 6
	DomainNKHash     uint64 = 7
	DomainDerive     uint64 = 8
	DomainRefLink    uint64 = 9
	DomainCreditNKey uint64 = 10
	// DomainChangeNKey is the domain tag for shielded-payment change derivation:
	// Poseidon2(D_CHANGE_NKEY, payerMasterKey, paymentCommitment). SPEC-029.
	DomainChangeNKey uint64 = 11

	TreeDepth = 16
)
