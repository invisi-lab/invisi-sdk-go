package ipp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HTTPOptions struct {
	HTTPClient     *http.Client
	RequestTimeout time.Duration
}
type ClientOptions struct {
	SessionURL     string
	Token          string
	HTTPClient     *http.Client
	RequestTimeout time.Duration
}
type WaitOptions struct {
	Interval time.Duration
	Timeout  time.Duration
}
type transport struct {
	client  *http.Client
	timeout time.Duration
}

func newTransport(o HTTPOptions) (transport, error) {
	if o.RequestTimeout < 0 || o.RequestTimeout > 2147483647*time.Millisecond {
		return transport{}, invalid()
	}
	if o.RequestTimeout == 0 {
		o.RequestTimeout = 15 * time.Second
	}
	c := http.Client{}
	if o.HTTPClient != nil {
		c = *o.HTTPClient
	}
	c.Jar = nil
	c.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return transport{&c, o.RequestTimeout}, nil
}
func retryAfter(v string) (time.Duration, bool) {
	if v == "" {
		return 0, false
	}
	var d time.Duration
	if regexp.MustCompile(`^[0-9]+$`).MatchString(v) {
		n, e := strconv.ParseUint(v, 10, 64)
		if e != nil || n > 2147483 {
			return 2147483647 * time.Millisecond, true
		}
		d = time.Duration(n) * time.Second
	} else {
		t, e := http.ParseTime(v)
		if e != nil {
			return 0, false
		}
		d = time.Until(t)
	}
	if d < 0 {
		d = 0
	}
	if d > 2147483647*time.Millisecond {
		d = 2147483647 * time.Millisecond
	}
	return d, true
}
func readResponse(res *http.Response) ([]byte, error) {
	if res == nil || res.Body == nil {
		return nil, invalid()
	}
	defer res.Body.Close()
	if length := res.Header.Get("Content-Length"); length != "" {
		n, e := strconv.ParseUint(length, 10, 64)
		if e != nil || !regexp.MustCompile(`^[0-9]+$`).MatchString(length) || n > MaxResponseBytes {
			return nil, invalid()
		}
	}
	media, _, e := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if e != nil || media != "application/json" {
		return nil, invalid()
	}
	b, e := io.ReadAll(io.LimitReader(res.Body, MaxResponseBytes+1))
	if e != nil || len(b) > MaxResponseBytes {
		return nil, invalid()
	}
	if _, e = DecodeJSON(b); e != nil {
		return nil, e
	}
	return b, nil
}
func (t transport) send(ctx context.Context, u, token, method string, body any, key string, statuses ...int) ([]byte, http.Header, error) {
	if ValidateAPIURL(u, "") != nil || !bearer(token) || !noSecret(u, token) || key != "" && !idempotency(key) {
		return nil, nil, invalid()
	}
	if ctx.Err() != nil {
		return nil, nil, &Error{Code: "aborted"}
	}
	parentContext := ctx
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	var reader io.Reader
	if body != nil {
		b, e := json.Marshal(body)
		if e != nil {
			return nil, nil, invalid()
		}
		reader = bytes.NewReader(b)
	}
	req, e := http.NewRequestWithContext(ctx, method, u, reader)
	if e != nil {
		return nil, nil, invalid()
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Cache-Control", "no-store")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	res, e := t.client.Do(req)
	if e != nil {
		if res != nil && res.Body != nil {
			res.Body.Close()
		}
		if parentContext.Err() != nil {
			return nil, nil, &Error{Code: "aborted"}
		}
		return nil, nil, &Error{Code: "transport_error", Retryable: true}
	}
	if res.StatusCode >= 300 && res.StatusCode < 400 || res.Request != nil && res.Request.URL.String() != u {
		res.Body.Close()
		return nil, nil, invalid()
	}
	b, e := readResponse(res)
	if e != nil {
		return nil, nil, e
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		v, _ := DecodeJSON(b)
		m, e := object(v, "ippVersion type error")
		if e != nil || !header(m, "error") {
			return nil, nil, invalid()
		}
		details, e := object(m["error"], "code message retryable")
		if e != nil || !text(details["code"]) || !text(details["message"]) {
			return nil, nil, invalid()
		}
		code := details["code"].(string)
		retry, ok := details["retryable"].(bool)
		if !ok || !regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`).MatchString(code) {
			return nil, nil, invalid()
		}
		if strings.Contains(code, token) {
			code = "remote_error"
		}
		wait, has := retryAfter(res.Header.Get("Retry-After"))
		return nil, nil, &Error{Code: code, Retryable: retry, Status: res.StatusCode, RetryAfter: wait, HasRetryAfter: has}
	}
	for _, s := range statuses {
		if res.StatusCode == s {
			return b, res.Header, nil
		}
	}
	return nil, nil, invalid()
}

// Client owns an opaque session credential. Methods return independent copies and serialize state changes.
type Client struct {
	mu                 *sync.Mutex
	url, token, origin string
	http               transport
	session            *PaymentSession
	request            *Request
	requestURL, option string
	retry              time.Duration
	hasRetry           bool
}

func NewClient(o ClientOptions) (*Client, error) {
	if ValidateAPIURL(o.SessionURL, "") != nil || !bearer(o.Token) || !noSecret(o.SessionURL, o.Token) {
		return nil, invalid()
	}
	t, e := newTransport(HTTPOptions{o.HTTPClient, o.RequestTimeout})
	if e != nil {
		return nil, e
	}
	return &Client{mu: &sync.Mutex{}, url: o.SessionURL, token: o.Token, origin: urlOrigin(o.SessionURL), http: t}, nil
}
func ClientFromCheckout(c *Checkout, o HTTPOptions) (*Client, error) {
	b, e := EncodeJSON(c)
	if e != nil {
		return nil, invalid()
	}
	v, e := ParseCheckout(b)
	if e != nil {
		return nil, e
	}
	entry, e := ParseCheckoutEntry(v.CheckoutURL)
	if e != nil {
		return nil, e
	}
	return NewClient(ClientOptions{entry.SessionURL, entry.Token, o.HTTPClient, o.RequestTimeout})
}

var linkPattern = regexp.MustCompile(`<([^<>]+)>\s*;\s*rel="ipp-(session|checkout)"`)

func ClientFromPaymentRequired(res *http.Response, o HTTPOptions) (*Client, error) {
	if res == nil || res.StatusCode != 402 {
		return nil, invalid()
	}
	matches := linkPattern.FindAllStringSubmatch(strings.Join(res.Header.Values("Link"), ","), -1)
	values := map[string]string{}
	for _, m := range matches {
		if _, ok := values[m[2]]; ok {
			return nil, invalid()
		}
		values[m[2]] = m[1]
	}
	if len(values) != 2 {
		return nil, invalid()
	}
	entry, e := ParseCheckoutEntry(values["checkout"])
	if e != nil || entry.SessionURL != values["session"] {
		return nil, invalid()
	}
	c, e := NewClient(ClientOptions{entry.SessionURL, entry.Token, o.HTTPClient, o.RequestTimeout})
	if e != nil {
		return nil, e
	}
	b, e := readResponse(res)
	if e != nil {
		return nil, e
	}
	s, e := ParsePaymentSession(b, c.origin)
	if e != nil {
		return nil, e
	}
	if e = c.acceptSession(s); e != nil {
		return nil, e
	}
	return c, nil
}
func (c *Client) sessionLinksSafe(s *PaymentSession, creditSharedKey string) bool {
	if s == nil {
		return false
	}
	secrets := []string{c.token}
	if creditSharedKey != "" {
		secrets = append(secrets, creditSharedKey)
	}
	for _, secret := range secrets {
		if !noSecret(c.url, secret) || !noSecret(s.Links.Confirm.Href, secret) ||
			s.Links.Request != nil && !noSecret(s.Links.Request.Href, secret) ||
			c.requestURL != "" && !noSecret(c.requestURL, secret) {
			return false
		}
	}
	return true
}
func (c *Client) acceptSession(s *PaymentSession) error {
	key := ""
	if c.request != nil && c.request.Credit != nil {
		key = c.request.Credit.CreditSharedKey
	}
	if !c.sessionLinksSafe(s, key) {
		return invalid()
	}
	old := c.session
	if old != nil {
		if old.SessionID != s.SessionID || old.Mode != s.Mode || old.CreatedAt != s.CreatedAt || old.ExpiresAt != s.ExpiresAt || !same(old.Merchant, s.Merchant) || !same(old.Description, s.Description) || !same(old.Options, s.Options) || old.ActiveRequestID != nil && !same(old.ActiveRequestID, s.ActiveRequestID) {
			return invalid()
		}
		if (old.Status == "succeeded" || old.Status == "failed") && (old.Status != s.Status || !same(old.FailureReason, s.FailureReason)) {
			return invalid()
		}
		if old.Receipt != nil && !same(old.Receipt, s.Receipt) {
			return invalid()
		}
	}
	if c.request != nil {
		if AssertRequestMatchesSession(c.request, s) != nil || s.ActiveRequestID == nil || *s.ActiveRequestID != c.request.RequestID {
			return invalid()
		}
		if s.Receipt != nil && AssertReceiptMatchesRequest(s.Receipt, c.request) != nil {
			return invalid()
		}
	}
	c.session = clone(s)
	if s.Links.Request != nil {
		c.requestURL = s.Links.Request.Href
	}
	return nil
}
func (c *Client) readSession(ctx context.Context) (*PaymentSession, error) {
	b, h, e := c.http.send(ctx, c.url, c.token, "GET", nil, "", 200)
	if e != nil {
		return nil, e
	}
	s, e := ParsePaymentSession(b, c.origin)
	if e != nil {
		return nil, e
	}
	if e = c.acceptSession(s); e != nil {
		return nil, e
	}
	c.retry, c.hasRetry = retryAfter(h.Get("Retry-After"))
	return clone(s), nil
}
func (c *Client) ReadSession(ctx context.Context) (*PaymentSession, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.readSession(ctx)
}
func (c *Client) acceptRequest(b []byte) (*Request, error) {
	r, e := ParsePaymentRequest(b, c.origin)
	if e != nil {
		return nil, e
	}
	if !noSecret(r.Links.Credentials.Href, c.token) || r.Credit != nil && !noSecret(r.Links.Credentials.Href, r.Credit.CreditSharedKey) {
		return nil, invalid()
	}
	if AssertRequestMatchesSession(r, c.session) != nil || c.option != "" && r.OptionID != c.option {
		return nil, invalid()
	}
	if r.Credit != nil && !c.sessionLinksSafe(c.session, r.Credit.CreditSharedKey) {
		return nil, invalid()
	}
	if c.session.Receipt != nil && AssertReceiptMatchesRequest(c.session.Receipt, r) != nil {
		return nil, invalid()
	}
	if old := c.request; old != nil {
		a, z := *r, *old
		a.Status = ""
		z.Status = ""
		a.FailureReason = nil
		z.FailureReason = nil
		a.Links = RequestLinks{}
		z.Links = RequestLinks{}
		if !same(a, z) || (old.Status == "succeeded" || old.Status == "failed") && (old.Status != r.Status || !same(old.FailureReason, r.FailureReason)) {
			return nil, invalid()
		}
	}
	c.request = clone(r)
	c.option = r.OptionID
	return clone(r), nil
}
func (c *Client) Confirm(ctx context.Context, option, key string) (*Request, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !idempotency(key) {
		return nil, invalid()
	}
	if c.session == nil {
		if _, e := c.readSession(ctx); e != nil {
			return nil, e
		}
	}
	found := false
	for _, o := range c.session.Options {
		if o.ID == option {
			found = true
		}
	}
	if !found || c.option != "" && c.option != option {
		return nil, invalid()
	}
	b, h, e := c.http.send(ctx, c.session.Links.Confirm.Href, c.token, "POST", struct {
		OptionID string `json:"optionId"`
	}{option}, key, 200, 201)
	if e != nil {
		return nil, e
	}
	location := h.Get("Location")
	if ValidateAPIURL(location, c.origin) != nil || !noSecret(location, c.token) {
		return nil, invalid()
	}
	candidate, e := ParsePaymentRequest(b, c.origin)
	if e != nil || candidate.OptionID != option || candidate.Credit != nil && !noSecret(location, candidate.Credit.CreditSharedKey) {
		return nil, invalid()
	}
	r, e := c.acceptRequest(b)
	if e != nil {
		return nil, e
	}
	c.requestURL = location
	return r, nil
}
func (c *Client) readRequest(ctx context.Context) (*Request, error) {
	if c.session == nil {
		if _, e := c.readSession(ctx); e != nil {
			return nil, e
		}
	}
	if c.requestURL == "" {
		return nil, invalid()
	}
	b, _, e := c.http.send(ctx, c.requestURL, c.token, "GET", nil, "", 200)
	if e != nil {
		return nil, e
	}
	return c.acceptRequest(b)
}
func (c *Client) ReadRequest(ctx context.Context) (*Request, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.readRequest(ctx)
}

// SubmitCredential sends optional evidence hints; it never creates or executes a transaction. An empty key omits the optional header.
func (c *Client) SubmitCredential(ctx context.Context, value Credential, key string) (*CredentialAcknowledgement, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if key != "" && !idempotency(key) {
		return nil, invalid()
	}
	if c.request == nil {
		if _, e := c.readRequest(ctx); e != nil {
			return nil, e
		}
	}
	b, e := EncodeJSON(value)
	if e != nil {
		return nil, invalid()
	}
	credential, e := ParseCredential(b, c.request)
	if e != nil {
		return nil, e
	}
	b, _, e = c.http.send(ctx, c.request.Links.Credentials.Href, c.token, "POST", credential, key, 202)
	if e != nil {
		return nil, e
	}
	v, e := DecodeJSON(b)
	if e != nil {
		return nil, e
	}
	ack, e := object(v, "ippVersion type sessionId requestId status")
	if e != nil || !header(ack, "credential_acknowledgement") || ack["status"] != "accepted" || ack["sessionId"] != c.request.SessionID || ack["requestId"] != c.request.RequestID {
		return nil, invalid()
	}
	return decodeTyped[CredentialAcknowledgement](ack)
}
func (c *Client) WaitForResult(ctx context.Context, o WaitOptions) (*PaymentSession, error) {
	if o.Interval < 0 || o.Timeout < 0 || o.Interval > 2147483647*time.Millisecond || o.Timeout > 2147483647*time.Millisecond {
		return nil, invalid()
	}
	if o.Interval == 0 {
		o.Interval = time.Second
	}
	if o.Timeout == 0 {
		o.Timeout = time.Minute
	}
	outer := ctx
	ctx, cancel := context.WithTimeout(ctx, o.Timeout)
	defer cancel()
	for {
		s, e := c.ReadSession(ctx)
		if ctx.Err() != nil {
			if outer.Err() != nil {
				return nil, &Error{Code: "aborted"}
			}
			return nil, &Error{Code: "timeout", Retryable: true}
		}
		wait := o.Interval
		if e == nil {
			if s.Status == "succeeded" || s.Status == "failed" {
				return s, nil
			}
			c.mu.Lock()
			if c.hasRetry {
				wait = c.retry
			}
			c.mu.Unlock()
		} else {
			var ie *Error
			if !errors.As(e, &ie) || !ie.Retryable || ie.Status != 429 && ie.Status != 503 {
				return nil, e
			}
			if ie.HasRetryAfter {
				wait = ie.RetryAfter
			}
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			if outer.Err() != nil {
				return nil, &Error{Code: "aborted"}
			}
			return nil, &Error{Code: "timeout", Retryable: true}
		case <-timer.C:
		}
	}
}

type MerchantClientOptions struct {
	CreateURL string
	Token     string
	HTTPOptions
}
type MerchantClient struct {
	url, token string
	http       transport
}

func NewMerchantClient(o MerchantClientOptions) (*MerchantClient, error) {
	if ValidateAPIURL(o.CreateURL, "") != nil || !bearer(o.Token) || !noSecret(o.CreateURL, o.Token) {
		return nil, invalid()
	}
	t, e := newTransport(o.HTTPOptions)
	if e != nil {
		return nil, e
	}
	return &MerchantClient{o.CreateURL, o.Token, t}, nil
}
func (m *MerchantClient) CreateCheckout(ctx context.Context, value CheckoutInput, key string) (*Checkout, error) {
	if !idempotency(key) {
		return nil, invalid()
	}
	b, e := EncodeJSON(value)
	if e != nil {
		return nil, invalid()
	}
	input, e := ParseCheckoutInput(b)
	if e != nil {
		return nil, e
	}
	b, _, e = m.http.send(ctx, m.url, m.token, "POST", input, key, 200, 201)
	if e != nil {
		return nil, e
	}
	out, e := ParseCheckout(b)
	if e != nil {
		return nil, e
	}
	if out.ExpiresAt != input.ExpiresAt {
		return nil, invalid()
	}
	return out, nil
}
func CreateCheckoutURL(sessionURL, token, checkoutPage string) (string, error) {
	if checkoutPage == "" {
		checkoutPage = "https://dapp.invisipay.fi/checkout"
	}
	if ValidateAPIURL(checkoutPage, "") != nil || ValidateAPIURL(sessionURL, "") != nil || !bearer(token) || !noSecret(checkoutPage, token) || !noSecret(sessionURL, token) {
		return "", invalid()
	}
	return checkoutPage + "#sessionUrl=" + url.QueryEscape(sessionURL) + "&token=" + url.QueryEscape(token), nil
}
func PaymentRequiredResponse(s *PaymentSession, c *Checkout) (*http.Response, error) {
	b, e := EncodeJSON(c)
	if e != nil {
		return nil, invalid()
	}
	co, e := ParseCheckout(b)
	if e != nil {
		return nil, e
	}
	b, e = EncodeJSON(s)
	if e != nil {
		return nil, invalid()
	}
	session, e := ParsePaymentSession(b, urlOrigin(co.SessionURL))
	if e != nil {
		return nil, e
	}
	if session.ExpiresAt != co.ExpiresAt || session.Status == "succeeded" || session.Status == "failed" {
		return nil, invalid()
	}
	return &http.Response{StatusCode: 402, Header: http.Header{"Content-Type": []string{"application/json"}, "Cache-Control": []string{"no-store"}, "Link": []string{"<" + co.SessionURL + ">; rel=\"ipp-session\", <" + co.CheckoutURL + ">; rel=\"ipp-checkout\""}}, Body: io.NopCloser(bytes.NewReader(b))}, nil
}
