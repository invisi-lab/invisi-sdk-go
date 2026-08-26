package ipp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"
)

// Resolving an IPP short link is the one step where a wallet follows a URL it
// got from an untrusted QR code, so every check that keeps that safe lives in
// this file rather than in each consumer. Three properties are load-bearing and
// easy to destroy while editing:
//
//   - Every URL check runs BEFORE the first byte leaves the process. Moving any
//     of them after the round trip turns "scan any QR code" into "GET any URL".
//   - ResolveLinkUnsafe exists as a separate exported name, never as a bool
//     argument, so lint/grep guards in consuming repos can ban it from release
//     builds. Collapsing the two entry points, or renaming this one, removes the
//     only thing those guards can match on.
//   - Every rule here is a WIRE rule shared with packages/ipp/src/link.ts and
//     the 06-ipp/6.6 conformance fixture. Loosening or tightening one on this
//     side alone produces a code that pays in one wallet and not in another,
//     which is the exact failure the shared rules exist to prevent. The written
//     form of a URL is refused or accepted as-is; it is never repaired, because
//     two parsers repair differently and the endpoint cross-check then compares
//     two different strings.

var (
	// wireURLPattern is scheme://authority[path] with no query and no fragment
	// anywhere in the string. Query and fragment are refused rather than
	// stripped, so they cannot survive as an unshown part of a URL the user
	// approved.
	wireURLPattern = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9+.\-]*)://([^/?#]*)(/[^?#]*)?$`)

	// linkPathPattern matches a trailing /ipp/v1/link/<token>; group 1 is the
	// endpoint path prefix, which may be empty. A prefix is allowed because a
	// third party mounting IPP under a shared gateway path is a normal
	// deployment and the endpoint cross-check binds it either way.
	linkPathPattern = regexp.MustCompile(`^(.*)/ipp/v1/link/([A-Za-z0-9_\-]{22})$`)

	// pathPrefixPattern is one or more non-empty segments. It is what rejects an
	// empty segment ("//") in a path.
	pathPrefixPattern = regexp.MustCompile(`^(/[^/]+)+$`)

	// dotSegmentPattern matches a path segment that IS "." or ".." in any
	// spelling. Percent-encoded forms count: matching only the literal dots
	// lets "%2e%2e" straight through. Such a segment is refused, never folded —
	// Go's url.Parse leaves dot segments alone while a WHATWG parser folds them,
	// so folding would derive two different endpoints from one written link.
	dotSegmentPattern = regexp.MustCompile(`^(?i:\.|%2e){1,2}$`)

	// canonicalPortPattern admits only the one spelling of a port: no leading
	// zero, no empty string. A leading zero is what a numeric parser silently
	// accepts and a string comparison does not, so ":0443" would drop the port
	// on one side and keep it on the other.
	canonicalPortPattern = regexp.MustCompile(`^[1-9][0-9]{0,4}$`)

	// pcharSegmentPattern is RFC 3986 pchar: unreserved / pct-encoded /
	// sub-delims / ":" / "@", with pct-encoded well formed.
	//
	// This is an EXPLICIT charset on purpose. The rule it replaced compared the
	// path against the local parser's normal form, which reads like one rule but
	// is really "whatever my URL library does": Go re-encodes "|" and "^" and
	// rejects a stray "%", a WHATWG parser passes all three through, so the same
	// printed code was payable in one wallet and not the other. A charset any
	// implementation can check without a URL library is the only shape that can
	// actually be shared.
	pcharSegmentPattern = regexp.MustCompile(`^(?:[A-Za-z0-9\-._~!$&'()*+,;=:@]|%[0-9A-F]{2})*$`)

	// pcharSegmentAnyCasePattern is pcharSegmentPattern with lowercase hex
	// allowed. It exists only to tell "this is not pchar at all" apart from
	// "this is pchar but the percent-encoding is not uppercase", so each of the
	// two rules can be pinned on its own reason.
	pcharSegmentAnyCasePattern = regexp.MustCompile(`^(?:[A-Za-z0-9\-._~!$&'()*+,;=:@]|%[0-9A-Fa-f]{2})*$`)

	// ldhLabelPattern is a letter-digit-hyphen label. Charset only: see
	// canonicalRegName for the restrictions deliberately left out.
	ldhLabelPattern = regexp.MustCompile(`^[A-Za-z0-9\-]+$`)

	// dottedQuadPattern is the one canonical spelling of an IPv4 literal: four
	// decimal octets, no leading zero. The octet range is checked numerically.
	dottedQuadPattern = regexp.MustCompile(`^(?:0|[1-9][0-9]{0,2})(?:\.(?:0|[1-9][0-9]{0,2})){3}$`)
)

const (
	// defaultLinkTimeout bounds the whole resolve, including DNS and TLS.
	defaultLinkTimeout = 10 * time.Second
	// maxLinkResponseBytes bounds the response body. An IPPMessage is ~1 KiB;
	// the slack is for meta, not for letting a hostile resolver stream forever.
	maxLinkResponseBytes = 64 << 10
	maxPort              = 65535
	// maxHostLabelLen is the RFC 1123 label limit.
	maxHostLabelLen = 63
)

// envelopeTopLevelKeys is the closed set of IPPMessage top-level keys
// (06-ipp/6.1 §JSON 格式). The payer rejecting anything else is what makes the
// forward-compatibility promise real: the envelope extension point is `meta`,
// and a sixth top-level key is a breaking change that must bump `v`. Note this
// is the opposite stance from machine error codes, which are an open set.
var envelopeTopLevelKeys = map[string]struct{}{
	"v": {}, "network": {}, "type": {}, "meta": {}, "payload": {},
}

// Errors reported by ResolveLink / ResolveLinkUnsafe for client-side rule
// violations. Server-reported failures keep the existing envelope error types
// (*NotFoundError, *InvalidRequestError, ...) so callers handle them the same
// way they already handle IPPClient failures.
//
// The wrapped message names WHICH check fired. That is not cosmetic: one bad
// URL usually trips several checks, so a test that only asserts "rejected"
// keeps passing when the rule it meant to cover is deleted.
var (
	// ErrLinkShape means the URL is not an IPP link URL at all: wrong scheme
	// family, no host, wrong path, a token that is not 22 base64url characters,
	// or a malformed endpoint path prefix.
	ErrLinkShape = errors.New("ipp: link URL does not end with /ipp/v1/link/<22-char token>")
	// ErrLinkSchemeNotHTTPS means the URL is http and this entry point requires
	// https. ResolveLinkUnsafe accepts it.
	ErrLinkSchemeNotHTTPS = errors.New("ipp: link URL scheme must be https")
	// ErrLinkURLForm means the URL is not in the one written form the wire rules
	// admit: userinfo, query, fragment, whitespace or control characters, a
	// non-ASCII or backslash-bearing authority, a non-canonical port, a dot
	// segment, or a spelling the parser would rewrite.
	ErrLinkURLForm = errors.New("ipp: link URL is not in canonical wire form")
	// ErrLinkPrivateHost means the target is, or resolves to, a non-public
	// address.
	ErrLinkPrivateHost = errors.New("ipp: link host resolves to a non-public address")
	// ErrLinkRedirect means the resolver answered with a 3xx. The JSON
	// representation must be served by the link URL's own host.
	ErrLinkRedirect = errors.New("ipp: link resolver returned a redirect")
	// ErrLinkResponseTooLarge means the response body exceeded the size limit.
	ErrLinkResponseTooLarge = errors.New("ipp: link response exceeds size limit")
	// ErrLinkResponseForm means the body is not a bare JSON IPPMessage: not an
	// object, or carrying a top-level key outside the closed envelope set.
	ErrLinkResponseForm = errors.New("ipp: link response is not a bare JSON IPPMessage")
	// ErrLinkEndpointMismatch means payload.endpoint is not the endpoint the
	// link was fetched from, or is not in a form that can be compared with it.
	// This is the nail that welds the short link to the protocol; without it the
	// link URL and the payment target can diverge.
	ErrLinkEndpointMismatch = errors.New("ipp: link URL does not match payload endpoint")
)

