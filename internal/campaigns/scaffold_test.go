package campaigns_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
)

func TestScaffoldCreatesTree(t *testing.T) {
	root := t.TempDir()
	camp, err := campaigns.Scaffold(root, "test-camp", "Test Campaign")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	if camp.ID() != "test-camp" || camp.Name() != "Test Campaign" {
		t.Errorf("unexpected campaign: id=%s name=%s", camp.ID(), camp.Name())
	}
	for _, dir := range []string{
		filepath.Join(root, "rulebook", "assets"),
		filepath.Join(root, "spatial"),
		filepath.Join(root, "state", "logs"),
	} {
		if _, err := os.Stat(dir); err != nil {
			t.Errorf("expected dir %s: %v", dir, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "campaign.toml")); err != nil {
		t.Errorf("expected campaign.toml: %v", err)
	}
}

func TestScaffoldRejectsNonEmpty(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "existing.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	_, err := campaigns.Scaffold(root, "camp", "Camp")
	if err == nil {
		t.Fatal("expected error for non-empty root, got nil")
	}
}

func makeTOMLFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, content := range files {
		dest := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			t.Fatalf("fixture mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
			t.Fatalf("fixture write %s: %v", rel, err)
		}
	}
	return dir
}

func TestComposeRulebookCopiesFiles(t *testing.T) {
	srcDir := makeTOMLFixture(t, map[string]string{
		"goals.toml":          "[goals]\n",
		"assets/cunning.toml": "[cunning]\n",
	})
	root := t.TempDir()
	camp, err := campaigns.Scaffold(root, "copy-test", "Copy Test")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	n, err := campaigns.ComposeRulebook(camp, srcDir, false)
	if err != nil {
		t.Fatalf("ComposeRulebook: %v", err)
	}
	if n != 2 {
		t.Errorf("copied %d files, want 2", n)
	}
	if _, err := os.Stat(filepath.Join(root, "rulebook", "goals.toml")); err != nil {
		t.Errorf("goals.toml not copied: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "rulebook", "assets", "cunning.toml")); err != nil {
		t.Errorf("assets/cunning.toml not copied: %v", err)
	}
}

func TestComposeRulebookConflictNoReplace(t *testing.T) {
	srcDir := makeTOMLFixture(t, map[string]string{"goals.toml": "[goals]\n"})
	root := t.TempDir()
	camp, err := campaigns.Scaffold(root, "conflict-test", "Conflict Test")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	if _, err := campaigns.ComposeRulebook(camp, srcDir, false); err != nil {
		t.Fatalf("first ComposeRulebook: %v", err)
	}
	_, err = campaigns.ComposeRulebook(camp, srcDir, false)
	if !errors.Is(err, campaigns.ErrRulebookConflict) {
		t.Errorf("expected ErrRulebookConflict, got %v", err)
	}
}

func TestComposeRulebookReplaceOverwrites(t *testing.T) {
	srcDir := makeTOMLFixture(t, map[string]string{"goals.toml": "[goals]\n"})
	root := t.TempDir()
	camp, err := campaigns.Scaffold(root, "replace-test", "Replace Test")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	if _, err := campaigns.ComposeRulebook(camp, srcDir, false); err != nil {
		t.Fatalf("first ComposeRulebook: %v", err)
	}
	n, err := campaigns.ComposeRulebook(camp, srcDir, true)
	if err != nil {
		t.Fatalf("ComposeRulebook with replace=true: %v", err)
	}
	if n != 1 {
		t.Errorf("copied %d files, want 1", n)
	}
}

func TestComposeRulebookCopiesIncludes(t *testing.T) {
	fixture := makeTOMLFixture(t, map[string]string{
		"swn/rulebook.toml": "includes = [\"../_core\"]\n",
		"swn/tags.toml":     "[tags]\n",
		"_core/goals.toml":  "[goals]\n",
	})
	root := t.TempDir()
	camp, err := campaigns.Scaffold(root, "include-test", "Include Test")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	n, err := campaigns.ComposeRulebook(camp, filepath.Join(fixture, "swn"), false)
	if err != nil {
		t.Fatalf("ComposeRulebook: %v", err)
	}
	if n != 2 {
		t.Errorf("copied %d files, want 2", n)
	}
	for _, name := range []string{"tags.toml", "goals.toml"} {
		if _, err := os.Stat(filepath.Join(root, "rulebook", name)); err != nil {
			t.Errorf("%s not copied: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "rulebook", "rulebook.toml")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("rulebook.toml copied into the campaign, want it left out (stat err = %v)", err)
	}
}

func TestComposeRulebookRefusesDuplicateEvenWithReplace(t *testing.T) {
	fixture := makeTOMLFixture(t, map[string]string{
		"swn/rulebook.toml": "includes = [\"../_core\"]\n",
		"swn/tags.toml":     "[tags]\n",
		"swn/goals.toml":    "[goals]\n",
		"_core/goals.toml":  "[goals]\n",
	})
	root := t.TempDir()
	camp, err := campaigns.Scaffold(root, "duplicate-test", "Duplicate Test")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	_, err = campaigns.ComposeRulebook(camp, filepath.Join(fixture, "swn"), true)
	if !errors.Is(err, campaigns.ErrRulebookDuplicate) {
		t.Fatalf("expected ErrRulebookDuplicate, got %v", err)
	}
	if !errors.Is(err, campaigns.ErrRulebookConflict) {
		t.Errorf("ErrRulebookDuplicate does not wrap ErrRulebookConflict: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "rulebook", "tags.toml")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("tags.toml written by a refused compose (stat err = %v)", err)
	}
}

func TestComposeRulebookRefusesMissingInclude(t *testing.T) {
	fixture := makeTOMLFixture(t, map[string]string{
		"swn/rulebook.toml": "includes = [\"../missing\"]\n",
		"swn/tags.toml":     "[tags]\n",
	})
	root := t.TempDir()
	camp, err := campaigns.Scaffold(root, "missing-test", "Missing Test")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	_, err = campaigns.ComposeRulebook(camp, filepath.Join(fixture, "swn"), false)
	if err == nil || !strings.Contains(err.Error(), `include "../missing"`) {
		t.Fatalf("err = %v, want a missing-include error", err)
	}
	if _, err := os.Stat(filepath.Join(root, "rulebook", "tags.toml")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("tags.toml written by a refused compose (stat err = %v)", err)
	}
}

func TestComposeRulebookRefusesUnknownManifestKey(t *testing.T) {
	fixture := makeTOMLFixture(t, map[string]string{
		"swn/rulebook.toml": "include = [\"../_core\"]\n",
		"swn/tags.toml":     "[tags]\n",
		"_core/goals.toml":  "[goals]\n",
	})
	root := t.TempDir()
	camp, err := campaigns.Scaffold(root, "manifest-test", "Manifest Test")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	_, err = campaigns.ComposeRulebook(camp, filepath.Join(fixture, "swn"), false)
	if err == nil || !strings.Contains(err.Error(), "unknown keys in") {
		t.Fatalf("err = %v, want an unknown-key error", err)
	}
}
