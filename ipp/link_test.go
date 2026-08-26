package ipp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

const (
	testLinkToken = "AbCdEfGhIjKlMnOpQrStUv" // 22 base64url chars, letters only
	// dashLinkToken and dashHost exist to pin two specific regexp accidents that
	// have actually happened here. "[\s -]" puts a LITERAL hyphen in the class
	// and rejects every hyphenated host plus roughly half of all base64url
	// tokens; "[ -,]" is a range over punctuation and swallows "+", ",", "%".
	// Neither shows up in a happy path built from letters only.
	dashLinkToken = "Ab-dEfGh_jKlMnOpQrStUv"
	dashHost      = "api-test.invisipay.fi"

	testLink = "https://pay.example.com/ipp/v1/link/" + testLinkToken
)

// countingRoundTripper records every outbound round trip so tests can assert
// that a rejected URL cost zero requests.
type countingRoundTripper struct {
	calls int
	last  *http.Request
	fn    func(*http.Request) (*http.Response, error)
}

func (rt *countingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.calls++
	rt.last = req
	if rt.fn == nil {
		return nil, errors.New("countingRoundTripper: unexpected request")
	}
	return rt.fn(req)
}

func stringResponse(status int, body string, header http.Header) *http.Response {
	if header == nil {
		header = http.Header{}
	}
	return &http.Response{
		StatusCode:    status,
		Header:        header,
		Body:          io.NopCloser(strings.NewReader(body)),
		ContentLength: -1,
	}
}

// withLinkLookupIP stubs DNS. It is deliberately test-only: production callers
// must not be able to talk the resolved-address check out of its answer.
func withLinkLookupIP(ips ...string) LinkOption {
	return func(c *linkConfig) {
		c.lookupIP = func(context.Context, string) ([]net.IP, error) {
			out := make([]net.IP, 0, len(ips))
			for _, s := range ips {
				out = append(out, net.ParseIP(s))
			}
			return out, nil
		}
	}
}

// failLookupIP makes any DNS use a test failure, for cases that must be
// rejected on the URL alone.
func failLookupIP(t *testing.T) LinkOption {
	t.Helper()
	return func(c *linkConfig) {
		c.lookupIP = func(_ context.Context, host string) ([]net.IP, error) {
			t.Errorf("unexpected DNS lookup for %q", host)
			return nil, errors.New("unexpected lookup")
		}
	}
}

func sampleLinkRequest(endpoint string) IPPRequest {
	return IPPRequest{
		V:       IPPCurrentVersion,
		Network: "eip155:97",
		Type:    MessageTypeRequest,
		Meta:    map[string]string{MetaKeyMerchantName: "Coffee Shop"},
		Payload: IPPPayloadV1{
			IdentityCommitment: "0xabc",
			TokenAddress:       "0xtoken",
			Operator:           "0xop",
			PaymentTypes:       []PaymentType{PaymentTypeOpen},
			Mode:               "multi",
			Endpoint:           strPtr(endpoint),
		},
	}
}

func serveLinkJSON(t *testing.T, payloadEndpoint string) *countingRoundTripper {
	t.Helper()
	body, err := Encode(sampleLinkRequest(payloadEndpoint), FormatJSON)
	if err != nil {
		t.Fatalf("encoding sample request: %v", err)
	}
	return serveRaw(body)
}

func serveRaw(body string) *countingRoundTripper {
	return &countingRoundTripper{
		fn: func(*http.Request) (*http.Response, error) {
			return stringResponse(200, body, http.Header{"Content-Type": []string{"application/json"}}), nil
		},
	}
}

func clientWith(rt http.RoundTripper) LinkOption {
	return WithLinkHTTPClient(&http.Client{Transport: rt})
}

// rejects asserts the URL is refused with want and that nothing was sent.
func rejects(t *testing.T, linkURL string, want error, extra ...LinkOption) {
	t.Helper()
	rt := &countingRoundTripper{}
	opts := append([]LinkOption{clientWith(rt)}, extra...)
	_, _, err := ResolveLink(context.Background(), linkURL, opts...)
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
	if rt.calls != 0 {
		t.Fatalf("round trips = %d, want 0 (the check must precede the request)", rt.calls)
	}
}

// rejectsBecause is rejects plus an assertion on WHICH check fired.
//
// One malformed URL usually trips several rules, so asserting only "it was
// rejected" is a false green: delete the rule under test and a neighbouring one
// catches the same input. Measured three times here - dot segments, C1 controls
// and backslash authorities were all being caught by the "path not in normal
// form" check, which exists for a different reason.
func rejectsBecause(t *testing.T, linkURL string, want error, reason string, extra ...LinkOption) {
	t.Helper()
	rt := &countingRoundTripper{}
	opts := append([]LinkOption{clientWith(rt)}, extra...)
	_, _, err := ResolveLink(context.Background(), linkURL, opts...)
	if err == nil {
		t.Fatalf("%q was accepted, want rejection because %q", linkURL, reason)
	}
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
	if !strings.Contains(err.Error(), reason) {
		t.Fatalf("rejected for the wrong reason:\n  got:  %v\n  want message containing: %q", err, reason)
	}
	if rt.calls != 0 {
		t.Fatalf("round trips = %d, want 0", rt.calls)
	}
}

// rejectsPayloadBecause serves payloadEndpoint against a fixed valid link URL
// and pins which wire rule refused it.
func rejectsPayloadBecause(t *testing.T, payloadEndpoint, reason string) {
	t.Helper()
	rt := serveLinkJSON(t, payloadEndpoint)
	_, _, err := ResolveLink(context.Background(), testLink,
		clientWith(rt), withLinkLookupIP("93.184.216.34"))
	if err == nil {
		t.Fatalf("payload endpoint %q was accepted, want rejection because %q", payloadEndpoint, reason)
	}
	// Callers branch on "the cross-check did not pass"...
	if !errors.Is(err, ErrLinkEndpointMismatch) {
		t.Fatalf("error = %v, want ErrLinkEndpointMismatch", err)
	}
	// ...while this pins the specific rule that refused it.
	if !strings.Contains(err.Error(), reason) {
		t.Fatalf("rejected for the wrong reason:\n  got:  %v\n  want message containing: %q", err, reason)
	}
}

// -- W1 / assertion 1: shape mismatch is rejected before any request --

func TestResolveLinkRejectsBadShapeWithoutRequest(t *testing.T) {
	cases := []struct {
		name string
		url  string
		want error
	}{
		{"http scheme", "http://pay.example.com/ipp/v1/link/" + testLinkToken, ErrLinkSchemeNotHTTPS},
		{"non-http scheme", "ipp://pay.example.com/ipp/v1/link/" + testLinkToken, ErrLinkShape},
		{"custom deeplink scheme", "invisi://pay/" + testLinkToken, ErrLinkShape},
		{"not a URL", "just-some-scanned-text", ErrLinkShape},
		{"compact envelope", "ipp1:gaFhAQ", ErrLinkShape},
		{"empty string", "", ErrLinkShape},
		{"scheme-relative", "//pay.example.com/ipp/v1/link/" + testLinkToken, ErrLinkShape},
		{"no host", "https:///ipp/v1/link/" + testLinkToken, ErrLinkShape},
		{"no path", "https://pay.example.com", ErrLinkShape},
		{"wrong path root", "https://pay.example.com/link/" + testLinkToken, ErrLinkShape},
		{"wrong version segment", "https://pay.example.com/ipp/v2/link/" + testLinkToken, ErrLinkShape},
		{"split link segment", "https://pay.example.com/ipp/v1//link/" + testLinkToken, ErrLinkShape},
		{"token too short", "https://pay.example.com/ipp/v1/link/tooshort", ErrLinkShape},
		{"token too long", "https://pay.example.com/ipp/v1/link/" + testLinkToken + "X", ErrLinkShape},
		{"token empty", "https://pay.example.com/ipp/v1/link/", ErrLinkShape},
		{"token has non-base64url char", "https://pay.example.com/ipp/v1/link/AbCdEfGhIjKlMnOpQrSt+v", ErrLinkShape},
		{"percent-encoded token", "https://pay.example.com/ipp/v1/link/%41bCdEfGhIjKlMnOpQrStUv", ErrLinkShape},
		{"trailing segment", "https://pay.example.com/ipp/v1/link/" + testLinkToken + "/pay", ErrLinkShape},
		{"trailing slash", "https://pay.example.com/ipp/v1/link/" + testLinkToken + "/", ErrLinkShape},
		{"empty leading prefix segment", "https://pay.example.com//ipp/v1/link/" + testLinkToken, ErrLinkShape},
		{"empty inner prefix segment", "https://pay.example.com/a//b/ipp/v1/link/" + testLinkToken, ErrLinkShape},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rejects(t, tc.url, tc.want, failLookupIP(t))
		})
	}
}

// W1: a path prefix is legal, and the endpoint carries it. Rejecting it would
// lock out every third party that mounts IPP behind a shared gateway path.
func TestResolveLinkAcceptsPathPrefix(t *testing.T) {
	cases := []struct {
		name         string
		linkURL      string
		wantEndpoint string
	}{
		{"no prefix", testLink, "https://pay.example.com"},
		{"one segment", "https://pay.example.com/api/ipp/v1/link/" + testLinkToken, "https://pay.example.com/api"},
		{"two segments", "https://pay.example.com/a/b/ipp/v1/link/" + testLinkToken, "https://pay.example.com/a/b"},
		{"prefix with a dot in a segment", "https://pay.example.com/api.v1/ipp/v1/link/" + testLinkToken, "https://pay.example.com/api.v1"},
		{"prefix with three dots", "https://pay.example.com/.../ipp/v1/link/" + testLinkToken, "https://pay.example.com/..."},
		{"prefix repeats the link path", "https://pay.example.com/ipp/v1/link/x/ipp/v1/link/" + testLinkToken, "https://pay.example.com/ipp/v1/link/x"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rt := serveLinkJSON(t, tc.wantEndpoint)
			_, endpoint, err := ResolveLink(context.Background(), tc.linkURL,
				clientWith(rt), withLinkLookupIP("93.184.216.34"))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if endpoint != tc.wantEndpoint {
				t.Fatalf("endpoint = %q, want %q", endpoint, tc.wantEndpoint)
			}
			// The prefix is folded into the endpoint, so it must not be
			// concatenated onto the request URL a second time.
			if got := rt.last.URL.String(); got != tc.linkURL {
				t.Fatalf("request URL = %q, want %q", got, tc.linkURL)
			}
		})
	}

	t.Run("prefix must match payload endpoint too", func(t *testing.T) {
		rt := serveLinkJSON(t, "https://pay.example.com")
		_, _, err := ResolveLink(context.Background(),
			"https://pay.example.com/api/ipp/v1/link/"+testLinkToken,
			clientWith(rt), withLinkLookupIP("93.184.216.34"))
		if !errors.Is(err, ErrLinkEndpointMismatch) {
			t.Fatalf("error = %v, want ErrLinkEndpointMismatch", err)
		}
	})
}

// -- W1 / assertion 2: userinfo, query, fragment --

