package provenance

import (
	public "github.com/invisi-lab/invisi-sdk-go/protocol/provenance"
	local "github.com/invisi-lab/invisi-sdk-go/prover/provenance"
	client "github.com/invisi-lab/invisi-sdk-go/rpc/provenance"
)

type RecognizedTree = public.RecognizedTree
type Membership = public.Membership
type ProvenanceProof = public.ProvenanceProof
type ProofContext = public.ProofContext
type Scope = public.Scope
type Source = public.Source
type RootCertificate = public.RootCertificate
type ManifestEntry = public.ManifestEntry
type Manifest = public.Manifest
type ManifestIdentity = public.ManifestIdentity
type ProofBundle = public.ProofBundle
type ProofSidecar = public.ProofSidecar
type TrustContext = public.TrustContext
type RotationRecord = public.RotationRecord
type KeyValidity = public.KeyValidity
type RotationCheckpoint = public.RotationCheckpoint
type TrustedKeyring = public.TrustedKeyring

var SupportedCapacity = public.SupportedCapacity
var Uint = public.Uint
var Field = public.Field
var ScopeHash = public.ScopeHash
var RecognizedLeaf = public.RecognizedLeaf
var InputDigest = public.InputDigest
var SourcesDigest = public.SourcesDigest
var BindingDigest = public.BindingDigest
var OperationDigest = public.OperationDigest
var ProvenanceIntentHash = public.ProvenanceIntentHash
var RootCertificatePayload = public.RootCertificatePayload
var RootCertificateMessage = public.RootCertificateMessage
var CheckpointHash = public.CheckpointHash
var VerifyRootCertificate = public.VerifyRootCertificate
var OperationDigestDecimal = public.OperationDigestDecimal
var PublicInputNames = public.PublicInputNames
var AuthenticateManifest = public.AuthenticateManifest
var AssertCertificateTrust = public.AssertCertificateTrust
var ValidateProofBundle = public.ValidateProofBundle
var ProofBundleToWire = public.ProofBundleToWire
var ProofSidecarsToWire = public.ProofSidecarsToWire
var CanonicalWord = public.CanonicalWord
var HashFields = public.HashFields
var DigestLimbs = public.DigestLimbs
var EqualNumber = public.EqualNumber
var CheckInteger = public.CheckInteger
var CertificateToWire = public.CertificateToWire
var AssertPublicSubmission = public.AssertPublicSubmission
var RotationPayload = public.RotationPayload
var RotationMessage = public.RotationMessage
var RotationHash = public.RotationHash
var VerifyRotationChain = public.VerifyRotationChain
var WithRotationHistory = public.WithRotationHistory

type Client = client.Client
type Result = client.Result
type TrackingMatch = client.TrackingMatch
type Tracking = client.Tracking

var TrackingFromResult = client.TrackingFromResult
var PreflightTracking = client.PreflightTracking

type LocalAssets = local.LocalAssets
type NativeProver = local.NativeProver
type WitnessInput = local.WitnessInput
type WitnessRequest = local.WitnessRequest
type Witness = local.Witness
type Opening = local.Opening
type ProofResult = local.ProofResult
type OfflineProver = local.OfflineProver

var BuildWitness = local.BuildWitness
var NoteOpening = local.NoteOpening
var ExportSidecar = local.ExportSidecar
var word = public.CanonicalWord
var hash = public.HashFields
var limbs = public.DigestLimbs
var equalNumber = public.EqualNumber
var check = public.CheckInteger
var certificateToWire = public.CertificateToWire
var assertPublicSubmission = public.AssertPublicSubmission
var publicInputNames = public.PublicInputNames()
var ErrTrackingStaleGeneration = client.ErrTrackingStaleGeneration

var decodeFixed = public.DecodeFixed
