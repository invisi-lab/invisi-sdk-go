package transaction

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestBroadcasterPlanFormattingRedactsCancellationSecret(t *testing.T) {
	secret := strings.Repeat("55", 32)
	plan := BroadcasterPlan{CancelToken: "0x" + secret}
	raw, err := json.Marshal(plan)
	if err != nil || !strings.Contains(string(raw), secret) {
		t.Fatal("durable JSON lost cancellation handle", err)
	}
	for _, format := range []string{"%v", "%+v", "%#v"} {
		if strings.Contains(fmt.Sprintf(format, plan), secret) || strings.Contains(fmt.Sprintf(format, &plan), secret) {
			t.Fatal("format exposed cancellation secret", format)
		}
	}
}
