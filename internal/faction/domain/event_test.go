package domain

import (
	"encoding/json"
	"testing"
)

func TestNewMutationRecord_CoinDelta(t *testing.T) {
	mutation := CoinDelta{FactionID: "f1", Delta: 5}
	record, err := NewMutationRecord(mutation)
	if err != nil {
		t.Fatalf("NewMutationRecord: %v", err)
	}
	if record.Type != "coin_delta" {
		t.Errorf("Type = %q, want %q", record.Type, "coin_delta")
	}
	var payload map[string]any
	if err := json.Unmarshal(record.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload["faction_id"] != "f1" {
		t.Errorf("faction_id = %v, want f1", payload["faction_id"])
	}
	if payload["delta"] != float64(5) {
		t.Errorf("delta = %v, want 5", payload["delta"])
	}
}

func TestNewMutationRecord_AssetRemoved(t *testing.T) {
	mutation := AssetRemoved{FactionID: "f1", AssetID: "a1"}
	record, err := NewMutationRecord(mutation)
	if err != nil {
		t.Fatalf("NewMutationRecord: %v", err)
	}
	if record.Type != "asset_removed" {
		t.Errorf("Type = %q, want %q", record.Type, "asset_removed")
	}
	var payload map[string]any
	if err := json.Unmarshal(record.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload["asset_id"] != "a1" {
		t.Errorf("asset_id = %v, want a1", payload["asset_id"])
	}
}

func TestNewMutationRecord_AssetMaintainedFlag(t *testing.T) {
	for _, maintained := range []bool{true, false} {
		mutation := AssetMaintainedFlag{FactionID: "f1", AssetID: "a1", Maintained: maintained}
		record, err := NewMutationRecord(mutation)
		if err != nil {
			t.Fatalf("NewMutationRecord (maintained=%v): %v", maintained, err)
		}
		if record.Type != "asset_maintained_flag" {
			t.Errorf("Type = %q, want %q", record.Type, "asset_maintained_flag")
		}
		var payload map[string]any
		if err := json.Unmarshal(record.Payload, &payload); err != nil {
			t.Fatalf("unmarshal payload (maintained=%v): %v", maintained, err)
		}
		if payload["maintained"] != maintained {
			t.Errorf("maintained = %v, want %v", payload["maintained"], maintained)
		}
	}
}
