package transaction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/provenance"
)

// ProvenanceCapability is explicitly local: only public proof evidence leaves the runtime.
type ProvenanceCapability struct {
	Client       *provenance.Client
	Manifest     provenance.ManifestIdentity
	Prover       provenance.OfflineProver
	MaxRootPages int
	MaxLeaves    uint64
}
type ProvenancePlan struct {
	PolicyID                                                          common.Hash
	Endpoint, Generation, ProviderID, ManifestSHA256, OperationDigest string
	Sidecars                                                          []provenance.ProofSidecar
	Tracking                                                          []provenance.Tracking
}

func (o *Execution) sourceCapability(policy common.Hash) (*ProvenanceCapability, error) {
	c := o.Runtime.Options.Provenance
	if c == nil || c.Client == nil || c.Prover == nil {
		return nil, &MissingCapabilityError{Capability: "local Provenance client, manifest and offline prover"}
	}
	t := c.Client.Trust
	if !sourceNumber(t.PolicyID, policy.Hex()) || !sourceNumber(t.ChainID, o.Plan.Context.ChainID.String()) || !sourceNumber(t.ProtocolAddress, o.Plan.Context.ProtocolAddress.Hex()) || c.Manifest.SHA256 != t.ManifestSHA256 {
		return nil, errors.New("provenance capability deployment/policy/manifest mismatch")
	}
	return c, nil
}
func sourceNumber(a, b string) bool {
	x, e := provenance.Uint(a, 256)
	y, f := provenance.Uint(b, 256)
	return e == nil && f == nil && x.Cmp(y) == 0
}
func (o *Execution) sourceDigest() (string, error) {
	w, e := o.Words()
	if e != nil {
		return "", e
	}
	values := make([]string, len(w))
	for i, v := range w {
		values[i] = common.BytesToHash(v[:]).Hex()
	}
	return provenance.OperationDigest(provenance.Scope{ChainID: o.Plan.Context.ChainID.String(), ProtocolAddress: o.Plan.Context.ProtocolAddress.Hex()}, o.Plan.Capacity, values)
}

type sourceTree struct {
	certificate provenance.RootCertificate
	tree        *provenance.RecognizedTree
}

