package foundry

import (
	"maps"
	"path/filepath"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

type gateStrike struct {
	id     string
	bonus  int
	damage string
	kind   string
}

type gateCase struct {
	pc           string
	attributes   map[string]int
	hp, ac       int
	saves        [3]int
	perception   int
	speed        int
	classDC      int
	spellcasting *creature.Spellcasting
	strikes      []gateStrike
	resist       map[string]int
}

var gateOne = []gateCase{
	{
		pc:         "kas",
		attributes: map[string]int{"str": 5, "dex": 0, "con": 4, "int": 0, "wis": 3, "cha": 4},
		hp:         230, ac: 36, saves: [3]int{26, 20, 23}, perception: 23, speed: 30, classDC: 33,
		strikes: []gateStrike{{"protectors-fury", 29, "3d12+9", "bludgeoning"}},
		resist:  map[string]int{"slashing": 4},
	},
	{
		pc:         "tavish",
		attributes: map[string]int{"str": 0, "dex": 4, "con": 3, "int": 4, "wis": 2, "cha": 3},
		hp:         164, ac: 35, saves: [3]int{24, 23, 21}, perception: 20, speed: 30, classDC: 32,
		strikes: []gateStrike{
			{"dagger", 22, "1d4+2", "piercing"},
			{"sling", 22, "1d6+2", "bludgeoning"},
			{"alchemists-fire-greater", 24, "3d8+2", "fire"},
			{"bottled-lightning-greater", 24, "3d6+2", "electricity"},
			{"tenderizer-grenade-greater", 24, "3d6+2", "acid"},
			{"nail-bomb-greater", 24, "6d4+2", "piercing"},
			{"star-grenade-greater", 24, "3d8+2", "fire"},
			{"sticky-algae-bomb-greater", 24, "3d8+2", "poison"},
			{"inflammation-flask-greater", 24, "3d6+2", "acid"},
			{"lodestone-bomb", 24, "3d4+2", "force"},
			{"crystal-shards-moderate", 23, "2d4+2", "piercing"},
		},
		resist: map[string]int{"poison": 7, "electricity": 5},
	},
	{
		pc:           "arabesque",
		attributes:   map[string]int{"str": 0, "dex": 4, "con": 1, "int": 4, "wis": 1, "cha": 5},
		hp:           104, ac: 34, saves: [3]int{21, 24, 21}, perception: 19, speed: 30,
		spellcasting: &creature.Spellcasting{Attack: 23, DC: 33, TopRank: 7},
		strikes:      []gateStrike{{"blink-blade", 24, "2d4+2", "piercing"}},
		resist:       map[string]int{},
	},
}

func TestGateOneStatParity(t *testing.T) {
	for _, want := range gateOne {
		t.Run(want.pc, func(t *testing.T) {
			layer, _ := importTestPC(t, want.pc)
			directory := filepath.Join(t.TempDir(), want.pc)
			if err := writeLayer(directory, "actors/"+want.pc+".json", layer); err != nil {
				t.Fatalf("writeLayer: %v", err)
			}
			loaded, err := creature.Load(directory)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if !maps.Equal(loaded.Attributes, want.attributes) {
				t.Errorf("attributes = %v, want %v", loaded.Attributes, want.attributes)
			}
			defense := loaded.Defense
			for _, check := range []struct {
				name      string
				got, want int
			}{
				{"hp", defense.HP, want.hp},
				{"ac", defense.AC, want.ac},
				{"fortitude", defense.Fortitude, want.saves[0]},
				{"reflex", defense.Reflex, want.saves[1]},
				{"will", defense.Will, want.saves[2]},
				{"perception", loaded.Perception, want.perception},
				{"land speed", loaded.Speed.Land, want.speed},
			} {
				if check.got != check.want {
					t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
				}
			}
			if want.classDC > 0 && loaded.ClassDC != want.classDC {
				t.Errorf("class DC = %d, want %d", loaded.ClassDC, want.classDC)
			}
			switch {
			case want.spellcasting == nil && loaded.Spellcasting != nil:
				t.Errorf("spellcasting = %+v, want none", *loaded.Spellcasting)
			case want.spellcasting != nil && (loaded.Spellcasting == nil || *loaded.Spellcasting != *want.spellcasting):
				t.Errorf("spellcasting = %+v, want %+v", loaded.Spellcasting, *want.spellcasting)
			}
			for _, strike := range want.strikes {
				got, found := loaded.Strikes[strike.id]
				if !found {
					t.Errorf("no Strike %s", strike.id)
					continue
				}
				part := got.Damage[0]
				if got.Bonus != strike.bonus || part.Roll.String() != strike.damage || part.Type != strike.kind {
					t.Errorf("%s = %+d %s %s, want %+d %s %s", strike.id, got.Bonus, part.Roll, part.Type, strike.bonus, strike.damage, strike.kind)
				}
			}
			resist := map[string]int{}
			for kind, adjustment := range defense.Resist {
				resist[kind] = adjustment.Value
			}
			if !maps.Equal(resist, want.resist) {
				t.Errorf("resistances = %v, want %v", resist, want.resist)
			}
		})
	}
}
