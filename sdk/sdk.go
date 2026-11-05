package sdk

import (
	"context"

	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
)

type AddressRegistry = transaction.AddressRegistry

type BroadcasterClient = transaction.BroadcasterClient

type BroadcasterClientOptions = transaction.BroadcasterClientOptions

type BroadcasterConfiguration = transaction.BroadcasterConfiguration

type BroadcasterFeeSource = transaction.BroadcasterFeeSource

type BroadcasterOptions = transaction.BroadcasterOptions

type CancelCreditPreparation = transaction.CancelCreditPreparation

type CreateCreditPreparation = transaction.CreateCreditPreparation

type CreditDrawInput = transaction.CreditDrawInput

type CreditInput = transaction.CreditInput

type CreditOpening = transaction.CreditOpening

type CreditPreview = transaction.CreditPreview

func DefineGasBenchmarks(data []byte) ([]GasBenchmarkProfile, error) {
	return transaction.DefineGasBenchmarks(data)
}

type Detector = transaction.Detector

type Disclosure = transaction.Disclosure

type DisclosureContext = transaction.DisclosureContext

type DisclosureEvidence = transaction.DisclosureEvidence

type DisclosureExport = transaction.DisclosureExport

func DisclosureFromCredit(note core.CreditNote) (*Disclosure, error) {
	return transaction.DisclosureFromCredit(note)
}

func DisclosureFromExport(data []byte) (*Disclosure, error) {
	return transaction.DisclosureFromExport(data)
}

func DisclosureFromNote(note core.Note) (*Disclosure, error) {
	return transaction.DisclosureFromNote(note)
}

type DisclosureOutput = transaction.DisclosureOutput

type DisclosureReference = transaction.DisclosureReference

type DisclosureResult = transaction.DisclosureResult

type DisclosureScope = transaction.DisclosureScope

type DisclosureVerifier = transaction.DisclosureVerifier

type DisclosureVerifierOptions = transaction.DisclosureVerifierOptions

type DrawCreditPreparation = transaction.DrawCreditPreparation

type EVMRecipient = transaction.EVMRecipient

type ExplicitCancelCreditPreparation = transaction.ExplicitCancelCreditPreparation

type ExplicitCreateCreditPreparation = transaction.ExplicitCreateCreditPreparation

type ExplicitDrawCreditPreparation = transaction.ExplicitDrawCreditPreparation

type ExplicitFunds = transaction.ExplicitFunds

type ExplicitShieldPreparation = transaction.ExplicitShieldPreparation

type ExplicitTopUpCreditPreparation = transaction.ExplicitTopUpCreditPreparation

type ExplicitTransferPreparation = transaction.ExplicitTransferPreparation

type FundingAuthorizationVerifier = transaction.FundingAuthorizationVerifier

type FundingPermitRequest = transaction.FundingPermitRequest

type FundingSigner = transaction.FundingSigner

type FundingSignerFunc = transaction.FundingSignerFunc

type FundingSource = transaction.FundingSource

type GasBenchmarkCircuit = transaction.GasBenchmarkCircuit

type GasBenchmarkDeployment = transaction.GasBenchmarkDeployment

type GasBenchmarkFingerprints = transaction.GasBenchmarkFingerprints

type GasBenchmarkMeasurement = transaction.GasBenchmarkMeasurement

type GasBenchmarkProfile = transaction.GasBenchmarkProfile

type GasBenchmarkRuntimeContract = transaction.GasBenchmarkRuntimeContract

type GasBenchmarkSample = transaction.GasBenchmarkSample

type GasBenchmarkShape = transaction.GasBenchmarkShape

type GasBenchmarkSource = transaction.GasBenchmarkSource

type GasBenchmarkToolchain = transaction.GasBenchmarkToolchain

type GasReference = transaction.GasReference

type GethFundingVerifier = transaction.GethFundingVerifier

type IdentityRecipient = transaction.IdentityRecipient

type Invisi = transaction.Invisi

type InvisiError = transaction.InvisiError

type InvisiErrorCode = transaction.InvisiErrorCode

type InvisiOptions = transaction.InvisiOptions

type InvisiRecipient = transaction.InvisiRecipient

const MaxUpdateBlocks = transaction.MaxUpdateBlocks

type MissingCapabilityError = transaction.MissingCapabilityError

func NewAddressRegistry(pool *Protocol, provider Provider) (*AddressRegistry, error) {
	return transaction.NewAddressRegistry(pool, provider)
}

