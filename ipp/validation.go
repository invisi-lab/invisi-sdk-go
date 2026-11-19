package ipp

import (
	"encoding/json"
	"math/big"
	"net/url"
	"reflect"
	"strings"

	"github.com/invisi-labs/invisi-sdk-go/primitives"
)

func validCredit(v any, extra string) bool {
	c, e := object(v, "spender availableCredit unlimited singleLimit dailyLimit monthlyLimit expiresAt "+extra, "initialPayment")
	if e != nil {
		return false
	}
	s, ok := c["spender"].(string)
	if !ok || !strings.HasPrefix(s, "0i") {
		return false
	}
	if _, e = primitives.DecodeAddress(s); e != nil {
		return false
	}
	unlimited, ok := c["unlimited"].(bool)
	if !ok || !integer(c["availableCredit"], 128, false) || unlimited != (c["availableCredit"] == "0") {
		return false
	}
	if initial, ok := c["initialPayment"]; ok {
		terms, e := object(initial, "amount")
		if e != nil || !integer(terms["amount"], 126, true) {
			return false
		}
	}
	return integer(c["singleLimit"], 126, false) && integer(c["dailyLimit"], 126, false) && integer(c["monthlyLimit"], 126, false) && timestamp(c["expiresAt"])
}
func validOptions(v any, m any) bool {
	a, ok := v.([]any)
	if !ok || len(a) < 1 || len(a) > 100 {
		return false
	}
	ids := map[string]bool{}
	for _, v := range a {
		extra := "amount"
		if m == "credit" {
			extra = "credit"
		}
		o, e := object(v, "id method network asset "+extra)
		if e != nil || !text(o["id"]) || o["method"] != "invisi" || !network(o["network"]) {
			return false
		}
		id := o["id"].(string)
		if ids[id] {
			return false
		}
		ids[id] = true
		a, e := object(o["asset"], "address symbol decimals")
		if e != nil || !hex(a["address"], 20, true) || !text(a["symbol"]) || !number(a["decimals"], 255) {
			return false
		}
		if m == "payment" {
			if !integer(o["amount"], 128, true) {
				return false
			}
		} else if !validCredit(o["credit"], "") {
			return false
		}
	}
	return true
}
func validReceipt(v any, m any) bool {
	extra := ""
	if m == "credit" {
		extra = "creditId creditStatus owner revocationShard revocationSlot verificationBlock expiresAt"
	}
	optional := ""
	if m == "credit" {
		optional = "initialPayment"
	}
	r, e := object(v, "network protocolAddress transactionHash commitment blockNumber blockHash completedAt "+extra, optional)
	if e != nil || !network(r["network"]) || !hex(r["protocolAddress"], 20, true) || !hex(r["transactionHash"], 32, true) || !field(r["commitment"]) || !integer(r["blockNumber"], 64, false) || !hex(r["blockHash"], 32, true) || !timestamp(r["completedAt"]) {
		return false
	}
	if m == "credit" {
		if !text(r["creditId"]) || (r["creditStatus"] != "active" && !(r["creditStatus"] == "depleted" && r["initialPayment"] != nil)) || !hex(r["owner"], 20, true) || !integer(r["revocationShard"], 64, false) || !number(r["revocationSlot"], 65535) || !timestamp(r["expiresAt"]) {
			return false
		}
		if initial, ok := r["initialPayment"]; ok {
			draw, e := object(initial, "noteCommitment grossAmount netAmount feeAmount feeRecipient")
			if e != nil || !field(draw["noteCommitment"]) || draw["noteCommitment"] == r["commitment"] || !integer(draw["grossAmount"], 126, true) || !integer(draw["netAmount"], 126, true) || !integer(draw["feeAmount"], 126, false) || !hex(draw["feeRecipient"], 20, false) {
				return false
			}
			gross, _ := new(big.Int).SetString(draw["grossAmount"].(string), 10)
			net, _ := new(big.Int).SetString(draw["netAmount"].(string), 10)
			fee, _ := new(big.Int).SetString(draw["feeAmount"].(string), 10)
			if gross.Cmp(new(big.Int).Add(net, fee)) != 0 || (fee.Sign() == 0) != (strings.Trim(draw["feeRecipient"].(string)[2:], "0") == "") {
				return false
			}
		}
		b, e := object(r["verificationBlock"], "number hash timestamp")
		if e != nil || !integer(b["number"], 64, false) || !hex(b["hash"], 32, true) || !timestamp(b["timestamp"]) {
			return false
		}
	}
	return true
}
func ParsePaymentSession(data []byte, origin string) (*PaymentSession, error) {
	v, e := DecodeJSON(data)
	if e != nil {
		return nil, e
	}
	s, e := object(v, "ippVersion type mode sessionId status merchant options activeRequestId createdAt expiresAt links failureReason receipt evidenceStatus exceptionSummary", "description")
	if e != nil || !header(s, "payment_session") || !mode(s["mode"]) || !text(s["sessionId"]) || !status(s, true) || !validOptions(s["options"], s["mode"]) {
		return nil, invalid()
	}
	merchant, e := object(s["merchant"], "name")
	if e != nil || !text(merchant["name"]) {
		return nil, invalid()
	}
	if d, ok := s["description"]; ok && !text(d) {
		return nil, invalid()
	}
	if !timestamp(s["createdAt"]) || !timestamp(s["expiresAt"]) || !before(s["createdAt"], s["expiresAt"]) {
		return nil, invalid()
	}
	links, e := object(s["links"], "confirm request")
	if e != nil || !link(links["confirm"], "POST", origin) {
		return nil, invalid()
	}
	if s["activeRequestId"] == nil {
		if links["request"] != nil || (s["status"] != "open" && s["status"] != "failed") {
			return nil, invalid()
		}
	} else if !text(s["activeRequestId"]) || s["status"] == "open" || !link(links["request"], "GET", origin) {
		return nil, invalid()
	}
	if s["status"] == "succeeded" {
		if !validReceipt(s["receipt"], s["mode"]) || (s["evidenceStatus"] != "confirmed" && s["evidenceStatus"] != "reconciling" && s["evidenceStatus"] != "disputed") {
			return nil, invalid()
		}
	} else if s["receipt"] != nil || s["evidenceStatus"] != nil {
		return nil, invalid()
	}
	ex, e := object(s["exceptionSummary"], "count unresolvedCount")
	if e != nil || !number(ex["count"], 9007199254740991) || !number(ex["unresolvedCount"], 9007199254740991) {
		return nil, invalid()
	}
	var counts ExceptionSummary
	b, _ := json.Marshal(ex)
	_ = json.Unmarshal(b, &counts)
	if counts.UnresolvedCount > counts.Count {
		return nil, invalid()
	}
	return decodeTyped[PaymentSession](s)
}
func ParsePaymentRequest(data []byte, origin string) (*Request, error) {
	v, e := DecodeJSON(data)
	if e != nil {
		return nil, e
	}
	r, e := object(v, "ippVersion type mode method sessionId requestId optionId status createdAt expiresAt links failureReason", "payment credit")
	if e != nil || !mode(r["mode"]) || r["method"] != "invisi" || !text(r["sessionId"]) || !text(r["requestId"]) || !text(r["optionId"]) || !status(r, false) || !timestamp(r["createdAt"]) || !timestamp(r["expiresAt"]) || !before(r["createdAt"], r["expiresAt"]) {
		return nil, invalid()
	}
	l, e := object(r["links"], "credentials")
	if e != nil || !link(l["credentials"], "POST", origin) {
		return nil, invalid()
	}
	if r["mode"] == "payment" {
		if !header(r, "payment_request") {
			return nil, invalid()
		}
		if _, ok := r["credit"]; ok {
			return nil, invalid()
		}
		p, e := object(r["payment"], "network protocolAddress tokenAddress amount recipient")
		if e != nil || !network(p["network"]) || !hex(p["protocolAddress"], 20, true) || !hex(p["tokenAddress"], 20, true) || !integer(p["amount"], 128, true) {
			return nil, invalid()
		}
		rec, e := object(p["recipient"], "type identity")
		if e != nil || rec["type"] != "identity" || !field(rec["identity"]) {
			return nil, invalid()
		}
	} else {
		if !header(r, "credit_request") {
			return nil, invalid()
		}
		if _, ok := r["payment"]; ok {
			return nil, invalid()
		}
		if !validCredit(r["credit"], "operation network protocolAddress tokenAddress creditSharedKey") {
			return nil, invalid()
		}
		c := r["credit"].(map[string]any)
		if c["operation"] != "create" || !network(c["network"]) || !hex(c["protocolAddress"], 20, true) || !hex(c["tokenAddress"], 20, true) || !field(c["creditSharedKey"]) || before(c["expiresAt"], r["expiresAt"]) {
			return nil, invalid()
		}
	}
	return decodeTyped[Request](r)
}
func AssertRequestMatchesSession(r *Request, s *PaymentSession) error {
	if r == nil || s == nil || r.SessionID != s.SessionID || r.Mode != s.Mode || s.ActiveRequestID != nil && *s.ActiveRequestID != r.RequestID {
		return invalid()
	}
	for _, o := range s.Options {
		if o.ID != r.OptionID {
			continue
		}
		if o.Method != r.Method {
			return invalid()
		}
		if r.Mode == Payment && r.Payment != nil && o.Credit == nil && r.Payment.Network == o.Network && r.Payment.TokenAddress == o.Asset.Address && r.Payment.Amount == o.Amount {
			return nil
		}
		if r.Mode == Credit && r.Credit != nil && o.Credit != nil && r.Credit.Network == o.Network && r.Credit.TokenAddress == o.Asset.Address && same(r.Credit.CreditTerms, *o.Credit) {
			return nil
		}
		return invalid()
	}
	return invalid()
}
func AssertReceiptMatchesRequest(receipt *Receipt, r *Request) error {
	if receipt == nil || r == nil {
		return invalid()
	}
	var n, p string
	if r.Payment != nil {
		n, p = r.Payment.Network, r.Payment.ProtocolAddress
	} else if r.Credit != nil {
		n, p = r.Credit.Network, r.Credit.ProtocolAddress
	}
	if receipt.Network != n || receipt.ProtocolAddress != p {
		return invalid()
	}
	if r.Mode == Credit && r.Credit != nil {
		terms, draw := r.Credit.InitialPayment, receipt.InitialPayment
		if (terms == nil) != (draw == nil) || (terms != nil && draw != nil && terms.Amount != draw.NetAmount) {
			return invalid()
		}
	}
	return nil
}
func ParseCheckoutEntry(s string) (*CheckoutEntry, error) {
	i := strings.IndexByte(s, '#')
	if i < 0 || ValidateAPIURL(s[:i], "") != nil {
		return nil, invalid()
	}
	q, e := url.ParseQuery(s[i+1:])
	if e != nil || len(q) != 2 || len(q["sessionUrl"]) != 1 || len(q["token"]) != 1 {
		return nil, invalid()
	}
	session, token := q.Get("sessionUrl"), q.Get("token")
	if ValidateAPIURL(session, "") != nil || !bearer(token) || !noSecret(s[:i], token) || !noSecret(session, token) {
		return nil, invalid()
	}
	return &CheckoutEntry{session, token}, nil
}
func ParseCheckout(data []byte) (*Checkout, error) {
	v, e := DecodeJSON(data)
	if e != nil {
		return nil, e
	}
	c, e := object(v, "ippVersion type sessionUrl checkoutUrl expiresAt")
	if e != nil || !header(c, "checkout") || !timestamp(c["expiresAt"]) {
		return nil, invalid()
	}
	s, ok := c["sessionUrl"].(string)
	u, ok2 := c["checkoutUrl"].(string)
	if !ok || !ok2 || !text(u) || ValidateAPIURL(s, "") != nil {
		return nil, invalid()
	}
	entry, e := ParseCheckoutEntry(u)
	if e != nil || entry.SessionURL != s {
		return nil, invalid()
	}
	return decodeTyped[Checkout](c)
}

