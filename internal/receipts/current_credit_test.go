package receipts

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"testing"
)

func fixture(t *testing.T) (*Decoder, func(string, uint, common.Hash, []any, ...any) types.Log) {
	t.Helper()
	d, err := NewDecoder()
	if err != nil {
		t.Fatal(err)
	}
	return d, func(name string, index uint, tx common.Hash, indexed []any, args ...any) types.Log {
		t.Helper()
		e := d.abi.Events[name]
		data, err := e.Inputs.NonIndexed().Pack(args...)
		if err != nil {
			t.Fatal(err)
		}
		topics := []common.Hash{e.ID}
		for _, v := range indexed {
			q, err := abi.MakeTopics([]any{v})
			if err != nil {
				t.Fatal(err)
			}
			topics = append(topics, q[0][0])
		}
		return types.Log{Topics: topics, Data: data, Index: index, TxHash: tx}
	}
}

type sourceTuple struct {
	TreeId     *big.Int
	MerkleRoot [32]byte
}

func TestCurrentCreditEventContract(t *testing.T) {
	d, err := NewDecoder()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"CreditCreated", "CreditDraw", "CreditUpdated", "CreditRevoked"} {
		if _, ok := d.abi.Events[name]; !ok {
			t.Errorf("missing canonical event %s", name)
		}
	}
	for _, name := range []string{"CreditCancelled", "CreditToppedUp"} {
		if _, ok := d.abi.Events[name]; ok {
			t.Errorf("obsolete prepaid event %s", name)
		}
	}
}

func TestCreditCreationHasNoCashNoteInputsAndKeepsOwnerFeeMetadata(t *testing.T) {
	d, log := fixture(t)
	tx := common.HexToHash("1")
	ins := log("CommitmentsInserted", 1, tx, []any{big.NewInt(1)}, uint32(0), [][32]byte{{1}})
	created := log("CreditCreated", 2, tx, []any{common.Address{}, common.Address{}, big.NewInt(1)}, uint64(1), uint16(0), [32]byte{1}, uint32(0), []delivery{{[]byte{1}, nil}, {[]byte{2}, nil}}, big.NewInt(99), common.HexToAddress("44"), big.NewInt(7))
	b, err := d.Decode(Block{}, []types.Log{ins, created})
	if err != nil {
		t.Fatal(err)
	}
	assertCreditFee(t, b.Events[1].Data, false, "7")
	if len(b.Outputs) != 0 || len(b.CreditStates) != 1 || len(b.CreditStates[0].Deliveries) != 2 || len(b.Nullifiers) != 0 || len(b.CreditNullifiers) != 0 {
		t.Fatalf("mixed Credit/Note projection: %+v", b)
	}
}

type payerTuple struct {
	Payer  common.Address
	Amount *big.Int
}
type refTuple struct {
	PayerIndex *big.Int
	Shard      uint64
	Slot       uint16
}
type currentDrawTuple struct {
	Token            common.Address
	Sources          []sourceTuple
	Payers           []payerTuple
	GrantRefs        []refTuple
	Nullifiers       [][32]byte
	StateCommitments [][32]byte
	RecoveryAmounts  [][32]byte
	NoteCommitment   [32]byte
	TotalAmount      *big.Int
	Deadline         *big.Int
	FeeRecipient     common.Address
	FeeAmount        *big.Int
}
type updateTuple struct {
	Token            common.Address
	Owner            common.Address
	Source           sourceTuple
	Nullifier        [32]byte
	CreditCommitment [32]byte
	RevocationShard  uint64
	RevocationSlot   uint16
	Deadline         *big.Int
	FeeRecipient     common.Address
	FeeAmount        *big.Int
}

