package provenance

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

type Client struct {
	Endpoint string
	Trust    TrustContext
	HTTP     *http.Client
	sequence atomic.Uint64
}
type Result map[string]json.RawMessage

func (c *Client) call(ctx context.Context, method string, query any) (Result, error) {
	u, e := url.Parse(c.Endpoint)
	if e != nil || u.Host == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("invalid service endpoint")
	}
	if !strings.HasPrefix(method, "provenance_") {
		return nil, errors.New("unsupported service method")
	}
	chain, e := Uint(c.Trust.ChainID, 64)
	if e != nil {
		return nil, e
	}
	if e := assertPublicSubmission(query); e != nil {
		return nil, e
	}
	id := c.sequence.Add(1)
	params := []any{}
	if query != nil {
		params = append(params, query)
	}
	body, e := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params})
	if e != nil {
		return nil, e
	}
	req, e := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(c.Endpoint, "/")+"/v1/chains/"+chain.String(), bytes.NewReader(body))
	if e != nil {
		return nil, e
	}
	req.Header.Set("Content-Type", "application/json")
	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	clientCopy := *client
	clientCopy.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	client = &clientCopy
	response, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("provenance HTTP %d", response.StatusCode)
	}
	var env struct {
		JSONRPC string `json:"jsonrpc"`
		ID      uint64 `json:"id"`
		Result  Result `json:"result"`
		Error   *struct {
			Message string `json:"message"`
			Data    struct {
				Code string `json:"code"`
			} `json:"data"`
		} `json:"error"`
	}
	if e := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(&env); e != nil {
		return nil, e
	}
	if env.JSONRPC != "2.0" || env.ID != id {
		return nil, errors.New("invalid JSON-RPC response")
	}
	if env.Error != nil {
		return nil, fmt.Errorf("provenance %s: %s", env.Error.Data.Code, env.Error.Message)
	}
	if method == "provenance_evaluateTransaction" || method == "provenance_submitProof" || method == "provenance_getProofJob" || method == "provenance_getTracking" {
		for _, key := range []string{"chainGeneration", "providerId", "signingKeyId", "manifestSha256"} {
			if len(env.Result[key]) == 0 {
				return nil, errors.New("missing response trust identity")
			}
		}
	}
	var scope Scope
	if len(env.Result["scope"]) == 0 {
		return nil, errors.New("missing response scope")
	}
	if e := json.Unmarshal(env.Result["scope"], &scope); e != nil {
		return nil, e
	}
	for _, p := range [][2]string{{scope.ChainID, c.Trust.ChainID}, {scope.ProtocolAddress, c.Trust.ProtocolAddress}, {scope.PolicyID, c.Trust.PolicyID}} {
		if !equalNumber(p[0], p[1]) {
			return nil, errors.New("deployment mismatch")
		}
	}
	for k, expected := range map[string]string{"chainGeneration": c.Trust.ChainGeneration, "providerId": c.Trust.ProviderID} {
		if raw, ok := env.Result[k]; ok {
			var v string
			if json.Unmarshal(raw, &v) != nil || !equalNumber(v, expected) {
				return nil, fmt.Errorf("%s mismatch", k)
			}
		}
	}
	if raw, ok := env.Result["manifestSha256"]; ok {
		var v string
		if json.Unmarshal(raw, &v) != nil || v != c.Trust.ManifestSHA256 {
			return nil, errors.New("manifest mismatch")
		}
	}
	if raw, ok := env.Result["signingKeyId"]; ok {
		var v string
		if json.Unmarshal(raw, &v) != nil {
			return nil, errors.New("invalid signing key")
		}
		key, e := word(v)
		if e != nil || (method != "provenance_getPolicy" && len(c.Trust.SigningKeys[key]) != 32) {
			return nil, errors.New("untrusted signing key")
		}
	}
	return env.Result, nil
}
func (c *Client) policyQuery(q map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range q {
		out[k] = v
	}
	out["policyId"], _ = word(c.Trust.PolicyID)
	return out
}
func (c *Client) GetVersion(ctx context.Context) (Result, error) {
	return c.call(ctx, "provenance_getVersion", c.policyQuery(nil))
}
func (c *Client) GetStatus(ctx context.Context) (Result, error) {
	return c.call(ctx, "provenance_getStatus", c.policyQuery(nil))
}
func (c *Client) GetPolicies(ctx context.Context, q map[string]any) (Result, error) {
	return c.call(ctx, "provenance_getPolicies", q)
}
func (c *Client) GetPolicy(ctx context.Context) (Result, error) {
	return c.call(ctx, "provenance_getPolicy", c.policyQuery(nil))
}
func (c *Client) GetRoots(ctx context.Context, q map[string]any) (Result, error) {
	r, e := c.call(ctx, "provenance_getRoots", c.policyQuery(q))
	if e != nil {
		return nil, e
	}
	var certs []RootCertificate
	if e := json.Unmarshal(r["items"], &certs); e != nil {
		return nil, e
	}
	for _, cert := range certs {
		if e := AssertCertificateTrust(cert, c.Trust); e != nil {
			return nil, e
		}
	}
	return r, nil
}

