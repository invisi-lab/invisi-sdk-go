package broadcaster

import (
	"errors"
	"strings"
)

// SubmissionRefusedError identifies a Broadcaster admission refusal before send.
// It does not resolve an earlier transport attempt whose outcome is unknown.
type SubmissionRefusedError struct {
	Code   int
	Reason string
	cause  error
}

func (e *SubmissionRefusedError) Error() string {
	if e.cause != nil {
		return e.cause.Error()
	}
	return e.Reason
}
func (e *SubmissionRefusedError) Unwrap() error { return e.cause }
func submissionRefusal(err error) error {
	var rpc *RPCError
	if !errors.As(err, &rpc) {
		return err
	}
	expected := map[int]string{-32024: "CREDIT_PERIOD_CHANGED", -32025: "INSUFFICIENT_EXECUTION_SLACK", -32026: "CREDIT_PERIOD_ENDING"}[rpc.Code]
	reason, _, _ := strings.Cut(rpc.Message, ":")
	if expected == "" || reason != expected {
		return err
	}
	return &SubmissionRefusedError{Code: rpc.Code, Reason: reason, cause: err}
}