func TestCreditDrawAndUpdateHaveDistinctStateAndNoteProjections(t *testing.T) {
	d, log := fixture(t)
	tx := common.HexToHash("1")
	token := common.HexToAddress("2")
	p := currentDrawTuple{Token: token, Sources: []sourceTuple{{big.NewInt(1), [32]byte{9}}}, Payers: []payerTuple{{common.HexToAddress("1"), big.NewInt(3)}, {common.HexToAddress("2"), big.NewInt(4)}}, GrantRefs: []refTuple{{big.NewInt(0), 1, 0}, {big.NewInt(1), 1, 0}}, Nullifiers: [][32]byte{{1}, {2}}, StateCommitments: [][32]byte{{3}, {4}}, RecoveryAmounts: [][32]byte{{5}, {6}}, NoteCommitment: [32]byte{7}, TotalAmount: big.NewInt(7), Deadline: big.NewInt(99), FeeRecipient: common.HexToAddress("44"), FeeAmount: big.NewInt(1)}
	ins := log("CommitmentsInserted", 1, tx, []any{big.NewInt(1)}, uint32(0), [][32]byte{{3}, {4}, {7}})
	draw := log("CreditDraw", 2, tx, []any{big.NewInt(1)}, p, []uint32{0, 1, 2}, []delivery{{[]byte{1}, nil}, {[]byte{2}, nil}, {[]byte{3}, nil}, {[]byte{4}, nil}, {[]byte{5}, nil}})
	b, e := d.Decode(Block{Number: 10, Timestamp: 100}, []types.Log{ins, draw})
	if e != nil {
		t.Fatal(e)
	}
	assertCreditFee(t, b.Events[1].Data, true, "1")
	if len(b.Outputs) != 1 || b.Outputs[0].Commitment != p.NoteCommitment || len(b.CreditStates) != 2 || len(b.CreditStates[1].Deliveries) != 2 || len(b.CreditNullifiers) != 2 || len(b.Nullifiers) != 0 {
		t.Fatalf("incorrect postpaid draw %+v", b)
	}
	bad := log("CreditDraw", 2, tx, []any{big.NewInt(1)}, p, []uint32{0, 1}, []delivery{{}, {}, {}, {}, {}})
	if _, e = d.Decode(Block{}, []types.Log{ins, bad}); e == nil {
		t.Fatal("accepted missing ordinary Note position")
	}
	up := updateTuple{Token: token, Owner: common.HexToAddress("1"), Source: sourceTuple{big.NewInt(1), [32]byte{9}}, Nullifier: [32]byte{8}, CreditCommitment: [32]byte{9}, RevocationShard: 1, Deadline: big.NewInt(99), FeeRecipient: common.HexToAddress("44"), FeeAmount: big.NewInt(2)}
	ins = log("CommitmentsInserted", 1, tx, []any{big.NewInt(1)}, uint32(3), [][32]byte{{9}})
	updated := log("CreditUpdated", 2, tx, []any{up.Owner, big.NewInt(1)}, up, uint32(3), []delivery{{[]byte{1}, nil}, {[]byte{2}, nil}})
	revoked := log("CreditRevoked", 3, tx, []any{up.Owner, token, uint64(1)}, uint16(0), common.HexToAddress("44"), big.NewInt(3))
	b, e = d.Decode(Block{}, []types.Log{ins, updated, revoked})
	if e != nil {
		t.Fatal(e)
	}
	assertCreditFee(t, b.Events[1].Data, true, "2")
	assertCreditFee(t, b.Events[2].Data, false, "3")
	if len(b.Outputs) != 0 || len(b.CreditStates) != 1 || len(b.CreditNullifiers) != 1 || len(b.Nullifiers) != 0 {
		t.Fatalf("incorrect update/revoke %+v", b)
	}
}

func assertCreditFee(t *testing.T, raw json.RawMessage, nested bool, amount string) {
	t.Helper()
	var data map[string]json.RawMessage
	if e := json.Unmarshal(raw, &data); e != nil {
		t.Fatal(e)
	}
	if nested {
		if e := json.Unmarshal(data["params"], &data); e != nil {
			t.Fatal(e)
		}
	}
	var recipient common.Address
	var value string
	if e := json.Unmarshal(data["feeRecipient"], &recipient); e != nil {
		t.Fatal(e)
	}
	if e := json.Unmarshal(data["feeAmount"], &value); e != nil {
		t.Fatal(e)
	}
	if recipient != common.HexToAddress("44") || value != amount {
		t.Fatalf("fee metadata lost %s", raw)
	}
}
