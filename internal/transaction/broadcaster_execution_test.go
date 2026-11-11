package transaction

import (
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
	"testing"
)

func TestResolveQuotedExecution(t *testing.T) {
	for _, q := range []string{"direct", "aggregate"} {
		want := "single"
		if q == "aggregate" {
			want = "aggregate"
		}
		o := Execution{Plan: Plan{Broadcaster: &BroadcasterPlan{Signed: broadcaster.SignedQuote{Quote: broadcaster.Quote{Execution: q}}}}}
		got, e := o.resolveExecution("")
		if e != nil || got != want {
			t.Fatalf("%s: %s %v", q, got, e)
		}
		other := "aggregate"
		if want == other {
			other = "single"
		}
		if _, e = o.resolveExecution(other); e == nil {
			t.Fatal("target mismatch accepted")
		}
	}
}
