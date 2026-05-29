package msgs

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

// Completion messages — emitted by sub-views, handled centrally by manage.Update.
type CreatedMsg struct{ Faction *domain.Faction }
type DeletedMsg struct{ ID string }
type CancelMsg struct{}
type SaveErrorMsg struct{ Err error }

// Transition messages — sub-views emit these to request a view change.
type RequestDetailMsg struct{ FactionID string }
type RequestCreateMsg struct{}
type RequestDeleteMsg struct{ Faction *domain.Faction }
