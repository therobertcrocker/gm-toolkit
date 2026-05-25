package campaigns_test

import (
	"errors"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
)

func TestLoadRegistryMissing(t *testing.T) {
	t.Setenv("GM_TOOLKIT_HOME", t.TempDir())
	reg, err := campaigns.LoadRegistry()
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	if reg.Active != "" || len(reg.Campaigns) != 0 {
		t.Errorf("expected empty registry, got %+v", reg)
	}
}

func TestRegistryRoundTrip(t *testing.T) {
	t.Setenv("GM_TOOLKIT_HOME", t.TempDir())
	reg := &campaigns.Registry{
		Active: "alpha",
		Campaigns: []campaigns.RegistryEntry{
			{ID: "alpha", Path: "/some/path/alpha"},
			{ID: "beta", Path: "/some/path/beta"},
		},
	}
	if err := campaigns.SaveRegistry(reg); err != nil {
		t.Fatalf("SaveRegistry: %v", err)
	}
	got, err := campaigns.LoadRegistry()
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	if got.Active != reg.Active {
		t.Errorf("Active = %q, want %q", got.Active, reg.Active)
	}
	if len(got.Campaigns) != 2 {
		t.Fatalf("len(Campaigns) = %d, want 2", len(got.Campaigns))
	}
	if got.Campaigns[0].ID != "alpha" || got.Campaigns[1].ID != "beta" {
		t.Errorf("unexpected entries: %+v", got.Campaigns)
	}
}

func TestRegisterDuplicate(t *testing.T) {
	reg := &campaigns.Registry{}
	entry := campaigns.RegistryEntry{ID: "my-camp", Path: "/some/path"}
	if err := reg.Register(entry); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	err := reg.Register(entry)
	if !errors.Is(err, campaigns.ErrCampaignAlreadyExists) {
		t.Errorf("expected ErrCampaignAlreadyExists, got %v", err)
	}
}

func TestSetActiveUnregistered(t *testing.T) {
	reg := &campaigns.Registry{}
	err := reg.SetActive("ghost")
	if !errors.Is(err, campaigns.ErrCampaignNotRegistered) {
		t.Errorf("expected ErrCampaignNotRegistered, got %v", err)
	}
}

func TestResolveActiveEmptyRegistry(t *testing.T) {
	reg := &campaigns.Registry{}
	_, err := campaigns.ResolveActive(reg, "")
	if !errors.Is(err, campaigns.ErrNoActiveCampaign) {
		t.Errorf("expected ErrNoActiveCampaign, got %v", err)
	}
}

func TestResolveActiveOverride(t *testing.T) {
	root := t.TempDir()
	camp, err := campaigns.Scaffold(root, "override-camp", "Override Campaign")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	reg := &campaigns.Registry{
		Campaigns: []campaigns.RegistryEntry{
			{ID: camp.ID(), Path: camp.Root},
		},
	}
	got, err := campaigns.ResolveActive(reg, camp.ID())
	if err != nil {
		t.Fatalf("ResolveActive: %v", err)
	}
	if got.ID() != camp.ID() {
		t.Errorf("got ID %q, want %q", got.ID(), camp.ID())
	}
}

func TestResolveActiveOverrideNotRegistered(t *testing.T) {
	reg := &campaigns.Registry{}
	_, err := campaigns.ResolveActive(reg, "ghost")
	if !errors.Is(err, campaigns.ErrCampaignNotRegistered) {
		t.Errorf("expected ErrCampaignNotRegistered, got %v", err)
	}
}

func TestResolveActivePathMissing(t *testing.T) {
	reg := &campaigns.Registry{
		Active: "gone",
		Campaigns: []campaigns.RegistryEntry{
			{ID: "gone", Path: "/nonexistent/path/that/does/not/exist"},
		},
	}
	_, err := campaigns.ResolveActive(reg, "")
	if !errors.Is(err, campaigns.ErrCampaignRootMissing) {
		t.Errorf("expected ErrCampaignRootMissing, got %v", err)
	}
}

func TestResolveActiveHappyPath(t *testing.T) {
	root := t.TempDir()
	camp, err := campaigns.Scaffold(root, "happy-camp", "Happy Campaign")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	reg := &campaigns.Registry{
		Active: camp.ID(),
		Campaigns: []campaigns.RegistryEntry{
			{ID: camp.ID(), Path: camp.Root},
		},
	}
	got, err := campaigns.ResolveActive(reg, "")
	if err != nil {
		t.Fatalf("ResolveActive: %v", err)
	}
	if got.ID() != camp.ID() || got.Root != camp.Root {
		t.Errorf("got id=%s root=%s, want id=%s root=%s", got.ID(), got.Root, camp.ID(), camp.Root)
	}
}
