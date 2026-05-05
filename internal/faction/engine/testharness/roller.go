package testharness

// FixedRoller returns values from a pre-set sequence in order. It panics when
// exhausted so tests fail loudly rather than silently returning zero.
type FixedRoller struct {
	Values []int
	index  int
}

func (r *FixedRoller) Roll(_ int) int {
	if r.index >= len(r.Values) {
		panic("FixedRoller exhausted")
	}
	v := r.Values[r.index]
	r.index++
	return v
}
