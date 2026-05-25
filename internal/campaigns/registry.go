package campaigns

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Registry is the in-memory form of $GM_TOOLKIT_HOME/campaigns.toml.
type Registry struct {
	Active    string          `toml:"active,omitempty"`
	Campaigns []RegistryEntry `toml:"campaigns"`
}

type RegistryEntry struct {
	ID   string `toml:"id"`
	Path string `toml:"path"`
}

// RegistryPath returns $GM_TOOLKIT_HOME/campaigns.toml. Errors if
// GM_TOOLKIT_HOME is not set.
func RegistryPath() (string, error) {
	home := os.Getenv("GM_TOOLKIT_HOME")
	if home == "" {
		return "", fmt.Errorf("GM_TOOLKIT_HOME is not set; export GM_TOOLKIT_HOME=<your-campaigns-dir> and try again")
	}
	return filepath.Join(home, "campaigns-registry.toml"), nil
}

// LoadRegistry reads the registry file. Returns an empty Registry if the file is missing.
func LoadRegistry() (*Registry, error) {
	path, err := RegistryPath()
	if err != nil {
		return nil, err
	}
	var reg Registry
	if _, err := toml.DecodeFile(path, &reg); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Registry{}, nil
		}
		return nil, fmt.Errorf("load registry %s: %w", path, err)
	}
	return &reg, nil
}

// SaveRegistry writes the registry, creating $GM_TOOLKIT_HOME if needed.
func SaveRegistry(reg *Registry) error {
	path, err := RegistryPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create registry dir: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create registry %s: %w", path, err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(reg); err != nil {
		return fmt.Errorf("encode registry %s: %w", path, err)
	}
	return nil
}

// Register adds a campaign to the registry. Returns ErrCampaignAlreadyExists
// if the id is already present.
func (r *Registry) Register(entry RegistryEntry) error {
	for _, e := range r.Campaigns {
		if e.ID == entry.ID {
			return fmt.Errorf("%w: %s", ErrCampaignAlreadyExists, entry.ID)
		}
	}
	r.Campaigns = append(r.Campaigns, entry)
	return nil
}

// SetActive sets the active campaign id. Returns ErrCampaignNotRegistered if
// the id is not present in the registry.
func (r *Registry) SetActive(id string) error {
	if _, ok := r.Lookup(id); !ok {
		return fmt.Errorf("%w: %s", ErrCampaignNotRegistered, id)
	}
	r.Active = id
	return nil
}

// Lookup returns the entry for id, or false if not registered.
func (r *Registry) Lookup(id string) (RegistryEntry, bool) {
	for _, e := range r.Campaigns {
		if e.ID == id {
			return e, true
		}
	}
	return RegistryEntry{}, false
}

// ResolveActive returns the campaign indicated by override (if non-empty) or
// by the registry's Active pointer. Errors:
//   - both empty: ErrNoActiveCampaign
//   - id not in registry: ErrCampaignNotRegistered
//   - registered path missing on disk: ErrCampaignRootMissing
func ResolveActive(reg *Registry, override string) (*Campaign, error) {
	id := override
	if id == "" {
		id = reg.Active
	}
	if id == "" {
		return nil, ErrNoActiveCampaign
	}
	entry, ok := reg.Lookup(id)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrCampaignNotRegistered, id)
	}
	if _, err := os.Stat(entry.Path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s (%s)", ErrCampaignRootMissing, id, entry.Path)
		}
		return nil, fmt.Errorf("stat campaign root %s: %w", entry.Path, err)
	}
	return LoadCampaign(entry.Path)
}
