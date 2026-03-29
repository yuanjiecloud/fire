package task

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Environment
// ---------------------------------------------------------------------------

func TestEnvironment_Clone(t *testing.T) {
	orig := Environment{"A": "1", "B": "2"}
	clone := orig.Clone()

	if len(clone) != len(orig) {
		t.Fatalf("Clone length mismatch: got %d want %d", len(clone), len(orig))
	}
	for k, v := range orig {
		if clone[k] != v {
			t.Errorf("Clone[%q] = %q, want %q", k, clone[k], v)
		}
	}
	// Mutations to clone must not affect original.
	clone["C"] = "3"
	if _, ok := orig["C"]; ok {
		t.Error("modifying clone affected original")
	}
}

func TestEnvironment_OverridePatch(t *testing.T) {
	base := Environment{"A": "1", "B": "2"}
	patch := Environment{"B": "overridden", "C": "3"}
	result := base.OverridePatch(patch)

	if result["A"] != "1" {
		t.Errorf("A should be unchanged, got %q", result["A"])
	}
	if result["B"] != "overridden" {
		t.Errorf("B should be overridden, got %q", result["B"])
	}
	if result["C"] != "3" {
		t.Errorf("C should be added, got %q", result["C"])
	}
	// Base must be unmodified.
	if base["B"] != "2" {
		t.Error("OverridePatch mutated the receiver")
	}
}

func TestEnvironment_MergeKeepExisting(t *testing.T) {
	base := Environment{"A": "original", "B": "2"}
	incoming := Environment{"A": "should-not-replace", "C": "new"}
	result := base.MergeKeepExisting(incoming)

	if result["A"] != "original" {
		t.Errorf("A should keep original value, got %q", result["A"])
	}
	if result["B"] != "2" {
		t.Errorf("B should be unchanged, got %q", result["B"])
	}
	if result["C"] != "new" {
		t.Errorf("C should be added, got %q", result["C"])
	}
}

func TestEnvironment_MergeKeepExisting_emptyBase(t *testing.T) {
	base := Environment{}
	incoming := Environment{"X": "val"}
	result := base.MergeKeepExisting(incoming)
	if result["X"] != "val" {
		t.Errorf("expected X=val in result, got %q", result["X"])
	}
}

func TestEnvironment_MergeKeepExisting_emptyIncoming(t *testing.T) {
	base := Environment{"X": "val"}
	result := base.MergeKeepExisting(Environment{})
	if result["X"] != "val" {
		t.Error("base value lost after merge with empty incoming")
	}
}

// ---------------------------------------------------------------------------
// EnvProvider
// ---------------------------------------------------------------------------

func TestEnvProvider_Clone_nil(t *testing.T) {
	var p EnvProvider
	clone := p.Clone()
	if clone == nil {
		t.Error("Clone of nil EnvProvider should return empty provider, not nil")
	}
	if len(clone) != 0 {
		t.Errorf("Clone of nil should be empty, got len=%d", len(clone))
	}
}

func TestEnvProvider_Clone(t *testing.T) {
	p := EnvProvider{
		"prod": Environment{"HOST": "prod.example.com"},
		"dev":  Environment{"HOST": "localhost"},
	}
	clone := p.Clone()
	if len(clone) != len(p) {
		t.Fatalf("Clone length mismatch: got %d want %d", len(clone), len(p))
	}
	// Mutations to clone should not affect original.
	clone["staging"] = Environment{"HOST": "staging.example.com"}
	if _, ok := p["staging"]; ok {
		t.Error("adding key to clone affected original")
	}
}

func TestEnvProvider_OverridePatch(t *testing.T) {
	base := EnvProvider{"env1": Environment{"K": "base"}}
	patch := EnvProvider{
		"env1": Environment{"K": "patched"},
		"env2": Environment{"K": "new"},
	}
	result := base.OverridePatch(patch)
	if result["env1"]["K"] != "patched" {
		t.Errorf("env1.K should be patched, got %q", result["env1"]["K"])
	}
	if result["env2"]["K"] != "new" {
		t.Errorf("env2.K should be new, got %q", result["env2"]["K"])
	}
	if base["env1"]["K"] != "base" {
		t.Error("OverridePatch mutated the receiver")
	}
}

func TestEnvProvider_MergeKeepExisting(t *testing.T) {
	base := EnvProvider{"env1": Environment{"K": "base"}}
	incoming := EnvProvider{
		"env1": Environment{"K": "incoming"},
		"env2": Environment{"K": "new"},
	}
	result := base.MergeKeepExisting(incoming)
	if result["env1"]["K"] != "base" {
		t.Errorf("env1 should keep base value, got %q", result["env1"]["K"])
	}
	if result["env2"]["K"] != "new" {
		t.Errorf("env2 should be added, got %q", result["env2"]["K"])
	}
}

func TestNewEnvProvider(t *testing.T) {
	p := NewEnvProvider()
	if p == nil {
		t.Error("NewEnvProvider should not return nil")
	}
	if len(p) != 0 {
		t.Errorf("NewEnvProvider should return empty map, got len=%d", len(p))
	}
}
