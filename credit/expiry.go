package credit

import "errors"

// ErrCreditExecutionExpired means a local, unsubmitted statement or quote needs
// fresh authorization and proof. It never resolves an earlier unknown send.
var ErrCreditExecutionExpired = errors.New("Credit operation or broadcaster quote expired; reauthorize and regenerate proof")
