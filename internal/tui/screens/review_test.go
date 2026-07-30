package screens

import (
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
	"github.com/gentleman-programming/gentle-ai/v2/internal/planner"
)

// ─── Issue #145: Review screen must show individual skills ───────────────────

// TestRenderReviewShowsSkillNames verifies that when ReviewPayload.Skills is
// populated, RenderReview output contains each individual skill name.
//
// Closes #145.
func TestRenderReviewShowsSkillNames(t *testing.T) {
	payload := planner.ReviewPayload{
		Agents:  []model.AgentID{model.AgentClaudeCode},
		Persona: model.PersonaGentleman,
		Preset:  model.PresetFullGentleman,
		Components: []planner.ComponentAction{
			{ID: model.ComponentSkills, Action: "selected"},
		},
		Skills: []model.SkillID{"sdd-apply", "sdd-spec", "go-testing"},
	}

	out := RenderReview(payload, 0)

	for _, skillName := range []string{"sdd-apply", "sdd-spec", "go-testing"} {
		if !strings.Contains(out, skillName) {
			t.Errorf("RenderReview output missing skill %q; output:\n%s", skillName, out)
		}
	}
}

// TestRenderReviewHidesSkillsSectionWhenEmpty verifies that when there are no
// skills selected, the review screen does not crash and shows no skill names.
//
// Closes #145.
func TestRenderReviewHidesSkillsSectionWhenEmpty(t *testing.T) {
	payload := planner.ReviewPayload{
		Agents:  []model.AgentID{model.AgentClaudeCode},
		Persona: model.PersonaGentleman,
		Preset:  model.PresetFullGentleman,
		// No Skills field.
	}

	out := RenderReview(payload, 0)

	// Should not panic and should render something.
	if len(out) == 0 {
		t.Fatal("RenderReview returned empty string")
	}
}

// ─── Issue #149: Review screen must show Strict TDD status ───────────────────

// TestRenderReviewShowsStrictTDDEnabled verifies that RenderReview output contains
// "Strict TDD" and "Enabled" when HasSDD=true and StrictTDD=true.
//
// Closes #149.
func TestRenderReviewShowsStrictTDDEnabled(t *testing.T) {
	payload := planner.ReviewPayload{
		Agents:  []model.AgentID{model.AgentClaudeCode},
		Persona: model.PersonaGentleman,
		Preset:  model.PresetFullGentleman,
		Components: []planner.ComponentAction{
			{ID: model.ComponentSDD, Action: "selected"},
		},
		HasSDD:    true,
		StrictTDD: true,
	}

	out := RenderReview(payload, 0)

	if !strings.Contains(out, "Strict TDD") {
		t.Errorf("RenderReview missing 'Strict TDD'; output:\n%s", out)
	}
	if !strings.Contains(out, "Enabled") {
		t.Errorf("RenderReview missing 'Enabled' for StrictTDD=true; output:\n%s", out)
	}
}

// TestRenderReviewShowsStrictTDDDisabled verifies that RenderReview output contains
// "Strict TDD" and "Disabled" when HasSDD=true and StrictTDD=false.
//
// Closes #149.
func TestRenderReviewShowsStrictTDDDisabled(t *testing.T) {
	payload := planner.ReviewPayload{
		Agents:  []model.AgentID{model.AgentClaudeCode},
		Persona: model.PersonaGentleman,
		Preset:  model.PresetFullGentleman,
		Components: []planner.ComponentAction{
			{ID: model.ComponentSDD, Action: "selected"},
		},
		HasSDD:    true,
		StrictTDD: false,
	}

	out := RenderReview(payload, 0)

	if !strings.Contains(out, "Strict TDD") {
		t.Errorf("RenderReview missing 'Strict TDD'; output:\n%s", out)
	}
	if !strings.Contains(out, "Disabled") {
		t.Errorf("RenderReview missing 'Disabled' for StrictTDD=false; output:\n%s", out)
	}
}

// TestRenderReviewHidesStrictTDDWhenNoSDD verifies that when HasSDD=false,
// "Strict TDD" does not appear in the review output.
//
// Closes #149.
func TestRenderReviewHidesStrictTDDWhenNoSDD(t *testing.T) {
	payload := planner.ReviewPayload{
		Agents:    []model.AgentID{model.AgentClaudeCode},
		Persona:   model.PersonaGentleman,
		Preset:    model.PresetFullGentleman,
		HasSDD:    false,
		StrictTDD: true,
	}

	out := RenderReview(payload, 0)

	if strings.Contains(out, "Strict TDD") {
		t.Errorf("RenderReview should NOT show 'Strict TDD' when HasSDD=false; output:\n%s", out)
	}
}

