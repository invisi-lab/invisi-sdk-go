// Managed handles delegate transaction algorithms to one shared implementation.
// Persistence, leases and recovery remain owned by this package.
package session

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/invisi-lab/invisi-sdk-go/broadcaster"
	"github.com/invisi-lab/invisi-sdk-go/core"
	gasmodel "github.com/invisi-lab/invisi-sdk-go/internal/gas"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

func (o *Operation) ExportAggregateMember(ctx context.Context) (sdk.OperationAggregateMember, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	execution := o.execution()
	result0, result1 := execution.ExportAggregateMember(ctx)
	return result0, result1
}
func (o *Operation) matchesTransactionData(data []byte) bool {
	execution := o.execution()
	result0 := execution.MatchesTransactionData(data)
	return result0
}
func (o *Operation) broadcasterIntent() ([]byte, error) {
	execution := o.execution()
	result0, result1 := execution.BroadcasterIntent()
	return result0, result1
}
func (o *Operation) applyBroadcasterFee(ctx context.Context, base transaction.Plan, recipient common.Address, fee *big.Int, opt sdk.WithBroadcasterOptions) error {
	execution := o.execution()
	result0 := execution.ApplyBroadcasterFee(ctx, base, recipient, fee, opt)
	o.apply(execution)
	return result0
}
func (o *Operation) withBroadcaster(ctx context.Context, c *broadcaster.Client, opt sdk.WithBroadcasterOptions) error {
	execution := o.execution()
	result0 := execution.WithBroadcaster(ctx, c, opt)
	o.apply(execution)
	return result0
}
func (o *Operation) extendBroadcasterInputs(ctx context.Context, base transaction.Plan, fee *big.Int) (transaction.Plan, error) {
	execution := o.execution()
	result0, result1 := execution.ExtendBroadcasterInputs(ctx, base, fee)
	return result0, result1
}
func (o *Operation) exportBroadcasterQuote() (broadcaster.SignedQuote, error) {
	execution := o.execution()
	result0, result1 := execution.ExportBroadcasterQuoteInternal()
	return result0, result1
}
func (o *Operation) ExportBroadcasterQuote(ctx context.Context) (broadcaster.SignedQuote, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	execution := o.execution()
	result0, result1 := execution.ExportBroadcasterQuote(ctx)
	return result0, result1
}
func (o *Operation) ExportBroadcasterIntent(ctx context.Context) ([]byte, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	execution := o.execution()
	result0, result1 := execution.ExportBroadcasterIntent(ctx)
	return result0, result1
}
func (o *Operation) submitBroadcaster(ctx context.Context, c *broadcaster.Client) (broadcaster.Status, error) {
	execution := o.execution()
	result0, result1 := execution.SubmitBroadcaster(ctx, c)
	o.apply(execution)
	return result0, result1
}
func (o *Operation) validateBroadcasterPlan() error {
	execution := o.execution()
	result0 := execution.ValidateBroadcasterPlan()
	return result0
}
func (o *Operation) creditPrivateInputs(private map[string]any) (map[string]any, error) {
	execution := o.execution()
	result0, result1 := execution.CreditPrivateInputs(private)
	return result0, result1
}
func (o *Operation) encodeCredit(proof []byte, caller *common.Address) (*protocol.Calldata, error) {
	execution := o.execution()
	result0, result1 := execution.EncodeCredit(proof, caller)
	return result0, result1
}
func (o *Operation) walletRequest(ctx context.Context) (core.WalletAuthorizationRequest, error) {
	execution := o.execution()
	result0, result1 := execution.WalletRequest(ctx)
	return result0, result1
}
func (o *Operation) ExportWalletAuthorization(ctx context.Context) (core.SigningRequest, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	execution := o.execution()
	result0, result1 := execution.ExportWalletAuthorization(ctx)
	return result0, result1
}
func (o *Operation) attachWallet(ctx context.Context, signature []byte) error {
	execution := o.execution()
	result0 := execution.AttachWallet(ctx, signature)
	o.apply(execution)
	return result0
}
func (o *Operation) fundingRequest(ctx context.Context) (apitypes.TypedData, error) {
	execution := o.execution()
	result0, result1 := execution.FundingRequest(ctx)
	return result0, result1
}
func (o *Operation) ExportFundingAuthorization(ctx context.Context) (core.SigningRequest, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	execution := o.execution()
	result0, result1 := execution.ExportFundingAuthorization(ctx)
	return result0, result1
}
func (o *Operation) attachFunding(ctx context.Context, signature []byte) error {
	execution := o.execution()
	result0 := execution.AttachFunding(ctx, signature)
	o.apply(execution)
	return result0
}
func (o *Operation) EstimateGas() (sdk.GasReference, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	execution := o.execution()
	result0, result1 := execution.EstimateGas()
	return result0, result1
}
func (o *Operation) gasRequest() (gasmodel.GasEstimateRequest, error) {
	execution := o.execution()
	result0, result1 := execution.GasRequest()
	return result0, result1
}
func (o *Operation) ID() string {
	o.metadataMu.RLock()
	defer o.metadataMu.RUnlock()
	return o.plan.ID
}
func (o *Operation) String() string {
	return "Operation(" + o.ID() + ",[private plan redacted])"
}
func (o *Operation) GoString() string { return o.String() }
func (o *Operation) MarshalJSON() ([]byte, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	execution := o.execution()
	result0, result1 := execution.MarshalJSON()
	return result0, result1
}
func (o *Operation) deliveries() []protocol.Delivery {
	execution := o.execution()
	result0 := execution.Deliveries()
	return result0
}
func (o *Operation) words() ([][32]byte, error) {
	execution := o.execution()
	result0, result1 := execution.Words()
	return result0, result1
}
func (o *Operation) authorizeInternal(ctx context.Context) (*Operation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	execution := o.execution()
	_, err := execution.AuthorizeInternal(ctx)
	o.apply(execution)
	return o, err
}
func (o *Operation) privateInputs() (map[string]any, error) {
	execution := o.execution()
	result0, result1 := execution.PrivateInputs()
	return result0, result1
}
func (o *Operation) circuitName() string {
	execution := o.execution()
	result0 := execution.CircuitName()
	return result0
}
func (o *Operation) proveInternal(ctx context.Context, options ...sdk.TransactionProveOptions) (*Operation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	execution := o.execution()
	_, err := execution.ProveInternal(ctx, options...)
	o.apply(execution)
	return o, err
}
func (o *Operation) readyInternal(ctx context.Context) (*Operation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	execution := o.execution()
	_, err := execution.ReadyInternal(ctx)
	o.apply(execution)
	return o, err
}
func (o *Operation) fundingTypedData() (apitypes.TypedData, error) {
	execution := o.execution()
	result0, result1 := execution.FundingTypedData()
	return result0, result1
}
func (o *Operation) authorizeFundingInternal(ctx context.Context) (*Operation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	execution := o.execution()
	_, err := execution.AuthorizeFundingInternal(ctx)
	o.apply(execution)
	return o, err
}
func (o *Operation) shieldParams() protocol.BuildShieldCalldataParams {
	execution := o.execution()
	result0 := execution.ShieldParams()
	return result0
}
func (o *Operation) encode() error {
	execution := o.execution()
	result0 := execution.Encode()
	o.apply(execution)
	return result0
}
func (o *Operation) sources() []protocol.InputSource {
	execution := o.execution()
	result0 := execution.Sources()
	return result0
}
func (o *Operation) validateExport(ctx context.Context) error {
	execution := o.execution()
	result0 := execution.ValidateExport(ctx)
	return result0
}
func (o *Operation) checkDrawPeriod(ctx context.Context) error {
	execution := o.execution()
	result0 := execution.CheckDrawPeriod(ctx)
	return result0
}
func (o *Operation) Kind() string {
	o.metadataMu.RLock()
	defer o.metadataMu.RUnlock()
	return transaction.PublicOperationKind(o.plan)
}
func (o *Operation) Preview() sdk.OperationPreview {
	o.mu.Lock()
	defer o.mu.Unlock()
	execution := o.execution()
	result0 := execution.Preview()
	return result0
}
func (o *Operation) validateRecovered() error {
	execution := o.execution()
	result0 := execution.ValidateRecovered()
	o.apply(execution)
	return result0
}
