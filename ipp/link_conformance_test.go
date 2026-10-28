package ipp

// Conformance run against the shared 05-ipp/5.6 fixture.
//
// This file exists for one failure shape that nothing else in either SDK can
// see. Six times now, this implementation and packages/ipp/src/link.ts have
// both answered a URL without erroring and answered it DIFFERENTLY. Each side's
// own suite passes in that situation no matter how dense it is, because "the
// other side says otherwise" is not a fact either suite has access to. Reading
// the other side's source found one of the six and cost 661 lines of review;
// the other five were only visible as two verdicts on one input.
//
// So: both sides read THE SAME BYTES and record accept/reject. The fixture is
// the medium for comparing answers, never for copying a decision -- an
// implementation written from the fixture's regexes instead of from the 6.6
// prose destroys the only property that made the six findings possible, which
// is that two readers hit the same ambiguity independently.
//
// testdata/link-transport-v1.json is a byte-identical copy of
// protocol-docs/05-ipp/conformance/link-transport-v1.json and is NEVER edited
// here. protocol-docs/05-ipp/conformance/vendored-copies.mjs fails when the two
// differ; --write is the only sanctioned way to change this copy.
//
// Verdicts only, on purpose. Pinning the REASON needs a language-neutral name
// per rule, which does not exist yet; every one of the six divergences shows up
// in the verdict alone.

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

const conformanceFixturePath = "testdata/link-transport-v1.json"

// publicStubAddress is what a name with no `resolves_to` in the fixture answers.
// The fixture lists it under host_policy.public_addresses, so it is public by
// the same document that supplies the cases.
const publicStubAddress = "93.184.216.34"

// ── fixture shape ───────────────────────────────────────────────────

type conformanceFixture struct {
	FixtureVersion int    `json:"fixture_version"`
	IPPVersion     int    `json:"ipp_version"`
	Spec           string `json:"spec"`

	HostPolicy struct {
		PublicAddresses []struct {
			Address string `json:"address"`
			Why     string `json:"why"`
		} `json:"public_addresses"`
	} `json:"host_policy"`

	Envelope struct {
		ErrorEnvelopeShapePairs []struct {
			Code   string `json:"code"`
			Flat   string `json:"flat"`
			Nested string `json:"nested"`
		} `json:"error_envelope_shape_pairs"`
	} `json:"envelope"`

	Messages map[string]json.RawMessage `json:"messages"`

	CodecPairs []struct {
		ID         string `json:"id"`
		MessageRef string `json:"message_ref"`
		Compact    string `json:"compact"`
		Note       string `json:"note"`
	} `json:"codec_pairs"`

	LinkURLs struct {
		Accept []json.RawMessage `json:"accept"`
		Reject []json.RawMessage `json:"reject"`
	} `json:"link_urls"`

	RequiredControls struct {
		IDs []struct {
			ID  string `json:"id"`
			Why string `json:"why"`
		} `json:"ids"`
	} `json:"required_controls"`

	Resolve []json.RawMessage `json:"resolve"`
}

type conformanceAccept struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Token    string `json:"token"`
	Endpoint string `json:"endpoint"`
	Note     string `json:"note"`
}

type conformanceReject struct {
	ID          string   `json:"id"`
	URL         string   `json:"url"`
	URLTemplate string   `json:"url_template"`
	Codepoint   int      `json:"codepoint"`
	Rule        int      `json:"rule"`
	RejectedBy  string   `json:"rejected_by"`
	Reason      string   `json:"reason"`
	Malicious   bool     `json:"malicious"`
	Isolates    bool     `json:"isolates"`
	Requires    string   `json:"requires"`
	ResolvesTo  []string `json:"resolves_to"`
	LocalName   bool     `json:"local_name"`
}

type conformanceResolve struct {
	ID        string                     `json:"id"`
	URL       string                     `json:"url"`
	Response  json.RawMessage            `json:"response"`
	Expect    map[string]json.RawMessage `json:"expect"`
	Malicious bool                       `json:"malicious"`
}

