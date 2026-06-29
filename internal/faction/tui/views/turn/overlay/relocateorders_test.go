package overlay

import "testing"

func TestCapValidator(t *testing.T) {
	t.Run("clearing a pick (empty value) always passes", func(t *testing.T) {
		data := &relocateData{dests: []string{"w1", ""}}
		if err := capValidator(data, 1, 1)(""); err != nil {
			t.Fatalf("empty pick should pass, got %v", err)
		}
	})

	t.Run("pick at the cap passes", func(t *testing.T) {
		data := &relocateData{dests: []string{"w1", "", ""}}
		if err := capValidator(data, 0, 1)("w1"); err != nil {
			t.Fatalf("single pick under cap 1 should pass, got %v", err)
		}
	})

	t.Run("pick that breaches the cap is rejected", func(t *testing.T) {
		// dest[0] already chosen; choosing dest[1] would make two against a cap of 1.
		data := &relocateData{dests: []string{"w1", "w2"}}
		if err := capValidator(data, 1, 1)("w2"); err == nil {
			t.Fatal("second pick against cap 1 should be rejected")
		}
	})

	t.Run("counts the current index once, not twice", func(t *testing.T) {
		// data.dests[0] is already bound to the incoming value; the validator must
		// not double-count it against the cap.
		data := &relocateData{dests: []string{"w1", "w2"}}
		if err := capValidator(data, 0, 2)("w1"); err != nil {
			t.Fatalf("two picks against cap 2 should pass, got %v", err)
		}
	})
}
