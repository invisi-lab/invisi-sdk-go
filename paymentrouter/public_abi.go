package paymentrouter

import (
	_ "embed"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

//go:embed payment-router.json
var paymentRouterJSON string

//go:embed zeroex-payment-router.json
var zeroExPaymentRouterJSON string

// PaymentRouterABI returns the canonical multi-mode Router ABI.
func PaymentRouterABI() abi.ABI { return CanonicalRouterABI() }

// LegacyPaymentRouterABI returns the historical V3-only ABI.
func LegacyPaymentRouterABI() abi.ABI { return mustABI(paymentRouterJSON) }

// ZeroExPaymentRouterABI returns the legacy native Router ABI with both swap channels.
func ZeroExPaymentRouterABI() abi.ABI { return mustABI(zeroExPaymentRouterJSON) }
