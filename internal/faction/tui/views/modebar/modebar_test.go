package modebar

import "testing"

func TestNewActiveIsManage(t *testing.T) {
	if got := New().Active(); got != ModeManage {
		t.Errorf("New().Active() = %v, want %v", got, ModeManage)
	}
}

func TestNextSkipsSpatialAndWraps(t *testing.T) {
	m := New()
	// Manage -> Turn
	m = m.Next()
	if m.Active() != ModeTurn {
		t.Fatalf("after Next: got %v, want %v", m.Active(), ModeTurn)
	}
	// Turn -> Manage (Spatial skipped, wrap)
	m = m.Next()
	if m.Active() != ModeManage {
		t.Fatalf("after Next x2: got %v, want %v", m.Active(), ModeManage)
	}
}

func TestPrevWrapsAndSkipsSpatial(t *testing.T) {
	m := New()
	// Manage -> Turn (wrap, Spatial skipped)
	m = m.Prev()
	if m.Active() != ModeTurn {
		t.Fatalf("after Prev: got %v, want %v", m.Active(), ModeTurn)
	}
	// Turn -> Manage
	m = m.Prev()
	if m.Active() != ModeManage {
		t.Fatalf("after Prev x2: got %v, want %v", m.Active(), ModeManage)
	}
}

func TestSetActiveDisabledIsNoOp(t *testing.T) {
	m := New().SetActive(ModeSpatial)
	if m.Active() != ModeManage {
		t.Errorf("SetActive(ModeSpatial): got %v, want %v (unchanged)", m.Active(), ModeManage)
	}
}

func TestSetActiveEnabledJumps(t *testing.T) {
	m := New().SetActive(ModeTurn)
	if m.Active() != ModeTurn {
		t.Errorf("SetActive(ModeTurn): got %v, want %v", m.Active(), ModeTurn)
	}
}