// ParseCheckoutInput allows expired input for idempotent create replay; new-session freshness is the server's responsibility.
func ParseCheckoutInput(data []byte) (*CheckoutInput, error) {
	v, e := DecodeJSON(data)
	if e != nil {
		return nil, e
	}
	c, e := object(v, "ippVersion mode options expiresAt", "description")
	if e != nil || c["ippVersion"] != json.Number("1") || !mode(c["mode"]) || !validOptions(c["options"], c["mode"]) || !timestamp(c["expiresAt"]) {
		return nil, invalid()
	}
	if d, ok := c["description"]; ok && !text(d) {
		return nil, invalid()
	}
	return decodeTyped[CheckoutInput](c)
}
func ParseCredential(data []byte, r *Request) (*Credential, error) {
	v, e := DecodeJSON(data)
	if e != nil {
		return nil, e
	}
	c, e := object(v, "ippVersion type mode method sessionId requestId proof")
	if e != nil || !header(c, "credential") || !mode(c["mode"]) || c["method"] != "invisi" || !text(c["sessionId"]) || !text(c["requestId"]) {
		return nil, invalid()
	}
	extra := ""
	kind := "note"
	if c["mode"] == "credit" {
		extra = "owner revocationShard revocationSlot"
		kind = "credit_created"
	}
	p, e := object(c["proof"], "type transactionHash commitment", extra)
	if e != nil || p["type"] != kind || !hex(p["transactionHash"], 32, true) || !field(p["commitment"]) {
		return nil, invalid()
	}
	if x, ok := p["owner"]; ok && !hex(x, 20, true) {
		return nil, invalid()
	}
	if x, ok := p["revocationShard"]; ok && !integer(x, 64, false) {
		return nil, invalid()
	}
	if x, ok := p["revocationSlot"]; ok && !number(x, 65535) {
		return nil, invalid()
	}
	out, e := decodeTyped[Credential](c)
	if e != nil {
		return nil, e
	}
	if r != nil && (out.Mode != r.Mode || out.Method != r.Method || out.SessionID != r.SessionID || out.RequestID != r.RequestID) {
		return nil, invalid()
	}
	return out, nil
}
func same(a, b any) bool { return reflect.DeepEqual(a, b) }
func clone[T any](v *T) *T {
	if v == nil {
		return nil
	}
	b, _ := json.Marshal(v)
	var out T
	_ = json.Unmarshal(b, &out)
	return &out
}
