package action

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrNoSelection_ErrorsIs(t *testing.T) {
	if !errors.Is(ErrNoSelection, ErrNoSelection) {
		t.Fatal("errors.Is(ErrNoSelection, ErrNoSelection) = false")
	}
	wrapped := fmt.Errorf("sell asset: %w", ErrNoSelection)
	if !errors.Is(wrapped, ErrNoSelection) {
		t.Fatal("errors.Is(wrapped ErrNoSelection, ErrNoSelection) = false")
	}
}
