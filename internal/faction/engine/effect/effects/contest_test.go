package effects

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

type pairRoller struct {
	values []int
	next   int
}

func (roller *pairRoller) Roll(int) int {
	value := roller.values[roller.next]
	roller.next++
	return value
}

func TestContestPasses(t *testing.T) {
	declaring := &domain.Faction{ID: "alpha", Cunning: 3}
	target := &domain.Faction{ID: "beta", Cunning: 2}
	cases := []struct {
		name   string
		roller domain.ContestRoller
		rolls  []int
		want   bool
	}{
		{"self: the declaring faction wins", domain.ContestRollerSelf, []int{6, 2}, true},
		{"self: the declaring faction loses", domain.ContestRollerSelf, []int{1, 10}, false},
		{"self: a tie fails", domain.ContestRollerSelf, []int{5, 6}, false},
		{"target: the target fails its test", domain.ContestRollerTarget, []int{1, 10}, true},
		{"target: the target passes its test", domain.ContestRollerTarget, []int{10, 1}, false},
		{"target: a tie passes the effect", domain.ContestRollerTarget, []int{6, 5}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			contest := domain.EffectContest{
				Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning},
				Roller:  tc.roller,
			}
			if got := contestPasses(contest, declaring, target, &pairRoller{values: tc.rolls}); got != tc.want {
				t.Errorf("contestPasses = %v, want %v", got, tc.want)
			}
		})
	}
}