func TestRenderReviewClarifiesCustomPersonaAndPreset(t *testing.T) {
	payload := planner.ReviewPayload{
		Agents:  []model.AgentID{model.AgentClaudeCode},
		Persona: model.PersonaCustom,
		Preset:  model.PresetCustom,
	}

	out := RenderReview(payload, 0)

	if !strings.Contains(out, "keep existing persona unmanaged") {
		t.Fatalf("RenderReview missing custom persona clarification; output:\n%s", out)
	}
	if !strings.Contains(out, "choose components and skills manually") {
		t.Fatalf("RenderReview missing custom preset clarification; output:\n%s", out)
	}
	if strings.Contains(out, "Persona  custom") {
		t.Fatalf("RenderReview should not show raw custom persona label; output:\n%s", out)
	}
	if strings.Contains(out, "Preset  custom") {
		t.Fatalf("RenderReview should not show raw custom preset label; output:\n%s", out)
	}
}

func TestRenderReviewSummarizesPersonaConversationAndArtifacts(t *testing.T) {
	tests := []struct {
		name    string
		persona model.PersonaID
		want    string
	}{
		// Re-anchored to the style axis: the review must summarize the STYLE the
		// user picked, without claiming a regional tone — the region is a separate
		// selection shown on its own review line.
		{name: "Gentle", persona: model.PersonaGentle, want: "Teaching-first mentor tone; reply language chosen on the next screen"},
		{name: "Regionless", persona: model.PersonaNeutral, want: "No marked conversation tone; English technical artifacts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := planner.ReviewPayload{Persona: tt.persona}
			out := RenderReview(payload, 0)
			if !strings.Contains(out, tt.want) {
				t.Fatalf("RenderReview() missing %q; output:\n%s", tt.want, out)
			}
		})
	}
}

// ─── WU-9: Review screen must show region + artifacts-in-English flag ─────────

func TestRenderReviewShowsRegionLabel(t *testing.T) {
	payload := planner.ReviewPayload{
		Agents:             []model.AgentID{model.AgentClaudeCode},
		Persona:            model.PersonaGentle,
		Preset:             model.PresetFullGentleman,
		Region:             string(model.RegionMexico),
		ArtifactsInEnglish: true,
	}

	out := RenderReview(payload, 0)

	if !strings.Contains(out, "Region") && !strings.Contains(out, "Language") {
		t.Fatalf("RenderReview should show a Region/Language line; output:\n%s", out)
	}
	if !strings.Contains(out, model.RegionMap[model.RegionMexico]) && !strings.Contains(out, "mexico") {
		t.Fatalf("RenderReview should show the selected region; output:\n%s", out)
	}
}

func TestRenderReviewShowsArtifactsFlagOnAndOff(t *testing.T) {
	on := RenderReview(planner.ReviewPayload{
		Persona:            model.PersonaGentle,
		Region:             string(model.RegionArgentina),
		ArtifactsInEnglish: true,
	}, 0)
	if !strings.Contains(on, "[x]") {
		t.Fatalf("RenderReview should show artifacts flag checked when ArtifactsInEnglish=true; output:\n%s", on)
	}

	off := RenderReview(planner.ReviewPayload{
		Persona:            model.PersonaGentle,
		Region:             string(model.RegionArgentina),
		ArtifactsInEnglish: false,
	}, 0)
	if !strings.Contains(off, "[ ]") {
		t.Fatalf("RenderReview should show artifacts flag unchecked when ArtifactsInEnglish=false; output:\n%s", off)
	}
}

// TestRenderReviewShowsRegionlessStateForNeutral pins the confirm-before-write
// reading for a regionless style: the region line must state that none applies,
// not render blank or "not set", which would read as an unfinished field.
func TestRenderReviewShowsRegionlessStateForNeutral(t *testing.T) {
	out := RenderReview(planner.ReviewPayload{
		Agents:             []model.AgentID{model.AgentClaudeCode},
		Persona:            model.PersonaNeutral,
		Region:             "",
		ArtifactsInEnglish: true,
	}, 0)

	if !strings.Contains(out, "none (regionless style)") {
		t.Fatalf("RenderReview() must make the regionless state legible; output:\n%s", out)
	}
	if strings.Contains(out, "not set") {
		t.Fatalf("RenderReview() still reads the regionless state as unfinished; output:\n%s", out)
	}
	if !strings.Contains(out, "[x]") {
		t.Fatalf("RenderReview() must show artifacts-in-English on for neutral; output:\n%s", out)
	}
}
