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
	// Turn -> Quit (Spatial skipped)
	m = m.Next()
	if m.Active() != ModeQuit {
		t.Fatalf("after Next x2: got %v, want %v", m.Active(), ModeQuit)
	}
	// Quit -> Manage (wrap)
	m = m.Next()
	if m.Active() != ModeManage {
		t.Fatalf("after Next x3: got %v, want %v", m.Active(), ModeManage)
	}
}

func TestPrevWrapsAndSkipsSpatial(t *testing.T) {
	m := New()
	// Manage -> Quit (wrap)
	m = m.Prev()
	if m.Active() != ModeQuit {
		t.Fatalf("after Prev: got %v, want %v", m.Active(), ModeQuit)
	}
	// Quit -> Turn (Spatial skipped)
	m = m.Prev()
	if m.Active() != ModeTurn {
		t.Fatalf("after Prev x2: got %v, want %v", m.Active(), ModeTurn)
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