// conformanceResponse is the stub response a resolve case declares. The body
// arrives in one of five carriers; body_text_template is the one that exists so
// an invisible character can be given as a code point instead of a literal, the
// same reason url_template exists.
type conformanceResponse struct {
	Status           int               `json:"status"`
	Headers          map[string]string `json:"headers"`
	BodyJSONRef      string            `json:"body_json_ref"`
	BodyJSON         json.RawMessage   `json:"body_json"`
	BodyText         string            `json:"body_text"`
	BodyTextTemplate string            `json:"body_text_template"`
	BodyBytesHex     string            `json:"body_bytes_hex"`
	Codepoint        int               `json:"codepoint"`
}

func loadConformanceFixture(t *testing.T) *conformanceFixture {
	t.Helper()
	raw, err := os.ReadFile(conformanceFixturePath)
	if err != nil {
		// Never a skip. A missing fixture means the cross-implementation check
		// did not run, and "did not run" must not read as "the two sides agree".
		t.Fatalf("reading %s: %v (regenerate with protocol-docs/05-ipp/conformance/vendored-copies.mjs --write)", conformanceFixturePath, err)
	}
	var fx conformanceFixture
	if err := json.Unmarshal(raw, &fx); err != nil {
		t.Fatalf("parsing %s: %v", conformanceFixturePath, err)
	}
	return &fx
}

