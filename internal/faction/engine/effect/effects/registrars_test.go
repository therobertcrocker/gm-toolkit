package effects_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
)

func TestRegistrars_DeclareARouteForEveryTrigger(t *testing.T) {
	for name, registrar := range effects.Registrars() {
		if len(registrar.Triggers) == 0 {
			t.Errorf("%s: a registrar must accept at least one trigger", name)
		}
		for trigger, accepts := range registrar.Triggers {
			if len(accepts.Owners) == 0 || len(accepts.Scopes) == 0 {
				t.Errorf("%s: trigger %q must accept at least one owner and scope", name, trigger)
			}
			if effect.IsPhaseTrigger(trigger) && registrar.Fire == nil {
				t.Errorf("%s: phase trigger %q needs Fire", name, trigger)
			}
			if !effect.IsPhaseTrigger(trigger) && registrar.Register == nil {
				t.Errorf("%s: hook trigger %q needs Register", name, trigger)
			}
		}
	}
}
