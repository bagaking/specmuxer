package adapters

import (
	"testing"
	"time"
)

func TestNewRegistryIncludesBuiltins(t *testing.T) {
	reg := NewRegistry()
	all := reg.List()
	if len(all) < 2 {
		t.Fatalf("expected at least 2 adapters, got %d", len(all))
	}

	if _, err := reg.Resolve("codex", nil); err != nil {
		t.Fatalf("Resolve codex: %v", err)
	}
	if _, err := reg.Resolve("claude", nil); err != nil {
		t.Fatalf("Resolve claude: %v", err)
	}
}

func TestResolveAppliesOverrides(t *testing.T) {
	reg := NewRegistry()
	override := map[string]LifecycleCommand{
		"start": {
			Exec:        []string{"custom", "start"},
			Timeout:     10 * time.Second,
			Description: "custom start",
		},
		"resume": {
			Exec: []string{"custom", "resume"},
		},
		"extract_state": {
			Exec: []string{"custom", "export"},
		},
		"post_launch": {
			Exec: []string{"post", "launch"},
		},
	}

	def, err := reg.Resolve("codex", override)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if def.Start.Exec[0] != "custom" {
		t.Fatalf("override start not applied: %#v", def.Start.Exec)
	}
	if def.Resume == nil || def.Resume.Exec[0] != "custom" {
		t.Fatalf("override resume not applied: %#v", def.Resume)
	}
	if def.ExtractState == nil || def.ExtractState.Exec[0] != "custom" {
		t.Fatalf("override extract_state not applied: %#v", def.ExtractState)
	}
	if def.Overrides["post_launch"].Exec[0] != "post" {
		t.Fatalf("custom override missing: %#v", def.Overrides)
	}

	// Ensure registry copy is isolated
	def.Start.Exec[0] = "mutated"
	resolved, err := reg.Resolve("codex", nil)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if resolved.Start.Exec[0] == "mutated" {
		t.Fatalf("registry definition mutated by caller")
	}
}

func TestRegisterAddsAdapter(t *testing.T) {
	reg := NewRegistry()
	custom := Definition{
		Name: "custom",
		Start: LifecycleCommand{
			Exec: []string{"custom", "run"},
		},
	}
	reg.Register(custom)

	def, err := reg.Resolve("custom", nil)
	if err != nil {
		t.Fatalf("Resolve custom: %v", err)
	}
	if def.Name != "custom" {
		t.Fatalf("unexpected definition: %+v", def)
	}
}
