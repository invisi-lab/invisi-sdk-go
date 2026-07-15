package core

import "testing"

func TestDomainConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant uint64
		expected uint64
	}{
		{"DomainNKey", DomainNKey, 1},
		{"DomainIdent", DomainIdent, 2},
		{"DomainPay", DomainPay, 3},
		{"DomainNull", DomainNull, 4},
		{"DomainCredit", DomainCredit, 5},
		{"DomainCNull", DomainCNull, 6},
		{"DomainNKHash", DomainNKHash, 7},
		{"DomainDerive", DomainDerive, 8},
		{"DomainRefLink", DomainRefLink, 9},
		{"DomainCreditNKey", DomainCreditNKey, 10},
		{"DomainChangeNKey", DomainChangeNKey, 11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("got %d, want %d", tt.constant, tt.expected)
			}
		})
	}
}

func TestTreeDepth(t *testing.T) {
	if TreeDepth != 16 {
		t.Errorf("TreeDepth = %d, want 16", TreeDepth)
	}
}