// decodeCase fills dst and reports any key the runner does not consume. An
// unrecognised key is a failure, not a skip: a case gaining an expectation this
// runner silently ignores is a green test that stopped checking something.
func decodeCase(t *testing.T, raw json.RawMessage, dst any, consumed ...string) {
	t.Helper()
	if err := json.Unmarshal(raw, dst); err != nil {
		t.Fatalf("decoding case %s: %v", raw, err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keys); err != nil {
		t.Fatalf("decoding case keys %s: %v", raw, err)
	}
	known := make(map[string]bool, len(consumed))
	for _, k := range consumed {
		known[k] = true
	}
	for k := range keys {
		if !known[k] {
			t.Errorf("case %s carries key %q this runner does not handle; handle it or the case is asserting nothing", raw, k)
		}
	}
}

// ── stubs ───────────────────────────────────────────────────────────

// conformanceLookup answers from the fixture. A name with `resolves_to` gets
// exactly those addresses; a name the fixture marks `local_name` gets loopback,
// which is what it answers on every real machine; anything else gets a public
// address. Nothing here reaches the network.
func conformanceLookup(resolvesTo []string, localName bool) LinkOption {
	answers := resolvesTo
	if len(answers) == 0 {
		if localName {
			answers = []string{"127.0.0.1"}
		} else {
			answers = []string{publicStubAddress}
		}
	}
	return func(c *linkConfig) {
		c.lookupIP = func(context.Context, string) ([]net.IP, error) {
			out := make([]net.IP, 0, len(answers))
			for _, s := range answers {
				ip := parseIPAddress(s)
				if ip == nil {
					return nil, fmt.Errorf("fixture address %q is not an IP", s)
				}
				out = append(out, ip)
			}
			return out, nil
		}
	}
}

func conformanceTransport(status int, header http.Header, body []byte) *countingRoundTripper {
	return &countingRoundTripper{
		fn: func(*http.Request) (*http.Response, error) {
			return stringResponse(status, string(body), header), nil
		},
	}
}

// ── the fixture copy itself ─────────────────────────────────────────

func TestConformanceFixtureIsIntact(t *testing.T) {
	fx := loadConformanceFixture(t)

	if fx.Spec != "05-ipp/5.6-transport-bindings.md" {
		t.Errorf("fixture spec = %q, want 05-ipp/5.6-transport-bindings.md", fx.Spec)
	}
	if fx.IPPVersion != IPPCurrentVersion {
		t.Errorf("fixture ipp_version = %d, want %d", fx.IPPVersion, IPPCurrentVersion)
	}
	if fx.FixtureVersion < 9 {
		t.Errorf("fixture_version = %d; this copy predates the version this runner was written against", fx.FixtureVersion)
	}

	// Every case id the fixture declares mandatory must still be in this copy.
	// The point of the list is that a MISSING positive control leaves a hole no
	// per-rule count reports: "reject any address with two equal zero runs"
	// passes every reject case in the file on its own.
	present := map[string]bool{}
	for _, raw := range fx.LinkURLs.Accept {
		var c conformanceAccept
		if err := json.Unmarshal(raw, &c); err == nil {
			present[c.ID] = true
		}
	}
	for _, raw := range fx.LinkURLs.Reject {
		var c conformanceReject
		if err := json.Unmarshal(raw, &c); err == nil {
			present[c.ID] = true
		}
	}
	for _, raw := range fx.Resolve {
		var c conformanceResolve
		if err := json.Unmarshal(raw, &c); err == nil {
			present[c.ID] = true
		}
	}
	for _, want := range fx.RequiredControls.IDs {
		if !present[want.ID] {
			t.Errorf("required control %q is missing from this copy of the fixture (%s)", want.ID, want.Why)
		}
	}

	// Every input string must be printable ASCII. The reject cases that carry an
	// invisible character supply it as `url_template` + `codepoint` for exactly
	// this reason: a literal U+200B or U+FEFF in a JSON file is one an editor
	// doing Unicode normalisation deletes on save, which leaves the case green
	// and its rule unguarded. A literal appearing here means that happened in
	// reverse -- someone inlined one.
	assertPrintableASCII := func(id, field, value string) {
		for i, r := range value {
			if r < 0x21 || r > 0x7e {
				t.Errorf("case %s field %s byte %d is U+%04X; inputs must be printable ASCII and invisible characters must ride url_template+codepoint", id, field, i, r)
			}
		}
	}
	for _, raw := range fx.LinkURLs.Accept {
		var c conformanceAccept
		if err := json.Unmarshal(raw, &c); err != nil {
			continue
		}
		assertPrintableASCII(c.ID, "url", c.URL)
		assertPrintableASCII(c.ID, "endpoint", c.Endpoint)
		assertPrintableASCII(c.ID, "token", c.Token)
	}
	for _, raw := range fx.LinkURLs.Reject {
		var c conformanceReject
		if err := json.Unmarshal(raw, &c); err != nil {
			continue
		}
		assertPrintableASCII(c.ID, "url", c.URL)
		assertPrintableASCII(c.ID, "url_template", c.URLTemplate)
	}
}

// ── link_urls.accept ────────────────────────────────────────────────

func TestConformanceLinkURLsAccept(t *testing.T) {
	fx := loadConformanceFixture(t)
	if len(fx.LinkURLs.Accept) == 0 {
		t.Fatal("fixture carries no accept cases")
	}

	for _, raw := range fx.LinkURLs.Accept {
		var c conformanceAccept
		decodeCase(t, raw, &c, "id", "url", "token", "endpoint", "note")
		t.Run(c.ID, func(t *testing.T) {
			// The served payload.endpoint is the endpoint the FIXTURE says this
			// URL derives. So rule 9 is a second, independent check on the same
			// expectation: derive anything else and the cross-check refuses.
			body, err := Encode(sampleLinkRequest(c.Endpoint), FormatJSON)
			if err != nil {
				t.Fatalf("building response body: %v", err)
			}
			rt := conformanceTransport(200, http.Header{"Content-Type": []string{"application/json"}}, []byte(body))

			_, endpoint, err := ResolveLink(context.Background(), c.URL,
				clientWith(rt), conformanceLookup(nil, false))
			if err != nil {
				t.Fatalf("MUST-ACCEPT case was rejected: %v", err)
			}
			if endpoint != c.Endpoint {
				t.Errorf("derived endpoint = %q, fixture says %q", endpoint, c.Endpoint)
			}
			if rt.calls != 1 {
				t.Fatalf("round trips = %d, want 1", rt.calls)
			}
			// The URL goes out AS SCANNED, byte for byte. Only the derived
			// endpoint is normalised, because only it is compared. The two SDKs
			// disagreed here until 6.6 pinned it: this side rebuilt the request
			// from the normal form, so "https://Pay.Example.COM:443/..." left as
			// "https://pay.example.com/...", and the receiver's access log held a
			// different line than the other SDK produced for one QR code.
			//
			// This also covers the token, which is the one part of the URL that
			// is neither case-folded nor percent-decoded (the fixture pins
			// case_sensitive_token): repairing it would request a resource that
			// does not exist while every other assertion here still passed.
			if got := rt.last.URL.String(); got != c.URL {
				t.Errorf("requested %q, want the URL as scanned %q", got, c.URL)
			}
			if !strings.HasSuffix(c.URL, "/ipp/v1/link/"+c.Token) {
				t.Errorf("fixture case is inconsistent: url %q does not end in token %q", c.URL, c.Token)
			}
		})
	}
}

// ── link_urls.reject ────────────────────────────────────────────────

func TestConformanceLinkURLsReject(t *testing.T) {
	fx := loadConformanceFixture(t)
	if len(fx.LinkURLs.Reject) == 0 {
		t.Fatal("fixture carries no reject cases")
	}

	for _, raw := range fx.LinkURLs.Reject {
		var c conformanceReject
		decodeCase(t, raw, &c,
			"id", "url", "url_template", "codepoint", "rule", "rejected_by",
			"reason", "malicious", "isolates", "requires", "resolves_to", "local_name")
		t.Run(c.ID, func(t *testing.T) {
			url, err := c.input()
			if err != nil {
				t.Fatal(err)
			}
			// fn is nil, so a round trip fails. That failure must NOT be read
			// as the rejection: a URL that got past every pre-flight rule would
			// then be reported as refused. Reaching the transport IS the
			// acceptance, and the call count is what says so.
			rt := &countingRoundTripper{}
			_, _, err = ResolveLink(context.Background(), url,
				clientWith(rt), conformanceLookup(c.ResolvesTo, c.LocalName))
			// reject_contract.http_calls_expected is 0 for every case in this
			// section: rules 1, 2 and 4 all decide before the first byte leaves
			// the process. A rejection that cost a request means a check moved
			// below the fetch, which turns "scan any QR code" into "GET any URL".
			if rt.calls != 0 {
				t.Fatalf("MUST-REJECT case reached the transport after %d request(s), so every pre-flight rule let it through (rule %d, %s)", rt.calls, c.Rule, c.Reason)
			}
			if err == nil {
				t.Fatalf("MUST-REJECT case was ACCEPTED (rule %d, %s)", c.Rule, c.Reason)
			}
		})
	}
}

// input builds the case's URL, substituting {CP} with the declared code point.
func (c *conformanceReject) input() (string, error) {
	switch {
	case c.URL != "" && c.URLTemplate != "":
		return "", fmt.Errorf("case %s carries both url and url_template", c.ID)
	case c.URL != "":
		return c.URL, nil
	case c.URLTemplate != "":
		if c.Codepoint == 0 {
			return "", fmt.Errorf("case %s has url_template but no codepoint", c.ID)
		}
		if !strings.Contains(c.URLTemplate, "{CP}") {
			return "", fmt.Errorf("case %s url_template has no {CP} slot", c.ID)
		}
		return strings.Replace(c.URLTemplate, "{CP}", string(rune(c.Codepoint)), 1), nil
	default:
		return "", fmt.Errorf("case %s carries neither url nor url_template", c.ID)
	}
}

// ── host_policy.public_addresses ────────────────────────────────────

// TestConformanceHostPolicyPublicAddresses is the half of rule 4 a reject-only
// suite cannot see. Every one of these addresses sits immediately outside a
// blocked prefix, so a prefix written one bit too wide -- 192.0.0.0/24 typed as
// /16, CGNAT typed as 100/8 -- still refuses every MUST-REJECT case in the file
// and only shows up here. That exact widening really shipped on the other side,
// and it survived two rounds of reading the list against the prose because the
// list, the comment and the report all said /24 while the code said /16.
func TestConformanceHostPolicyPublicAddresses(t *testing.T) {
	fx := loadConformanceFixture(t)
	if len(fx.HostPolicy.PublicAddresses) == 0 {
		t.Fatal("fixture carries no public_addresses")
	}
	token := conformanceToken(t, fx)

	for _, pub := range fx.HostPolicy.PublicAddresses {
		t.Run(pub.Address, func(t *testing.T) {
			authority := pub.Address
			if strings.Contains(authority, ":") {
				authority = "[" + authority + "]"
			}
			endpoint := "https://" + authority
			linkURL := endpoint + "/ipp/v1/link/" + token

			body, err := Encode(sampleLinkRequest(endpoint), FormatJSON)
			if err != nil {
				t.Fatalf("building response body: %v", err)
			}
			rt := conformanceTransport(200, http.Header{"Content-Type": []string{"application/json"}}, []byte(body))

			_, got, err := ResolveLink(context.Background(), linkURL,
				clientWith(rt), conformanceLookup(nil, false))
			if err != nil {
				t.Fatalf("public address was refused (%s): %v", pub.Why, err)
			}
			if got != endpoint {
				t.Errorf("endpoint = %q, want %q", got, endpoint)
			}
		})
	}
}

// conformanceToken borrows a legal token from the fixture rather than declaring
// one here: a token spelled in this file could drift out of the grammar the
// fixture pins.
func conformanceToken(t *testing.T, fx *conformanceFixture) string {
	t.Helper()
	for _, raw := range fx.LinkURLs.Accept {
		var c conformanceAccept
		if err := json.Unmarshal(raw, &c); err == nil && c.Token != "" {
			return c.Token
		}
	}
	t.Fatal("fixture carries no accept case with a token")
	return ""
}

// ── codec_pairs ─────────────────────────────────────────────────────

func TestConformanceCodecPairs(t *testing.T) {
	fx := loadConformanceFixture(t)
	if len(fx.CodecPairs) == 0 {
		t.Fatal("fixture carries no codec pairs")
	}
	for _, pair := range fx.CodecPairs {
		t.Run(pair.ID, func(t *testing.T) {
			ref, ok := fx.Messages[pair.MessageRef]
			if !ok {
				t.Fatalf("message_ref %q is not in the fixture", pair.MessageRef)
			}
			wantMsg, err := Decode(string(ref))
			if err != nil {
				t.Fatalf("decoding messages[%s]: %v", pair.MessageRef, err)
			}
			gotMsg, err := Decode(pair.Compact)
			if err != nil {
				t.Fatalf("decoding compact carrier: %v", err)
			}
			// Value, never the re-encoded string: MessagePack map ordering is
			// implementation-defined, so a byte comparison would fail for a
			// reason that is not a divergence.
			if !reflect.DeepEqual(gotMsg, wantMsg) {
				t.Errorf("compact carrier decoded to a different envelope\n got: %+v\nwant: %+v", gotMsg, wantMsg)
			}
		})
	}
}

// ── resolve ─────────────────────────────────────────────────────────

func TestConformanceResolve(t *testing.T) {
	fx := loadConformanceFixture(t)
	if len(fx.Resolve) == 0 {
		t.Fatal("fixture carries no resolve cases")
	}

	for _, raw := range fx.Resolve {
		var c conformanceResolve
		decodeCase(t, raw, &c, "id", "url", "response", "expect", "malicious")
		t.Run(c.ID, func(t *testing.T) {
			r := c.response(t)
			body := c.body(t, fx, r)
			header := http.Header{}
			for k, v := range r.Headers {
				header.Set(k, v)
			}
			rt := conformanceTransport(r.Status, header, body)

			request, endpoint, err := ResolveLink(context.Background(), c.URL,
				clientWith(rt), conformanceLookup(nil, false))

			c.assertExpectations(t, fx, request, endpoint, err, rt)
		})
	}
}

// response decodes the case's stub response and reports any key this runner does
// not consume. The key check matters here specifically: validate.mjs accepts a
// body_text_template carrier that this runner did not originally handle, so a
// case using it would have been fed an EMPTY body, rejected for that reason, and
// reported as a pass for the rule it names.
func (c *conformanceResolve) response(t *testing.T) *conformanceResponse {
	t.Helper()
	var r conformanceResponse
	if err := json.Unmarshal(c.Response, &r); err != nil {
		t.Fatalf("case %s: decoding response: %v", c.ID, err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(c.Response, &keys); err != nil {
		t.Fatalf("case %s: decoding response keys: %v", c.ID, err)
	}
	known := map[string]bool{
		"status": true, "headers": true, "body_json_ref": true, "body_json": true,
		"body_text": true, "body_text_template": true, "body_bytes_hex": true,
		"codepoint": true,
	}
	for k := range keys {
		if !known[k] {
			t.Errorf("case %s: response carries key %q this runner does not handle; a body it cannot read is an empty body, and an empty body is refused for the wrong reason", c.ID, k)
		}
	}
	return &r
}

// body assembles the response body from whichever of the five carriers the case
// uses. Hex is one of them for a reason: the BOM and the leading-byte rules are
// decided on BYTES, and handing the resolver an already-decoded string skips the
// layer where a decoder strips the BOM -- which is the layer under test.
func (c *conformanceResolve) body(t *testing.T, fx *conformanceFixture, r *conformanceResponse) []byte {
	t.Helper()
	forms := 0
	var out []byte
	if r.BodyJSONRef != "" {
		forms++
		msg, ok := fx.Messages[r.BodyJSONRef]
		if !ok {
			t.Fatalf("body_json_ref %q is not in the fixture", r.BodyJSONRef)
		}
		out = msg
	}
	if len(r.BodyJSON) > 0 {
		forms++
		out = r.BodyJSON
	}
	if r.BodyText != "" {
		forms++
		out = []byte(r.BodyText)
	}
	if r.BodyTextTemplate != "" {
		forms++
		if r.Codepoint == 0 {
			t.Fatalf("case %s: body_text_template with no codepoint", c.ID)
		}
		if !strings.Contains(r.BodyTextTemplate, "{CP}") {
			t.Fatalf("case %s: body_text_template has no {CP} slot", c.ID)
		}
		out = []byte(strings.Replace(r.BodyTextTemplate, "{CP}", string(rune(r.Codepoint)), 1))
	}
	if r.BodyBytesHex != "" {
		forms++
		decoded, err := hex.DecodeString(r.BodyBytesHex)
		if err != nil {
			t.Fatalf("body_bytes_hex is not hex: %v", err)
		}
		out = decoded
	}
	if forms > 1 {
		t.Fatalf("case %s declares %d response bodies; exactly one or none is allowed", c.ID, forms)
	}
	return out
}

func (c *conformanceResolve) assertExpectations(
	t *testing.T,
	fx *conformanceFixture,
	request *IPPRequest,
	endpoint string,
	err error,
	rt *countingRoundTripper,
) {
	t.Helper()

	var outcome string
	mustUnmarshal(t, c.Expect["outcome"], &outcome)
	switch outcome {
	case "accept":
		if err != nil {
			t.Fatalf("MUST-ACCEPT case was rejected: %v", err)
		}
	case "reject":
		if err == nil {
			t.Fatalf("MUST-REJECT case was ACCEPTED (endpoint %q)", endpoint)
		}
	default:
		t.Fatalf("unknown outcome %q", outcome)
	}

	for key, rawVal := range c.Expect {
		switch key {
		case "outcome":
			// handled above.

		case "endpoint":
			var want string
			mustUnmarshal(t, rawVal, &want)
			if endpoint != want {
				t.Errorf("endpoint = %q, fixture says %q", endpoint, want)
			}

		case "message_ref":
			var ref string
			mustUnmarshal(t, rawVal, &ref)
			msg, ok := fx.Messages[ref]
			if !ok {
				t.Fatalf("message_ref %q is not in the fixture", ref)
			}
			want, decErr := Decode(string(msg))
			if decErr != nil {
				t.Fatalf("decoding messages[%s]: %v", ref, decErr)
			}
			if request == nil || !reflect.DeepEqual(*request, want) {
				t.Errorf("resolved envelope != messages[%s]\n got: %+v\nwant: %+v", ref, request, want)
			}

		case "envelope_expired":
			var expired bool
			mustUnmarshal(t, rawVal, &expired)
			if !expired {
				t.Fatalf("envelope_expired: false is not a shape this runner handles")
			}
			// Neither SDK carries a dedicated expiry signal out of the resolve.
			// What both do deliver is meta.expiresAt intact, which is what a
			// caller needs to warn before display. Asserting that keeps the
			// case honest: drop `meta` on the way through and this goes red.
			if request == nil {
				t.Fatal("expected an envelope")
			}
			at, ok := request.Meta[MetaKeyExpiresAt]
			if !ok {
				t.Fatalf("resolved envelope carries no meta.%s, so a caller cannot warn about expiry", MetaKeyExpiresAt)
			}
			parsed, perr := time.Parse(time.RFC3339, at)
			if perr != nil {
				t.Fatalf("meta.%s = %q is not RFC3339: %v", MetaKeyExpiresAt, at, perr)
			}
			if !parsed.Before(time.Now()) {
				t.Errorf("meta.%s = %q is not in the past; this case is no longer testing an expired envelope", MetaKeyExpiresAt, at)
			}

		case "rule":
			var rule int
			mustUnmarshal(t, rawVal, &rule)
			// The verdict is what this file compares across implementations, but
			// WHEN a rule fires is also observable without naming it: rules 1, 2
			// and 4 decide before any I/O, so they cost zero requests, and
			// everything from rule 6 on has to have read a response first.
			switch rule {
			case 1, 2, 4:
				if rt.calls != 0 {
					t.Errorf("rule %d is a pre-flight rule but the resolve cost %d request(s)", rule, rt.calls)
				}
			default:
				if rt.calls != 1 {
					t.Errorf("rule %d decides on the response but the resolve cost %d request(s)", rule, rt.calls)
				}
			}

		case "error_code":
			var code string
			mustUnmarshal(t, rawVal, &code)
			got, ok := errorCodeOf(err)
			if !ok {
				t.Errorf("no machine code surfaced, fixture says %q (err: %v)", code, err)
			} else if got != code {
				t.Errorf("machine code = %q, fixture says %q", got, code)
			}

		case "error_code_absent":
			var absent bool
			mustUnmarshal(t, rawVal, &absent)
			if !absent {
				t.Fatalf("error_code_absent: false is not a shape this runner handles")
			}
			// The other half of the shape negotiation. Both envelope shapes are
			// accepted, so the boundary has to be pinned too: a body matching
			// NEITHER must not be turned into an envelope error at all.
			//
			// The assertion is "not a typed envelope error", NOT "surfaced no
			// code" -- those are different, and the weaker one is a false green.
			// MEASURED here: widening the flat reading to any object carrying a
			// `message` types this body with Code == "", which "surfaced no
			// code" reads as a pass. An empty code is exactly what must not be
			// manufactured: the caller cannot tell it from a code the SDK failed
			// to read.
			var typed *IPPError
			if errors.As(err, &typed) {
				t.Errorf("body matching neither envelope shape became a typed %T with code %q; it must stay a transport failure (err: %v)", err, typed.Code, err)
			}

		case "http_status":
			var status int
			mustUnmarshal(t, rawVal, &status)
			if got := statusOf(err); got != status {
				t.Errorf("error carries HTTP status %d, fixture says %d (err: %v)", got, status, err)
			}

		case "reason", "note":
			// Prose for a human reading the fixture. There is nothing to assert
			// and pretending otherwise would be the false green this file exists
			// to remove.

		default:
			// resolve_contract.unknown_expect_key: an expectation this runner
			// does not understand must fail loudly, never be skipped.
			t.Errorf("expect key %q is one this runner does not handle; handle it or the case is asserting less than it says", key)
		}
	}
}

func errorCodeOf(err error) (string, bool) {
	var ipp *IPPError
	if errors.As(err, &ipp) && ipp.Code != "" {
		return ipp.Code, true
	}
	return "", false
}

func statusOf(err error) int {
	var ipp *IPPError
	if errors.As(err, &ipp) {
		return ipp.StatusCode
	}
	var conn *ConnectionError
	if errors.As(err, &conn) {
		return conn.StatusCode
	}
	return 0
}

func mustUnmarshal(t *testing.T, raw json.RawMessage, dst any) {
	t.Helper()
	if len(raw) == 0 {
		t.Fatalf("expected a value, got nothing")
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		t.Fatalf("decoding %s: %v", raw, err)
	}
}

// ── envelope.error_envelope_shape_pairs ─────────────────────────────

// TestConformanceErrorEnvelopeShapePairs asserts that one machine code arrives
// intact through BOTH accepted error-envelope shapes.
//
// Asserting each case against its own literal leaves "both happen to be spelled
// link_not_found" as a coincidence in the fixture rather than a property a runner
// checks. It is the pair that is the control: a reader understanding only one
// shape passes every case written in that shape, which is exactly how the
// nested-only reading survived here while dropping the machine code of every
// answer a third-party resolver sends in 6.2's flat shape.
//
// The third pair carries an unrecognised code, because widening the accepted
// SHAPES must not widen anything about the code: it stays an open string,
// surfaced verbatim, never mapped onto a known one and never turned into a
// rejection of its own.
func TestConformanceErrorEnvelopeShapePairs(t *testing.T) {
	fx := loadConformanceFixture(t)
	pairs := fx.Envelope.ErrorEnvelopeShapePairs
	if len(pairs) == 0 {
		t.Fatal("envelope.error_envelope_shape_pairs is empty; the shape negotiation has no control")
	}

	byID := map[string]*conformanceResolve{}
	for _, raw := range fx.Resolve {
		var c conformanceResolve
		if err := json.Unmarshal(raw, &c); err != nil {
			t.Fatalf("decoding resolve case: %v", err)
		}
		byID[c.ID] = &c
	}

	// runCase replays one resolve case and returns the machine code it surfaced.
	runCase := func(t *testing.T, id string) (string, bool) {
		t.Helper()
		c, ok := byID[id]
		if !ok {
			t.Fatalf("pair names case %q, which is not in the fixture", id)
		}
		r := c.response(t)
		body := c.body(t, fx, r)
		header := http.Header{}
		for k, v := range r.Headers {
			header.Set(k, v)
		}
		rt := conformanceTransport(r.Status, header, body)
		_, _, err := ResolveLink(context.Background(), c.URL,
			clientWith(rt), conformanceLookup(nil, false))
		if err == nil {
			t.Fatalf("case %s was accepted; an error envelope must be refused", id)
		}
		return errorCodeOf(err)
	}

	for _, pair := range pairs {
		t.Run(pair.Code, func(t *testing.T) {
			flatCode, flatOK := runCase(t, pair.Flat)
			nestedCode, nestedOK := runCase(t, pair.Nested)
			if !flatOK {
				t.Errorf("flat shape (%s) surfaced no machine code; a third-party resolver following 6.2 would have its code dropped", pair.Flat)
			}
			if !nestedOK {
				t.Errorf("nested shape (%s) surfaced no machine code; the gateway's own shape would have its code dropped", pair.Nested)
			}
			if flatCode != nestedCode {
				t.Errorf("one code, two shapes, two answers: flat (%s) gave %q, nested (%s) gave %q",
					pair.Flat, flatCode, pair.Nested, nestedCode)
			}
			if flatOK && flatCode != pair.Code {
				t.Errorf("machine code = %q, fixture says %q", flatCode, pair.Code)
			}
		})
	}
}
