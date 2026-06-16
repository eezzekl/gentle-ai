package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
	"github.com/gentleman-programming/gentle-ai/v2/internal/system"
	"github.com/gentleman-programming/gentle-ai/v2/internal/tui/screens"
)

func personaCursor(t *testing.T, persona model.PersonaID) int {
	t.Helper()
	for i, p := range screens.PersonaOptions() {
		if p == persona {
			return i
		}
	}
	t.Fatalf("persona %q not in PersonaOptions()", persona)
	return -1
}

func TestPersonaGentleRoutesToLanguageScreen(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenPersona
	m.Cursor = personaCursor(t, model.PersonaGentle)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if got := updated.(Model).Screen; got != ScreenPersonaLanguage {
		t.Fatalf("after selecting gentle, Screen = %v, want ScreenPersonaLanguage", got)
	}
}

func TestPersonaNeutralRoutesToLanguageScreen(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenPersona
	m.Cursor = personaCursor(t, model.PersonaNeutral)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if got := updated.(Model).Screen; got != ScreenPersonaLanguage {
		t.Fatalf("after selecting neutral, Screen = %v, want ScreenPersonaLanguage", got)
	}
}

func TestPersonaCustomSkipsLanguageScreen(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenPersona
	m.Cursor = personaCursor(t, model.PersonaCustom)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if got := updated.(Model).Screen; got != ScreenPreset {
		t.Fatalf("after selecting custom, Screen = %v, want ScreenPreset", got)
	}
}

func TestArtifactsInEnglishDefaultsTrue(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	if !m.Selection.ArtifactsInEnglish {
		t.Fatal("Selection.ArtifactsInEnglish should default to true on a fresh model")
	}
}

func TestPersonaLanguageSelectingUserLanguageSetsRegionAndAdvances(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenPersonaLanguage
	// user-language is index 5 in LanguageRegionOptions.
	m.Cursor = 5
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	st := updated.(Model)
	if st.Selection.Region != string(model.RegionUserLanguage) {
		t.Fatalf("Region = %q, want %q", st.Selection.Region, model.RegionUserLanguage)
	}
	if st.Screen != ScreenPreset {
		t.Fatalf("after selecting region, Screen = %v, want ScreenPreset", st.Screen)
	}
}

func TestPersonaLanguageFreeTextSetsRegionVerbatim(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenPersonaLanguage
	freeTextIdx := len(screens.LanguageRegionOptions()) - 1
	m.Cursor = freeTextIdx
	// Type "klingon" while the free-text radio is focused.
	for _, r := range "klingon" {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	st := updated.(Model)
	if st.Selection.Region != "klingon" {
		t.Fatalf("Region = %q, want verbatim free text %q", st.Selection.Region, "klingon")
	}
	if st.Screen != ScreenPreset {
		t.Fatalf("after free-text select, Screen = %v, want ScreenPreset", st.Screen)
	}
}

func TestPersonaLanguageCheckboxToggle(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenPersonaLanguage
	m.Selection.ArtifactsInEnglish = true
	// Checkbox sits right after the region options.
	m.Cursor = len(screens.LanguageRegionOptions())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if updated.(Model).Selection.ArtifactsInEnglish {
		t.Fatal("enter on the checkbox should toggle ArtifactsInEnglish to false")
	}
}