func TestResolveLinkRejectsUserinfoQueryFragment(t *testing.T) {
	cases := []struct {
		name, url, reason string
	}{
		{"host-shaped userinfo", "https://good.com@evil.com/ipp/v1/link/" + testLinkToken, "carries userinfo"},
		{"user and password", "https://user:pass@pay.example.com/ipp/v1/link/" + testLinkToken, "carries userinfo"},
		{"empty userinfo", "https://@pay.example.com/ipp/v1/link/" + testLinkToken, "carries userinfo"},
		{"query", testLink + "?fmt=json", "carries a query or fragment"},
		{"bare question mark", testLink + "?", "carries a query or fragment"},
		{"fragment", testLink + "#anchor", "carries a query or fragment"},
		{"bare hash", testLink + "#", "carries a query or fragment"},
		{"userinfo and query", "https://good.com@evil.com/ipp/v1/link/" + testLinkToken + "?a=1", "carries a query or fragment"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rejectsBecause(t, tc.url, ErrLinkURLForm, tc.reason, failLookupIP(t))
		})
	}

	// ResolveLinkUnsafe relaxes transport security only; the wire-form rules
	// decide what a link URL means and hold on both entry points.
	rt := &countingRoundTripper{}
	_, _, err := ResolveLinkUnsafe(context.Background(), "https://good.com@evil.com/ipp/v1/link/"+testLinkToken, clientWith(rt))
	if !errors.Is(err, ErrLinkURLForm) {
		t.Fatalf("unsafe entry point: error = %v, want ErrLinkURLForm", err)
	}
	if rt.calls != 0 {
		t.Fatalf("unsafe entry point: round trips = %d, want 0", rt.calls)
	}
}

// -- W1: dot segments, in every spelling --

func TestResolveLinkRejectsDotSegments(t *testing.T) {
	for _, seg := range []string{".", "..", "%2e", "%2E", "%2e%2e", "%2E%2E", ".%2e", "%2e."} {
		t.Run("prefix segment "+seg, func(t *testing.T) {
			rejectsBecause(t, "https://pay.example.com/"+seg+"/ipp/v1/link/"+testLinkToken,
				ErrLinkURLForm, "is a dot segment", failLookupIP(t))
		})
		t.Run("deep prefix segment "+seg, func(t *testing.T) {
			rejectsBecause(t, "https://pay.example.com/api/"+seg+"/ipp/v1/link/"+testLinkToken,
				ErrLinkURLForm, "is a dot segment", failLookupIP(t))
		})
	}
}

// cp renders a code point as a string. Tests build every invisible or non-ASCII
// character this way rather than pasting it: a literal C1 byte or a zero-width
// space in the source is unreviewable, and an editor that normalises it silently
// deletes the case.
func cp(r rune) string { return string(r) }

// -- W1: whitespace, C0, DEL, C1 --