func NewBroadcasterClient(config BroadcasterConfiguration, options ...BroadcasterClientOptions) (*BroadcasterClient, error) {
	return transaction.NewBroadcasterClient(config, options...)
}

func NewDisclosureVerifier(options DisclosureVerifierOptions) (*DisclosureVerifier, error) {
	return transaction.NewDisclosureVerifier(options)
}

func NewInvisi(options InvisiOptions) *Invisi { return transaction.NewInvisi(options) }

func NewNoteOutput(o NoteOutputOptions) (NoteOutput, error) { return transaction.NewNoteOutput(o) }

func NewProtocol(deployment ProtocolDeployment) (*Protocol, error) {
	return transaction.NewProtocol(deployment)
}

func NewProtocolRpcClient(endpoint string, options ...protocolrpc.ClientOptions) (*ProtocolRpcClient, error) {
	return transaction.NewProtocolRpcClient(endpoint, options...)
}

func NewViewingCapability(identity common.Address, key []byte, public []byte) (*ViewingCapability, error) {
	return transaction.NewViewingCapability(identity, key, public)
}

type NoteInput = transaction.NoteInput

type NoteOutput = transaction.NoteOutput

type NoteOutputContext = transaction.NoteOutputContext

type NoteOutputOptions = transaction.NoteOutputOptions

type NoteRecipient = transaction.NoteRecipient

type NullifierCapability = transaction.NullifierCapability

type OperationAggregateMember = transaction.OperationAggregateMember

type OperationPermissions = transaction.OperationPermissions

type OperationPreview = transaction.OperationPreview

type PaymentPreview = transaction.PaymentPreview

type PreparedTransaction = transaction.PreparedTransaction

type ProofBundle = transaction.ProofBundle

type Protocol = transaction.Protocol

type ProtocolCreditTransition = transaction.ProtocolCreditTransition

type ProtocolDeployment = transaction.ProtocolDeployment

func ProtocolFromRPC(ctx context.Context, rpc *ProtocolRpcClient, provider protocolrpc.ChainProvider) (*Protocol, error) {
	return transaction.ProtocolFromRPC(ctx, rpc, provider)
}

type ProtocolRpcClient = transaction.ProtocolRpcClient

type ProtocolTransactionUpdates = transaction.ProtocolTransactionUpdates

type Prover = transaction.Prover

type ProverOptions = transaction.ProverOptions

type Provider = transaction.Provider

type PublicWallet = transaction.PublicWallet

func PublicWalletFromAddress(address string) (*PublicWallet, error) {
	return transaction.PublicWalletFromAddress(address)
}

type RegistrationOperation = transaction.RegistrationOperation

type RegistrationPreparation = transaction.RegistrationPreparation

type RegistrationPreview = transaction.RegistrationPreview

type ShieldPreparation = transaction.ShieldPreparation

type TopUpCreditPreparation = transaction.TopUpCreditPreparation

type TransactionCall = transaction.TransactionCall

type TransactionIdentity = transaction.TransactionIdentity

type TransactionProveOptions = transaction.TransactionProveOptions

type TransactionUpdateStatus = transaction.TransactionUpdateStatus

type TransferPreparation = transaction.TransferPreparation

type VerifiedDisclosure = transaction.VerifiedDisclosure

type ViewingCapability = transaction.ViewingCapability

type Wallet = transaction.Wallet

type WalletAuthorization = transaction.WalletAuthorization

type WalletAuthorizationProvider = transaction.WalletAuthorizationProvider

func WalletFromPrivateKey(privateKey string) (*Wallet, error) {
	return transaction.WalletFromPrivateKey(privateKey)
}

type WalletReference = transaction.WalletReference

type WithBroadcasterOptions = transaction.WithBroadcasterOptions

type Withdrawal = transaction.Withdrawal

const PreparationFailed = transaction.PreparationFailed
const MissingCapability = transaction.MissingCapability
const InvalidRecipient = transaction.InvalidRecipient
const EVMAddressNotRegistered = transaction.EVMAddressNotRegistered
const RecipientLookupFailed = transaction.RecipientLookupFailed
const BroadcasterFeePublicFunding = transaction.BroadcasterFeePublicFunding
const BroadcasterFeePrivateInputs = transaction.BroadcasterFeePrivateInputs
const BroadcasterFeePayout = transaction.BroadcasterFeePayout
const TransactionNotFound = transaction.TransactionNotFound
const TransactionPendingIndex = transaction.TransactionPendingIndex
const TransactionIndexed = transaction.TransactionIndexed
