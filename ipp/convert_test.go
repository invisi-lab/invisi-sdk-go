package ipp

import (
	"reflect"
	"testing"
)

func TestPaymentTypesFromStrings(t *testing.T) {
	t.Run("known values", func(t *testing.T) {
		got := PaymentTypesFromStrings([]string{"open", "blind", "credit"})
		want := []PaymentType{PaymentTypeOpen, PaymentTypeBlind, PaymentTypeCredit}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("empty input yields empty slice", func(t *testing.T) {
		if got := PaymentTypesFromStrings(nil); len(got) != 0 {
			t.Errorf("got %v, want len 0", got)
		}
	})

	// The cast is parse-don't-validate: it must NOT drop or reject unknown
	// values — legality is the caller's job via PaymentType.Valid().
	t.Run("preserves unknown values without validating", func(t *testing.T) {
		got := PaymentTypesFromStrings([]string{"open", "bogus"})
		want := []PaymentType{PaymentTypeOpen, PaymentType("bogus")}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestPaymentTypesToStrings(t *testing.T) {
	t.Run("typed values to strings", func(t *testing.T) {
		got := PaymentTypesToStrings([]PaymentType{PaymentTypeOpen, PaymentTypeBlind})
		want := []string{"open", "blind"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("empty input yields empty slice", func(t *testing.T) {
		if got := PaymentTypesToStrings(nil); len(got) != 0 {
			t.Errorf("got %v, want len 0", got)
		}
	})
}

func TestPaymentTypesRoundTrip(t *testing.T) {
	got := PaymentTypesFromStrings(PaymentTypesToStrings(AllPaymentTypes))
	if !reflect.DeepEqual(got, AllPaymentTypes) {
		t.Errorf("round-trip mismatch: got %v, want %v", got, AllPaymentTypes)
	}
}
