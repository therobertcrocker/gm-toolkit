package engine

import "github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"

// Action is a type alias for action.Action so that existing callers of
// the top-level engine package continue to compile unchanged through the
// refactor. Remove in commit 8 when the top-level package is collapsed.
type Action = action.Action
