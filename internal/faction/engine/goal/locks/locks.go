package locks

// LockType classifies the constraint that a faction's active goal imposes on its turn.
type LockType int

const (
	LockNone            LockType = iota // normal flow
	LockSkip                            // skip faction entirely (Change Homeworld in-progress)
	LockRestrictActions                 // limit available actions (Seize Planet combat phase)
)

// GoalLock describes how the active goal constrains this faction's turn.
type GoalLock struct {
	Type           LockType
	AllowedActions []string // non-nil only when Type == LockRestrictActions
}
