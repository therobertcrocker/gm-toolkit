package creature

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrCreatureNotFound  = errors.New("creature not found")
	ErrCreatureAmbiguous = errors.New("creature found in more than one folder")
	ErrBadCreatureID     = errors.New("bad creature id")
)

func Locate(partyDir, encountersDir, id string) (string, error) {
	if !IsFolderName(id) {
		return "", fmt.Errorf("%w: %q", ErrBadCreatureID, id)
	}
	candidates := []string{filepath.Join(partyDir, id)}
	fights, err := os.ReadDir(encountersDir)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("locate %s: %w", id, err)
	}
	for _, fight := range fights {
		if fight.IsDir() {
			candidates = append(candidates, filepath.Join(encountersDir, fight.Name(), id))
		}
	}
	var found []string
	for _, candidate := range candidates {
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			found = append(found, candidate)
		}
	}
	switch len(found) {
	case 0:
		return "", fmt.Errorf("%w: %s (looked in %s and %s/*)", ErrCreatureNotFound, id, partyDir, encountersDir)
	case 1:
		return found[0], nil
	default:
		return "", fmt.Errorf("%w: %s: %s", ErrCreatureAmbiguous, id, strings.Join(found, ", "))
	}
}
