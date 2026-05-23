package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
)

func TestIsRecoverable(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"ErrNoSelection direct", action.ErrNoSelection, true},
		{"ErrNoSelection wrapped", fmt.Errorf("sell asset: %w", action.ErrNoSelection), true},
		{"arbitrary error", errors.New("something else"), false},
		{"nil", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsRecoverable(tc.err); got != tc.want {
				t.Errorf("IsRecoverable(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
