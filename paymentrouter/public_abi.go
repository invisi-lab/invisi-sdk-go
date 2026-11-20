package paymentrouter

import (
	_ "embed"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

//go:embed payment-router.json
var paymentRouterJSON string

//go:embed zeroex-payment-router.json
var zeroExPaymentRouterJSON string

// PaymentRouterABI returns a fresh ABI snapshot of the V3-only Router.
func PaymentRouterABI() abi.ABI { return mustABI(paymentRouterJSON) }

// ZeroExPaymentRouterABI returns a fresh ABI snapshot of the Router with both channels.
func ZeroExPaymentRouterABI() abi.ABI { return mustABI(zeroExPaymentRouterJSON) }