func (c *Client) GetRoot(ctx context.Context, rootID string) (Result, error) {
	r, e := c.call(ctx, "provenance_getRoot", c.policyQuery(map[string]any{"rootId": rootID}))
	if e != nil {
		return nil, e
	}
	if raw, ok := r["certificate"]; ok {
		var cert RootCertificate
		if e := json.Unmarshal(raw, &cert); e != nil {
			return nil, e
		}
		if e := AssertCertificateTrust(cert, c.Trust); e != nil {
			return nil, e
		}
	}
	return r, nil
}
func (c *Client) GetLeaves(ctx context.Context, rootID, from string, limit int) (Result, error) {
	if limit < 1 || limit > 200 {
		return nil, errors.New("page limit")
	}
	n, e := Uint(from, 64)
	if e != nil {
		return nil, e
	}
	return c.call(ctx, "provenance_getLeaves", c.policyQuery(map[string]any{"rootId": rootID, "fromIndex": n.String(), "limit": limit}))
}
func (c *Client) GetNoteRecognition(ctx context.Context, commitment string) (Result, error) {
	n, e := Field(commitment)
	if e != nil {
		return nil, e
	}
	v, _ := word(n.String())
	return c.call(ctx, "provenance_getNoteRecognition", c.policyQuery(map[string]any{"commitment": v}))
}
func (c *Client) GetOperationRecognition(ctx context.Context, locator map[string]any) (Result, error) {
	return c.call(ctx, "provenance_getOperationRecognition", c.policyQuery(locator))
}
func (c *Client) GetTransactionRecognition(ctx context.Context, tx string) (Result, error) {
	v, e := word(tx)
	if e != nil {
		return nil, e
	}
	return c.call(ctx, "provenance_getTransactionRecognition", c.policyQuery(map[string]any{"txHash": v}))
}
func (c *Client) EvaluateTransaction(ctx context.Context, data any, sidecars []ProofSidecar) (Result, error) {
	if e := assertPublicSubmission(data); e != nil {
		return nil, e
	}
	wireSidecars := make([]ProofSidecar, len(sidecars))
	for i, sidecar := range sidecars {
		if _, e := Uint(sidecar.OperationIndex, 32); e != nil {
			return nil, e
		}
		if e := ValidateProofBundle(sidecar.ProofBundle, c.Trust); e != nil {
			return nil, e
		}
		normalized, e := ProofBundleToWire(sidecar.ProofBundle)
		if e != nil {
			return nil, e
		}
		index, _ := Uint(sidecar.OperationIndex, 32)
		wireSidecars[i] = ProofSidecar{OperationIndex: index.String(), ProofBundle: normalized}
	}
	return c.call(ctx, "provenance_evaluateTransaction", c.policyQuery(map[string]any{"data": data, "proofSidecars": wireSidecars}))
}
func (c *Client) SubmitProof(ctx context.Context, locator map[string]any, bundle ProofBundle, idempotencyKey string) (Result, error) {
	if strings.HasPrefix(idempotencyKey, "internal/") {
		return nil, errors.New("reserved idempotency key prefix")
	}
	if idempotencyKey == "" {
		return nil, errors.New("idempotency key required")
	}
	if e := ValidateProofBundle(bundle, c.Trust); e != nil {
		return nil, e
	}
	normalized, e := ProofBundleToWire(bundle)
	if e != nil {
		return nil, e
	}
	q := c.policyQuery(locator)
	q["proofBundle"] = normalized
	q["idempotencyKey"] = idempotencyKey
	return c.call(ctx, "provenance_submitProof", q)
}
func (c *Client) GetProofJob(ctx context.Context, jobID string) (Result, error) {
	return c.call(ctx, "provenance_getProofJob", map[string]any{"jobId": jobID})
}
func (c *Client) WaitProofJob(ctx context.Context, jobID string, interval time.Duration) (Result, error) {
	if interval <= 0 || interval > time.Minute {
		return nil, errors.New("poll interval")
	}
	for {
		r, e := c.GetProofJob(ctx, jobID)
		if e != nil {
			return nil, e
		}
		var status string
		if e := json.Unmarshal(r["status"], &status); e != nil {
			return nil, e
		}
		switch status {
		case "valid", "invalid", "retryable_error", "stale_generation":
			return r, nil
		case "queued", "verifying":
		default:
			return nil, errors.New("unsupported proof job status")
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

type recognizedLeafItem struct {
	Scope              Scope  `json:"scope"`
	Commitment         string `json:"noteCommitment"`
	Leaf               string `json:"leaf"`
	Generation         string `json:"chainGeneration"`
	TreeID             string `json:"treeId"`
	LeafIndex          string `json:"leafIndex"`
	CheckpointSequence string `json:"checkpointSequence"`
}

func (c *Client) SyncRecognizedTree(ctx context.Context, rootID string, certificate RootCertificate, maxLeaves uint64) (*RecognizedTree, error) {
	if e := AssertCertificateTrust(certificate, c.Trust); e != nil {
		return nil, e
	}
	count, _ := Uint(certificate.LeafCount, 64)
	if maxLeaves == 0 || count.Uint64() > maxLeaves {
		return nil, errors.New("local tree replay exceeds configured leaf budget")
	}
	expectedCheckpoint, e := CheckpointHash(certificate)
	if e != nil {
		return nil, e
	}
	scope, e := ScopeHash(c.Trust.Scope)
	if e != nil {
		return nil, e
	}
	tree := &RecognizedTree{}
	for tree.LeafCount() < count.Uint64() {
		result, e := c.GetLeaves(ctx, rootID, fmt.Sprint(tree.LeafCount()), 200)
		if e != nil {
			return nil, e
		}
		var page struct {
			RootID      string               `json:"rootId"`
			Certificate RootCertificate      `json:"certificate"`
			Items       []recognizedLeafItem `json:"items"`
			NextIndex   string               `json:"nextIndex"`
		}
		data, _ := json.Marshal(result)
		if e := json.Unmarshal(data, &page); e != nil {
			return nil, e
		}
		checkpoint, e := CheckpointHash(page.Certificate)
		if e != nil {
			return nil, e
		}
		if page.RootID != rootID || checkpoint != expectedCheckpoint || len(page.Items) == 0 {
			return nil, errors.New("root snapshot changed or incomplete leaf page")
		}
		if e := AssertCertificateTrust(page.Certificate, c.Trust); e != nil {
			return nil, e
		}
		for _, item := range page.Items {
			for _, p := range [][2]string{{item.Scope.ChainID, c.Trust.ChainID}, {item.Scope.ProtocolAddress, c.Trust.ProtocolAddress}, {item.Scope.PolicyID, c.Trust.PolicyID}, {item.Generation, certificate.ChainGeneration}, {item.TreeID, certificate.TreeID}} {
				if !equalNumber(p[0], p[1]) {
					return nil, errors.New("recognized leaf deployment mismatch")
				}
			}
			index, e := Uint(item.LeafIndex, 32)
			if e != nil {
				return nil, e
			}
			sequence, e := Uint(item.CheckpointSequence, 64)
			if e != nil {
				return nil, e
			}
			maxSequence, _ := Uint(certificate.RootSequence, 64)
			if index.Uint64() != tree.LeafCount() || tree.LeafCount() >= count.Uint64() || sequence.Cmp(maxSequence) > 0 {
				return nil, errors.New("noncontiguous leaf page")
			}
			expected, e := RecognizedLeaf(scope, item.Commitment)
			if e != nil {
				return nil, e
			}
			if !equalNumber(expected, item.Leaf) {
				return nil, errors.New("recognized leaf hash mismatch")
			}
			if _, e := tree.Append(item.Leaf); e != nil {
				return nil, e
			}
		}
		if page.NextIndex != fmt.Sprint(tree.LeafCount()) {
			return nil, errors.New("noncontiguous leaf cursor")
		}
	}
	root, e := tree.Root()
	if e != nil {
		return nil, e
	}
	if !equalNumber(root, certificate.Root) {
		return nil, errors.New("rebuilt root mismatch")
	}
	return tree, nil
}

func (c *Client) GetMembership(ctx context.Context, commitment, rootID string, allowInterestDisclosure bool) (Result, error) {
	if !allowInterestDisclosure {
		return nil, errors.New("membership lookup requires explicit interest disclosure")
	}
	requested, e := Field(commitment)
	if e != nil {
		return nil, e
	}
	encoded, _ := word(requested.String())
	result, e := c.call(ctx, "provenance_getMembership", c.policyQuery(map[string]any{"commitment": encoded, "rootId": rootID}))
	if e != nil {
		return nil, e
	}
	var response struct {
		Certificate RootCertificate `json:"certificate"`
		Commitment  string          `json:"commitment"`
		LeafIndex   string          `json:"leafIndex"`
		Path        []string        `json:"path"`
	}
	data, _ := json.Marshal(result)
	if e := json.Unmarshal(data, &response); e != nil {
		return nil, e
	}
	if e := AssertCertificateTrust(response.Certificate, c.Trust); e != nil {
		return nil, e
	}
	if len(response.Path) != 32 || !equalNumber(response.Commitment, requested.String()) {
		return nil, errors.New("invalid membership response")
	}
	index, e := Uint(response.LeafIndex, 32)
	if e != nil {
		return nil, e
	}
	count, e := Uint(response.Certificate.LeafCount, 64)
	if e != nil {
		return nil, e
	}
	if index.Cmp(count) >= 0 {
		return nil, errors.New("membership index outside certified prefix")
	}
	scope, e := ScopeHash(c.Trust.Scope)
	if e != nil {
		return nil, e
	}
	leaf, e := RecognizedLeaf(scope, requested.String())
	if e != nil {
		return nil, e
	}
	for level, sibling := range response.Path {
		if index.Bit(level) == 0 {
			leaf, e = hash(leaf, sibling)
		} else {
			leaf, e = hash(sibling, leaf)
		}
		if e != nil {
			return nil, e
		}
	}
	if !equalNumber(leaf, response.Certificate.Root) {
		return nil, errors.New("membership root mismatch")
	}
	return result, nil
}
