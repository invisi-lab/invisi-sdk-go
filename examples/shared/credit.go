package shared

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func CreditDeliveries(snapshot credit.DeliveryState) ([]protocol.Delivery, error) {
	commitment, err := snapshot.State.Commitment()
	if err != nil {
		return nil, err
	}
	result := []protocol.Delivery{}
	for _, role := range []credit.Role{credit.Owner, credit.Spender} {
		identity := snapshot.State.OwnerWalletIdentity
		keys := snapshot.OwnerReceivingKeys
		if role == credit.Spender {
			identity = snapshot.State.SpenderWalletIdentity
			keys = snapshot.SpenderReceivingKeys
		}
		cipher, err := credit.Encrypt(snapshot, credit.DeliveryContext{ProtocolContext: snapshot.State.ProtocolContext, TokenAddress: snapshot.State.TokenAddress, Commitment: commitment, RecipientIdentity: identity}, role)
		if err != nil {
			return nil, err
		}
		clue, err := core.CreateClue(keys.ClueKey)
		if err != nil {
			return nil, err
		}
		result = append(result, protocol.Delivery{Ciphertext: cipher, Clue: clue})
	}
	return result, nil
}

// RunCreditExample uses locally initialized accounts and never submits without --execute.
func RunCreditExample(kind string, actorIndex int) {
	execute := flag.Bool("execute", false, "authorize local proving and chain submission after reviewing Settings")
	shard := flag.Uint64("shard", 1, "original grant shard for revocation")
	slot := flag.Uint("slot", 0, "original grant slot for revocation")
	available := flag.String("available-credit", "400", "initial finite available credit, uint128")
	unlimited := flag.Bool("unlimited", false, "create unlimited grant; available-credit must be zero")
	increase := flag.String("increase-available-credit", "0", "update quota increase, uint128")
	decrease := flag.String("decrease-available-credit", "0", "update quota decrease, uint128")
	mode := flag.String("new-unlimited", "keep", "update mode: keep, true or false")
	flag.Parse()
	if !*execute {
		log.Fatal("Review Settings and code, then add --execute; no network request or transaction was made")
	}
	if *slot > 65535 {
		log.Fatal("slot exceeds uint16")
	}
	parseQuota := func(text string) *big.Int {
		value, ok := new(big.Int).SetString(text, 10)
		if !ok || value.Sign() < 0 || value.BitLen() > 128 || value.String() != text {
			log.Fatal("quota must be canonical uint128 decimal")
		}
		return value
	}
	initial, increaseAmount, decreaseAmount := parseQuota(*available), parseQuota(*increase), parseQuota(*decrease)
	var changedMode *bool
	switch *mode {
	case "keep":
	case "true":
		value := true
		changedMode = &value
	case "false":
		value := false
		changedMode = &value
	default:
		log.Fatal("new-unlimited must be keep, true or false")
	}
	if kind == "create" && *unlimited && initial.Sign() != 0 {
		log.Fatal("unlimited create requires --available-credit=0")
	}
	ctx := context.Background()
	accounts, err := GetAccounts(Settings.AccountsFile)
	if err != nil {
		log.Fatal(err)
	}
	if len(accounts) < 4 || actorIndex >= len(accounts) {
		log.Fatal("initialize at least four local accounts")
	}
	actor, alice, recipient := accounts[actorIndex], accounts[0], accounts[3]
	provider, err := ethclient.DialContext(ctx, Settings.ChainRPCURL)
	if err != nil {
		log.Fatal(err)
	}
	defer provider.Close()
	chain, err := provider.ChainID(ctx)
	if err != nil || chain.Uint64() != Settings.ChainID {
		log.Fatal("chain mismatch", err)
	}
	client, err := sdk.NewProtocolRpcClient(Settings.ProtocolRPCURL)
	if err != nil {
		log.Fatal(err)
	}
	pool, err := sdk.ProtocolFromRPC(ctx, client, provider)
	if err != nil {
		log.Fatal(err)
	}
	head, err := provider.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	token := common.HexToAddress(Settings.TokenAddress)
	coreContext := core.ProtocolContext{ChainID: core.BigIntToElement(chain), ProtocolAddress: core.BigIntToElement(pool.Context().ProtocolAddress.Big())}
	c := credit.Context{ProtocolContext: coreContext, TokenAddress: token, Timestamp: head.Time}
	deadline := new(big.Int).SetUint64(head.Time + 300)
	var prepared *credit.Prepared
	switch kind {
	case "create":
		contract, e := contracts.NewInvisiProtocolCaller(pool.Context().ProtocolAddress, provider)
		if e != nil {
			log.Fatal(e)
		}
		allocation, e := contract.CreditAllocation(&bind.CallOpts{Context: ctx, BlockNumber: head.Number}, actor.PublicWallet.OwnerAddress(), token)
		if e != nil {
			log.Fatal(e)
		}
		nonce, e := core.GenerateNonce()
		if e != nil {
			log.Fatal(e)
		}
		key, e := credit.GenerateSharedKey()
		if e != nil {
			log.Fatal(e)
		}
		state := credit.State{AvailableCredit: core.BigIntToElement(initial), Unlimited: *unlimited, ProtocolContext: coreContext, TokenAddress: core.BigIntToElement(token.Big()), OwnerWalletIdentity: core.BigIntToElement(actor.PublicWallet.OwnerAddress().Big()), SpenderWalletIdentity: core.BigIntToElement(alice.PublicWallet.OwnerAddress().Big()), RevocationShard: core.ElementFromUint64(allocation.Shard), RevocationSlot: core.ElementFromUint64(uint64(allocation.Slot)), Nonce: nonce, Expiry: core.ElementFromUint64(head.Time + 7*86400), SharedKey: key, Policy: credit.Policy{SingleLimit: core.ElementFromUint64(100), DailyLimit: core.ElementFromUint64(1000), MonthlyLimit: core.ElementFromUint64(10000)}}
		snapshot := credit.DeliveryState{State: state, OwnerReceivingKeys: core.ReceivingKeys{ViewingPublicKey: actor.PublicWallet.ViewingPublicKey(), ClueKey: actor.PublicWallet.ClueKey()}, SpenderReceivingKeys: core.ReceivingKeys{ViewingPublicKey: alice.PublicWallet.ViewingPublicKey(), ClueKey: alice.PublicWallet.ClueKey()}}
		ds, e := CreditDeliveries(snapshot)
		if e != nil {
			log.Fatal(e)
		}
		commitment, e := state.Commitment()
		if e != nil {
			log.Fatal(e)
		}
		prepared, err = credit.PrepareCreate(c, credit.CreateParams{Token: token, Owner: actor.PublicWallet.OwnerAddress(), RevocationShard: allocation.Shard, RevocationSlot: allocation.Slot, CreditCommitment: common.Hash(commitment.Bytes()), Deadline: deadline}, ds, state)
		fmt.Printf("Retain original revocation reference: owner=%s token=%s shard=%d slot=%d\n", actor.PublicWallet.OwnerAddress(), token, allocation.Shard, allocation.Slot)
	case "draw":
		candidates, e := DiscoverCredits(ctx, provider, client, pool, actor.PublicWallet, actor.ViewingKey, token, "spender")
		if e != nil {
			log.Fatal(e)
		}
		if len(candidates) == 0 {
			log.Fatal("no active grants; run owner creation examples and approve the Protocol for each payer first")
		}
		sort.Slice(candidates, func(i, j int) bool {
			a, b := candidates[i].Input.State, candidates[j].Input.State
			if x := a.OwnerWalletIdentity.Cmp(&b.OwnerWalletIdentity); x != 0 {
				return x < 0
			}
			if x := a.RevocationShard.Cmp(&b.RevocationShard); x != 0 {
				return x < 0
			}
			return a.RevocationSlot.Cmp(&b.RevocationSlot) < 0
		})
		if len(candidates) > 2 {
			candidates = candidates[:2]
		}
		p := credit.DrawParams{Token: token, TotalAmount: new(big.Int), Deadline: deadline}
		inputs := []credit.LocatedState{}
		ds := []protocol.Delivery{}
		sources := map[uint64]common.Hash{}
		debit := core.ElementFromUint64(7)
		for _, candidate := range candidates {
			s := candidate.Input.State
			if core.ElementToBigInt(s.Expiry).Cmp(deadline) < 0 {
				log.Fatal("selected grant expires before request deadline")
			}
			owner := common.BigToAddress(core.ElementToBigInt(s.OwnerWalletIdentity))
			pi := len(p.Payers) - 1
			if pi < 0 || p.Payers[pi].Payer != owner {
				p.Payers = append(p.Payers, credit.Payer{Payer: owner, Amount: new(big.Int)})
				pi++
			}
			p.Payers[pi].Amount.Add(p.Payers[pi].Amount, big.NewInt(7))
			next, e := s.Draw(debit, head.Time)
			if e != nil {
				log.Fatal(e)
			}
			p.GrantRefs = append(p.GrantRefs, credit.GrantRef{PayerIndex: big.NewInt(int64(pi)), Shard: core.ElementToBigInt(s.RevocationShard).Uint64(), Slot: uint16(core.ElementToBigInt(s.RevocationSlot).Uint64())})
			p.Nullifiers = append(p.Nullifiers, common.Hash(next.Nullifier.Bytes()))
			p.CreditCommitments = append(p.CreditCommitments, common.Hash(next.StateCommitment.Bytes()))
			p.RecoveryAmounts = append(p.RecoveryAmounts, common.Hash(next.RecoveryAmount.Bytes()))
			p.TotalAmount.Add(p.TotalAmount, big.NewInt(7))
			candidate.Input.Debit = debit
			inputs = append(inputs, candidate.Input)
			sources[candidate.Input.TreeID] = candidate.Membership.Root
			snapshot := candidate.Delivery
			snapshot.State = next.State
			deliveries, e := CreditDeliveries(snapshot)
			if e != nil {
				log.Fatal(e)
			}
			ds = append(ds, deliveries...)
		}
		ids := []uint64{}
		for id := range sources {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		for _, id := range ids {
			p.Sources = append(p.Sources, protocol.InputSource{TreeId: new(big.Int).SetUint64(id), MerkleRoot: sources[id]})
		}
		nonce, e := core.GenerateNonce()
		if e != nil {
			log.Fatal(e)
		}
		note := core.Note{ProtocolContext: coreContext, OwnerAddress: core.BigIntToElement(recipient.PublicWallet.OwnerAddress().Big()), TokenAddress: core.BigIntToElement(token.Big()), Amount: core.BigIntToElement(p.TotalAmount), Nonce: nonce}
		identity, e := core.ComputeNoteIdentity(note.OwnerAddress, note.Nonce)
		if e != nil {
			log.Fatal(e)
		}
		commitment, e := core.ComputeNoteCommitment(note.ChainID, identity, note.TokenAddress, note.Amount)
		if e != nil {
			log.Fatal(e)
		}
		p.NoteCommitment = common.Hash(commitment.Bytes())
		cipher, e := core.EncryptPrivateTransferNote(recipient.PublicWallet.ViewingPublicKey(), note, core.NoteEncryptionContext{ProtocolContext: coreContext, OutputCommitment: commitment})
		if e != nil {
			log.Fatal(e)
		}
		clue, e := core.CreateClue(recipient.PublicWallet.ClueKey())
		if e != nil {
			log.Fatal(e)
		}
		ds = append(ds, protocol.Delivery{Ciphertext: cipher, Clue: clue})
		prepared, err = credit.PrepareDraw(c, p, ds, inputs, identity)
		fmt.Printf("Gross payer debit=%s; ordinary Note recipient=%s\n", p.TotalAmount, recipient.PublicWallet.OwnerAddress())
	case "update":
		candidates, e := DiscoverCredits(ctx, provider, client, pool, actor.PublicWallet, actor.ViewingKey, token, "owner")
		if e != nil {
			log.Fatal(e)
		}
		if len(candidates) == 0 {
			log.Fatal("no active owner grant")
		}
		selected := candidates[0]
		old := selected.Input.State
		policy := old.Policy
		expiry := core.ElementFromUint64(head.Time + 14*86400)
		change := credit.Change{Policy: policy, Expiry: expiry, Unlimited: changedMode, IncreaseAvailableCredit: core.BigIntToElement(increaseAmount), DecreaseAvailableCredit: core.BigIntToElement(decreaseAmount)}
		next, e := old.Update(change, deadline.Uint64())
		if e != nil {
			log.Fatal(e)
		}
		snapshot := selected.Delivery
		snapshot.State = next.State
		ds, e := CreditDeliveries(snapshot)
		if e != nil {
			log.Fatal(e)
		}
		prepared, err = credit.PrepareUpdate(c, credit.UpdateParams{Token: token, Owner: actor.PublicWallet.OwnerAddress(), Source: protocol.InputSource{TreeId: new(big.Int).SetUint64(selected.Input.TreeID), MerkleRoot: selected.Membership.Root}, Nullifier: common.Hash(next.Nullifier.Bytes()), CreditCommitment: common.Hash(next.StateCommitment.Bytes()), RevocationShard: core.ElementToBigInt(old.RevocationShard).Uint64(), RevocationSlot: uint16(core.ElementToBigInt(old.RevocationSlot).Uint64()), Deadline: deadline}, ds, selected.Input, change)
	case "revoke":
		prepared, err = credit.PrepareRevoke(c, credit.RevokeParams{Token: token, Owner: actor.PublicWallet.OwnerAddress(), Shard: *shard, Slot: uint16(*slot), Deadline: deadline})
	default:
		log.Fatal("unknown Credit example")
	}
	if err != nil {
		log.Fatal(err)
	}
	if kind != "revoke" {
		typed, e := prepared.SigningRequest()
		if e != nil {
			log.Fatal(e)
		}
		digest, _, e := apitypes.TypedDataAndHash(typed)
		if e != nil {
			log.Fatal(e)
		}
		sig, e := crypto.Sign(digest, actor.Key)
		if e != nil {
			log.Fatal(e)
		}
		if e = prepared.WithAuthority(credit.Authority{PublicKey: crypto.FromECDSAPub(&actor.Key.PublicKey), Signature: sig[:64]}); e != nil {
			log.Fatal(e)
		}
		backend, e := native.NewNativeProver(Settings.Prover)
		if e != nil {
			log.Fatal(e)
		}
		if e = prepared.Prove(ctx, backend, prover.Options{}); e != nil {
			log.Fatal(e)
		}
	}
	data, err := prepared.Calldata(actor.PublicWallet.OwnerAddress())
	if err != nil {
		log.Fatal(err)
	}
	signer, err := bind.NewKeyedTransactorWithChainID(actor.Key, chain)
	if err != nil {
		log.Fatal(err)
	}
	signer.Context = ctx
	tx, err := bind.NewBoundContract(pool.Context().ProtocolAddress, abi.ABI{}, provider, provider, provider).RawTransact(signer, data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Submitted:", tx.Hash())
	receipt, err := bind.WaitMined(ctx, provider, tx)
	if err != nil {
		log.Fatal(err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Fatal("Credit operation reverted")
	}
	canonical, err := provider.HeaderByNumber(ctx, receipt.BlockNumber)
	if err != nil || canonical.Hash() != receipt.BlockHash {
		log.Fatal("receipt is no longer canonical; retain transaction hash and reconcile", err)
	}
	fmt.Println("Confirmed block:", receipt.BlockNumber)
}
