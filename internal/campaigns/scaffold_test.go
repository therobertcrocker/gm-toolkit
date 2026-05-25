package campaigns_test

import (
	"errors"
	"os"
	"path/filepath"
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

func TestCopyRulebookCopiesFiles(t *testing.T) {
	srcDir := makeTOMLFixture(t, map[string]string{
		"goals.toml":          "[goals]\n",
		"assets/cunning.toml": "[cunning]\n",
	})
	root := t.TempDir()
	camp, err := campaigns.Scaffold(root, "copy-test", "Copy Test")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	n, err := campaigns.CopyRulebook(camp, srcDir, false)
	if err != nil {
		t.Fatalf("CopyRulebook: %v", err)
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

func TestCopyRulebookConflictNoReplace(t *testing.T) {
	srcDir := makeTOMLFixture(t, map[string]string{"goals.toml": "[goals]\n"})
	root := t.TempDir()
	camp, err := campaigns.Scaffold(root, "conflict-test", "Conflict Test")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	if _, err := campaigns.CopyRulebook(camp, srcDir, false); err != nil {
		t.Fatalf("first CopyRulebook: %v", err)
	}
	_, err = campaigns.CopyRulebook(camp, srcDir, false)
	if !errors.Is(err, campaigns.ErrRulebookConflict) {
		t.Errorf("expected ErrRulebookConflict, got %v", err)
	}
}

func TestCopyRulebookReplaceOverwrites(t *testing.T) {
	srcDir := makeTOMLFixture(t, map[string]string{"goals.toml": "[goals]\n"})
	root := t.TempDir()
	camp, err := campaigns.Scaffold(root, "replace-test", "Replace Test")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	if _, err := campaigns.CopyRulebook(camp, srcDir, false); err != nil {
		t.Fatalf("first CopyRulebook: %v", err)
	}
	n, err := campaigns.CopyRulebook(camp, srcDir, true)
	if err != nil {
		t.Fatalf("CopyRulebook with replace=true: %v", err)
	}
	if n != 1 {
		t.Errorf("copied %d files, want 1", n)
	}
}
