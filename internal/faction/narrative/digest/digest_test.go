package digest

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func TestBuildStub(t *testing.T) {
	_, err := Build(nil, 1, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil records, got nil")
	}

	oneRecord := []domain.EventRecord{{Cycle: 1, FactionID: "faction-a"}}
	got, err := Build(oneRecord, 1, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Cycle != 1 {
		t.Fatalf("expected Cycle 1, got %d", got.Cycle)
	}
}
