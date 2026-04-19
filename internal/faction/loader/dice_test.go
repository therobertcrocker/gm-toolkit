package loader

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func TestParseDice(t *testing.T) {
	tests := []struct {
		input   string
		want    *domain.DiceRoll
		wantErr bool
	}{
		{"1d6", &domain.DiceRoll{NumDice: 1, Sides: 6, Modifier: 0}, false},
		{"2d8", &domain.DiceRoll{NumDice: 2, Sides: 8, Modifier: 0}, false},
		{"1d3+1", &domain.DiceRoll{NumDice: 1, Sides: 3, Modifier: 1}, false},
		{"2d6+2", &domain.DiceRoll{NumDice: 2, Sides: 6, Modifier: 2}, false},
		{"1d6-1", &domain.DiceRoll{NumDice: 1, Sides: 6, Modifier: -1}, false},
		{"", nil, false},
		{"invalid", nil, true},
		{"d6", nil, true},
		{"1d", nil, true},
		{"1d6+", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseDice(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseDice(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseDice(%q) = %+v, want nil", tt.input, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("parseDice(%q) = nil, want %+v", tt.input, tt.want)
			}
			if *got != *tt.want {
				t.Errorf("parseDice(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}
