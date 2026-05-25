package campaigns

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/BurntSushi/toml"
)

var (
	ErrNoActiveCampaign      = errors.New("no active campaign")
	ErrCampaignNotRegistered = errors.New("campaign not registered")
	ErrCampaignRootMissing   = errors.New("campaign root missing on disk")
	ErrCampaignAlreadyExists = errors.New("campaign id already registered")
	ErrRulebookEmpty         = errors.New("rulebook is empty")
	ErrRulebookConflict      = errors.New("rulebook file conflict")
)

// Manifest is the on-disk form of campaign.toml.
type Manifest struct {
	Campaign ManifestCampaign `toml:"campaign"`
}

type ManifestCampaign struct {
	ID   string `toml:"id"`
	Name string `toml:"name"`
}

// Campaign is the in-memory resolved campaign: its manifest plus the root path
// it was loaded from.
type Campaign struct {
	Manifest Manifest
	Root     string
}

func (c *Campaign) ID() string   { return c.Manifest.Campaign.ID }
func (c *Campaign) Name() string { return c.Manifest.Campaign.Name }

// Paths is the runtime path-bundle engine consumers receive. Each field is
// tagged with path (relative to campaign root) and kind (dir or file).
// Scaffold and Paths() both derive their behavior from these tags, so adding
// a new entry means adding one field — no other code changes required.
type Paths struct {
	FactionDataDir string `path:"rulebook"             kind:"dir"`
	AssetsDir      string `path:"rulebook/assets"      kind:"dir"`
	SpatialDataDir string `path:"spatial"              kind:"dir"`
	StatePath      string `path:"state/state.toml"     kind:"file"`
	HistoryPath    string `path:"state/history.jsonl"  kind:"file"`
	NarrativesDir  string `path:"narratives"           kind:"dir"`
	LogsDir        string `path:"state/logs"           kind:"dir"`
}

// Paths derives the runtime path bundle from the campaign root by reflecting
// over the Paths struct tags.
func (c *Campaign) Paths() Paths {
	var paths Paths
	rv := reflect.ValueOf(&paths).Elem()
	rt := rv.Type()
	for i := range rt.NumField() {
		rel := rt.Field(i).Tag.Get("path")
		if rel == "" {
			panic(fmt.Sprintf("campaigns.Paths: field %s missing path tag", rt.Field(i).Name))
		}
		rv.Field(i).SetString(filepath.Join(c.Root, rel))
	}
	return paths
}

// ValidateID checks that a campaign id is non-empty and kebab-case:
// lowercase letters, digits, and hyphens, with no leading or trailing hyphen.
func ValidateID(id string) error {
	if id == "" {
		return errors.New("campaign id must not be empty")
	}
	for i, r := range id {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return fmt.Errorf("campaign id %q: only lowercase letters, digits, and hyphens are allowed", id)
		}
		if r == '-' && (i == 0 || i == len(id)-1) {
			return fmt.Errorf("campaign id %q: hyphens may not appear at the start or end", id)
		}
	}
	return nil
}

// LoadManifest reads <root>/campaign.toml.
func LoadManifest(root string) (Manifest, error) {
	var m Manifest
	path := filepath.Join(root, "campaign.toml")
	if _, err := toml.DecodeFile(path, &m); err != nil {
		return Manifest{}, fmt.Errorf("load manifest %s: %w", path, err)
	}
	if m.Campaign.ID == "" {
		return Manifest{}, fmt.Errorf("load manifest %s: missing campaign.id", path)
	}
	return m, nil
}

// SaveManifest writes <root>/campaign.toml, creating the directory if needed.
func SaveManifest(root string, m Manifest) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("create campaign root %s: %w", root, err)
	}
	path := filepath.Join(root, "campaign.toml")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create manifest %s: %w", path, err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(m); err != nil {
		return fmt.Errorf("encode manifest %s: %w", path, err)
	}
	return nil
}

// LoadCampaign reads <root>/campaign.toml and returns a resolved Campaign.
func LoadCampaign(root string) (*Campaign, error) {
	m, err := LoadManifest(root)
	if err != nil {
		return nil, err
	}
	return &Campaign{Manifest: m, Root: root}, nil
}
