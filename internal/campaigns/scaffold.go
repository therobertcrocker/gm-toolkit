package campaigns

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

const rulebookManifestFile = "rulebook.toml"

// rulebookManifest is the on-disk form of a rulebook folder's rulebook.toml.
type rulebookManifest struct {
	Includes []string `toml:"includes"`
}

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

// ComposeRulebook copies every .toml file under srcDir, and under each folder
// that srcDir's rulebook.toml names in includes, into <campaign-root>/rulebook/
// at the matching relative path. The manifest itself is not copied. Every check
// runs before any write, so a refused compose copies nothing: a file that two
// folders both supply errors with ErrRulebookDuplicate whatever replace says,
// and with replace=false an existing destination errors with
// ErrRulebookConflict. Returns the count of files copied.
func ComposeRulebook(c *Campaign, srcDir string, replace bool) (int, error) {
	folders, err := composeFolders(srcDir)
	if err != nil {
		return 0, err
	}
	destBase := filepath.Join(c.Root, "rulebook")

	suppliedBy := make(map[string]string)
	var duplicates, conflicts []string
	var toCopy []struct{ src, dest string }

	for _, folder := range folders {
		err := filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(strings.ToLower(d.Name()), ".toml") {
				return nil
			}
			rel, relErr := filepath.Rel(folder, path)
			if relErr != nil {
				return relErr
			}
			// deliberate: every folder's manifest is skipped, not only srcDir's, because includes are one level deep (A.001.4 D-1) and a manifest is never campaign data.
			if rel == rulebookManifestFile {
				return nil
			}
			if first, supplied := suppliedBy[rel]; supplied {
				duplicates = append(duplicates, fmt.Sprintf("%s (%s and %s)", rel, first, folder))
				return nil
			}
			suppliedBy[rel] = folder
			dest := filepath.Join(destBase, rel)
			if _, statErr := os.Stat(dest); statErr == nil {
				conflicts = append(conflicts, rel)
			}
			toCopy = append(toCopy, struct{ src, dest string }{path, dest})
			return nil
		})
		if err != nil {
			return 0, fmt.Errorf("walk %s: %w", folder, err)
		}
	}

	if len(duplicates) > 0 {
		return 0, fmt.Errorf("%w: %s", ErrRulebookDuplicate, strings.Join(duplicates, ", "))
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

// composeFolders returns srcDir followed by each folder its rulebook.toml
// includes, resolved from srcDir. With no manifest, srcDir is a plain folder.
func composeFolders(srcDir string) ([]string, error) {
	manifestPath := filepath.Join(srcDir, rulebookManifestFile)
	var manifest rulebookManifest
	metadata, err := toml.DecodeFile(manifestPath, &manifest)
	if errors.Is(err, fs.ErrNotExist) {
		return []string{srcDir}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", manifestPath, err)
	}
	if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, len(undecoded))
		for i, key := range undecoded {
			keys[i] = key.String()
		}
		return nil, fmt.Errorf("unknown keys in %s: %s", manifestPath, strings.Join(keys, ", "))
	}
	folders := []string{srcDir}
	for _, include := range manifest.Includes {
		folder := filepath.Join(srcDir, include)
		info, statErr := os.Stat(folder)
		if statErr != nil {
			return nil, fmt.Errorf("include %q in %s: %w", include, manifestPath, statErr)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("include %q in %s: %s is not a directory", include, manifestPath, folder)
		}
		folders = append(folders, folder)
	}
	return folders, nil
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
