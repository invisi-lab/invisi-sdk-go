package ipp

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/netip"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"
)

func text(v any) bool {
	s, ok := v.(string)
	if !ok || s == "" || !utf8.ValidString(s) || len(utf16.Encode([]rune(s))) > 8192 {
		return false
	}
	for _, r := range s {
		if r < 32 || r == 127 {
			return false
		}
	}
	return true
}
func integer(v any, bits int, positive bool) bool {
	s, ok := v.(string)
	if !ok || len(s) > 80 || !canonicalNumber.MatchString(s) {
		return false
	}
	n, ok := new(big.Int).SetString(s, 10)
	return ok && n.BitLen() <= bits && (!positive || n.Sign() > 0)
}
func number(v any, max uint64) bool {
	n, ok := v.(json.Number)
	if !ok || !canonicalNumber.MatchString(string(n)) {
		return false
	}
	x, ok := new(big.Int).SetString(string(n), 10)
	return ok && x.IsUint64() && x.Uint64() <= max
}
func timestamp(v any) bool {
	s, ok := v.(string)
	if !ok || len(s) != 20 {
		return false
	}
	t, e := time.Parse("2006-01-02T15:04:05Z", s)
	return e == nil && t.Format("2006-01-02T15:04:05Z") == s
}
func before(a, b any) bool { return a.(string) < b.(string) }
func hex(v any, n int, nonzero bool) bool {
	s, ok := v.(string)
	if !ok || len(s) != 2+n*2 || !strings.HasPrefix(s, "0x") {
		return false
	}
	for _, c := range s[2:] {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return !nonzero || strings.Trim(s[2:], "0") != ""
}
func field(v any) bool {
	if !hex(v, 32, true) {
		return false
	}
	n, _ := new(big.Int).SetString(v.(string)[2:], 16)
	p, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
	return n.Cmp(p) < 0
}
func network(v any) bool {
	s, ok := v.(string)
	return ok && strings.HasPrefix(s, "eip155:") && integer(strings.TrimPrefix(s, "eip155:"), 64, true)
}

var secretKey = regexp.MustCompile(`(?i)token|secret|key|opening|credential|authorization`)

func ValidateAPIURL(s, origin string) error {
	if !text(s) {
		return invalid()
	}
	for _, c := range s {
		if c < 33 || c > 126 {
			return invalid()
		}
	}
	u, e := url.Parse(s)
	if e != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Fragment != "" || strings.Contains(s, "#") || u.Opaque != "" || u.RawPath != "" && u.EscapedPath() != u.RawPath {
		return invalid()
	}
	if strings.Contains(s, "\\") || u.Host != strings.ToLower(u.Host) || (strings.HasSuffix(u.Host, ":443") || strings.HasSuffix(u.Host, ":")) || u.Path == "" || u.String() != s {
		return invalid()
	}
	if decoded, err := url.PathUnescape(s); err != nil || !utf8.ValidString(decoded) {
		return invalid()
	}
	for _, part := range strings.Split(u.Path, "/") {
		if part == "." || part == ".." {
			return invalid()
		}
	}
	if u.Port() != "" {
		p, ok := new(big.Int).SetString(u.Port(), 10)
		if !ok || p.Sign() < 0 || p.Cmp(big.NewInt(65535)) > 0 || p.String() != u.Port() {
			return invalid()
		}
	}
	if origin != "" && u.Scheme+"://"+u.Host != origin {
		return invalid()
	}

	host := u.Hostname()
	if strings.Contains(host, ":") {
		ip, err := netip.ParseAddr(host)
		if err != nil || ip.Zone() != "" {
			return invalid()
		}
		canonical := ip.String()
		if ip.Is4In6() {
			raw := ip.As16()
			canonical = fmt.Sprintf("::ffff:%x:%x", uint16(raw[12])<<8|uint16(raw[13]), uint16(raw[14])<<8|uint16(raw[15]))
		}
		if canonical != host {
			return invalid()
		}
	} else if regexp.MustCompile(`(?i)(^|\.)([0-9]+|0x[0-9a-f]+)\.?$`).MatchString(host) {
		ip, err := netip.ParseAddr(host)
		if err != nil || !ip.Is4() || ip.String() != host {
			return invalid()
		}
	}
	if strings.Contains(u.RawQuery, "'") {
		return invalid()
	}
	for _, part := range strings.Split(u.RawQuery, "&") {
		key := strings.SplitN(part, "=", 2)[0]
		decoded, err := url.QueryUnescape(key)
		if err != nil || secretKey.MatchString(decoded) {
			return invalid()
		}
	}
	return nil
}
func urlOrigin(s string) string { u, _ := url.Parse(s); return u.Scheme + "://" + u.Host }
func noSecret(s, secret string) bool {
	if ValidateAPIURL(s, "") != nil {
		return false
	}
	u, _ := url.Parse(s)
	route := u.EscapedPath()
	if u.RawQuery != "" {
		route += "?" + u.RawQuery
	}
	decoded, e := url.PathUnescape(route)
	return e == nil && !strings.Contains(route, secret) && !strings.Contains(decoded, secret)
}

var bearerPattern = regexp.MustCompile(`^[A-Za-z0-9\-._~+/]+=*$`)

func bearer(s string) bool { return text(s) && bearerPattern.MatchString(s) }
func idempotency(s string) bool {
	if len(s) < 1 || len(s) > 128 {
		return false
	}
	for _, c := range s {
		if c < 33 || c > 126 {
			return false
		}
	}
	return true
}
func link(v any, method, origin string) bool {
	m, e := object(v, "href method")
	if e != nil || m["method"] != method {
		return false
	}
	s, ok := m["href"].(string)
	return ok && ValidateAPIURL(s, origin) == nil
}
func status(m map[string]any, open bool) bool {
	s, ok := m["status"].(string)
	if !ok {
		return false
	}
	if !(s == "requires_action" || s == "processing" || s == "succeeded" || s == "failed" || open && s == "open") {
		return false
	}
	if s == "failed" {
		return m["failureReason"] == "expired" || m["failureReason"] == "closed"
	}
	return m["failureReason"] == nil
}
func mode(v any) bool { return v == "payment" || v == "credit" }
func header(m map[string]any, t string) bool {
	return m["ippVersion"] == json.Number("1") && m["type"] == t
}
