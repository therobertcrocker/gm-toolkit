package overlay

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

func TestSelectAssetToLoseIgnoresEsc(t *testing.T) {
	assets := []*domain.Asset{{ID: "b-1", DefinitionID: "machine"}}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{"machine": {ID: "machine", Name: "Machine"}}}

	updated, cmd := NewSelectAssetToLose("lose which asset?", assets, rb).Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Errorf("Esc produced a command (%T), want none", cmd())
	}
	if _, ok := updated.(SelectAssetToLose); !ok {
		t.Errorf("Update returned %T, want SelectAssetToLose", updated)
	}
}

func TestConfirmPaymentIgnoresEsc(t *testing.T) {
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{"machine": {ID: "machine", Name: "Machine"}}}
	for _, reason := range []hooks.PaymentReason{hooks.PaymentAvoidDisable, hooks.PaymentLiftDisable, hooks.PaymentRestore} {
		offer := hooks.PaymentOffer{
			Payer:  &domain.Faction{ID: "f2", Name: "Beta"},
			Asset:  &domain.Asset{ID: "b-1", DefinitionID: "machine"},
			Amount: 3,
			Reason: reason,
		}

		updated, cmd := NewConfirmPayment(offer, rb).Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cmd != nil {
			t.Errorf("%s: Esc produced a command (%T), want none", reason, cmd())
		}
		if _, ok := updated.(NoCancelConfirm); !ok {
			t.Errorf("%s: Update returned %T, want NoCancelConfirm", reason, updated)
		}
	}
}

func TestConfirmSacrificeIgnoresEsc(t *testing.T) {
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"front":  {ID: "front", Name: "False Front"},
		"target": {ID: "target", Name: "Safehouse"},
	}}
	offer := hooks.SacrificeOffer{
		Owner:     &domain.Faction{ID: "f2", Name: "Beta"},
		Sacrifice: &domain.Asset{ID: "b-1", DefinitionID: "front"},
		Saves:     &domain.Asset{ID: "b-2", DefinitionID: "target"},
	}

	updated, cmd := NewConfirmSacrifice(offer, rb).Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Errorf("Esc produced a command (%T), want none", cmd())
	}
	if _, ok := updated.(NoCancelConfirm); !ok {
		t.Errorf("Update returned %T, want NoCancelConfirm", updated)
	}
}

func TestRequestPermissionIgnoresEsc(t *testing.T) {
	request := hooks.PermissionRequest{
		Requester: &domain.Faction{ID: "f1", Name: "Alpha"},
		Ruler:     &domain.Faction{ID: "f2", Name: "Bravo"},
		WorldID:   "krylos",
		Def:       &domain.AssetDefinition{ID: "gunships", Name: "Gunships"},
	}

	updated, cmd := NewRequestPermission(request, "Krylos").Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Errorf("Esc produced a command (%T), want none", cmd())
	}
	if _, ok := updated.(NoCancelConfirm); !ok {
		t.Errorf("Update returned %T, want NoCancelConfirm", updated)
	}
}

func TestPermissionTitle(t *testing.T) {
	request := hooks.PermissionRequest{
		Requester: &domain.Faction{ID: "f1", Name: "Alpha"},
		WorldID:   "krylos",
		Def:       &domain.AssetDefinition{ID: "troops", Name: "Troops"},
		Cause:     domain.PermissionCauseArrival,
	}
	if got, want := permissionTitle(request, "Krylos"), "Permit Alpha to land Troops on Krylos?"; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
	request.Cause = "buy"
	if got, want := permissionTitle(request, "Krylos"), "Permit Alpha to raise Troops on Krylos?"; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
}

func TestContestTitle(t *testing.T) {
	offer := hooks.PermissionContestOffer{
		Owner:     &domain.Faction{ID: "f2", Name: "Bravo"},
		Source:    &domain.Asset{ID: "lob"},
		Requester: &domain.Faction{ID: "f1", Name: "Alpha"},
		WorldID:   "krylos",
		Def:       &domain.AssetDefinition{ID: "troops", Name: "Troops"},
		Cause:     domain.PermissionCauseRelocate,
	}
	if got, want := contestTitle(offer, "Krylos"), "Contest Alpha's permission to land Troops on Krylos?"; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
}
