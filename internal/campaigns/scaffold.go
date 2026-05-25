package campaigns

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// Scaffold creates a fresh campaign directory tree at root and writes a
// campaign.toml manifest. Directory layout is derived from Paths() so the two
// stay in sync automatically. Errors if root exists and is non-empty.
func Scaffold(root, id, name string) (*Campaign, error) {
	if entries, err := os.ReadDir(root); err == nil && len(entries) > 0 {
		return nil, fmt.Errorf("scaffold: %s is not empty", root)
	} else if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("scaffold: stat %s: %w", root, err)
	}
	camp := &Campaign{Manifest: Manifest{Campaign: ManifestCampaign{ID: id, Name: name}}, Root: root}
	paths := camp.Paths()
	pv := reflect.ValueOf(paths)
	pt := pv.Type()
	for i := range pt.NumField() {
		field := pt.Field(i)
		full := pv.Field(i).String()
		switch field.Tag.Get("kind") {
		case "dir":
			if err := os.MkdirAll(full, 0o755); err != nil {
				return nil, fmt.Errorf("scaffold: create %s: %w", full, err)
			}
		case "file":
			if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
				return nil, fmt.Errorf("scaffold: create parent for %s: %w", full, err)
			}
		}
	}
	if err := SaveManifest(root, camp.Manifest); err != nil {
		return nil, err
	}
	return camp, nil
}

// CopyRulebook copies every .toml file under srcDir into <campaign-root>/rulebook/
// at the matching relative path. With replace=false, errors with ErrRulebookConflict
// (wrapped, message listing all conflicts) if any destination already exists.
// Returns the count of files copied.
func CopyRulebook(c *Campaign, srcDir string, replace bool) (int, error) {
	destBase := filepath.Join(c.Root, "rulebook")

	var conflicts []string
	var toCopy []struct{ src, dest string }

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".toml") {
			return nil
		}
		rel, relErr := filepath.Rel(srcDir, path)
		if relErr != nil {
			return relErr
		}
		dest := filepath.Join(destBase, rel)
		if _, statErr := os.Stat(dest); statErr == nil {
			conflicts = append(conflicts, rel)
		}
		toCopy = append(toCopy, struct{ src, dest string }{path, dest})
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("walk %s: %w", srcDir, err)
	}

	if len(conflicts) > 0 && !replace {
		return 0, fmt.Errorf("%w in %s: %s", ErrRulebookConflict, destBase, strings.Join(conflicts, ", "))
	}

	copied := 0
	for _, p := range toCopy {
		if err := copyFile(p.src, p.dest); err != nil {
			return copied, fmt.Errorf("copy %s -> %s: %w", p.src, p.dest, err)
		}
		copied++
	}
	return copied, nil
}

func copyFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}