// linkConfig holds per-call resolve settings.
type linkConfig struct {
	httpClient *http.Client
	timeout    time.Duration
	// lookupIP is the pre-flight resolver. The connect-time re-check lives in
	// the guarded transport's dialer; both exist because DNS can answer
	// differently between the two.
	lookupIP func(ctx context.Context, host string) ([]net.IP, error)
}

// LinkOption customises a ResolveLink / ResolveLinkUnsafe call.
type LinkOption func(*linkConfig)

// WithLinkHTTPClient overrides the *http.Client used for the resolve. Redirect
// following and the cookie jar are always overridden on a copy of the supplied
// client, and an *http.Transport is cloned with the non-public-address dial
// guard installed — a transport of any other type dials on its own terms, so
// supply one only when it does not reach the network (tests) or when the caller
// owns the egress policy. Passing nil is a no-op.
func WithLinkHTTPClient(hc *http.Client) LinkOption {
	return func(c *linkConfig) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithLinkTimeout overrides the overall resolve timeout. Non-positive is a no-op.
func WithLinkTimeout(d time.Duration) LinkOption {
	return func(c *linkConfig) {
		if d > 0 {
			c.timeout = d
		}
	}
}

// ResolveLink fetches the IPPRequest behind an IPP short link and returns it
// together with the endpoint derived from the link URL (feed that to
// NewIPPClient). It requires https and refuses targets that resolve to a
// non-public address.
func ResolveLink(ctx context.Context, linkURL string, opts ...LinkOption) (*IPPRequest, string, error) {
	return resolveLink(ctx, linkURL, false, opts)
}

// ResolveLinkUnsafe is ResolveLink with the transport-security rules relaxed:
// it accepts http and targets on loopback / private / link-local addresses. The
// wire-form rules are NOT relaxed — they decide what a link URL means, not how
// safe the target is.
//
// It exists for local deployments and dev builds only. Release builds must not
// reach it — consuming repos ban this symbol in their own lint gates, which is
// why the relaxation is a separate exported name and not an argument. The TS
// counterpart is resolveIPPLinkUnsafe; both carry "Unsafe" so one guard literal
// covers the pair.
func ResolveLinkUnsafe(ctx context.Context, linkURL string, opts ...LinkOption) (*IPPRequest, string, error) {
	return resolveLink(ctx, linkURL, true, opts)
}

func resolveLink(ctx context.Context, linkURL string, unsafe bool, opts []LinkOption) (*IPPRequest, string, error) {
	cfg := &linkConfig{
		timeout:  defaultLinkTimeout,
		lookupIP: defaultLookupIP,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	// Rules 1-3: shape, wire form, endpoint derivation. All of it before any
	// network use, so an unrecognised QR code costs zero requests.
	endpoint, target, requestURL, err := parseLinkURL(linkURL, unsafe)
	if err != nil {
		return nil, "", err
	}

	ctx, cancel := context.WithTimeout(ctx, cfg.timeout)
	defer cancel()

	// Rule 4: judged on the resolved address, not the literal host.
	if !unsafe {
		if err := assertPublicTarget(ctx, cfg.lookupIP, target); err != nil {
			return nil, "", err
		}
	}

	// Rule 5: anonymous. No credentials, no cookie jar, no custom headers —
	// a fingerprintable header set would identify the payer to a receiver it
	// has not paid yet.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("ipp: link: create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := hardenedClient(cfg.httpClient, unsafe).Do(req)
	if err != nil {
		return nil, "", &ConnectionError{Message: "link request failed", Err: err}
	}
	defer resp.Body.Close()

	// Rule 6: a 3xx is a failure, never a hop. The whole reason the derived
	// endpoint can be trusted is that the JSON came from that exact host.
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		return nil, "", fmt.Errorf("ipp: link: HTTP %d: %w", resp.StatusCode, ErrLinkRedirect)
	}

	// Rule 7: bounded body. The declared length is checked first so an oversized
	// response is refused without being read.
	if resp.ContentLength > maxLinkResponseBytes {
		return nil, "", fmt.Errorf("ipp: link: declared body of %d bytes over %d: %w",
			resp.ContentLength, maxLinkResponseBytes, ErrLinkResponseTooLarge)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxLinkResponseBytes+1))
	if err != nil {
		return nil, "", &ConnectionError{
			Message:    "reading link response body failed",
			StatusCode: resp.StatusCode,
			Err:        err,
		}
	}
	if len(body) > maxLinkResponseBytes {
		return nil, "", fmt.Errorf("ipp: link: body over %d bytes: %w", maxLinkResponseBytes, ErrLinkResponseTooLarge)
	}

	// Rule 10: server failures reuse the shared envelope decoder, so link
	// errors surface as the same typed errors as every other IPP call.
	if resp.StatusCode >= 400 {
		return nil, "", newEnvelopeError(resp.StatusCode, body)
	}

	// Rule 8.
	request, err := decodeLinkBody(body)
	if err != nil {
		return nil, "", err
	}

	// Rule 9: cross-check.
	if request.Payload.Endpoint == nil || *request.Payload.Endpoint == "" {
		return nil, "", fmt.Errorf("ipp: link: payload carries no endpoint to cross-check against %q: %w", endpoint, ErrLinkEndpointMismatch)
	}
	declared, err := normalizeEndpoint(*request.Payload.Endpoint, unsafe)
	if err != nil {
		// Both sentinels: callers branching on "the cross-check did not pass"
		// need ErrLinkEndpointMismatch, while a test pinning which wire rule
		// fired needs the specific one.
		return nil, "", fmt.Errorf("ipp: link: payload endpoint %q cannot be compared: %w (%w)",
			*request.Payload.Endpoint, err, ErrLinkEndpointMismatch)
	}
	if declared != endpoint {
		return nil, "", fmt.Errorf("ipp: link: derived endpoint %q != payload endpoint %q: %w", endpoint, declared, ErrLinkEndpointMismatch)
	}

	return request, endpoint, nil
}

// ── Rules 1-3: shape, wire form, endpoint derivation ────────────────

