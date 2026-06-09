package msgs

// StartCycleMsg requests the turn router to build a fresh adapter and begin a
// cycle. Emitted by the setup view on the Start key; handled by turn.Model.
type StartCycleMsg struct{}

// ReturnToSetupMsg requests the router to discard a finished cycle and return
// to the setup roster. Emitted by the execution view on a keypress after a
// fatal cycle error — the one terminal state with no automatic return (the
// happy path returns via StreamClosedMsg). Handled by turn.Model.
type ReturnToSetupMsg struct{}
