package foundry

import "testing"

func TestStrikeRange(t *testing.T) {
	cases := []struct {
		traits []string
		want   int
	}{
		{[]string{"agile", "finesse"}, 0},
		{[]string{"range-increment-60", "deadly-d10"}, 60},
		{[]string{"range-30"}, 30},
	}
	for _, testCase := range cases {
		if got := strikeRange(testCase.traits); got != testCase.want {
			t.Errorf("strikeRange(%v) = %d, want %d", testCase.traits, got, testCase.want)
		}
	}
}

func TestPersistentDamageDice(t *testing.T) {
	six := 6
	if got := (persistentDamage{Number: 1, Faces: &six, Type: "bleed"}).dice(); got != "1d6" {
		t.Errorf("with faces: %q, want 1d6", got)
	}
	if got := (persistentDamage{Number: 2, Type: "fire"}).dice(); got != "2" {
		t.Errorf("without faces: %q, want 2", got)
	}
}
