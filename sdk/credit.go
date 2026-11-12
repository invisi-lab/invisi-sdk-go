package sdk

import (
	"github.com/invisi-lab/invisi-sdk-go/credit"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
)

type CreateCreditPreparation = credit.CreatePreparation
type UpdateCreditPreparation = credit.UpdatePreparation
type DrawCreditPreparation = credit.DrawPreparation
type RevokeCreditPreparation = credit.RevokePreparation
type CreditDrawMember = credit.DrawMember
type CreditUpdatePatch = credit.UpdatePatch

type ExplicitCreateCreditPreparation = transaction.ExplicitCreateCreditPreparation
type ExplicitUpdateCreditPreparation = transaction.ExplicitUpdateCreditPreparation
type ExplicitDrawCreditPreparation = transaction.ExplicitDrawCreditPreparation
type ExplicitRevokeCreditPreparation = transaction.ExplicitRevokeCreditPreparation
type CreditInput = transaction.CreditInput