// wireURL is a URL that passed every written-form rule, reduced to the parts the
// endpoint comparison is built from.
type wireURL struct {
	// origin is scheme://authority with the host lowercased and a default port
	// dropped.
	origin string
	// host is the lowercased hostname, with no port and no brackets.
	host string
	// hostIP is the parsed address when host is an IP literal, nil for a
	// registered name. Carried so rule 4 never parses the host a second time.
	hostIP net.IP
	// rawPath is the path exactly as written.
	rawPath string
}

// parseWireURL applies every rule that governs the written form of a URL on this
// wire, and is deliberately shared by the link URL and payload.endpoint paths.
//
// Sharing it is the rule, not an optimisation: the earlier split — where
// payload.endpoint skipped the dot-segment and backslash checks — left the
// canonical form of those strings undefined, and the two parsers answered
// differently. "https://h/api/../api" stays as written here and folds to
// "https://h/api" under WHATWG, so one wallet paid a code the other refused.
// The only rule that does NOT belong here is the /ipp/v1/link/<token> suffix,
// which describes the scanned URL alone.
func parseWireURL(raw string, allowHTTP bool) (*wireURL, error) {
	// A WHATWG parser strips whitespace and control characters before parsing,
	// so a string that only becomes a valid URL after the stripping would reach
	// the network under a spelling the user was never shown. C0, C1 and DEL are
	// all in: Go and JS disagree about which of them count as whitespace, and
	// the disagreement is per-code-point.
	if i := indexControlOrSpace(raw); i >= 0 {
		return nil, fmt.Errorf("whitespace or control character at byte %d: %w", i, ErrLinkURLForm)
	}

	m := wireURLPattern.FindStringSubmatch(raw)
	if m == nil {
		// A query or fragment is the common reason the wire form fails, and it
		// is a different mistake from "this is not a URL at all".
		if strings.ContainsAny(raw, "?#") {
			return nil, fmt.Errorf("carries a query or fragment: %w", ErrLinkURLForm)
		}
		return nil, fmt.Errorf("%q is not scheme://host/path: %w", raw, ErrLinkShape)
	}
	scheme, rawAuthority, rawPath := m[1], m[2], m[3]

	// The scheme is refused when it is not already lowercase, not normalised into
	// lowercase. RFC 3986 makes the scheme case-insensitive, so "normalise it"
	// has a defensible reading -- which is exactly the problem: every other
	// component here (port, host spelling, path) refuses non-canonical writing
	// rather than repairing it, and one component going the other way is a
	// second spelling of the same link. The host keeps the opposite treatment
	// on purpose (accepted and lowercased), because DNS is case-insensitive and
	// a payload endpoint may legitimately arrive uppercased.
	if lower := strings.ToLower(scheme); lower != scheme {
		return nil, fmt.Errorf("scheme %q is not lowercase: %w", scheme, ErrLinkURLForm)
	}
	if scheme != "https" && scheme != "http" {
		return nil, fmt.Errorf("scheme %q: %w", scheme, ErrLinkShape)
	}
	if scheme == "http" && !allowHTTP {
		return nil, fmt.Errorf("%q: %w", raw, ErrLinkSchemeNotHTTPS)
	}

	// "https://good.com@evil.com/ipp/v1/link/<t>" fetches from evil.com and
	// derives the endpoint "good.com@evil.com" — which the attacker controls on
	// both sides, so the cross-check passes — while the confirmation screen
	// shows a host the user reads as good.com. That is the last defence against
	// an unsigned payload, so userinfo is refused outright, not stripped.
	if strings.Contains(rawAuthority, "@") {
		return nil, fmt.Errorf("carries userinfo: %w", ErrLinkURLForm)
	}
	if rawAuthority == "" {
		return nil, fmt.Errorf("%q has no host: %w", raw, ErrLinkShape)
	}
	// A WHATWG parser reads a backslash in the authority as a path separator, so
	// "https://good.example\evil.example/..." is host good.example there and a
	// parse error here. Refusing it keeps one written link from meaning two
	// different origins.
	if strings.Contains(rawAuthority, `\`) {
		return nil, fmt.Errorf("authority %q contains a backslash: %w", rawAuthority, ErrLinkURLForm)
	}
	// ASCII-only closes three measured divergences at once: an IDN host is
	// punycode on one side and raw UTF-8 on the other (and Go cannot even
	// resolve raw UTF-8), and UTS-46 mapping DELETES U+200B — so a WHATWG
	// parser derives, and requests, a host that is not the one in the URL. The
	// interoperable spelling of an IDN endpoint is punycode.
	if i := indexNonASCII(rawAuthority); i >= 0 {
		return nil, fmt.Errorf("authority %q has a non-ASCII code point at byte %d: %w", rawAuthority, i, ErrLinkURLForm)
	}
	hostText, bracketed, portText, hasPort, err := splitAuthority(rawAuthority)
	if err != nil {
		return nil, err
	}
	if hostText == "" {
		return nil, fmt.Errorf("%q has no host: %w", raw, ErrLinkShape)
	}
	host, hostIP, err := canonicalHost(hostText, bracketed)
	if err != nil {
		return nil, err
	}
	port, err := canonicalPort(portText, hasPort, scheme)
	if err != nil {
		return nil, err
	}

	if seg := findDotSegment(rawPath); seg != "" {
		return nil, fmt.Errorf("path segment %q is a dot segment: %w", seg, ErrLinkURLForm)
	}
	// The written path is checked against a charset, never against what a URL
	// library would rewrite it to, and every value below is taken from the
	// written string. Nothing here consults net/url, so no part of the derived
	// endpoint can depend on this language's URL implementation.
	if seg, reason, ok := findNonPcharSegment(rawPath); !ok {
		return nil, fmt.Errorf("path segment %q %s: %w", seg, reason, ErrLinkURLForm)
	}

	return &wireURL{
		origin:  scheme + "://" + joinAuthority(host, port),
		host:    host,
		hostIP:  hostIP,
		rawPath: rawPath,
	}, nil
}

// splitAuthority separates the host text from the port text. It admits exactly
// one layout: an optional bracketed IPv6 literal or a bare host, followed by at
// most one ":port".
func splitAuthority(rawAuthority string) (hostText string, bracketed bool, portText string, hasPort bool, err error) {
	if strings.HasPrefix(rawAuthority, "[") {
		i := strings.LastIndex(rawAuthority, "]")
		if i < 0 {
			return "", false, "", false, fmt.Errorf("authority %q has an unterminated IPv6 literal: %w", rawAuthority, ErrLinkURLForm)
		}
		hostText, bracketed = rawAuthority[1:i], true
		if rest := rawAuthority[i+1:]; rest != "" {
			if !strings.HasPrefix(rest, ":") {
				return "", false, "", false, fmt.Errorf("authority %q has trailing text after the IPv6 literal: %w", rawAuthority, ErrLinkURLForm)
			}
			portText, hasPort = rest[1:], true
		}
		return hostText, bracketed, portText, hasPort, nil
	}
	if strings.ContainsAny(rawAuthority, "[]") {
		return "", false, "", false, fmt.Errorf("authority %q has a misplaced bracket: %w", rawAuthority, ErrLinkURLForm)
	}
	if i := strings.IndexByte(rawAuthority, ':'); i >= 0 {
		return rawAuthority[:i], false, rawAuthority[i+1:], true, nil
	}
	return rawAuthority, false, "", false, nil
}

// canonicalHost returns the comparable form of a host, refusing every other
// spelling of the same host.
//
// Only refusing non-canonical spellings makes the endpoint cross-check work
// across implementations. A WHATWG host parser rewrites "[2001:db8:0:0:0:0:0:1]"
// to "[2001:db8::1]", "0177.0.0.1" to "127.0.0.1" and "203.0.113" to
// "203.0.0.113"; this side only lowercased. Same printed code, and the endpoint
// derived from it differed -- so one wallet paid and the other refused. Nothing
// legitimate is lost: every one of those hosts has exactly one canonical
// spelling that both sides accept.
// The second result is the parsed address for an IP literal, nil for a
// registered name. It is returned rather than re-derived so rule 4 judges the
// very bytes this function accepted: parsing the host a second time somewhere
// else is another place for two answers to appear.
func canonicalHost(hostText string, bracketed bool) (string, net.IP, error) {
	if bracketed {
		return canonicalIPv6Literal(hostText)
	}
	// A host made only of digits and dots is what a numeric host parser
	// reinterprets, so it must be the one canonical dotted-quad spelling.
	if onlyDigitsAndDots(hostText) {
		octets, ok := parseCanonicalIPv4(hostText)
		if !ok {
			return "", nil, fmt.Errorf("host %q is not a canonical IPv4 literal (four decimal octets 0-255, no leading zero): %w", hostText, ErrLinkURLForm)
		}
		return hostText, net.IPv4(octets[0], octets[1], octets[2], octets[3]), nil
	}
	name, err := canonicalRegName(hostText)
	return name, nil, err
}

// canonicalRegName validates a registered name and returns it lowercased.
//
// A label is an RFC 1123 LDH label in full: letter-digit-hyphen, 1-63
// characters, no leading or trailing hyphen, and never empty (so a trailing dot
// is refused too). Those last three are stricter than a WHATWG host parser,
// which checks none of them. That is deliberate and pinned: they are part of the
// LDH definition anyone reading the wire rules would implement, and a trailing
// dot in particular is a second spelling of one host - which breaks a
// cross-check that compares strings.
//
// The final label MUST NOT look numeric. That is not a DNS rule but a
// cross-implementation one: a WHATWG parser reads "0x7f000001" and "3405803625"
// as IPv4 numbers and rewrites the host, while a charset check alone accepts
// them verbatim. No real top-level label is all digits or "0x"-prefixed.
func canonicalRegName(hostText string) (string, error) {
	labels := strings.Split(hostText, ".")
	for _, label := range labels {
		if label == "" {
			return "", fmt.Errorf("host %q has an empty label: %w", hostText, ErrLinkURLForm)
		}
		if len(label) > maxHostLabelLen {
			return "", fmt.Errorf("host label %q is longer than %d characters: %w", label, maxHostLabelLen, ErrLinkURLForm)
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return "", fmt.Errorf("host label %q starts or ends with a hyphen: %w", label, ErrLinkURLForm)
		}
		if !ldhLabelPattern.MatchString(label) {
			return "", fmt.Errorf("host label %q is not a letter-digit-hyphen label: %w", label, ErrLinkURLForm)
		}
	}
	// WHATWG's ends-in-a-number checker skips a trailing empty label before
	// picking the last one. That step is unreachable here: an empty label is
	// already refused above, so this only ever sees a non-empty final label.
	if last := labels[len(labels)-1]; endsInANumber(last) {
		return "", fmt.Errorf("host %q ends in the numeric label %q, which a WHATWG host parser would reread as an IPv4 address: %w", hostText, last, ErrLinkURLForm)
	}
	return strings.ToLower(hostText), nil
}

// endsInANumber reports whether a final reg-name label is one a WHATWG host
// parser would reread as an IPv4 number, per the URL Standard's
// "ends in a number checker" applied to a single label.
//
// Written as the algorithm rather than as a description of it: a descriptive
// paraphrase has to be re-derived by each implementation, and the octal branch
// is where those derivations part company. The two branches below are the
// checker's own: all-ASCII-digits, or the IPv4 number parser succeeding. The
// octal case needs no branch of its own - the IPv4 number parser only reaches
// radix 8 for a string whose every character is already an ASCII digit, so the
// first branch has claimed it.
func endsInANumber(label string) bool {
	if label == "" {
		return false
	}
	if isAllASCIIDigits(label) {
		return true
	}
	// IPv4 number parser: an "0x"/"0X" prefix selects radix 16, and the rest
	// must be hex digits (an empty remainder parses as 0, so "0x" counts).
	if len(label) >= 2 && label[0] == '0' && (label[1] == 'x' || label[1] == 'X') {
		return isAllHexDigits(label[2:])
	}
	return false
}

func isAllASCIIDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func isAllHexDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

func onlyDigitsAndDots(s string) bool {
	for i := 0; i < len(s); i++ {
		if (s[i] < '0' || s[i] > '9') && s[i] != '.' {
			return false
		}
	}
	return true
}

// parseCanonicalIPv4 accepts only four decimal octets with no leading zero and
// returns them.
func parseCanonicalIPv4(s string) ([4]byte, bool) {
	var out [4]byte
	if !dottedQuadPattern.MatchString(s) {
		return out, false
	}
	for i, octet := range strings.Split(s, ".") {
		n, err := strconv.Atoi(octet)
		if err != nil || n > 255 {
			return out, false
		}
		out[i] = byte(n)
	}
	return out, true
}

// canonicalIPv6Literal accepts only the RFC 5952 spelling of an IPv6 address:
// lowercase hex, no leading zeros in a group, and "::" covering the longest run
// of zero groups (never a single group).
//
// Parsing is done here rather than by net.ParseIP. The comparison against the
// re-serialised form would have framed most of the risk either way, but a
// contract that requires two languages to answer identically cannot leave the
// accepting side of the decision inside one language's standard library: what
// that library accepts is not stated by the wire rules and cannot be checked
// against them.
func canonicalIPv6Literal(inner string) (string, net.IP, error) {
	if inner != strings.ToLower(inner) {
		return "", nil, fmt.Errorf("IPv6 literal %q must be lowercase: %w", inner, ErrLinkURLForm)
	}
	if !strings.Contains(inner, ":") {
		return "", nil, fmt.Errorf("bracketed host %q is not an IPv6 literal: %w", inner, ErrLinkURLForm)
	}
	ip, ok := parseIPv6Literal(inner)
	if !ok {
		return "", nil, fmt.Errorf("%q is not an IPv6 literal: %w", inner, ErrLinkURLForm)
	}
	if canon := canonicalIPv6Text(ip); inner != canon {
		return "", nil, fmt.Errorf("IPv6 literal %q is not RFC 5952 canonical form (%q): %w", inner, canon, ErrLinkURLForm)
	}
	return inner, ip, nil
}

// parseIPv6Literal parses the RFC 4291 textual forms: up to eight groups of one
// to four lowercase hex digits separated by ":", at most one "::" standing for
// one or more all-zero groups, and an optional trailing dotted-quad occupying
// the last two groups. Uppercase is refused here as well as by the caller, so
// the two checks cannot drift apart.
func parseIPv6Literal(s string) (net.IP, bool) {
	head, tail := s, ""
	hasEllipsis := false
	if i := strings.Index(s, "::"); i >= 0 {
		if strings.Contains(s[i+2:], "::") {
			return nil, false
		}
		head, tail, hasEllipsis = s[:i], s[i+2:], true
	}
	// The dotted-quad tail may only end the whole literal.
	headGroups, ok := parseIPv6Groups(head, !hasEllipsis)
	if !ok {
		return nil, false
	}
	tailGroups, ok := parseIPv6Groups(tail, hasEllipsis)
	if !ok {
		return nil, false
	}
	total := len(headGroups) + len(tailGroups)
	if hasEllipsis {
		// "::" stands for at least one group, so a full eight leaves it nothing.
		if total > 7 {
			return nil, false
		}
	} else if total != 8 {
		return nil, false
	}
	var groups [8]uint16
	copy(groups[:], headGroups)
	copy(groups[8-len(tailGroups):], tailGroups)
	ip := make(net.IP, net.IPv6len)
	for i, g := range groups {
		ip[2*i], ip[2*i+1] = byte(g>>8), byte(g)
	}
	return ip, true
}

// parseIPv6Groups parses a colon-separated run of hex groups, optionally ending
// in a dotted quad when allowDottedQuad is set. An empty string is zero groups.
func parseIPv6Groups(s string, allowDottedQuad bool) ([]uint16, bool) {
	if s == "" {
		return nil, true
	}
	parts := strings.Split(s, ":")
	out := make([]uint16, 0, 8)
	for i, part := range parts {
		if strings.Contains(part, ".") {
			if !allowDottedQuad || i != len(parts)-1 {
				return nil, false
			}
			octets, ok := parseCanonicalIPv4(part)
			if !ok {
				return nil, false
			}
			out = append(out,
				uint16(octets[0])<<8|uint16(octets[1]),
				uint16(octets[2])<<8|uint16(octets[3]))
			continue
		}
		if len(part) < 1 || len(part) > 4 {
			return nil, false
		}
		var v uint16
		for j := 0; j < len(part); j++ {
			c := part[j]
			var d byte
			switch {
			case c >= '0' && c <= '9':
				d = c - '0'
			case c >= 'a' && c <= 'f':
				d = c - 'a' + 10
			default:
				return nil, false
			}
			v = v<<4 | uint16(d)
		}
		out = append(out, v)
	}
	if len(out) > 8 {
		return nil, false
	}
	return out, true
}

// canonicalIPv6Text renders 16 bytes in RFC 5952 form. It is spelled out rather
// than delegated to net.IP.String because that method renders an IPv4-mapped
// address as a bare dotted quad, which is not a valid IPv6 literal.
func canonicalIPv6Text(ip net.IP) string {
	// RFC 5952 section 5: an IPv4-mapped address keeps the mixed notation.
	if isZeroBytes(ip[:10]) && ip[10] == 0xff && ip[11] == 0xff {
		return fmt.Sprintf("::ffff:%d.%d.%d.%d", ip[12], ip[13], ip[14], ip[15])
	}
	var groups [8]uint16
	for i := range groups {
		groups[i] = uint16(ip[2*i])<<8 | uint16(ip[2*i+1])
	}
	// RFC 5952 section 4.2, all three clauses - each one is a spelling the other
	// clauses would leave ambiguous, so dropping any of them lets one address be
	// written two ways:
	//   4.2.1 a single zero group is never compressed
	//   4.2.2 only a run of two or more is compressed
	//   4.2.3 on a tie the LEFTMOST run wins ("1::1:0:0:1:1", not
	//         "1:0:0:1::1:1" - same address)
	// The strict > below is what implements 4.2.3: a later run of equal length
	// never displaces an earlier one.
	bestStart, bestLen := -1, 0
	for i := 0; i < 8; {
		if groups[i] != 0 {
			i++
			continue
		}
		j := i
		for j < 8 && groups[j] == 0 {
			j++
		}
		if j-i >= 2 && j-i > bestLen {
			bestStart, bestLen = i, j-i
		}
		i = j
	}
	out := ""
	for i := 0; i < 8; {
		if i == bestStart {
			out += "::"
			i += bestLen
			continue
		}
		if out != "" && !strings.HasSuffix(out, ":") {
			out += ":"
		}
		out += strconv.FormatUint(uint64(groups[i]), 16)
		i++
	}
	return out
}

// findNonPcharSegment reports the first path segment that is not RFC 3986 pchar,
// and separately the first whose percent-encoding is not uppercase.
//
// The uppercase requirement is not cosmetic. Rule 9 compares endpoints for
// string equality, while RFC 3986 section 6.2.2.1 says a comparison should first
// normalise percent-encodings to uppercase - a defensible reading. So with
// "…/a%2Fb" on the link and "…/a%2fb" in the payload, strict equality refuses
// payment and the normalising reading allows it. Both spellings match a
// case-insensitive pchar test, so the shape check cannot see the difference:
// the casing has to be pinned separately. Refused rather than normalised, in
// line with scheme, port and host here; Go's url.QueryEscape and JS's
// encodeURIComponent both emit uppercase already.
func findNonPcharSegment(rawPath string) (segment string, reason string, ok bool) {
	for _, seg := range strings.Split(rawPath, "/") {
		if !pcharSegmentAnyCasePattern.MatchString(seg) {
			return seg, "is not RFC 3986 pchar (unreserved / pct-encoded / sub-delims / \":\" / \"@\")", false
		}
		if !pcharSegmentPattern.MatchString(seg) {
			return seg, "has a percent-encoding that is not uppercase hex", false
		}
	}
	return "", "", true
}

// parseLinkURL returns the derived endpoint (scheme://authority{prefix}), the
// hostname for the address check, and the URL to request.
//
// The requested URL is the string AS SCANNED, not a re-rendering of it. Only the
// derived endpoint is normalised, because only it is compared. Rebuilding the
// request URL from the normal form put a different string on the wire than the
// user was shown -- "https://Pay.Example.COM/..." went out lowercased and an
// explicit ":443" was dropped -- so the receiver's access log did not contain
// the URL that was scanned, and the two SDKs logged different request lines for
// one QR code. Nothing here needs a rebuilt URL: every rule already passed on
// the written string.
func parseLinkURL(raw string, unsafe bool) (endpoint string, target *wireURL, requestURL string, err error) {
	w, err := parseWireURL(raw, unsafe)
	if err != nil {
		return "", nil, "", fmt.Errorf("ipp: link: %w", err)
	}
	pm := linkPathPattern.FindStringSubmatch(w.rawPath)
	if pm == nil {
		return "", nil, "", fmt.Errorf("ipp: link: path %q: %w", w.rawPath, ErrLinkShape)
	}
	prefix := pm[1]
	if prefix != "" && !pathPrefixPattern.MatchString(prefix) {
		return "", nil, "", fmt.Errorf("ipp: link: malformed endpoint path prefix %q: %w", prefix, ErrLinkShape)
	}
	return w.origin + prefix, w, raw, nil
}

// normalizeEndpoint reduces a payload endpoint to the same normal form
// parseLinkURL derives, so rule 9 compares like with like.
func normalizeEndpoint(raw string, unsafe bool) (string, error) {
	w, err := parseWireURL(raw, unsafe)
	if err != nil {
		// ErrLinkSchemeNotHTTPS describes the SCANNED url, and callers branch on
		// it to tell the user to reach for the unsafe entry point. A payload
		// endpoint saying http is a different fault - the scanned link was
		// https - so pointing the user at the unsafe entry point would send them
		// to relax a rule that was never the problem. Translated here rather
		// than by a flag inside parseWireURL: the shared parser has to stay
		// side-agnostic, or the two sides can drift apart again.
		if errors.Is(err, ErrLinkSchemeNotHTTPS) {
			return "", fmt.Errorf("scheme is http while the scanned link is https: %w", ErrLinkURLForm)
		}
		return "", err
	}
	// The ORDER here is itself a wire rule: strip trailing slashes first, then
	// check for empty segments. Reversed, "https://h//" would be refused here
	// while another implementation collapses it to "https://h" and pays — the
	// one-code-two-verdicts split these rules exist to prevent.
	path := strings.TrimRight(w.rawPath, "/")
	if path != "" && !pathPrefixPattern.MatchString(path) {
		return "", fmt.Errorf("malformed path %q: %w", path, ErrLinkURLForm)
	}
	return w.origin + path, nil
}

// canonicalPort validates the port text splitAuthority pulled off the authority
// and returns it, empty when absent or when it is the scheme's default. Only one spelling is admitted: a numeric
// parser and a string comparison disagree about ":0443" (dropped as the default
// on one side, kept verbatim on the other), and about a port above 65535
// (rejected by a WHATWG parser, kept by Go's).
func canonicalPort(portPart string, hasPort bool, scheme string) (string, error) {
	if !hasPort {
		return "", nil
	}
	if !canonicalPortPattern.MatchString(portPart) {
		return "", fmt.Errorf("port %q is not canonical (no leading zero, no empty port): %w", portPart, ErrLinkURLForm)
	}
	n, convErr := strconv.Atoi(portPart)
	if convErr != nil || n > maxPort {
		return "", fmt.Errorf("port %q is out of range: %w", portPart, ErrLinkURLForm)
	}
	if (scheme == "https" && n == 443) || (scheme == "http" && n == 80) {
		return "", nil
	}
	return portPart, nil
}

// joinAuthority renders host[:port], bracketing an IPv6 literal.
func joinAuthority(host, port string) string {
	if port != "" {
		return net.JoinHostPort(host, port)
	}
	if strings.Contains(host, ":") {
		return "[" + host + "]"
	}
	return host
}

// indexControlOrSpace returns the byte offset of the first code point a link URL
// may not contain, or -1.
func indexControlOrSpace(s string) int {
	for i, r := range s {
		if isRejectedURLCodePoint(r) {
			return i
		}
	}
	return -1
}

// isRejectedURLCodePoint enumerates the set member by member.
//
// It MUST NOT be written as unicode.IsSpace, \s, or a Unicode category test.
// Those are environment-defined sets: the tail of Zs moves with the engine's
// Unicode version (U+180E was Zs through Unicode 6.2 and is not now), so two
// implementations - or the same one on two toolchain versions - can disagree
// about a URL while both "implement the rule". The set below is the rule.
//
// It is also not written as a regexp. Both shapes that have actually gone wrong
// here are regexp accidents: "[\s -]" puts a literal hyphen in the class and so
// rejects every hyphenated host and roughly half of all base64url tokens, while
// "[ -,]" is a range over punctuation that swallows "+", ",", "%" and "*".
// Neither shows up on a happy path built from letters.
func isRejectedURLCodePoint(r rune) bool {
	switch {
	case r <= 0x1f: // C0 controls, which covers TAB / LF / VT / FF / CR
		return true
	case r == 0x7f: // DEL
		return true
	case r >= 0x80 && r <= 0x9f: // C1 controls, which covers NEL (U+0085)
		return true
	}
	switch r {
	case 0x0020, // SPACE
		0x00a0, // NO-BREAK SPACE
		0x1680, // OGHAM SPACE MARK
		0x2000, // EN QUAD
		0x2001, // EM QUAD
		0x2002, // EN SPACE
		0x2003, // EM SPACE
		0x2004, // THREE-PER-EM SPACE
		0x2005, // FOUR-PER-EM SPACE
		0x2006, // SIX-PER-EM SPACE
		0x2007, // FIGURE SPACE
		0x2008, // PUNCTUATION SPACE
		0x2009, // THIN SPACE
		0x200a, // HAIR SPACE
		0x2028, // LINE SEPARATOR
		0x2029, // PARAGRAPH SEPARATOR
		0x202f, // NARROW NO-BREAK SPACE
		0x205f, // MEDIUM MATHEMATICAL SPACE
		0x3000, // IDEOGRAPHIC SPACE
		0xfeff: // ZERO WIDTH NO-BREAK SPACE (BOM)
		return true
	}
	return false
}

// indexNonASCII returns the byte offset of the first rune >= U+0080, or -1.
func indexNonASCII(s string) int {
	for i, r := range s {
		if r >= 0x80 {
			return i
		}
	}
	return -1
}

// findDotSegment returns the first "." / ".." path segment in any spelling, or "".
func findDotSegment(rawPath string) string {
	for _, seg := range strings.Split(rawPath, "/") {
		if dotSegmentPattern.MatchString(seg) {
			return seg
		}
	}
	return ""
}

// ── Rule 8: response body ──────────────────────────────────────────

// decodeLinkBody parses the link response into an IPPRequest.
//
// The body contract is a BARE JSON IPPMessage. The compact "ipp1:" string is
// deliberately not accepted even though Decode understands it: the request
// already declared Accept: application/json, so a second accepted encoding buys
// nothing and gives two implementations something to disagree about.
//
// IMPORTANT: every judgement here is made on BYTES, and the body must not be
// handed to a text decoder first. Two earlier spellings of this rule both
// diverged across implementations:
//
//   - "trim, then the first character must be '{'" leaves the whitespace set to
//     whatever the local standard library trims. Go strips U+0085 and not
//     U+FEFF; JS strips U+FEFF and not U+0085. A body with either prefix was
//     paid by one wallet and refused by the other.
//   - Respelling it on code points is not enough either, because stripping a
//     BOM happens in the bytes-to-text step and decoders disagree there:
//     TextDecoder and fetch().text() drop it, Buffer.toString("utf8") and Go's
//     string(b) keep it. Two faithful implementations of a code-point rule still
//     split.
//
// So: no BOM, and only the four RFC 8259 whitespace bytes may precede 0x7B.
func decodeLinkBody(body []byte) (*IPPRequest, error) {
	// Well-formed UTF-8, judged here and not after a decoder has had a look.
	// A non-fatal decoder rewrites what every later rule then examines: an
	// invalid byte becomes U+FFFD, so two bodies differing only in an illegal
	// byte decode to the same text, and a key that is illegal on the wire can
	// decode into a legal one. Requiring well-formed input makes decoding
	// lossless, which is what lets the rules below sit at any layer safely.
	// Go's string(b) happens not to substitute, but the rule cannot rest on
	// that: "this language does not corrupt it" is not a wire rule.
	if !utf8.Valid(body) {
		return nil, fmt.Errorf("ipp: link: body is not well-formed UTF-8: %w", ErrLinkResponseForm)
	}
	// RFC 8259 section 8.1: a JSON text must not begin with a byte order mark.
	if len(body) >= 3 && body[0] == 0xEF && body[1] == 0xBB && body[2] == 0xBF {
		return nil, fmt.Errorf("ipp: link: body starts with a UTF-8 BOM (EF BB BF), which RFC 8259 forbids: %w", ErrLinkResponseForm)
	}
	i := 0
	for i < len(body) && isJSONWhitespaceByte(body[i]) {
		i++
	}
	if i >= len(body) || body[i] != '{' {
		return nil, fmt.Errorf("ipp: link: body is not a JSON object (the compact \"ipp1:\" form is not a link representation): %w", ErrLinkResponseForm)
	}
	doc := body[i:]

	if err := checkEnvelopeTopLevelKeys(doc, body); err != nil {
		return nil, err
	}
	// Same decoder as the ipp1: carrier from here on, so the two carriers
	// produce one shape by construction. It runs ValidateIPPRequest and rejects
	// a voucher envelope (a voucher carries nullifier_key and must never be
	// handed out by an anonymous public GET).
	request, err := Decode(string(doc))
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// isJSONWhitespaceByte reports whether b is one of the four bytes RFC 8259
// section 2 allows as whitespace. Deliberately not unicode.IsSpace or
// strings.TrimSpace: the whole point is that the set is fixed by the wire rule,
// not by this language's idea of whitespace.
func isJSONWhitespaceByte(b byte) bool {
	return b == 0x20 || b == 0x09 || b == 0x0A || b == 0x0D
}

// checkEnvelopeTopLevelKeys walks the top-level object as a token stream and
// enforces both that the key set is the closed envelope set and that no key
// repeats.
//
// The duplicate check has to happen here rather than after unmarshalling into a
// map: a map has already folded repeats, so {"v":1,...,"v":2} would show a legal
// five-key set. Which of the two values survives folding is a parser decision
// (last-wins, first-wins, or an error depending on the library), so a document
// that repeats a key has no single meaning and is refused outright.
func checkEnvelopeTopLevelKeys(doc, raw []byte) error {
	malformed := func(err error) error {
		return &ConnectionError{Message: "link response body was not valid JSON", Raw: raw, Err: err}
	}
	dec := json.NewDecoder(bytes.NewReader(doc))
	tok, err := dec.Token()
	if err != nil {
		return malformed(err)
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("ipp: link: body is not a JSON object: %w", ErrLinkResponseForm)
	}
	seen := make(map[string]struct{}, len(envelopeTopLevelKeys))
	for dec.More() {
		// The RAW spelling of the key, not the decoded one. A decoded key hides
		// the escape: "\u0076" decodes to "v", so a set check on the decoded form
		// accepts it and a duplicate check on the decoded form folds it together
		// with a literal "v" - while a check on raw bytes without this rule would
		// call the same key unknown. Both readings were implementable from the
		// earlier wording and they disagreed on the same body. Refusing escapes
		// outright removes the unescaping step, and no real server spells "v" as
		// "\u0076".
		before := dec.InputOffset()
		keyTok, err := dec.Token()
		if err != nil {
			return malformed(err)
		}
		if _, ok := keyTok.(string); !ok {
			return fmt.Errorf("ipp: link: body has a non-string top-level key: %w", ErrLinkResponseForm)
		}
		raw := trimJSONKeyPrefix(doc[before:dec.InputOffset()])
		if len(raw) < 2 || raw[0] != '"' || raw[len(raw)-1] != '"' {
			return fmt.Errorf("ipp: link: top-level key is not a plain quoted string: %w", ErrLinkResponseForm)
		}
		key := string(raw[1 : len(raw)-1])
		if strings.Contains(key, `\`) {
			return fmt.Errorf("ipp: link: top-level key %q uses a backslash escape; the envelope keys are written literally: %w", key, ErrLinkResponseForm)
		}
		if _, dup := seen[key]; dup {
			return fmt.Errorf("ipp: link: top-level key %q appears more than once: %w", key, ErrLinkResponseForm)
		}
		seen[key] = struct{}{}
		if _, known := envelopeTopLevelKeys[key]; !known {
			return fmt.Errorf("ipp: link: unexpected top-level key %q (the envelope carries only v / network / type / meta / payload): %w", key, ErrLinkResponseForm)
		}
		var skip json.RawMessage
		if err := dec.Decode(&skip); err != nil {
			return malformed(err)
		}
	}
	if _, err := dec.Token(); err != nil {
		return malformed(err)
	}
	// Anything after the envelope but the four JSON whitespace bytes. Without
	// this the leftovers reach the shared decoder, which reports them with a
	// bare error that unwraps to no sentinel at all - so a caller cannot even
	// tell whether it was refused for a safety reason.
	for _, b := range doc[dec.InputOffset():] {
		if !isJSONWhitespaceByte(b) {
			return fmt.Errorf("ipp: link: body carries trailing content after the envelope: %w", ErrLinkResponseForm)
		}
	}
	return nil
}

// trimJSONKeyPrefix strips the separator run a token stream leaves in front of a
// key: the comma from the previous pair and any JSON whitespace.
func trimJSONKeyPrefix(b []byte) []byte {
	for len(b) > 0 && (b[0] == ',' || isJSONWhitespaceByte(b[0])) {
		b = b[1:]
	}
	return b
}

// ── Rule 4: non-public targets, judged on resolved addresses ────────

// defaultLookupIP is the production resolver.
//
// Obfuscated literal spellings ("0x7f000001", "2130706433", "127.1") are not IPs
// to net.ParseIP, so they arrive here and the system resolver unwraps them —
// every address it returns is then checked, so they are refused. Note the error
// class differs from a WHATWG runtime, which canonicalises them before any
// lookup and always reports a safety verdict: under Go's pure-Go resolver the
// same input can surface as *ConnectionError instead. Refused either way.
func defaultLookupIP(ctx context.Context, host string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(ctx, "ip", host)
}

// assertPublicTarget implements rule 4's pre-flight half: reject if the literal
// host is a non-public address, otherwise reject if ANY address it resolves to
// is. Checking the literal host alone would let evil.example.com with an A
// record of 127.0.0.1 turn the wallet into the attacker's intranet probe.
//
// Go has name resolution, so this never degrades to a weaker check: a resolver
// failure is reported as a transport error and the resolve stops. There is no
// branch that proceeds without an answer.
func assertPublicTarget(ctx context.Context, lookupIP func(context.Context, string) ([]net.IP, error), target *wireURL) error {
	host := target.host
	// An IP literal was already parsed while the written form was validated, so
	// it is judged here without being parsed again.
	if target.hostIP != nil {
		if isNonPublicIP(target.hostIP) {
			return fmt.Errorf("ipp: link: host %s: %w", host, ErrLinkPrivateHost)
		}
		return nil
	}
	ips, err := lookupIP(ctx, host)
	if err != nil {
		// Being offline is a transport failure, not a verdict about this host.
		// Reporting it as ErrLinkPrivateHost would have the wallet tell the user
		// "refused to resolve", which reads as "this code is dangerous".
		return &ConnectionError{Message: fmt.Sprintf("resolving link host %q failed", host), Err: err}
	}
	if len(ips) == 0 {
		return &ConnectionError{Message: fmt.Sprintf("link host %q resolved to no addresses", host)}
	}
	for _, ip := range ips {
		// Every answer must be public, not merely the first: a name returning
		// one public and one internal address would otherwise be a coin flip
		// decided by whichever record the connect path happens to pick.
		if isNonPublicIP(ip) {
			return fmt.Errorf("ipp: link: host %q resolves to %s: %w", host, ip, ErrLinkPrivateHost)
		}
	}
	return nil
}

// parseIPAddress parses a dial address, dropping an IPv6 zone id (fe80::1%en0)
// — the zone is routing information, not part of the address.
//
// This is the one place net.ParseIP is still used, and it must stay the only
// one: the input here is an address the operating system chose at connect time,
// not anything written on the wire, so no cross-implementation agreement is at
// stake. Every judgement about a written host goes through parseIPv6Literal /
// parseCanonicalIPv4 instead.
func parseIPAddress(host string) net.IP {
	if i := strings.IndexByte(host, '%'); i >= 0 {
		host = host[:i]
	}
	return net.ParseIP(host)
}

// isNonPublicIP is the shared rejection set. It is deliberately a closed list
// rather than "everything net.IP has a predicate for": each widening is another
// way for two implementations to disagree about one address. Documentation-only
// ranges (TEST-NET) are left out on purpose — nothing routes there, so they are
// not an SSRF target.
func isNonPublicIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil {
		return isNonPublicIPv4(ip4)
	}
	ip16 := ip.To16()
	if ip16 == nil {
		return true
	}
	return isNonPublicIPv6(ip16)
}

func isNonPublicIPv4(ip net.IP) bool {
	a, b := ip[0], ip[1]
	switch {
	case a == 0: // 0.0.0.0/8 "this network"
		return true
	case a == 10: // RFC 1918
		return true
	case a == 127: // loopback
		return true
	case a == 100 && b >= 64 && b <= 127: // RFC 6598 CGNAT
		return true
	case a == 169 && b == 254: // link-local, incl. cloud metadata
		return true
	case a == 172 && b >= 16 && b <= 31: // RFC 1918
		return true
	// Strictly /24. Widening this to /16 also swallows 192.0.2.0/24 (TEST-NET-1),
	// which must stay public — the enclosing block's TEST-NET note is the reason.
	case a == 192 && b == 0 && ip[2] == 0: // 192.0.0.0/24 IETF protocol assignments
		return true
	case a == 192 && b == 168: // RFC 1918
		return true
	case a == 198 && (b == 18 || b == 19): // RFC 2544 benchmarking
		return true
	case a >= 224: // multicast, reserved, broadcast
		return true
	}
	return false
}

func isNonPublicIPv6(ip net.IP) bool {
	if ip.IsUnspecified() || ip.IsLoopback() {
		return true
	}
	// An IPv4 address wearing an IPv6 costume: judge the address inside, never
	// the wrapper. All three wrappers reach the same IPv4 host, and this also
	// judges AAAA records — a hostile NAT64 answer is a real way to hand a
	// wallet an internal target that looks like a public IPv6 address.
	if inner := embeddedIPv4(ip); inner != nil {
		return isNonPublicIPv4(inner)
	}
	switch {
	case ip[0]&0xfe == 0xfc: // fc00::/7 unique local
		return true
	case ip[0] == 0xfe && ip[1]&0xc0 == 0x80: // fe80::/10 link-local
		return true
	case ip[0] == 0xff: // ff00::/8 multicast
		return true
	}
	return false
}

// embeddedIPv4 returns the IPv4 address carried inside ::ffff:0:0/96 (mapped),
// ::/96 (deprecated compat, e.g. ::127.0.0.1) or 64:ff9b::/96 (NAT64), or nil.
// net.IP.To4 only recognises the mapped form, so the other two reach here.
func embeddedIPv4(ip net.IP) net.IP {
	if len(ip) != net.IPv6len {
		return nil
	}
	if isZeroBytes(ip[:10]) && ((ip[10] == 0 && ip[11] == 0) || (ip[10] == 0xff && ip[11] == 0xff)) {
		return ip[12:16]
	}
	if ip[0] == 0x00 && ip[1] == 0x64 && ip[2] == 0xff && ip[3] == 0x9b && isZeroBytes(ip[4:12]) {
		return ip[12:16]
	}
	return nil
}

func isZeroBytes(b []byte) bool {
	for _, x := range b {
		if x != 0 {
			return false
		}
	}
	return true
}

// denyNonPublicDial is rule 4's connect-time half. The pre-flight lookup and
// the dialer's own lookup are separate resolutions; only this one sees the
// address actually connected to, which is why this side does not need the
// after-the-response re-check a fetch-based implementation has to approximate.
func denyNonPublicDial(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("ipp: link: unparseable dial address %q: %w", address, ErrLinkPrivateHost)
	}
	ip := parseIPAddress(host)
	if ip == nil {
		return fmt.Errorf("ipp: link: dial address %q is not an IP: %w", address, ErrLinkPrivateHost)
	}
	if isNonPublicIP(ip) {
		return fmt.Errorf("ipp: link: refusing to connect to %s: %w", ip, ErrLinkPrivateHost)
	}
	return nil
}

func guardedDialContext() func(context.Context, string, string) (net.Conn, error) {
	return (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Control:   denyNonPublicDial,
	}).DialContext
}

var (
	guardedTransportOnce sync.Once
	guardedTransportInst *http.Transport
)

func guardedTransport() *http.Transport {
	guardedTransportOnce.Do(func() {
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.DialContext = guardedDialContext()
		guardedTransportInst = tr
	})
	return guardedTransportInst
}

// hardenedClient returns a copy of base with redirect following and cookies
// disabled, and — unless relaxed — the dial guard installed. It copies rather
// than mutates so a caller-owned client keeps its own behaviour elsewhere.
func hardenedClient(base *http.Client, unsafe bool) *http.Client {
	var c http.Client
	if base != nil {
		c = *base
	}
	c.Jar = nil
	c.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

	if unsafe {
		return &c
	}
	switch tr := c.Transport.(type) {
	case nil:
		c.Transport = guardedTransport()
	case *http.Transport:
		clone := tr.Clone()
		clone.DialContext = guardedDialContext()
		c.Transport = clone
	}
	return &c
}