func TestResolveLinkRejectsWhitespaceAndControlCharacters(t *testing.T) {
	inPath := func(r rune) string {
		return "https://pay.example.com/" + cp(r) + "/ipp/v1/link/" + testLinkToken
	}
	cases := map[string]string{
		"leading space":   " " + testLink,
		"trailing space":  testLink + " ",
		"leading newline": cp(0x0a) + testLink,
		"inner tab":       "https://pay.example.com/ipp" + cp(0x09) + "/v1/link/" + testLinkToken,
		"inner space":     "https://pay.example.com/ipp/v1/link/" + testLinkToken[:10] + " " + testLinkToken[10:],
		"carriage return": testLink + cp(0x0d),
		"vertical tab":    testLink + cp(0x0b),
		"form feed":       testLink + cp(0x0c),
		"NUL":             testLink + cp(0x00),
		"unit separator":  testLink + cp(0x1f),
		"DEL":             testLink + cp(0x7f),

		"non-breaking space": inPath(0x00a0),
		"BOM":                inPath(0xfeff),
		"ideographic space":  inPath(0x3000),
		"line separator":     inPath(0x2028),

		// C1 (U+0080-U+009F) in the path. Go's unicode.IsSpace reaches exactly
		// ONE of these (U+0085 NEL) and JS's \s reaches none, so without the
		// explicit range the two implementations disagree per code point.
		"C1 PAD":            inPath(0x0080),
		"C1 BPH":            inPath(0x0082),
		"C1 NEL":            inPath(0x0085),
		"C1 DCS":            inPath(0x0090),
		"C1 APC":            inPath(0x009f),
		"C1 upper boundary": inPath(0x009e),
	}
	for name, u := range cases {
		t.Run(name, func(t *testing.T) {
			rejectsBecause(t, u, ErrLinkURLForm, "whitespace or control character at byte", failLookupIP(t))
		})
	}

	// The reviewer's probe: this exact URL used to be ACCEPTED, because
	// unicode.IsSpace covers only U+0085 out of the whole C1 block.
	t.Run("C1 in the authority", func(t *testing.T) {
		rejectsBecause(t, "https://pay"+cp(0x0090)+".example.com/ipp/v1/link/"+testLinkToken,
			ErrLinkURLForm, "whitespace or control character at byte", failLookupIP(t))
	})

	// Regression guard for the two regexp accidents: "-", "_", "+", "," and "%"
	// must not be swept up as whitespace.
	t.Run("hyphens and other punctuation are not whitespace", func(t *testing.T) {
		linkURL := "https://" + dashHost + "/a-b/ipp/v1/link/" + dashLinkToken
		wantEndpoint := "https://" + dashHost + "/a-b"
		rt := serveLinkJSON(t, wantEndpoint)
		_, endpoint, err := ResolveLink(context.Background(), linkURL,
			clientWith(rt), withLinkLookupIP("93.184.216.34"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if endpoint != wantEndpoint {
			t.Fatalf("endpoint = %q, want %q", endpoint, wantEndpoint)
		}
	})
}

// -- W1: the authority must be all-ASCII --

func TestResolveLinkRejectsNonASCIIAuthority(t *testing.T) {
	inAuthority := func(s string) string {
		return "https://pay" + s + ".example.com/ipp/v1/link/" + testLinkToken
	}
	cases := map[string]string{
		// Go keeps raw UTF-8 and lowercases per Unicode; a WHATWG parser emits
		// punycode. The two endpoints are never equal, and Go cannot resolve the
		// raw form at all.
		"latin-1 supplement": inAuthority(cp(0x00c4)),
		"CJK":                inAuthority(cp(0x65e5) + cp(0x672c)),
		// The real hole: UTS-46 mapping DELETES U+200B, so a WHATWG parser
		// derives and then requests "pay.example.com" -- a host that is not the
		// one in the URL it was handed. U+200B is in neither side's whitespace
		// class, so only the ASCII rule catches it.
		"zero-width space":      inAuthority(cp(0x200b)),
		"soft hyphen":           inAuthority(cp(0x00ad)),
		"zero-width non-joiner": inAuthority(cp(0x200c)),
	}
	for name, u := range cases {
		t.Run(name, func(t *testing.T) {
			rejectsBecause(t, u, ErrLinkURLForm, "non-ASCII code point", failLookupIP(t))
		})
	}

	t.Run("punycode is the interoperable spelling and is accepted", func(t *testing.T) {
		linkURL := "https://xn--py-via.example.com/ipp/v1/link/" + testLinkToken
		rt := serveLinkJSON(t, "https://xn--py-via.example.com")
		_, endpoint, err := ResolveLink(context.Background(), linkURL,
			clientWith(rt), withLinkLookupIP("93.184.216.34"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if endpoint != "https://xn--py-via.example.com" {
			t.Fatalf("endpoint = %q", endpoint)
		}
	})
}

// -- W1: only one spelling of a port --

func TestResolveLinkRejectsNonCanonicalPort(t *testing.T) {
	// ":0443" is the load-bearing one: a numeric parser reads 443 and drops it as
	// the scheme default, a string comparison keeps "0443". Same printed code,
	// one wallet pays and the other refuses.
	for _, authority := range []string{
		"pay.example.com:0443",
		"pay.example.com:08443",
		"pay.example.com:00",
		"pay.example.com:0",
		"pay.example.com:",
		"pay.example.com:abc",
		"[::ffff:1.2.3.4]:0443",
	} {
		t.Run(authority, func(t *testing.T) {
			u := "https://" + authority + "/ipp/v1/link/" + testLinkToken
			rejectsBecause(t, u, ErrLinkURLForm, "port", failLookupIP(t))
		})
	}

	t.Run("port above 65535", func(t *testing.T) {
		rejectsBecause(t, "https://pay.example.com:99999/ipp/v1/link/"+testLinkToken,
			ErrLinkURLForm, "out of range", failLookupIP(t))
	})

	// A port written exactly as the scheme default is still dropped (W2).
	accepted := []struct {
		linkURL, payloadEndpoint, wantEndpoint string
	}{
		{"https://pay.example.com:443/ipp/v1/link/" + testLinkToken, "https://pay.example.com", "https://pay.example.com"},
		{"https://pay.example.com:8443/ipp/v1/link/" + testLinkToken, "https://pay.example.com:8443", "https://pay.example.com:8443"},
		{"https://pay.example.com:65535/ipp/v1/link/" + testLinkToken, "https://pay.example.com:65535", "https://pay.example.com:65535"},
	}
	for _, tc := range accepted {
		t.Run("accepts "+tc.linkURL, func(t *testing.T) {
			rt := serveLinkJSON(t, tc.payloadEndpoint)
			_, endpoint, err := ResolveLink(context.Background(), tc.linkURL,
				clientWith(rt), withLinkLookupIP("93.184.216.34"))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if endpoint != tc.wantEndpoint {
				t.Fatalf("endpoint = %q, want %q", endpoint, tc.wantEndpoint)
			}
		})
	}
}

func TestResolveLinkRejectsBackslashAuthority(t *testing.T) {
	// A WHATWG parser reads this as host good.example; Go's url.Parse errors on
	// it. Refusing it explicitly is what stops one written link from meaning two
	// origins -- and keeps the rejection reason stable rather than "not a URL".
	rejectsBecause(t, `https://good.example\evil.example/ipp/v1/link/`+testLinkToken,
		ErrLinkURLForm, "contains a backslash", failLookupIP(t))
	rejectsBecause(t, `https://good.example\evil.example:8443/ipp/v1/link/`+testLinkToken,
		ErrLinkURLForm, "contains a backslash", failLookupIP(t))

	// An authority carrying BOTH "@" and "\\" is refused for the userinfo, not the
	// backslash: the userinfo check runs first, and it runs first on the TS side
	// too. Pinning the order keeps the two implementations reporting the same
	// reason for the same string.
	rejectsBecause(t, `https://pay.example.com\@evil.example/ipp/v1/link/`+testLinkToken,
		ErrLinkURLForm, "carries userinfo", failLookupIP(t))
}

// -- W1: path segments are RFC 3986 pchar, checked as a charset --

func TestResolveLinkRejectsNonPcharPath(t *testing.T) {
	// The rule this replaced compared the path against Go's normal form, which
	// is not a shareable judgement: Go re-encodes "|" and "^" and rejects a
	// stray "%", while a WHATWG parser passes all three through. Same printed
	// code, two verdicts. The first three entries are exactly the measured
	// divergence set.
	for _, ch := range []string{"|", "^", "%", `"`, "<", ">", "{", "}", "`", `\`, "[", "]"} {
		t.Run("path contains "+ch, func(t *testing.T) {
			rejectsBecause(t, "https://pay.example.com/a"+ch+"b/ipp/v1/link/"+testLinkToken,
				ErrLinkURLForm, "is not RFC 3986 pchar", failLookupIP(t))
		})
	}

	for _, bad := range []string{"%", "%b", "%zz", "%1", "%g0", "%0", "a%", "a%2"} {
		t.Run("malformed percent-encoding "+bad, func(t *testing.T) {
			rejectsBecause(t, "https://pay.example.com/"+bad+"/ipp/v1/link/"+testLinkToken,
				ErrLinkURLForm, "is not RFC 3986 pchar", failLookupIP(t))
		})
	}

	// Lowercase percent-encoding is pchar-shaped but still refused: rule 9
	// compares endpoints as strings, and RFC 3986 section 6.2.2.1 says a
	// comparison should uppercase percent-encodings first - so "%2F" and "%2f"
	// are the same endpoint under one reading and different under the other.
	// Both spellings pass a case-insensitive pchar test, so the shape check
	// cannot see this and it needs its own case.
	for _, seg := range []string{"a%2fb", "%2f", "a%4fb", "%ab", "%aB", "%Ab"} {
		t.Run("lowercase percent-encoding "+seg, func(t *testing.T) {
			rejectsBecause(t, "https://pay.example.com/"+seg+"/ipp/v1/link/"+testLinkToken,
				ErrLinkURLForm, "not uppercase hex", failLookupIP(t))
		})
	}

	t.Run("uppercase percent-encoding is accepted", func(t *testing.T) {
		prefix := "/a%2FB%41"
		linkURL := "https://pay.example.com" + prefix + "/ipp/v1/link/" + testLinkToken
		wantEndpoint := "https://pay.example.com" + prefix
		rt := serveLinkJSON(t, wantEndpoint)
		_, endpoint, err := ResolveLink(context.Background(), linkURL,
			clientWith(rt), withLinkLookupIP("93.184.216.34"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if endpoint != wantEndpoint {
			t.Fatalf("endpoint = %q, want %q", endpoint, wantEndpoint)
		}
	})

	// Positive control, and the reason the charset must be written out in full: a
	// pchar set copied down too narrow would reject legitimate prefixes, and
	// nothing else in the suite would notice.
	t.Run("the whole pchar set is accepted in a prefix", func(t *testing.T) {
		prefix := "/a!$&'()*+,;=:@-._~%41"
		linkURL := "https://pay.example.com" + prefix + "/ipp/v1/link/" + testLinkToken
		wantEndpoint := "https://pay.example.com" + prefix
		rt := serveLinkJSON(t, wantEndpoint)
		_, endpoint, err := ResolveLink(context.Background(), linkURL,
			clientWith(rt), withLinkLookupIP("93.184.216.34"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if endpoint != wantEndpoint {
			t.Fatalf("endpoint = %q, want %q", endpoint, wantEndpoint)
		}
	})
}

// -- W1: the authority has one canonical spelling --

func TestResolveLinkRejectsNonCanonicalAuthority(t *testing.T) {
	// Every entry here is a host a WHATWG parser silently rewrites. Without this
	// rule the endpoint derived from one printed code differed between the two
	// implementations, so one wallet paid and the other refused.
	cases := map[string]struct{ authority, reason string }{
		"IPv6 uncompressed":       {"[2001:db8:0:0:0:0:0:1]", "not RFC 5952 canonical form"},
		"IPv6 partly compressed":  {"[0::1]", "not RFC 5952 canonical form"},
		"IPv6 leading zero group": {"[::01]", "not RFC 5952 canonical form"},
		"IPv6 padded group":       {"[2001:0db8::1]", "not RFC 5952 canonical form"},
		"IPv6 zero not collapsed": {"[::0]", "not RFC 5952 canonical form"},
		// The form a WHATWG parser produces for an IPv4-mapped address. RFC 5952
		// section 5 keeps the mixed notation, so this spelling is the wrong one.
		"IPv6 v4-mapped as hex": {"[::ffff:7f00:1]", "not RFC 5952 canonical form"},
		"IPv6 uppercase":        {"[2001:DB8::1]", "must be lowercase"},
		"IPv6 with a zone id":   {"[fe80::1%25en0]", "is not an IPv6 literal"},
		"bracketed IPv4":        {"[1.2.3.4]", "is not an IPv6 literal"},
		"unterminated bracket":  {"[2001:db8::1", "unterminated IPv6 literal"},
		"misplaced bracket":     {"pay]example.com", "misplaced bracket"},

		"IPv4 leading zero":       {"01.2.3.4", "not a canonical IPv4 literal"},
		"IPv4 padded last octet":  {"1.2.3.04", "not a canonical IPv4 literal"},
		"IPv4 octal":              {"0177.0.0.1", "not a canonical IPv4 literal"},
		"IPv4 three parts":        {"203.0.113", "not a canonical IPv4 literal"},
		"IPv4 as one integer":     {"3405803625", "not a canonical IPv4 literal"},
		"IPv4 five parts":         {"1.2.3.4.5", "not a canonical IPv4 literal"},
		"IPv4 octet out of range": {"256.1.1.1", "not a canonical IPv4 literal"},
		"IPv4 octet way too big":  {"999.1.1.1", "not a canonical IPv4 literal"},
		"IPv4 trailing dot":       {"1.2.3.", "not a canonical IPv4 literal"},
		"IPv4 hex":                {"0x7f000001", "reread as an IPv4 address"},
		"numeric last label":      {"pay.example.123", "reread as an IPv4 address"},
		"hex last label":          {"pay.example.0x1", "reread as an IPv4 address"},
		"percent-encoded dot":     {"h%2e", "not a letter-digit-hyphen label"},
		"underscore in a label":   {"pay_example.com", "not a letter-digit-hyphen label"},
		"empty inner label":       {"h..com", "empty label"},
		"trailing dot":            {"pay.example.com.", "empty label"},
		"leading dot":             {".pay.example.com", "empty label"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rejectsBecause(t, "https://"+tc.authority+"/ipp/v1/link/"+testLinkToken,
				ErrLinkURLForm, tc.reason, failLookupIP(t))
		})
	}

	// Positive controls. Without these the rules above could be tightened into
	// rejecting ordinary hosts and nothing would catch it.
	accepted := map[string]string{
		"ordinary reg-name":        "pay.example.com",
		"hyphenated label":         "api-test.invisipay.fi",
		"punycode label":           "xn--py-via.example.com",
		"all-digit FIRST label":    "1.example.com",
		"canonical IPv4":           "1.2.3.4",
		"canonical IPv6":           "[2606:4700:4700::1111]",
		"canonical v4-mapped IPv6": "[::ffff:1.2.3.4]",
	}
	for name, authority := range accepted {
		t.Run("accepts "+name, func(t *testing.T) {
			wantEndpoint := "https://" + authority
			rt := serveLinkJSON(t, wantEndpoint)
			_, endpoint, err := ResolveLink(context.Background(),
				wantEndpoint+"/ipp/v1/link/"+testLinkToken,
				clientWith(rt), withLinkLookupIP("93.184.216.34"))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if endpoint != wantEndpoint {
				t.Fatalf("endpoint = %q, want %q", endpoint, wantEndpoint)
			}
		})
	}
}

// -- W2: payload.endpoint gets the SAME written-form rules --

func TestResolveLinkAppliesWireFormRulesToPayloadEndpoint(t *testing.T) {
	cases := map[string]struct{ endpoint, reason string }{
		// The two that were measured to diverge while these rules were scoped to
		// the link URL only. Go leaves "/api/../api" as written; a WHATWG parser
		// folds it to "/api" and pays a code Go refuses.
		"dot segment":         {"https://pay.example.com/api/../api", "is a dot segment"},
		"encoded dot segment": {"https://pay.example.com/api/%2e%2e/api", "is a dot segment"},
		"single dot segment":  {"https://pay.example.com/./api", "is a dot segment"},
		"backslash authority": {`https://good.com\evil.com`, "contains a backslash"},
		// The rest were already in scope; they stay pinned here because the
		// shared parser is now the only place they live.
		"userinfo":            {"https://good.com@pay.example.com", "carries userinfo"},
		"query":               {"https://pay.example.com?x=1", "carries a query or fragment"},
		"fragment":            {"https://pay.example.com#x", "carries a query or fragment"},
		"trailing space":      {"https://pay.example.com ", "whitespace or control character at byte"},
		"C1 control":          {"https://pay.example.com/" + cp(0x0090), "whitespace or control character at byte"},
		"non-ASCII authority": {"https://pay" + cp(0x00c4) + ".example.com", "non-ASCII code point"},
		"zero-width space":    {"https://pay" + cp(0x200b) + ".example.com", "non-ASCII code point"},
		"leading-zero port":   {"https://pay.example.com:0443", "port"},
		"empty port":          {"https://pay.example.com:", "port"},
		"empty path segment":  {"https://pay.example.com//api", "malformed path"},
		"not a URL":           {"pay.example.com", "is not scheme://host/path"},
		"no host":             {"https://", "has no host"},
		"uppercase scheme":    {"HTTPS://pay.example.com", "is not lowercase"},
		"mixed-case scheme":   {"Https://pay.example.com", "is not lowercase"},
		// These three had no reason-pinned coverage on the payload side. The
		// mechanism is the shared parser, but without them a bypass added to
		// normalizeEndpoint would only turn the link-side tests red.
		"non-pchar path":       {"https://pay.example.com/a|b", "is not RFC 3986 pchar"},
		"malformed percent":    {"https://pay.example.com/a%zz", "is not RFC 3986 pchar"},
		"lowercase percent":    {"https://pay.example.com/a%2fb", "not uppercase hex"},
		"non-canonical IPv6":   {"https://[2001:db8:0:0:0:0:0:1]", "not RFC 5952 canonical form"},
		"non-canonical IPv4":   {"https://0177.0.0.1", "not a canonical IPv4 literal"},
		"numeric-looking host": {"https://0x7f000001", "reread as an IPv4 address"},
		"port out of range":    {"https://pay.example.com:99999", "out of range"},
		"non-http scheme":      {"ftp://pay.example.com", `scheme "ftp"`},
		"http scheme":          {"http://pay.example.com", "scheme is http while the scanned link is https"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rejectsPayloadBecause(t, tc.endpoint, tc.reason)
		})
	}

	// ErrLinkSchemeNotHTTPS is reserved for the scanned link. A caller that
	// branches on it to say "use the unsafe entry point" must not see it here:
	// the link WAS https, and relaxing the entry point would not fix a payload
	// that declares http.
	t.Run("the scheme sentinel does not leak from the payload side", func(t *testing.T) {
		rt := serveLinkJSON(t, "http://pay.example.com")
		_, _, err := ResolveLink(context.Background(), testLink,
			clientWith(rt), withLinkLookupIP("93.184.216.34"))
		if err == nil {
			t.Fatal("an http payload endpoint was accepted under the strict entry point")
		}
		if errors.Is(err, ErrLinkSchemeNotHTTPS) {
			t.Fatalf("ErrLinkSchemeNotHTTPS leaked from payload.endpoint; a caller would tell the user to switch entry points over a link that was already https: %v", err)
		}
		if !errors.Is(err, ErrLinkURLForm) || !errors.Is(err, ErrLinkEndpointMismatch) {
			t.Fatalf("error = %v, want both ErrLinkURLForm and ErrLinkEndpointMismatch", err)
		}
	})

	// By contrast, the scanned link saying http MUST carry the sentinel.
	t.Run("the scheme sentinel is still raised for the scanned link", func(t *testing.T) {
		rejectsBecause(t, "http://pay.example.com/ipp/v1/link/"+testLinkToken,
			ErrLinkSchemeNotHTTPS, "scheme must be https", failLookupIP(t))
	})

	// The only rule that does not apply: a payload endpoint is not required to
	// end in /ipp/v1/link/<token>. It is not the URL that was scanned.
	t.Run("the link path suffix rule does not apply", func(t *testing.T) {
		rt := serveLinkJSON(t, "https://pay.example.com")
		if _, _, err := ResolveLink(context.Background(), testLink,
			clientWith(rt), withLinkLookupIP("93.184.216.34")); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// -- W4 / assertion 3: non-public targets --

func TestIsNonPublicIPRejectionSet(t *testing.T) {
	blocked := []string{
		"0.0.0.0", "0.1.2.3",
		"10.0.0.5", "10.255.255.255",
		"100.64.0.1", "100.127.255.254",
		"127.0.0.1", "127.53.0.1",
		"169.254.169.254",
		"172.16.3.9", "172.31.255.1",
		"192.0.0.1", "192.0.0.170",
		"192.168.1.1",
		"198.18.0.1", "198.19.5.5",
		"224.0.0.1", "239.1.2.3", "255.255.255.255",
		"::", "::1", "fc00::1", "fd12:3456::1", "fe80::1", "ff02::1", "ff01::1",
		// IPv4 wearing an IPv6 costume -- judge the inner address
		"::ffff:127.0.0.1", "::ffff:10.0.0.1",
		"::127.0.0.1", "::192.168.0.1",
		"64:ff9b::127.0.0.1", "64:ff9b::10.0.0.1", "64:ff9b::1:1",
	}
	for _, s := range blocked {
		ip := net.ParseIP(s)
		if ip == nil {
			t.Fatalf("test bug: %q is not a parseable IP", s)
		}
		if !isNonPublicIP(ip) {
			t.Errorf("isNonPublicIP(%s) = false, want true", s)
		}
	}

	public := []string{
		"1.1.1.1", "8.8.8.8", "93.184.216.34",
		"100.63.255.255", "100.128.0.1", // just outside CGNAT
		// 192.0.1.1 is the canary for the /16-instead-of-/24 bug: widening
		// 192.0.0.0/24 also swallows 192.0.2.0/24, which must stay public.
		"192.0.1.1",
		"192.0.2.1", "198.51.100.1", "203.0.113.1", // TEST-NET: documentation-only, deliberately NOT blocked
		"198.17.255.255", "198.20.0.1",
		"172.15.0.1", "172.32.0.1",
		"223.255.255.255",
		"2606:4700:4700::1111",
		"64:ff9b::93.184.216.34", // NAT64 prefix wrapping a PUBLIC inner address
		"::ffff:93.184.216.34",
	}
	for _, s := range public {
		ip := net.ParseIP(s)
		if ip == nil {
			t.Fatalf("test bug: %q is not a parseable IP", s)
		}
		if isNonPublicIP(ip) {
			t.Errorf("isNonPublicIP(%s) = true, want false", s)
		}
	}
}

func TestResolveLinkRejectsNonPublicResolvedAddress(t *testing.T) {
	for _, addr := range []string{
		"127.0.0.1", "10.0.0.5", "100.64.1.1", "169.254.169.254",
		"192.0.0.1", "198.18.0.1", "224.0.0.1", "0.0.0.0",
		"::1", "fd00::1", "fe80::1", "::ffff:127.0.0.1", "::127.0.0.1", "64:ff9b::10.0.0.1",
	} {
		t.Run("resolves to "+addr, func(t *testing.T) {
			rejectsBecause(t, "https://evil.example.com/ipp/v1/link/"+testLinkToken,
				ErrLinkPrivateHost, "resolves to", withLinkLookupIP(addr))
		})
	}

	t.Run("every answer must be public, not just the first", func(t *testing.T) {
		rejectsBecause(t, "https://evil.example.com/ipp/v1/link/"+testLinkToken,
			ErrLinkPrivateHost, "resolves to 10.1.2.3", withLinkLookupIP("93.184.216.34", "10.1.2.3"))
	})

	t.Run("a non-public answer anywhere in the list", func(t *testing.T) {
		rejectsBecause(t, "https://evil.example.com/ipp/v1/link/"+testLinkToken,
			ErrLinkPrivateHost, "resolves to ::1", withLinkLookupIP("1.1.1.1", "8.8.8.8", "::1"))
	})

	t.Run("literal non-public host", func(t *testing.T) {
		rejects(t, "https://192.168.1.10/ipp/v1/link/"+testLinkToken, ErrLinkPrivateHost, failLookupIP(t))
	})

	t.Run("literal non-public IPv6 host", func(t *testing.T) {
		rejects(t, "https://[::1]/ipp/v1/link/"+testLinkToken, ErrLinkPrivateHost, failLookupIP(t))
	})

	t.Run("public address passes", func(t *testing.T) {
		rt := serveLinkJSON(t, "https://pay.example.com")
		got, endpoint, err := ResolveLink(context.Background(), testLink,
			clientWith(rt), withLinkLookupIP("93.184.216.34"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if endpoint != "https://pay.example.com" {
			t.Fatalf("endpoint = %q", endpoint)
		}
		if got.Payload.IdentityCommitment != "0xabc" {
			t.Fatalf("payload not decoded: %+v", got.Payload)
		}
	})
}

func TestDenyNonPublicDialRejectsResolvedAddress(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:443", "[fd00::1]:443", "[::ffff:10.0.0.1]:443", "169.254.169.254:80"} {
		if err := denyNonPublicDial("tcp", addr, nil); !errors.Is(err, ErrLinkPrivateHost) {
			t.Errorf("denyNonPublicDial(%s) = %v, want ErrLinkPrivateHost", addr, err)
		}
	}
	if err := denyNonPublicDial("tcp", "93.184.216.34:443", nil); err != nil {
		t.Errorf("public: err = %v, want nil", err)
	}
}

// TestHardenedClientWiresDialGuard covers the WIRING, not the predicate.
//
// This is rule 4's connect-time half, and the reason this implementation does not
// need the after-the-response re-check a fetch-based one has to approximate. It
// was previously untested: the assertions only looked at Transport != nil, and
// the *http.Transport branch was never even entered, because every other test
// injects a bare RoundTripper. Deleting the DialContext assignment kept the
// whole suite green.
func TestHardenedClientWiresDialGuard(t *testing.T) {
	bases := map[string]*http.Client{
		"nil client":                       nil,
		"client with nil transport":        {},
		"client with *http.Transport":      {Transport: &http.Transport{}},
		"client with configured transport": {Transport: &http.Transport{MaxIdleConns: 7}},
	}
	for name, base := range bases {
		t.Run(name, func(t *testing.T) {
			c := hardenedClient(base, false)
			if c.CheckRedirect == nil {
				t.Fatal("strict client must never follow redirects")
			}
			tr, ok := c.Transport.(*http.Transport)
			if !ok {
				t.Fatalf("transport = %T, want *http.Transport carrying the dial guard", c.Transport)
			}
			if tr.DialContext == nil {
				t.Fatal("no DialContext: the connect-time address check is not wired up")
			}
			// Port 9 (discard) on loopback: the guard must refuse before any
			// connect syscall, so nothing needs to be listening.
			if _, err := tr.DialContext(context.Background(), "tcp", "127.0.0.1:9"); !errors.Is(err, ErrLinkPrivateHost) {
				t.Fatalf("dial to loopback: err = %v, want ErrLinkPrivateHost", err)
			}
			if _, err := tr.DialContext(context.Background(), "tcp", "[fd00::1]:9"); !errors.Is(err, ErrLinkPrivateHost) {
				t.Fatalf("dial to ULA: err = %v, want ErrLinkPrivateHost", err)
			}
		})
	}

	t.Run("caller transport is cloned, not mutated", func(t *testing.T) {
		base := &http.Transport{}
		c := hardenedClient(&http.Client{Transport: base}, false)
		if c.Transport == http.RoundTripper(base) {
			t.Fatal("the caller's transport was mutated in place")
		}
		if base.DialContext != nil {
			t.Fatal("the guard leaked onto the caller's own transport")
		}
	})

	// The unsafe entry point is the one place the guard is intentionally absent.
	// If these start failing, the two entry points have merged.
	t.Run("unsafe installs no guard", func(t *testing.T) {
		if got := hardenedClient(nil, true).Transport; got != nil {
			t.Fatalf("transport = %T, want nil (no guard on the unsafe path)", got)
		}
		base := &http.Transport{}
		if got := hardenedClient(&http.Client{Transport: base}, true).Transport.(*http.Transport); got.DialContext != nil {
			t.Fatal("unsafe path installed a dial guard")
		}
		if hardenedClient(nil, true).CheckRedirect == nil {
			t.Fatal("unsafe client must never follow redirects either")
		}
	})
}

// -- W5: a resolver failure is a transport verdict, not a safety one --

func TestResolveLinkDNSFailureIsTransportError(t *testing.T) {
	rt := &countingRoundTripper{}
	_, _, err := ResolveLink(context.Background(), testLink,
		clientWith(rt),
		func(c *linkConfig) {
			c.lookupIP = func(context.Context, string) ([]net.IP, error) {
				return nil, errors.New("dial udp 1.1.1.1:53: network is unreachable")
			}
		},
	)
	var ce *ConnectionError
	if !errors.As(err, &ce) {
		t.Fatalf("error = %v, want *ConnectionError", err)
	}
	// Being offline must not read to the user as "this code is dangerous".
	if errors.Is(err, ErrLinkPrivateHost) {
		t.Fatalf("a resolver failure was reported as a safety verdict: %v", err)
	}
	if rt.calls != 0 {
		t.Fatalf("round trips = %d, want 0 (fail closed, no unresolved request)", rt.calls)
	}
}

func TestResolveLinkEmptyDNSAnswerFailsClosed(t *testing.T) {
	rt := &countingRoundTripper{}
	_, _, err := ResolveLink(context.Background(), testLink,
		clientWith(rt), withLinkLookupIP())
	if err == nil {
		t.Fatal("an empty DNS answer must not proceed to the request")
	}
	if rt.calls != 0 {
		t.Fatalf("round trips = %d, want 0", rt.calls)
	}
}

// -- Assertion 4: 3xx is a failure, never a hop --

func TestResolveLinkDoesNotFollowRedirects(t *testing.T) {
	for _, status := range []int{301, 302, 303, 307, 308} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			rt := &countingRoundTripper{
				fn: func(*http.Request) (*http.Response, error) {
					return stringResponse(status, "", http.Header{
						"Location": []string{"https://evil.example.com/ipp/v1/link/" + testLinkToken},
					}), nil
				},
			}
			_, _, err := ResolveLink(context.Background(), testLink,
				clientWith(rt), withLinkLookupIP("93.184.216.34"))
			if !errors.Is(err, ErrLinkRedirect) {
				t.Fatalf("error = %v, want ErrLinkRedirect", err)
			}
			if rt.calls != 1 {
				t.Fatalf("round trips = %d, want 1 (the redirect must not be followed)", rt.calls)
			}
		})
	}
}

// -- W2 / assertion 5: endpoint cross-check --

func TestResolveLinkCrossChecksEndpoint(t *testing.T) {
	rejected := map[string]string{
		"different host":         "https://evil.example.com",
		"different port":         "https://pay.example.com:8443",
		"different scheme":       "http://pay.example.com",
		"subdomain":              "https://api.pay.example.com",
		"host prefix":            "https://pay.example.co",
		"unexpected path prefix": "https://pay.example.com/api",
	}
	for name, payloadEndpoint := range rejected {
		t.Run("rejects "+name, func(t *testing.T) {
			rt := serveLinkJSON(t, payloadEndpoint)
			_, _, err := ResolveLink(context.Background(), testLink,
				clientWith(rt), withLinkLookupIP("93.184.216.34"))
			if !errors.Is(err, ErrLinkEndpointMismatch) {
				t.Fatalf("error = %v, want ErrLinkEndpointMismatch", err)
			}
		})
	}

	t.Run("rejects a missing payload endpoint", func(t *testing.T) {
		req := sampleLinkRequest("https://pay.example.com")
		req.Payload.Mode = "" // ValidateIPPRequest only demands endpoint for single/multi
		req.Payload.Endpoint = nil
		body, err := Encode(req, FormatJSON)
		if err != nil {
			t.Fatalf("encode: %v", err)
		}
		rt := serveRaw(body)
		_, _, err = ResolveLink(context.Background(), testLink,
			clientWith(rt), withLinkLookupIP("93.184.216.34"))
		if !errors.Is(err, ErrLinkEndpointMismatch) {
			t.Fatalf("error = %v, want ErrLinkEndpointMismatch", err)
		}
		if !strings.Contains(err.Error(), "carries no endpoint") {
			t.Fatalf("wrong reason: %v", err)
		}
	})

	equivalents := []struct {
		name            string
		linkURL         string
		payloadEndpoint string
		wantEndpoint    string
	}{
		{"uppercase host in link", "https://Pay.Example.COM/ipp/v1/link/" + testLinkToken, "https://pay.example.com", "https://pay.example.com"},
		{"default port in link", "https://pay.example.com:443/ipp/v1/link/" + testLinkToken, "https://pay.example.com", "https://pay.example.com"},
		{"uppercase host in payload", testLink, "https://PAY.example.com", "https://pay.example.com"},
		{"default port in payload", testLink, "https://pay.example.com:443", "https://pay.example.com"},
		{"trailing slash in payload", testLink, "https://pay.example.com/", "https://pay.example.com"},
		// These two pin the normalisation ORDER: trailing slashes are stripped
		// before the empty-segment check, so "https://h//" collapses to the bare
		// origin instead of being refused as "//".
		{"double trailing slash in payload", testLink, "https://pay.example.com//", "https://pay.example.com"},
		{"many trailing slashes in payload", testLink, "https://pay.example.com///", "https://pay.example.com"},
		{"explicit non-default port on both", "https://pay.example.com:8443/ipp/v1/link/" + testLinkToken, "https://pay.example.com:8443", "https://pay.example.com:8443"},
		{"prefix with a trailing slash in payload", "https://pay.example.com/api/ipp/v1/link/" + testLinkToken, "https://pay.example.com/api/", "https://pay.example.com/api"},
	}
	for _, tc := range equivalents {
		t.Run("accepts "+tc.name, func(t *testing.T) {
			rt := serveLinkJSON(t, tc.payloadEndpoint)
			_, endpoint, err := ResolveLink(context.Background(), tc.linkURL,
				clientWith(rt), withLinkLookupIP("93.184.216.34"))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if endpoint != tc.wantEndpoint {
				t.Fatalf("endpoint = %q, want %q", endpoint, tc.wantEndpoint)
			}
		})
	}
}

// -- Assertion 6: the two entry points differ, and only in transport --

func TestResolveLinkUnsafeRelaxesSchemeAndPrivateTargets(t *testing.T) {
	cases := []struct {
		name            string
		linkURL         string
		payloadEndpoint string
		wantStrictErr   error
	}{
		{
			name:            "http on a private address",
			linkURL:         "http://192.168.1.10:8080/ipp/v1/link/" + testLinkToken,
			payloadEndpoint: "http://192.168.1.10:8080",
			wantStrictErr:   ErrLinkSchemeNotHTTPS,
		},
		{
			name:            "https on a loopback address",
			linkURL:         "https://127.0.0.1:8443/ipp/v1/link/" + testLinkToken,
			payloadEndpoint: "https://127.0.0.1:8443",
			wantStrictErr:   ErrLinkPrivateHost,
		},
		{
			name:            "http on a public host",
			linkURL:         "http://pay.example.com/ipp/v1/link/" + testLinkToken,
			payloadEndpoint: "http://pay.example.com",
			wantStrictErr:   ErrLinkSchemeNotHTTPS,
		},
		{
			name:            "http behind a path prefix",
			linkURL:         "http://10.0.0.7:8080/api/ipp/v1/link/" + testLinkToken,
			payloadEndpoint: "http://10.0.0.7:8080/api",
			wantStrictErr:   ErrLinkSchemeNotHTTPS,
		},
		{
			name:            "http on its default port is still dropped",
			linkURL:         "http://10.0.0.7:80/ipp/v1/link/" + testLinkToken,
			payloadEndpoint: "http://10.0.0.7",
			wantStrictErr:   ErrLinkSchemeNotHTTPS,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rejects(t, tc.linkURL, tc.wantStrictErr, withLinkLookupIP("127.0.0.1"))

			ok := serveLinkJSON(t, tc.payloadEndpoint)
			got, endpoint, err := ResolveLinkUnsafe(context.Background(), tc.linkURL, clientWith(ok))
			if err != nil {
				t.Fatalf("ResolveLinkUnsafe error = %v, want nil", err)
			}
			if got == nil || got.Payload.IdentityCommitment != "0xabc" {
				t.Fatalf("ResolveLinkUnsafe returned %+v", got)
			}
			if endpoint != tc.payloadEndpoint {
				t.Fatalf("endpoint = %q, want %q", endpoint, tc.payloadEndpoint)
			}
		})
	}

	// The wire-form rules are not part of the relaxation: a non-canonical port or
	// a non-ASCII authority is refused on both entry points.
	t.Run("wire form still applies on the unsafe path", func(t *testing.T) {
		for _, u := range []string{
			"http://10.0.0.7:0443/ipp/v1/link/" + testLinkToken,
			"http://10.0.0" + cp(0x00c4) + ".7/ipp/v1/link/" + testLinkToken,
			"http://10.0.0.7/../ipp/v1/link/" + testLinkToken,
		} {
			rt := &countingRoundTripper{}
			if _, _, err := ResolveLinkUnsafe(context.Background(), u, clientWith(rt)); !errors.Is(err, ErrLinkURLForm) {
				t.Errorf("ResolveLinkUnsafe(%q) = %v, want ErrLinkURLForm", u, err)
			}
			if rt.calls != 0 {
				t.Errorf("ResolveLinkUnsafe(%q) sent %d requests, want 0", u, rt.calls)
			}
		}
	})
}

// -- The two carriers must decode to the same value --

func TestResolveLinkMatchesCompactDecode(t *testing.T) {
	want := sampleLinkRequest("https://pay.example.com")
	compact, err := Encode(want, FormatCompact)
	if err != nil {
		t.Fatalf("encode compact: %v", err)
	}
	if !strings.HasPrefix(compact, "ipp1:") {
		t.Fatalf("compact form = %q, want an ipp1: envelope", compact)
	}
	fromCompact, err := Decode(compact)
	if err != nil {
		t.Fatalf("decode compact: %v", err)
	}

	rt := serveLinkJSON(t, "https://pay.example.com")
	fromLink, _, err := ResolveLink(context.Background(), testLink,
		clientWith(rt), withLinkLookupIP("93.184.216.34"))
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !reflect.DeepEqual(*fromLink, fromCompact) {
		t.Fatalf("link form %+v != compact form %+v", *fromLink, fromCompact)
	}
}

// -- W3: the response body contract --

func TestResolveLinkRejectsNonEnvelopeBody(t *testing.T) {
	compact, err := Encode(sampleLinkRequest("https://pay.example.com"), FormatCompact)
	if err != nil {
		t.Fatalf("encode compact: %v", err)
	}
	jsonBody, err := Encode(sampleLinkRequest("https://pay.example.com"), FormatJSON)
	if err != nil {
		t.Fatalf("encode json: %v", err)
	}

	// One extra top-level key. The payer refusing it is what makes "the envelope
	// is a closed set, meta is the extension point" a real promise rather than a
	// wish about server behaviour.
	var withExtra map[string]any
	if err := json.Unmarshal([]byte(jsonBody), &withExtra); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	withExtra["sponsored"] = true
	extraBody, err := json.Marshal(withExtra)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	cases := map[string]struct {
		body, reason string
	}{
		"compact ipp1 string":   {compact, "not a JSON object"},
		"json array":            {`[{"v":1}]`, "not a JSON object"},
		"json string":           {`"hello"`, "not a JSON object"},
		"html page":             {"<html>ok</html>", "not a JSON object"},
		"empty body":            {"", "not a JSON object"},
		"unknown top-level key": {string(extraBody), `unexpected top-level key "sponsored"`},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rt := serveRaw(tc.body)
			_, _, err := ResolveLink(context.Background(), testLink,
				clientWith(rt), withLinkLookupIP("93.184.216.34"))
			if !errors.Is(err, ErrLinkResponseForm) {
				t.Fatalf("error = %v, want ErrLinkResponseForm", err)
			}
			if !strings.Contains(err.Error(), tc.reason) {
				t.Fatalf("rejected for the wrong reason:\n  got:  %v\n  want message containing: %q", err, tc.reason)
			}
		})
	}

	// Trimming is not optional: a valid document with a leading newline must
	// resolve, or the same body succeeds in one implementation and fails in
	// another.
	t.Run("leading and trailing whitespace is tolerated", func(t *testing.T) {
		rt := serveRaw("\n  " + jsonBody + "\n")
		if _, _, err := ResolveLink(context.Background(), testLink,
			clientWith(rt), withLinkLookupIP("93.184.216.34")); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestResolveLinkRejectsVoucherEnvelope(t *testing.T) {
	// A voucher carries nullifier_key; an anonymous public GET must never hand
	// one out, and the payer must not accept one if it does.
	voucher := IPPVoucher{
		V:       IPPCurrentVersion,
		Network: "eip155:97",
		Type:    MessageTypeVoucher,
		Payload: IPPVoucherV1{
			ReceivingAddress:  "0xrecv",
			PaymentCommitment: "0xpc",
			TokenAddress:      "0xtoken",
			Operator:          "0xop",
			Amount:            "1",
			Nonce:             "0x1",
			NullifierKey:      "0xnk",
			TreeID:            "1",
			Endpoint:          "https://pay.example.com",
		},
	}
	body, err := EncodeVoucher(voucher, FormatJSON)
	if err != nil {
		t.Fatalf("encode voucher: %v", err)
	}
	rt := serveRaw(body)
	_, _, err = ResolveLink(context.Background(), testLink,
		clientWith(rt), withLinkLookupIP("93.184.216.34"))
	if err == nil {
		t.Fatal("a voucher envelope must not resolve as a payment request")
	}
}

// -- Rules 5, 7, 10 --

func TestResolveLinkSendsAnonymousRequestForTheScannedURL(t *testing.T) {
	scanned := "https://Pay.Example.com:443/ipp/v1/link/" + testLinkToken
	rt := serveLinkJSON(t, "https://pay.example.com")
	_, _, err := ResolveLink(context.Background(), scanned,
		clientWith(rt), withLinkLookupIP("93.184.216.34"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req := rt.last
	if req.Method != http.MethodGet {
		t.Errorf("method = %s, want GET", req.Method)
	}
	// The scanned string, not the normal form. Only the DERIVED ENDPOINT is
	// normalised, because only it is compared; the request goes out as written
	// so the receiver's access log holds the URL the user actually scanned.
	// Rebuilding it here lowercased the host and dropped an explicit ":443",
	// which put a different line in that log than packages/ipp/src/link.ts did
	// for the same QR code.
	if got := req.URL.String(); got != scanned {
		t.Errorf("request URL = %q, want the scanned string %q", got, scanned)
	}
	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Errorf("Accept = %q, want application/json", got)
	}
	for _, h := range []string{"Authorization", "Cookie", "X-Api-Key", "User-Agent"} {
		if v := req.Header.Get(h); v != "" {
			t.Errorf("%s = %q, want unset (the resolve must not be attributable)", h, v)
		}
	}
	if len(req.Header) != 1 {
		t.Errorf("headers = %v, want only Accept", req.Header)
	}
}

func TestResolveLinkRejectsOversizedBody(t *testing.T) {
	t.Run("streamed past the cap", func(t *testing.T) {
		rt := serveRaw(strings.Repeat("x", maxLinkResponseBytes+1))
		_, _, err := ResolveLink(context.Background(), testLink,
			clientWith(rt), withLinkLookupIP("93.184.216.34"))
		if !errors.Is(err, ErrLinkResponseTooLarge) {
			t.Fatalf("error = %v, want ErrLinkResponseTooLarge", err)
		}
	})

	t.Run("declared over the cap is refused unread", func(t *testing.T) {
		rt := &countingRoundTripper{fn: func(*http.Request) (*http.Response, error) {
			resp := stringResponse(200, "{}", nil)
			resp.ContentLength = maxLinkResponseBytes + 1
			return resp, nil
		}}
		_, _, err := ResolveLink(context.Background(), testLink,
			clientWith(rt), withLinkLookupIP("93.184.216.34"))
		if !errors.Is(err, ErrLinkResponseTooLarge) {
			t.Fatalf("error = %v, want ErrLinkResponseTooLarge", err)
		}
	})
}

func TestResolveLinkMapsEnvelopeErrors(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		check  func(*testing.T, error)
	}{
		{
			name:   "unknown token",
			status: 404,
			body:   `{"error":{"type":"not_found_error","code":"link_not_found","message":"no such link"}}`,
			check: func(t *testing.T, err error) {
				var nf *NotFoundError
				if !errors.As(err, &nf) {
					t.Fatalf("error = %v, want *NotFoundError", err)
				}
				if nf.Code != "link_not_found" {
					t.Fatalf("code = %q", nf.Code)
				}
			},
		},
		{
			name:   "inactive request",
			status: 404,
			body:   `{"error":{"type":"not_found_error","code":"link_inactive","message":"cancelled"}}`,
			check: func(t *testing.T, err error) {
				var nf *NotFoundError
				if !errors.As(err, &nf) || nf.Code != "link_inactive" {
					t.Fatalf("error = %v, want *NotFoundError with link_inactive", err)
				}
			},
		},
		{
			name:   "expired request",
			status: 400,
			body:   `{"error":{"type":"invalid_request_error","code":"link_expired","message":"expired"}}`,
			check: func(t *testing.T, err error) {
				var ir *InvalidRequestError
				if !errors.As(err, &ir) {
					t.Fatalf("error = %v, want *InvalidRequestError", err)
				}
				if ir.Code != "link_expired" {
					t.Fatalf("code = %q", ir.Code)
				}
			},
		},
		{
			name:   "html error page",
			status: 502,
			body:   "<html>bad gateway</html>",
			check: func(t *testing.T, err error) {
				var ce *ConnectionError
				if !errors.As(err, &ce) {
					t.Fatalf("error = %v, want *ConnectionError", err)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, status := tc.body, tc.status
			rt := &countingRoundTripper{fn: func(*http.Request) (*http.Response, error) {
				return stringResponse(status, body, nil), nil
			}}
			_, _, err := ResolveLink(context.Background(), testLink,
				clientWith(rt), withLinkLookupIP("93.184.216.34"))
			tc.check(t, err)
		})
	}
}

func TestResolveLinkPropagatesTransportFailure(t *testing.T) {
	rt := &countingRoundTripper{fn: func(*http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("dial tcp: boom")
	}}
	_, _, err := ResolveLink(context.Background(), testLink,
		clientWith(rt), withLinkLookupIP("93.184.216.34"))
	var ce *ConnectionError
	if !errors.As(err, &ce) {
		t.Fatalf("error = %v, want *ConnectionError", err)
	}
}

// -- W1: the WHATWG "ends in a number" judgement, cross-checked --

// referenceEndsInANumber transcribes the WHATWG URL Standard's "ends in a number
// checker" (applied to a single label) and its IPv4 number parser, structured
// after the SPEC TEXT -- explicit radix selection, then a radix-R digit test --
// rather than after the implementation it is checking.
//
// It exists because the pinned rule names that algorithm, and the octal branch
// is exactly where a paraphrase ("all digits, or 0x-prefixed hex") and the real
// algorithm could part company. Two independently shaped predicates agreeing
// over the corpus below is the evidence; reading them side by side is not.
func referenceEndsInANumber(label string) bool {
	if label == "" {
		return false // parts is one empty item: the checker returns false
	}
	if refAllDigits(label, 10) {
		return true // step 4: last is non-empty and all ASCII digits
	}
	return refIPv4NumberParses(label) // step 5
}

func refIPv4NumberParses(input string) bool {
	if input == "" {
		return false
	}
	radix := 10
	switch {
	case len(input) >= 2 && (input[0] == '0') && (input[1] == 'x' || input[1] == 'X'):
		input, radix = input[2:], 16
	case len(input) >= 2 && input[0] == '0':
		input, radix = input[1:], 8
	}
	if input == "" {
		return true // the parser returns (0, true), which is not failure
	}
	return refAllDigits(input, radix)
}

func refAllDigits(s string, radix int) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		var v int
		switch {
		case c >= '0' && c <= '9':
			v = int(c - '0')
		case c >= 'a' && c <= 'f':
			v = int(c-'a') + 10
		case c >= 'A' && c <= 'F':
			v = int(c-'A') + 10
		default:
			return false
		}
		if v >= radix {
			return false
		}
	}
	return true
}

func TestEndsInANumberMatchesWHATWG(t *testing.T) {
	corpus := []string{
		// octal shapes -- the branch the pinned rule singles out
		"0", "00", "01", "07", "08", "09", "0177", "0777", "0800", "018", "0o17",
		// decimal
		"1", "9", "10", "113", "255", "3405803625", "999999999999999999999",
		// hex
		"0x", "0X", "0x0", "0xff", "0xFF", "0x7f000001", "0XABCDEF",
		"0xzz", "0xg", "0x1g", "0x-1", "0x0x0", "x0",
		// ordinary labels
		"com", "fi", "example", "invisipay", "a1", "1a", "ff", "abc", "xn--py-via",
		"pay", "api-test", "localhost", "e", "-", "--", "a-1",
		// empty
		"",
	}
	for _, label := range corpus {
		if got, want := endsInANumber(label), referenceEndsInANumber(label); got != want {
			t.Errorf("endsInANumber(%q) = %v, WHATWG checker says %v", label, got, want)
		}
	}

	// The specific answers the pinned rule depends on, asserted outright so the
	// cross-check cannot pass by both sides being wrong together.
	mustBeNumber := []string{"0", "0177", "08", "113", "3405803625", "0x", "0xff", "0x7f000001"}
	for _, label := range mustBeNumber {
		if !endsInANumber(label) {
			t.Errorf("endsInANumber(%q) = false, want true", label)
		}
	}
	mustNotBeNumber := []string{"com", "ff", "1a", "a1", "0xzz", "0x1g", "x0", "xn--py-via", "api-test"}
	for _, label := range mustNotBeNumber {
		if endsInANumber(label) {
			t.Errorf("endsInANumber(%q) = true, want false", label)
		}
	}

	// Exhaustive over the radix-selection boundary rather than over a hand-picked
	// list, because the corpus above is only as good as whoever wrote it. The
	// alphabet is chosen for the boundaries: "0" and "1"/"8"/"9" separate the
	// octal branch from decimal, "x"/"X" select hex, "a"/"f"/"F" are hex digits,
	// "z"/"g" are not, and "-"/"." are the non-digits a label may otherwise hold.
	t.Run("exhaustive over the radix boundaries", func(t *testing.T) {
		alphabet := []byte("0189xXafFzg-.")
		var buf []byte
		var rec func(depth int)
		rec = func(depth int) {
			s := string(buf)
			if got, want := endsInANumber(s), referenceEndsInANumber(s); got != want {
				t.Fatalf("endsInANumber(%q) = %v, WHATWG checker says %v", s, got, want)
			}
			if depth == 0 {
				return
			}
			for _, c := range alphabet {
				buf = append(buf, c)
				rec(depth - 1)
				buf = buf[:len(buf)-1]
			}
		}
		rec(4)
	})

	// The octal branch specifically: a leading-zero run and the same run bare.
	// WHATWG only reaches radix 8 for a string that is already all ASCII digits,
	// so its all-digits branch has claimed the input first - these must agree for
	// every one of them, and a divergence here would be invisible in a spot check.
	t.Run("exhaustive over octal-shaped labels", func(t *testing.T) {
		for n := 0; n < 50000; n++ {
			d := strconv.Itoa(n)
			for _, s := range []string{d, "0" + d, "00" + d} {
				if got, want := endsInANumber(s), referenceEndsInANumber(s); got != want {
					t.Fatalf("endsInANumber(%q) = %v, WHATWG checker says %v", s, got, want)
				}
			}
		}
	})
}

// -- W1: RFC 5952 section 4.2, all three clauses --

func TestResolveLinkEnforcesRFC5952ZeroCompression(t *testing.T) {
	// A table of PAIRS, not two independent lists: every non-canonical spelling
	// is stored next to the canonical spelling of the same address, and the
	// subtest asserts all three things at once - the two denote one address, the
	// first is refused, the second is accepted. Two separate lists let a clause
	// end up with only a rejection or only an acceptance, which is a one-way
	// test: it catches the rule going slack but not the rule tightening onto the
	// wrong spelling.
	pairs := []struct {
		clause       string
		nonCanonical string
		canonical    string
	}{
		{"4.2.1 a single zero group is never compressed",
			"1::2:3:4:5:6:7", "1:0:2:3:4:5:6:7"},
		{"4.2.1 single zero group, at the end",
			"1:2:3:4:5:6::8", "1:2:3:4:5:6:0:8"},
		{"4.2.2 a run of two or more must be compressed",
			"1:0:0:2:0:0:3:4", "1::2:0:0:3:4"},
		{"4.2.2 a long run must be compressed",
			"1:0:0:0:0:0:0:2", "1::2"},
		{"4.2.3 on a tie the leftmost run wins",
			"1:0:0:1::1:1", "1::1:0:0:1:1"},
	}

	for _, pair := range pairs {
		t.Run(pair.clause, func(t *testing.T) {
			// net.ParseIP is the oracle here, independent of the implementation
			// under test: it confirms the pair really is one address written two
			// ways, which is what makes the rest of the subtest meaningful.
			lhs, rhs := net.ParseIP(pair.nonCanonical), net.ParseIP(pair.canonical)
			if lhs == nil || rhs == nil || !lhs.Equal(rhs) {
				t.Fatalf("test bug: %q and %q are not the same address", pair.nonCanonical, pair.canonical)
			}

			rejectsBecause(t, "https://["+pair.nonCanonical+"]/ipp/v1/link/"+testLinkToken,
				ErrLinkURLForm, "not RFC 5952 canonical form", failLookupIP(t))

			wantEndpoint := "https://[" + pair.canonical + "]"
			rt := serveLinkJSON(t, wantEndpoint)
			_, endpoint, err := ResolveLink(context.Background(),
				wantEndpoint+"/ipp/v1/link/"+testLinkToken,
				clientWith(rt), withLinkLookupIP("93.184.216.34"))
			if err != nil {
				t.Fatalf("canonical spelling %q was refused: %v", pair.canonical, err)
			}
			if endpoint != wantEndpoint {
				t.Fatalf("endpoint = %q, want %q", endpoint, wantEndpoint)
			}

			if got := canonicalIPv6Text(rhs.To16()); got != pair.canonical {
				t.Errorf("canonicalIPv6Text = %q, want %q", got, pair.canonical)
			}
		})
	}

	// Spellings with no zero group to argue about, so no clause applies.
	for _, authority := range []string{"[1:2:3:4:5:6:7:8]", "[1:2:3:4:5:6:7:0]", "[::ffff:1.2.3.4]"} {
		t.Run("accepts "+authority, func(t *testing.T) {
			rt := serveLinkJSON(t, "https://"+authority)
			_, endpoint, err := ResolveLink(context.Background(),
				"https://"+authority+"/ipp/v1/link/"+testLinkToken,
				clientWith(rt), withLinkLookupIP("93.184.216.34"))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if endpoint != "https://"+authority {
				t.Fatalf("endpoint = %q", endpoint)
			}
		})
	}
}

// The IPv6 literal parser is now on the accepting side of the decision, so it
// needs its own rejection coverage: before, net.ParseIP decided what an IPv6
// literal was, and whatever it let through was caught only by the canonical-form
// comparison. Each case below pins the PARSER (reason "is not an IPv6 literal"),
// not that comparison.
func TestResolveLinkRejectsMalformedIPv6Literals(t *testing.T) {
	for _, inner := range []string{
		"1::2::3",           // two runs of "::"
		"1:::2",             // three colons
		":1:2:3:4:5:6:7",    // leading single colon
		"1:2:3:4:5:6:7",     // seven groups, no "::"
		"1:2:3:4:5:6:7:8:9", // nine groups
		"1:2:3:4:5:6:7:8::", // "::" with nothing left to stand for
		"12345::1",          // five hex digits in a group
		"1:2:3:4:5:6:7:g",   // not a hex digit
		"::ffff:1.2.3.4.5",  // five octets in the embedded quad
		"::ffff:01.2.3.4",   // leading zero in the embedded quad
		"::ffff:256.1.1.1",  // octet out of range
		"1.2.3.4::1",        // dotted quad before the "::"
		"::1.2.3.4:5",       // dotted quad not last
		"1:2",               // too few groups
	} {
		t.Run(inner, func(t *testing.T) {
			rejectsBecause(t, "https://["+inner+"]/ipp/v1/link/"+testLinkToken,
				ErrLinkURLForm, "is not an IPv6 literal", failLookupIP(t))
		})
	}
}

// -- W1: RFC 1123 label length and hyphen position --

func TestResolveLinkEnforcesLDHLabelShape(t *testing.T) {
	label63 := strings.Repeat("a", 63)
	label64 := strings.Repeat("a", 64)

	rejected := map[string]struct{ authority, reason string }{
		"label of 64":              {label64 + ".example.com", "longer than 63 characters"},
		"last label of 64":         {"example." + label64, "longer than 63 characters"},
		"leading hyphen":           {"-pay.example.com", "starts or ends with a hyphen"},
		"trailing hyphen":          {"pay-.example.com", "starts or ends with a hyphen"},
		"leading hyphen on a TLD":  {"pay.example.-com", "starts or ends with a hyphen"},
		"trailing hyphen on a TLD": {"pay.example.com-", "starts or ends with a hyphen"},
		"hyphen-only label":        {"pay.-.com", "starts or ends with a hyphen"},
	}
	for name, tc := range rejected {
		t.Run("rejects "+name, func(t *testing.T) {
			rejectsBecause(t, "https://"+tc.authority+"/ipp/v1/link/"+testLinkToken,
				ErrLinkURLForm, tc.reason, failLookupIP(t))
		})
	}

	accepted := map[string]string{
		"label of exactly 63": label63 + ".example.com",
		"inner hyphens":       "api-test.invisipay.fi",
		"double hyphen":       "xn--py-via.example.com",
		"single character":    "a.example.com",
	}
	for name, authority := range accepted {
		t.Run("accepts "+name, func(t *testing.T) {
			wantEndpoint := "https://" + authority
			rt := serveLinkJSON(t, wantEndpoint)
			_, endpoint, err := ResolveLink(context.Background(),
				wantEndpoint+"/ipp/v1/link/"+testLinkToken,
				clientWith(rt), withLinkLookupIP("93.184.216.34"))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if endpoint != wantEndpoint {
				t.Fatalf("endpoint = %q, want %q", endpoint, wantEndpoint)
			}
		})
	}
}

// -- W3: the scheme is a lowercase literal, the host is normalised --

// These two rules point opposite ways on purpose, so both need pinning: an
// uppercase scheme is refused, an uppercase host is accepted and lowercased.
// RFC 3986 makes both case-insensitive, so "normalise it" reads as defensible
// for either -- and that ambiguity is what splits implementations.
func TestResolveLinkSchemeIsALowercaseLiteral(t *testing.T) {
	for _, scheme := range []string{"HTTPS", "Https", "httpS", "HTTP", "Http"} {
		t.Run("rejects scheme "+scheme, func(t *testing.T) {
			rejectsBecause(t, scheme+"://pay.example.com/ipp/v1/link/"+testLinkToken,
				ErrLinkURLForm, "is not lowercase", failLookupIP(t))
		})
	}

	t.Run("rejects an uppercase scheme on the unsafe entry point too", func(t *testing.T) {
		rt := &countingRoundTripper{}
		_, _, err := ResolveLinkUnsafe(context.Background(),
			"HTTP://10.0.0.7/ipp/v1/link/"+testLinkToken, clientWith(rt))
		if !errors.Is(err, ErrLinkURLForm) {
			t.Fatalf("error = %v, want ErrLinkURLForm", err)
		}
		if rt.calls != 0 {
			t.Fatalf("round trips = %d, want 0", rt.calls)
		}
	})

	// The other half: the host IS accepted and lowercased. If this ever starts
	// failing, the scheme rule has been copied onto the host by mistake.
	t.Run("accepts and lowercases an uppercase host", func(t *testing.T) {
		rt := serveLinkJSON(t, "https://PAY.Example.COM")
		_, endpoint, err := ResolveLink(context.Background(),
			"https://Pay.Example.COM/ipp/v1/link/"+testLinkToken,
			clientWith(rt), withLinkLookupIP("93.184.216.34"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if endpoint != "https://pay.example.com" {
			t.Fatalf("endpoint = %q, want the lowercased host", endpoint)
		}
	})
}

// -- W3: the body contract is judged on bytes --

func TestResolveLinkBodyIsJudgedOnBytes(t *testing.T) {
	jsonBody, err := Encode(sampleLinkRequest("https://pay.example.com"), FormatJSON)
	if err != nil {
		t.Fatalf("encode json: %v", err)
	}

	// resolveBytes serves an exact byte sequence. Building the prefix as bytes is
	// the whole point of these cases: handing the implementation an already
	// decoded string skips the bytes-to-text step where BOM stripping happens,
	// which is the layer under test. That would be a false green.
	resolveBytes := func(t *testing.T, prefix []byte) error {
		t.Helper()
		rt := serveRaw(string(append(append([]byte{}, prefix...), jsonBody...)))
		_, _, err := ResolveLink(context.Background(), testLink,
			clientWith(rt), withLinkLookupIP("93.184.216.34"))
		return err
	}

	t.Run("rejects a UTF-8 BOM", func(t *testing.T) {
		// EF BB BF is stripped by TextDecoder and fetch().text() but kept by
		// Buffer.toString and Go's string(b), so a code-point rule splits here.
		err := resolveBytes(t, []byte{0xEF, 0xBB, 0xBF})
		if !errors.Is(err, ErrLinkResponseForm) {
			t.Fatalf("error = %v, want ErrLinkResponseForm", err)
		}
		if !strings.Contains(err.Error(), "UTF-8 BOM") {
			t.Fatalf("rejected for the wrong reason: %v", err)
		}
	})

	// Bytes that some standard library trims but RFC 8259 does not allow.
	notJSONWhitespace := map[string][]byte{
		"NEL (U+0085), which Go's TrimSpace strips": {0xC2, 0x85},
		"vertical tab":               {0x0B},
		"form feed":                  {0x0C},
		"NBSP (U+00A0)":              {0xC2, 0xA0},
		"line separator (U+2028)":    {0xE2, 0x80, 0xA8},
		"ideographic space (U+3000)": {0xE3, 0x80, 0x80},
		"NUL":                        {0x00},
	}
	for name, prefix := range notJSONWhitespace {
		t.Run("rejects a leading "+name, func(t *testing.T) {
			err := resolveBytes(t, prefix)
			if !errors.Is(err, ErrLinkResponseForm) {
				t.Fatalf("error = %v, want ErrLinkResponseForm", err)
			}
			if !strings.Contains(err.Error(), "not a JSON object") {
				t.Fatalf("rejected for the wrong reason: %v", err)
			}
		})
	}

	// The four RFC 8259 section 2 whitespace bytes, which MUST be skipped.
	t.Run("accepts the four RFC 8259 whitespace bytes", func(t *testing.T) {
		if err := resolveBytes(t, []byte{0x20, 0x09, 0x0A, 0x0D, 0x20}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestResolveLinkRejectsDuplicateTopLevelKeys(t *testing.T) {
	// A repeated key has no single meaning: folding it is a parser decision
	// (last-wins, first-wins, or an error), and after folding the key set still
	// looks legal, so the closed-set check cannot catch it.
	base, err := Encode(sampleLinkRequest("https://pay.example.com"), FormatJSON)
	if err != nil {
		t.Fatalf("encode json: %v", err)
	}
	for name, body := range map[string]string{
		"duplicate v":       `{"v":1,` + base[1:len(base)-1] + `,"v":2}`,
		"duplicate type":    `{"type":"request",` + base[1:len(base)-1] + `,"type":"voucher"}`,
		"duplicate network": `{"network":"eip155:1",` + base[1:len(base)-1] + `,"network":"eip155:97"}`,
	} {
		t.Run(name, func(t *testing.T) {
			rt := serveRaw(body)
			_, _, err := ResolveLink(context.Background(), testLink,
				clientWith(rt), withLinkLookupIP("93.184.216.34"))
			if !errors.Is(err, ErrLinkResponseForm) {
				t.Fatalf("error = %v, want ErrLinkResponseForm", err)
			}
			if !strings.Contains(err.Error(), "appears more than once") {
				t.Fatalf("rejected for the wrong reason: %v", err)
			}
		})
	}

	// A duplicate nested key is NOT this rule's business: the envelope contract
	// is about top-level keys, and refusing more would be a rule no other
	// implementation was told to apply.
	t.Run("a duplicate key inside payload is not this rule", func(t *testing.T) {
		body := strings.Replace(base, `"identity_commitment":"0xabc"`,
			`"identity_commitment":"0xabc","identity_commitment":"0xabc"`, 1)
		if body == base {
			t.Fatal("test bug: payload key not found to duplicate")
		}
		rt := serveRaw(body)
		_, _, err := ResolveLink(context.Background(), testLink,
			clientWith(rt), withLinkLookupIP("93.184.216.34"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// -- W1: the rejected code-point set is enumerated, not borrowed --

// This pins the SET itself, member by member. The end-to-end URL cases above
// exercise the interesting members through the real path, but they cannot notice
// a member quietly going missing from the enumeration.
//
// It is also the guard W16 asks for. The reason the set may not be spelled
// unicode.IsSpace is that the tail of Zs moves with the toolchain's Unicode
// version (U+180E was Zs through Unicode 6.2), and today the two agree - so no
// end-to-end case can tell them apart. The membership table can.
func TestIsRejectedURLCodePointEnumeration(t *testing.T) {
	rejected := []rune{
		0x0000, 0x0001, 0x0009, 0x000a, 0x000b, 0x000c, 0x000d, 0x001f, // C0
		0x007f,                         // DEL
		0x0080, 0x0085, 0x0090, 0x009f, // C1
		0x0020,         // SPACE
		0x00a0, 0x1680, // NO-BREAK, OGHAM
		0x2000, 0x2001, 0x2002, 0x2003, 0x2004, 0x2005, // EN QUAD..
		0x2006, 0x2007, 0x2008, 0x2009, 0x200a, // ..HAIR SPACE
		0x2028, 0x2029, 0x202f, 0x205f, 0x3000, // separators, narrow, medium, ideographic
		0xfeff, // BOM
	}
	for _, r := range rejected {
		if !isRejectedURLCodePoint(r) {
			t.Errorf("isRejectedURLCodePoint(U+%04X) = false, want true", r)
		}
	}

	// Near misses on every boundary of the enumeration. A member accidentally
	// widened into a range would show up here.
	accepted := []rune{
		0x0021, 0x0041, 0x007a, 0x007e, // ordinary ASCII
		0x00a1, 0x1681, 0x1fff, 0x200b, // just past NO-BREAK, OGHAM, before EN QUAD, ZWSP
		0x200c, 0x200d, 0x200e, 0x2027, // ZWNJ, ZWJ, LRM, just before LINE SEPARATOR
		0x202a, 0x202e, 0x2030, 0x205e, // just past PARAGRAPH SEP, before/after narrow+medium
		0x2060, 0x2fff, 0x3001, 0xfefe, // WORD JOINER, before/after IDEOGRAPHIC SPACE, before BOM
		0xff00, 0x10000,
		// U+180E: Zs through Unicode 6.2, not Zs now. It is NOT in the pinned
		// set, so it must stay accepted no matter what the toolchain thinks.
		0x180e,
	}
	for _, r := range accepted {
		if isRejectedURLCodePoint(r) {
			t.Errorf("isRejectedURLCodePoint(U+%04X) = true, want false", r)
		}
	}
}

// -- W3: the body must be well-formed UTF-8, judged on bytes --

func TestResolveLinkRejectsInvalidUTF8Body(t *testing.T) {
	jsonBody, err := Encode(sampleLinkRequest("https://pay.example.com"), FormatJSON)
	if err != nil {
		t.Fatalf("encode json: %v", err)
	}

	// A non-fatal decoder maps every one of these to U+FFFD, so a rule that
	// looked at decoded text would see a legal document - and 0xFF and 0xFE
	// would become the same character.
	cases := map[string][]byte{
		"0xFF inside a string value": {0xFF},
		"0xFE inside a string value": {0xFE},
		"lone continuation byte":     {0x80},
		"truncated two-byte seq":     {0xC2},
		"truncated three-byte seq":   {0xE2, 0x80},
		"overlong encoding of '/'":   {0xC0, 0xAF},
		"UTF-16 surrogate encoded":   {0xED, 0xA0, 0x80},
		"out of range four-byte":     {0xF5, 0x80, 0x80, 0x80},
	}
	for name, bad := range cases {
		t.Run(name, func(t *testing.T) {
			// Splice the illegal bytes into the merchantName value, where they
			// would survive into the parsed request if the body were accepted.
			body := strings.Replace(jsonBody, "Coffee Shop", "Coffee "+string(bad)+"Shop", 1)
			if body == jsonBody {
				t.Fatal("test bug: splice point not found")
			}
			rt := serveRaw(body)
			_, _, err := ResolveLink(context.Background(), testLink,
				clientWith(rt), withLinkLookupIP("93.184.216.34"))
			if !errors.Is(err, ErrLinkResponseForm) {
				t.Fatalf("error = %v, want ErrLinkResponseForm", err)
			}
			if !strings.Contains(err.Error(), "well-formed UTF-8") {
				t.Fatalf("rejected for the wrong reason: %v", err)
			}
		})
	}

	// 0xFF and 0xFE decode to the same character under a non-fatal decoder, so
	// two bodies that differ only there must not both be accepted - and here
	// neither is.
	t.Run("bodies differing only in an illegal byte are both refused", func(t *testing.T) {
		for _, b := range []byte{0xFF, 0xFE} {
			body := strings.Replace(jsonBody, "Coffee Shop", "Coffee"+string([]byte{b}), 1)
			rt := serveRaw(body)
			if _, _, err := ResolveLink(context.Background(), testLink,
				clientWith(rt), withLinkLookupIP("93.184.216.34")); !errors.Is(err, ErrLinkResponseForm) {
				t.Errorf("byte %#x: error = %v, want ErrLinkResponseForm", b, err)
			}
		}
	})

	t.Run("multi-byte UTF-8 in a value is fine when well formed", func(t *testing.T) {
		body := strings.Replace(jsonBody, "Coffee Shop", "Caf"+cp(0x00e9)+" "+cp(0x5317)+cp(0x4eac), 1)
		rt := serveRaw(body)
		if _, _, err := ResolveLink(context.Background(), testLink,
			clientWith(rt), withLinkLookupIP("93.184.216.34")); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// -- W3: top-level keys are written literally, with no escapes --

func TestResolveLinkRejectsEscapedTopLevelKeys(t *testing.T) {
	jsonBody, err := Encode(sampleLinkRequest("https://pay.example.com"), FormatJSON)
	if err != nil {
		t.Fatalf("encode json: %v", err)
	}

	// "v" is "v". A set check on the DECODED key accepts it; a set check on
	// the raw bytes calls it unknown. Both readings were implementable from the
	// earlier wording, and they disagree on this exact body - so the spelling is
	// refused instead.
	// jsonEsc builds a JSON \uXXXX escape from code points, so that no editor,
	// formatter or transport can normalise the escape away. If these were written
	// literally and something collapsed them, the cases would still compile and
	// still pass - against an unescaped body that proves nothing (W7).
	jsonEsc := func(hex string) string { return cp(0x5c) + "u" + hex }
	cases := map[string]string{
		"escaped v":       strings.Replace(jsonBody, `"v":`, `"`+jsonEsc("0076")+`":`, 1),
		"escaped network": strings.Replace(jsonBody, `"network":`, `"`+jsonEsc("006e")+`etwork":`, 1),
		"escaped type":    strings.Replace(jsonBody, `"type":`, `"`+jsonEsc("0074")+`ype":`, 1),
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if body == jsonBody {
				t.Fatal("test bug: splice point not found")
			}
			rt := serveRaw(body)
			_, _, err := ResolveLink(context.Background(), testLink,
				clientWith(rt), withLinkLookupIP("93.184.216.34"))
			if !errors.Is(err, ErrLinkResponseForm) {
				t.Fatalf("error = %v, want ErrLinkResponseForm", err)
			}
			if !strings.Contains(err.Error(), "backslash escape") {
				t.Fatalf("rejected for the wrong reason: %v", err)
			}
		})
	}

	// The escape is also how a duplicate could have hidden from both a decoded
	// check and a raw check at once.
	t.Run("a duplicate spelled with an escape", func(t *testing.T) {
		body := `{"v":1,"` + cp(0x5c) + `u0076":1,` + jsonBody[len(`{"v":1,`):]
		rt := serveRaw(body)
		_, _, err := ResolveLink(context.Background(), testLink,
			clientWith(rt), withLinkLookupIP("93.184.216.34"))
		if !errors.Is(err, ErrLinkResponseForm) {
			t.Fatalf("error = %v, want ErrLinkResponseForm", err)
		}
	})

	// Escapes inside a VALUE are ordinary JSON and must keep working: the rule is
	// about key spelling only.
	t.Run("an escape inside a value is untouched", func(t *testing.T) {
		body := strings.Replace(jsonBody, `"Coffee Shop"`, `"Caf`+cp(0x5c)+`u00e9"`, 1)
		if body == jsonBody {
			t.Fatal("test bug: splice point not found")
		}
		rt := serveRaw(body)
		got, _, err := ResolveLink(context.Background(), testLink,
			clientWith(rt), withLinkLookupIP("93.184.216.34"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Meta[MetaKeyMerchantName] != "Caf"+cp(0x00e9) {
			t.Fatalf("merchantName = %q, want the unescaped value", got.Meta[MetaKeyMerchantName])
		}
	})
}

// -- W3: trailing content lands in a named sentinel --

func TestResolveLinkRejectsTrailingContent(t *testing.T) {
	jsonBody, err := Encode(sampleLinkRequest("https://pay.example.com"), FormatJSON)
	if err != nil {
		t.Fatalf("encode json: %v", err)
	}

	// Before this rule existed these reached the shared envelope decoder, which
	// reported them with a bare error unwrapping to NO sentinel: not
	// ErrLinkResponseForm, not *ConnectionError. A caller could not tell whether
	// it had been refused for a safety reason.
	cases := map[string]string{
		"NEL after the envelope":           jsonBody + cp(0x0085),
		"letter after the envelope":        jsonBody + "x",
		"second object after the envelope": jsonBody + "{}",
		"NBSP after the envelope":          jsonBody + cp(0x00a0),
		"comma after the envelope":         jsonBody + ",",
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			rt := serveRaw(body)
			_, _, err := ResolveLink(context.Background(), testLink,
				clientWith(rt), withLinkLookupIP("93.184.216.34"))
			if !errors.Is(err, ErrLinkResponseForm) {
				t.Fatalf("error = %v, want ErrLinkResponseForm", err)
			}
			if !strings.Contains(err.Error(), "trailing content") {
				t.Fatalf("rejected for the wrong reason: %v", err)
			}
		})
	}

	t.Run("trailing RFC 8259 whitespace is fine", func(t *testing.T) {
		rt := serveRaw(jsonBody + " \t\r\n")
		if _, _, err := ResolveLink(context.Background(), testLink,
			clientWith(rt), withLinkLookupIP("93.184.216.34")); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
