package scenarios

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
)

// composedRulebookDir composes rulebooks/swn/ and the folders it includes into a
// temp campaign and returns that campaign's rulebook dir. The source is resolved
// from this file so it works regardless of the working directory when tests run.
func composedRulebookDir(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "..")
	camp := &campaigns.Campaign{Root: t.TempDir()}
	if _, err := campaigns.ComposeRulebook(camp, filepath.Join(repoRoot, "rulebooks", "swn"), false); err != nil {
		t.Fatalf("compose rulebook: %v", err)
	}
	return camp.Paths().FactionDataDir
}
