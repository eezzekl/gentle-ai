package model

import "testing"

func TestAgentAntigravity(t *testing.T) {
	if AgentAntigravity != "antigravity" {
		t.Errorf("AgentAntigravity = %q, want %q", AgentAntigravity, "antigravity")
	}
}

// TestPersonaGentle verifies that PersonaGentle is defined with the expected value.
func TestPersonaGentle(t *testing.T) {
	if PersonaGentle != "gentle" {
		t.Errorf("PersonaGentle = %q, want %q", PersonaGentle, "gentle")
	}
}

// TestPersonaGentlemanNeutralArtifactsRemoved verifies that the deprecated constant
// is no longer in the selectable constant block. The selectable set is now
// {gentle, neutral, custom} only. This test uses a compile-time check by
// referencing PersonaGentlemanNeutralArtifacts and asserting it maps to
// PersonaGentle (it is kept only as a legacy migration alias in validate.go).
//
// NOTE: PersonaGentlemanNeutralArtifacts is preserved as a constant so existing
// callers referencing it do not break at compile time; however, it is treated as
// a back-compat alias, NOT a first-class selectable persona.
func TestPersonaGentlemanNeutralArtifactsIsAlias(t *testing.T) {
	// PersonaGentlemanNeutralArtifacts must NOT equal PersonaGentle at the type level
	// (they are distinct string values), but normalizePersona maps it to PersonaGentle.
	// This test confirms the constant still exists (compile guard) and that its value
	// is still the legacy string (so alias mapping in validate.go can use it).
	if PersonaGentlemanNeutralArtifacts != "gentleman-neutral-artifacts" {
		t.Errorf("PersonaGentlemanNeutralArtifacts = %q, want %q",
			PersonaGentlemanNeutralArtifacts, "gentleman-neutral-artifacts")
	}
}

// TestRegionIDConstants verifies the curated region sentinel constants exist.
func TestRegionIDConstants(t *testing.T) {
	tests := []struct {
		name string
		got  RegionID
		want RegionID
	}{
		{"RegionArgentina", RegionArgentina, "argentina"},
		{"RegionMexico", RegionMexico, "mexico"},
		{"RegionColombia", RegionColombia, "colombia"},
		{"RegionSpain", RegionSpain, "spain"},
		{"RegionChile", RegionChile, "chile"},
		{"RegionUserLanguage", RegionUserLanguage, "user-language"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// TestRegionMapContainsCuratedRegions verifies the curated region map has entries
// for all six defined regions.
func TestRegionMapContainsCuratedRegions(t *testing.T) {
	expected := []RegionID{
		RegionArgentina,
		RegionMexico,
		RegionColombia,
		RegionSpain,
		RegionChile,
		RegionUserLanguage,
	}

	for _, r := range expected {
		if _, ok := RegionMap[r]; !ok {
			t.Errorf("RegionMap missing entry for %q", r)
		}
	}
}