func (o *Execution) prepareProvenance(ctx context.Context, policy common.Hash) (*ProvenancePlan, error) {
	if policy == (common.Hash{}) {
		return nil, nil
	}
	out := &ProvenancePlan{PolicyID: policy}
	if o.Plan.Kind != "transact" || len(o.Plan.Inputs) == 0 {
		return out, nil
	}
	c, e := o.sourceCapability(policy)
	if e != nil {
		return nil, e
	}
	trust := c.Client.Trust
	digest, e := o.sourceDigest()
	if e != nil {
		return nil, e
	}
	out.Endpoint = c.Client.Endpoint
	out.Generation = trust.ChainGeneration
	out.ProviderID = trust.ProviderID
	out.ManifestSHA256 = trust.ManifestSHA256
	out.OperationDigest = digest
	openings := make([]provenance.Opening, len(o.Plan.Inputs))
	commitments := make([]string, len(openings))
	if len(o.Plan.Params.Nullifiers) != len(openings) {
		return nil, errors.New("provenance input ordering mismatch")
	}
	for i, input := range o.Plan.Inputs {
		note := input.Note
		openings[i] = provenance.Opening{Owner: core.ElementToBigInt(note.OwnerAddress).String(), Nonce: core.ElementToBigInt(note.Nonce).String(), Amount: core.ElementToBigInt(note.Amount).String()}
		commitment, nf, err := provenance.NoteOpening(trust.Scope, o.Plan.Context.TokenAddress.Hex(), openings[i])
		if err != nil {
			return nil, err
		}
		if !sourceNumber(commitment, input.Commitment.Hex()) || !sourceNumber(nf, input.Nullifier.Hex()) || input.Nullifier != o.Plan.Params.Nullifiers[i] {
			return nil, errors.New("provenance opening/nullifier differs from frozen input")
		}
		commitments[i] = commitment
	}
	status, e := c.Client.GetStatus(ctx)
	if e != nil {
		return nil, e
	}
	var sequence string
	if e = json.Unmarshal(status["checkpointSequence"], &sequence); e != nil {
		return nil, e
	}
	if n, e := provenance.Uint(sequence, 64); e != nil || n.Sign() == 0 {
		return nil, errors.New("recognized source checkpoint unavailable")
	}
	maxPages := c.MaxRootPages
	if maxPages == 0 {
		maxPages = 10
	}
	budget := c.MaxLeaves
	if budget == 0 {
		budget = 100000
	}
	if maxPages < 1 || maxPages > 1000 || budget > 1<<32 {
		return nil, errors.New("invalid provenance replay budget")
	}
	latest, e := c.Client.GetRoot(ctx, trust.ChainGeneration+":"+sequence)
	if e != nil {
		return nil, e
	}
	var certificate provenance.RootCertificate
	if e = json.Unmarshal(latest["certificate"], &certificate); e != nil {
		return nil, e
	}
	var allowed bool
	if json.Unmarshal(latest["allowedForNewProof"], &allowed) != nil || !allowed {
		return nil, errors.New("source checkpoint not available for new proof")
	}
	selected := map[string]sourceTree{}
	remaining := budget
	load := func(cert provenance.RootCertificate) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := provenance.AssertCertificateTrust(cert, trust); err != nil {
			return err
		}
		height, err := provenance.Uint(cert.ConfirmedBlockNumber, 64)
		if err != nil {
			return err
		}
		header, err := o.Runtime.Options.Provider.HeaderByNumber(ctx, height)
		if err != nil {
			return err
		}
		if header == nil || header.Number == nil || header.Number.Cmp(height) != 0 || header.Hash() != common.HexToHash(cert.ConfirmedBlockHash) {
			return errors.New("provenance root anchor is not canonical on primary")
		}
		count, err := provenance.Uint(cert.LeafCount, 64)
		if err != nil || !count.IsUint64() || count.Uint64() > remaining {
			return errors.New("provenance tree replay budget exceeded")
		}
		tree, err := c.Client.SyncRecognizedTree(ctx, cert.ChainGeneration+":"+cert.RootSequence, cert, remaining)
		if err != nil {
			return err
		}
		remaining -= count.Uint64()
		treeID, err := provenance.Uint(cert.TreeID, 64)
		if err != nil {
			return err
		}
		cert.TreeID = treeID.String()
		selected[cert.TreeID] = sourceTree{cert, tree}
		return nil
	}
	treeID, e := provenance.Uint(certificate.TreeID, 64)
	if e != nil {
		return nil, e
	}
	certificate.TreeID = treeID.String()
	if !sourceNumber(certificate.RootSequence, sequence) {
		return nil, errors.New("source checkpoint differs from requested latest sequence")
	}
	if e = load(certificate); e != nil {
		return nil, e
	}
	find := func() (map[int]struct {
		tree       string
		membership provenance.Membership
	}, error) {
		found := map[int]struct {
			tree       string
			membership provenance.Membership
		}{}
		keys := make([]string, 0, len(selected))
		for key := range selected {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			a, _ := new(big.Int).SetString(keys[i], 10)
			b, _ := new(big.Int).SetString(keys[j], 10)
			return a.Cmp(b) < 0
		})
		for i, commitment := range commitments {
			for _, key := range keys {
				member, ok, err := selected[key].tree.MembershipForCommitment(trust.Scope, commitment)
				if err != nil {
					return nil, err
				}
				if ok {
					found[i] = struct {
						tree       string
						membership provenance.Membership
					}{key, member}
					break
				}
			}
		}
		return found, nil
	}
	found, e := find()
	if e != nil {
		return nil, e
	}
	if len(found) != len(openings) {
		newest := map[string]provenance.RootCertificate{certificate.TreeID: certificate}
		cursor := ""
		complete := false
		for page := 0; page < maxPages; page++ {
			query := map[string]any{"limit": "200"}
			if cursor == "" {
				query["after"] = "0"
			} else {
				query["cursor"] = cursor
			}
			response, err := c.Client.GetRoots(ctx, query)
			if err != nil {
				return nil, err
			}
			var certs []provenance.RootCertificate
			if err = json.Unmarshal(response["items"], &certs); err != nil {
				return nil, err
			}
			for _, cert := range certs {
				if err := provenance.AssertCertificateTrust(cert, trust); err != nil {
					return nil, err
				}
				treeID, err := provenance.Uint(cert.TreeID, 64)
				if err != nil {
					return nil, err
				}
				cert.TreeID = treeID.String()
				old, ok := newest[cert.TreeID]
				a, err := provenance.Uint(cert.RootSequence, 64)
				if err != nil {
					return nil, err
				}
				b, _ := provenance.Uint(old.RootSequence, 64)
				if !ok || a.Cmp(b) > 0 {
					newest[cert.TreeID] = cert
				}
			}
			cursor = ""
			if value, ok := response["nextCursor"]; ok {
				if err = json.Unmarshal(value, &cursor); err != nil {
					return nil, err
				}
			}
			if cursor == "" {
				complete = true
				break
			}
		}
		if !complete {
			return nil, errors.New("provenance root directory replay budget exceeded")
		}
		keys := make([]string, 0, len(newest))
		for key := range newest {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			a, _ := provenance.Uint(keys[i], 64)
			b, _ := provenance.Uint(keys[j], 64)
			return a.Cmp(b) < 0
		})
		for _, key := range keys {
			cert := newest[key]
			if old, ok := selected[key]; ok && old.certificate.RootSequence == cert.RootSequence {
				continue
			}
			if e = load(cert); e != nil {
				return nil, e
			}
		}
		found, e = find()
		if e != nil {
			return nil, e
		}
	}
	if len(found) != len(openings) {
		return nil, errors.New("private input Note is not recognized under required policy")
	}
	used := map[string]bool{}
	for _, item := range found {
		used[item.tree] = true
	}
	keys := make([]string, 0, len(used))
	for key := range used {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, _ := provenance.Uint(keys[i], 64)
		b, _ := provenance.Uint(keys[j], 64)
		return a.Cmp(b) < 0
	})
	indices := map[string]int{}
	certs := []provenance.RootCertificate{}
	sources := []provenance.Source{}
	for i, key := range keys {
		indices[key] = i
		cert := selected[key].certificate
		certs = append(certs, cert)
		sources = append(sources, provenance.Source{TreeID: cert.TreeID, Root: cert.Root, LeafCount: cert.LeafCount})
	}
	inputs := make([]provenance.WitnessInput, len(openings))
	for i, opening := range openings {
		member := found[i]
		inputs[i] = provenance.WitnessInput{Opening: opening, SourceIndex: indices[member.tree], LeafIndex: member.membership.LeafIndex, Path: member.membership.Path}
	}
	capacity := 0
	for k := 1; k <= 190; k++ {
		if provenance.SupportedCapacity(k) && k >= len(inputs) {
			capacity = k
			break
		}
	}
	if capacity == 0 {
		return nil, errors.New("unsupported provenance input capacity")
	}
	sidecar, e := provenance.ExportSidecar(ctx, provenance.WitnessRequest{Scope: trust.Scope, Capacity: capacity, TokenAddress: o.Plan.Context.TokenAddress.Hex(), ChainGeneration: trust.ChainGeneration, OperationDigest: digest, Inputs: inputs, Sources: sources}, certs, trust, c.Manifest, c.Prover, 0)
	if e != nil {
		return nil, e
	}
	out.Sidecars = []provenance.ProofSidecar{sidecar}
	return out, nil
}
func (o *Execution) acceptProvenanceTracking(ctx context.Context, saved *ProvenancePlan, ack *provenance.Tracking) error {
	if saved == nil || len(saved.Sidecars) == 0 {
		if ack != nil {
			return errors.New("unexpected source tracking on operation without private inputs")
		}
		return nil
	}
	if ack == nil {
		return errors.New("broadcaster quote omitted provenance tracking")
	}
	raw, e := json.Marshal(ack)
	if e != nil {
		return e
	}
	quoted, e := provenance.TrackingFromResult(provenance.Result{"tracking": raw}, ack.ID)
	if e != nil {
		return e
	}
	if e = allowedSourceTracking(quoted, saved.OperationDigest); e != nil {
		return e
	}
	c, e := o.sourceCapability(saved.PolicyID)
	if e != nil {
		return e
	}
	result, e := c.Client.GetTracking(ctx, quoted.ID)
	if e != nil {
		return e
	}
	current, e := provenance.TrackingFromResult(result, quoted.ID)
	if e != nil {
		return e
	}
	if e = allowedSourceTracking(current, saved.OperationDigest); e != nil {
		return e
	}
	saved.Tracking = []provenance.Tracking{current}
	return nil
}
func allowedSourceTracking(t provenance.Tracking, digest string) error {
	if t.OperationDigest != digest {
		return errors.New("source tracking statement mismatch")
	}
	switch t.Status {
	case "pending_confirmation":
		expires, e := provenance.Uint(t.ExpiresAt, 64)
		if e != nil || expires.Cmp(big.NewInt(time.Now().Unix())) <= 0 {
			return errors.New("source tracking acknowledgement expired")
		}
	case "pending_source", "recognized":
	default:
		return errors.New("source tracking no longer permits broadcast")
	}
	return nil
}

