package scenarios

import (
	"path/filepath"
	"runtime"
)

// testDataDir is the rulebooks/swn/ directory, resolved from this source file
// so it works regardless of the working directory when tests run.
var testDataDir = func() string {
	_, file, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "..")
	return filepath.Join(repoRoot, "rulebooks", "swn")
}()
