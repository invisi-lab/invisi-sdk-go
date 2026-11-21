package session

import (
	"fmt"
	protocolrpc "github.com/invisi-labs/invisi-sdk-go/rpc"
)

// SetGrantStatus records an authenticated status observation. Unknown is distinct
// from permanently disabled; only status 1 can authorize an inventory selection.
func (s *StoredState) SetGrantStatus(status uint8, block *protocolrpc.CanonicalBlock) error {
	if status > 2 {
		return fmt.Errorf("invalid Credit grant status")
	}
	s.GrantStatus = status
	s.Revoked = status == 2
	s.GrantStatusBlock = nil
	if block != nil {
		copy := *block
		s.GrantStatusBlock = &copy
	}
	return nil
}

// MergeGrantStatus preserves a newer canonical observation when replaying older
// receipts. A terminal disable survives every later private state of its grant.
func (s StoredState) MergeGrantStatus(previous StoredState) StoredState {
	a, b := s.Delivery.State, previous.Delivery.State
	if a.ProtocolContext == b.ProtocolContext && a.OwnerWalletIdentity == b.OwnerWalletIdentity && a.TokenAddress == b.TokenAddress && a.GrantID == b.GrantID {
		if s.GrantStatus == 2 {
			if previous.GrantStatus == 2 && previous.GrantStatusBlock != nil && (s.GrantStatusBlock == nil || previous.GrantStatusBlock.Number > s.GrantStatusBlock.Number) {
				_ = s.SetGrantStatus(2, previous.GrantStatusBlock)
			}
			s.Revoked = true
			return s
		}
		preserve := previous.GrantStatus == 2 || previous.GrantStatusBlock != nil && (s.GrantStatusBlock == nil || previous.GrantStatusBlock.Number > s.GrantStatusBlock.Number || previous.GrantStatusBlock.Number == s.GrantStatusBlock.Number && (previous.GrantStatus == 0 || s.GrantStatus != 0)) || s.GrantStatus == 0 && s.GrantStatusBlock == nil && previous.GrantStatus != 0
		if preserve {
			_ = s.SetGrantStatus(previous.GrantStatus, previous.GrantStatusBlock)
		}
	}
	s.Revoked = s.GrantStatus == 2
	return s
}

// MergeGrantStatusFrom carries public grant authority across different private
// commitments of the same scoped grant, including first discovery of a child.
func (s StoredState) MergeGrantStatusFrom(records map[string]StoredState) StoredState {
	for _, previous := range records {
		s = s.MergeGrantStatus(previous)
	}
	return s
}
