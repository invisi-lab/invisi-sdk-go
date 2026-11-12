package provenance

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	protocol "github.com/invisi-lab/invisi-sdk-go/protocol/provenance"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const DefaultEndpoint = "https://provenance.invisipay.fi"

type officialEntry struct {
	done   chan struct{}
	client *Client
	err    error
}

// OfficialClients pins the official HTTPS identity independently for each deployment.
type OfficialClients struct {
	HTTP    *http.Client
	mu      sync.Mutex
	clients map[string]*officialEntry
}

func (c *OfficialClients) Get(ctx context.Context, scope Scope) (*Client, error) {
	chain, e := Uint(scope.ChainID, 64)
	if e != nil || chain.Sign() == 0 {
		return nil, errors.New("invalid Provenance deployment")
	}
	address, e := Uint(scope.ProtocolAddress, 160)
	if e != nil || address.Sign() == 0 {
		return nil, errors.New("invalid Provenance deployment")
	}
	canonical, _ := word(address.String())
	scope = Scope{ChainID: chain.String(), ProtocolAddress: canonical}
	key := scope.ChainID + "/" + canonical
	c.mu.Lock()
	if c.clients == nil {
		c.clients = map[string]*officialEntry{}
	}
	entry := c.clients[key]
	if entry != nil {
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-entry.done:
			return entry.client, entry.err
		}
	}
	if len(c.clients) >= 128 {
		c.mu.Unlock()
		return nil, errors.New("official Provenance deployment cache limit")
	}
	entry = &officialEntry{done: make(chan struct{})}
	c.clients[key] = entry
	c.mu.Unlock()
	entry.client, entry.err = bootstrapOfficial(ctx, scope, c.HTTP)
	c.mu.Lock()
	if entry.err != nil {
		delete(c.clients, key)
	}
	close(entry.done)
	c.mu.Unlock()
	return entry.client, entry.err
}
func bootstrapOfficial(ctx context.Context, scope Scope, transport *http.Client) (*Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	body, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "provenance_getRules", "params": []any{map[string]any{}}})
	req, e := http.NewRequestWithContext(ctx, "POST", DefaultEndpoint+"/v1/chains/"+scope.ChainID, bytes.NewReader(body))
	if e != nil {
		return nil, e
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	if transport != nil {
		client = *transport
	}
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Provenance bootstrap HTTP %d", response.StatusCode)
	}
	raw, e := io.ReadAll(io.LimitReader(response.Body, 256*1024+1))
	if e != nil {
		return nil, e
	}
	if len(raw) > 256*1024 {
		return nil, errors.New("Provenance bootstrap response too large")
	}
	var env struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      int             `json:"id"`
		Error   json.RawMessage `json:"error"`
		Result  *struct {
			Scope      Scope            `json:"scope"`
			Manifest   string           `json:"manifestSha256"`
			Generation string           `json:"chainGeneration"`
			Provider   string           `json:"providerId"`
			Initial    string           `json:"initialSigningPublicKey"`
			Current    string           `json:"signingKeyId"`
			Rotations  []RotationRecord `json:"keyRotations"`
		} `json:"result"`
	}
	if !utf8.Valid(raw) || json.Unmarshal(raw, &env) != nil || env.JSONRPC != "2.0" || env.ID != 1 || (len(env.Error) > 0 && string(env.Error) != "null") || env.Result == nil {
		return nil, errors.New("invalid Provenance bootstrap response")
	}
	rules := env.Result
	digest := sha256.Sum256(protocol.ManifestBytes())
	manifest := hex.EncodeToString(digest[:])
	generation, e := Uint(rules.Generation, 64)
	if !equalNumber(rules.Scope.ChainID, scope.ChainID) || !equalNumber(rules.Scope.ProtocolAddress, scope.ProtocolAddress) || rules.Manifest != manifest || e != nil || generation.Sign() == 0 || rules.Rotations == nil || len(rules.Rotations) > 128 {
		return nil, errors.New("Provenance bootstrap deployment or release mismatch")
	}
	initial, e := hex.DecodeString(strings.TrimPrefix(rules.Initial, "0x"))
	if e != nil || !strings.HasPrefix(rules.Initial, "0x") || len(initial) != 32 {
		return nil, errors.New("invalid initial Provenance signing key")
	}
	provider, e := word(rules.Provider)
	if e != nil {
		return nil, e
	}
	id := sha256.Sum256(initial)
	zero, _ := word("0")
	keyring, e := VerifyRotationChain(RotationCheckpoint{ProviderID: provider, Keys: map[string]ed25519.PublicKey{"0x" + hex.EncodeToString(id[:]): initial}, PreviousRotationHash: zero, LastActivatedAt: "0", Now: fmt.Sprint(time.Now().Unix())}, rules.Rotations)
	if e != nil {
		return nil, e
	}
	current, e := word(rules.Current)
	if e != nil || len(keyring.Keys[current]) != 32 || keyring.Validity[current].RetiredAt != "" {
		return nil, errors.New("untrusted active Provenance signing key")
	}
	return &Client{Endpoint: DefaultEndpoint, HTTP: transport, Trust: TrustContext{Scope: scope, ChainGeneration: rules.Generation, ProviderID: provider, ManifestSHA256: manifest, SigningKeys: keyring.Keys, SigningKeyValidity: keyring.Validity}}, nil
}
