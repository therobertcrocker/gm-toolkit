package foundry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

var ErrNotAnActor = errors.New("not an actor export")

type valueOf[T any] struct {
	Value T `json:"value"`
}

type actor struct {
	Name  string          `json:"name"`
	Type  string          `json:"type"`
	Raw   json.RawMessage `json:"system"`
	Items []item          `json:"items"`
}

type item struct {
	ID     string          `json:"_id"`
	Name   string          `json:"name"`
	Type   string          `json:"type"`
	Raw    json.RawMessage `json:"system"`
	System itemCommon      `json:"-"`
	Stats  itemStats       `json:"_stats"`
	Flags  itemFlags       `json:"flags"`
}

type itemStats struct {
	CompendiumSource string `json:"compendiumSource"`
}

type itemFlags struct {
	Core coreFlags `json:"core"`
}

type coreFlags struct {
	SourceID string `json:"sourceId"`
}

type itemCommon struct {
	Slug       string           `json:"slug"`
	Traits     itemTraits       `json:"traits"`
	Rules      []map[string]any `json:"rules"`
	Equipped   equipped         `json:"equipped"`
	Quantity   int              `json:"quantity"`
	ActionType valueOf[string]  `json:"actionType"`
}

type itemTraits struct {
	Value  []string `json:"value"`
	Rarity string   `json:"rarity"`
}

type equipped struct {
	CarryType string `json:"carryType"`
	Invested  *bool  `json:"invested"`
}

type damageRoll struct {
	Damage     string `json:"damage"`
	DamageType string `json:"damageType"`
	Category   string `json:"category"`
}

type damageRolls []damageRoll

func (rolls *damageRolls) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	opening, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, isDelimiter := opening.(json.Delim); !isDelimiter || delimiter != '{' {
		return fmt.Errorf("damageRolls: want an object, got %v", opening)
	}
	for decoder.More() {
		if _, err := decoder.Token(); err != nil {
			return err
		}
		var roll damageRoll
		if err := decoder.Decode(&roll); err != nil {
			return err
		}
		*rolls = append(*rolls, roll)
	}
	_, err = decoder.Token()
	return err
}

func readActor(path string) (*actor, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read actor %s: %w", path, err)
	}
	var source actor
	if err := json.Unmarshal(content, &source); err != nil {
		return nil, fmt.Errorf("read actor %s: %w", path, err)
	}
	if source.Type != "character" && source.Type != "npc" {
		return nil, fmt.Errorf("read actor %s: %w: type %q", path, ErrNotAnActor, source.Type)
	}
	for index := range source.Items {
		if err := source.Items[index].decodeCommon(); err != nil {
			return nil, fmt.Errorf("read actor %s: %w", path, err)
		}
	}
	return &source, nil
}

func readItem(path string) (*item, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read item %s: %w", path, err)
	}
	var loose item
	if err := json.Unmarshal(content, &loose); err != nil {
		return nil, fmt.Errorf("read item %s: %w", path, err)
	}
	if err := loose.decodeCommon(); err != nil {
		return nil, fmt.Errorf("read item %s: %w", path, err)
	}
	return &loose, nil
}

func (source *actor) decode(target any) error {
	if err := json.Unmarshal(source.Raw, target); err != nil {
		return fmt.Errorf("actor %s: %w", source.Name, err)
	}
	return nil
}

func (current *item) decodeCommon() error {
	return current.decode(&current.System)
}

func (current *item) decode(target any) error {
	if err := json.Unmarshal(current.Raw, target); err != nil {
		return fmt.Errorf("item %s (%s): %w", current.Name, current.Type, err)
	}
	return nil
}
