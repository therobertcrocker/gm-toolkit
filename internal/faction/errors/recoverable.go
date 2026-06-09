package errors

import (
	"errors"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
)

// recoverableErrors enumerates sentinels whose return value is treated as a
// gameplay outcome rather than an engine failure. The orchestrator's
// Recoverable/Fatal branch logs at Warn level and continues past the failing
// sub-engine call for any error matching one of these via errors.Is.
//
// Adding to this list is a deliberate decision — it changes the orchestrator's
// abort semantics for the new error category.
var recoverableErrors = []error{
	action.ErrNoSelection,
	action.ErrTurnCanceled,
	action.ErrActionUnavailable,
}

func IsRecoverable(err error) bool {
	for _, sentinel := range recoverableErrors {
		if errors.Is(err, sentinel) {
			return true
		}
	}
	return false
}
