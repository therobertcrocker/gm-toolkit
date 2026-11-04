package dicetest

// Sequence returns Values in order and panics when they run out, so a test that
// rolls more than it planned fails loudly.
type Sequence struct {
	Values []int
	next   int
}

func (roller *Sequence) Roll(int) int {
	if roller.next >= len(roller.Values) {
		panic("dicetest: Sequence exhausted")
	}
	value := roller.Values[roller.next]
	roller.next++
	return value
}

// Rolled returns how many dice have been rolled.
func (roller *Sequence) Rolled() int { return roller.next }

// Cycle returns Values in order and starts again when they run out.
type Cycle struct {
	Values []int
	next   int
}

func (roller *Cycle) Roll(int) int {
	value := roller.Values[roller.next%len(roller.Values)]
	roller.next++
	return value
}

// Rolled returns how many dice have been rolled.
func (roller *Cycle) Rolled() int { return roller.next }

// Constant returns Value on every roll.
type Constant struct {
	Value int
}

func (roller Constant) Roll(int) int { return roller.Value }