func (o *Execution) validateProvenance(ctx context.Context, saved *ProvenancePlan, remote bool) error {
	if saved == nil {
		return nil
	}
	if len(saved.Sidecars) == 0 {
		if len(o.Plan.Inputs) > 0 && o.Plan.Kind == "transact" {
			return errors.New("missing stored provenance proof")
		}
		return nil
	}
	c, e := o.sourceCapability(saved.PolicyID)
	if e != nil {
		return e
	}
	trust := c.Client.Trust
	if c.Client.Endpoint != saved.Endpoint || trust.ChainGeneration != saved.Generation || !sourceNumber(trust.ProviderID, saved.ProviderID) || trust.ManifestSHA256 != saved.ManifestSHA256 {
		return errors.New("stored provenance trust changed; explicit preparation required")
	}
	digest, e := o.sourceDigest()
	if e != nil || digest != saved.OperationDigest {
		return errors.New("stored provenance statement changed")
	}
	if len(saved.Sidecars) != 1 || len(saved.Tracking) != 1 {
		return errors.New("stored provenance attachments missing")
	}
	if e = provenance.ValidateProofBundle(saved.Sidecars[0].ProofBundle, trust); e != nil {
		return e
	}
	bundle := saved.Sidecars[0].ProofBundle
	if saved.Sidecars[0].OperationIndex != "0" || !sourceNumber(bundle.PublicInputs[2], o.Plan.Context.TokenAddress.Hex()) || !sourceNumber(bundle.PublicInputs[8], fmt.Sprint(len(o.Plan.Inputs))) {
		return errors.New("stored provenance operation/token/input count mismatch")
	}
	nullifiers := make([]string, len(o.Plan.Params.Nullifiers))
	for i, n := range o.Plan.Params.Nullifiers {
		nullifiers[i] = common.Hash(n).Hex()
	}
	inputDigest, err := provenance.InputDigest(nullifiers)
	if err != nil {
		return err
	}
	if !sourceNumber(bundle.PublicInputs[11], inputDigest) {
		return errors.New("stored provenance ordered nullifiers mismatch")
	}
	trackingData, err := json.Marshal(saved.Tracking[0])
	if err != nil {
		return err
	}
	tracking, err := provenance.TrackingFromResult(provenance.Result{"tracking": trackingData}, saved.Tracking[0].ID)
	if err != nil || tracking.OperationDigest != digest || (tracking.Status != "pending_confirmation" && tracking.Status != "pending_source" && tracking.Status != "recognized") {
		return errors.New("stored provenance tracking mismatch")
	}

	if !sourceNumber(bundle.PublicInputs[9], new(big.Int).Rsh(common.HexToHash(digest).Big(), 128).String()) || !sourceNumber(bundle.PublicInputs[10], new(big.Int).And(common.HexToHash(digest).Big(), new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))).String()) {
		return errors.New("stored provenance digest mismatch")
	}
	if remote {
		for _, cert := range bundle.RootCertificates {
			height, err := provenance.Uint(cert.ConfirmedBlockNumber, 64)
			if err != nil {
				return err
			}
			header, err := o.Runtime.Options.Provider.HeaderByNumber(ctx, height)
			if err != nil {
				return err
			}
			if header == nil || header.Number == nil || header.Number.Cmp(height) != 0 || header.Hash() != common.HexToHash(cert.ConfirmedBlockHash) {
				return errors.New("stored provenance root anchor is no longer canonical on primary")
			}
		}
		if _, e = c.Client.GetStatus(ctx); e != nil {
			return e
		}
		for _, tracking := range saved.Tracking {
			result, err := c.Client.GetTracking(ctx, tracking.ID)
			if err != nil {
				return err
			}
			current, err := provenance.TrackingFromResult(result, tracking.ID)
			if err != nil {
				return err
			}
			if current.OperationDigest != digest || current.Status == "expired" || current.Status == "superseded" {
				return fmt.Errorf("stored provenance tracking cannot be submitted: %s", current.Status)
			}
		}
	}
	return ctx.Err()
}

func (o *Execution) exportProvenanceTracking() ([]provenance.Tracking, error) {
	if o.Plan.Broadcaster == nil || o.Plan.Broadcaster.Provenance == nil {
		return nil, errors.New("provenance tracking unavailable")
	}
	return Frozen(o.Plan.Broadcaster.Provenance.Tracking)
}

// ExportProvenanceTracking returns detached public acknowledgements; local proving capabilities are never exported.
func (o *Execution) ExportProvenanceTracking(ctx context.Context) ([]provenance.Tracking, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.CheckLease(ctx); e != nil {
		return nil, e
	}
	return o.exportProvenanceTracking()
}
func (p *PreparedTransaction) ExportProvenanceTracking(ctx context.Context) ([]provenance.Tracking, error) {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	if e := p.current(ctx); e != nil {
		return nil, e
	}
	return p.operation.exportProvenanceTracking()
}
