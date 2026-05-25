package campaigns_test

import (
	"path/filepath"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
)

func TestCampaignPaths(t *testing.T) {
	root := "/some/root"
	camp := &campaigns.Campaign{
		Manifest: campaigns.Manifest{Campaign: campaigns.ManifestCampaign{ID: "test", Name: "Test"}},
		Root:     root,
	}
	paths := camp.Paths()
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"FactionDataDir", paths.FactionDataDir, filepath.Join(root, "rulebook")},
		{"SpatialDataDir", paths.SpatialDataDir, filepath.Join(root, "spatial")},
		{"StatePath", paths.StatePath, filepath.Join(root, "state", "state.toml")},
		{"HistoryPath", paths.HistoryPath, filepath.Join(root, "state", "history.jsonl")},
		{"NarrativesDir", paths.NarrativesDir, filepath.Join(root, "narratives")},
		{"LogsDir", paths.LogsDir, filepath.Join(root, "state", "logs")},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s = %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}

func TestManifestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	m := campaigns.Manifest{Campaign: campaigns.ManifestCampaign{ID: "my-camp", Name: "My Campaign"}}
	if err := campaigns.SaveManifest(dir, m); err != nil {
		t.Fatalf("SaveManifest: %v", err)
	}
	got, err := campaigns.LoadManifest(dir)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if got.Campaign.ID != m.Campaign.ID || got.Campaign.Name != m.Campaign.Name {
		t.Errorf("got %+v, want %+v", got.Campaign, m.Campaign)
	}
}

func TestLoadManifestMissing(t *testing.T) {
	dir := t.TempDir()
	_, err := campaigns.LoadManifest(dir)
	if err == nil {
		t.Fatal("expected error for missing campaign.toml, got nil")
	}
}

func TestValidateID(t *testing.T) {
	valid := []string{"my-camp", "camp1", "a", "foo-bar-baz"}
	for _, id := range valid {
		if err := campaigns.ValidateID(id); err != nil {
			t.Errorf("ValidateID(%q) = %v, want nil", id, err)
		}
	}

	invalid := []string{"", "-start", "end-", "has space", "CamelCase", "under_score"}
	for _, id := range invalid {
		if err := campaigns.ValidateID(id); err == nil {
			t.Errorf("ValidateID(%q) = nil, want error", id)
		}
	}
}

func TestLoadManifestEmptyID(t *testing.T) {
	dir := t.TempDir()
	m := campaigns.Manifest{Campaign: campaigns.ManifestCampaign{ID: "", Name: "No ID"}}
	if err := campaigns.SaveManifest(dir, m); err != nil {
		t.Fatalf("SaveManifest: %v", err)
	}
	_, err := campaigns.LoadManifest(dir)
	if err == nil {
		t.Fatal("expected error for empty campaign.id, got nil")
	}
}
