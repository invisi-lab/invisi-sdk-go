package ipp

// String and GoString keep routine diagnostics from exposing bearer credentials or private openings.
// JSON serialization intentionally retains the explicit protocol wire representation.
func (Client) String() string           { return "ipp.Client{redacted}" }
func (Client) GoString() string         { return "ipp.Client{redacted}" }
func (MerchantClient) String() string   { return "ipp.MerchantClient{redacted}" }
func (MerchantClient) GoString() string { return "ipp.MerchantClient{redacted}" }
func (Checkout) String() string         { return "ipp.Checkout{redacted}" }
func (Checkout) GoString() string       { return "ipp.Checkout{redacted}" }
func (CheckoutEntry) String() string    { return "ipp.CheckoutEntry{redacted}" }
func (CheckoutEntry) GoString() string  { return "ipp.CheckoutEntry{redacted}" }
func (CreditDetails) String() string    { return "ipp.CreditDetails{redacted}" }
func (CreditDetails) GoString() string  { return "ipp.CreditDetails{redacted}" }
