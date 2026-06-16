package screens

import (
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

func TestPersonaOptionsAreStyleAxisOnly(t *testing.T) {
	got := PersonaOptions()
	want := []model.PersonaID{model.PersonaGentle, model.PersonaNeutral, model.PersonaCustom}
	if len(got) != len(want) {
		t.Fatalf("PersonaOptions() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("PersonaOptions()[%d] = %q, want %q (full: %v)", i, got[i], want[i], got)
		}
	}
}

func TestPersonaOptionsDropHybrid(t *testing.T) {
	for _, option := range PersonaOptions() {
		if option == model.PersonaGentlemanNeutralArtifacts {
			t.Fatalf("PersonaOptions() still includes the removed hybrid persona %q", option)
		}
	}
}

// TestPersonaDescriptionsNeverReuseNeutral pins the invariant issue #833 is
// actually about: "neutral" named two unrelated axes at once, artifact language
// and conversational tone, so the selector could not be read unambiguously.
// Asserting the negative space is what keeps that ambiguity from returning; a
// positive assertion against the description constants cannot, because it
// passes for any wording those constants happen to hold, including a
// reintroduced "neutral".
func TestPersonaDescriptionsNeverReuseNeutral(t *testing.T) {
	for persona, description := range personaDescriptions {
		for _, field := range strings.Fields(strings.ToLower(description)) {
			if strings.Trim(field, ".,;:()") == "neutral" {
				t.Fatalf("persona %q description reuses the ambiguous word %q: %q", persona, "neutral", description)
			}
		}
	}
}

// TestPersonaDescriptionsCarryNoRegionalTone is the two-axis successor to
// TestPersonaDescriptionsSeparateToneFromArtifactLanguage. That test asserted
// only the Gentleman personas claimed a regional tone; under the decoupled model
// NO style may claim one, because the region is a separate selection made on the
// next screen. A description promising voseo would misreport what the selector
// installs for a user who then picks Mexico.
func TestPersonaDescriptionsCarryNoRegionalTone(t *testing.T) {
	regionalWords := []string{"voseo", "rioplatense", "argentine", "mexican", "colombian", "chilean", "castilian"}
	for persona, description := range personaDescriptions {
		lowered := strings.ToLower(description)
		for _, word := range regionalWords {
			if strings.Contains(lowered, word) {
				t.Fatalf("persona %q description claims regional tone %q, which now belongs to the region axis: %q",
					persona, word, description)
			}
		}
	}
}

// TestPersonaDescriptionsStayDistinguishable keeps the selector readable: the
// managed styles must not collapse into the same sentence, and the style that
// forces English artifacts must say so.
func TestPersonaDescriptionsStayDistinguishable(t *testing.T) {
	for _, persona := range PersonaOptions() {
		if _, ok := personaDescriptions[persona]; !ok {
			t.Fatalf("persona %q has no description", persona)
		}
	}

	if personaDescriptions[model.PersonaGentle] == personaDescriptions[model.PersonaNeutral] {
		t.Fatal("the gentle and neutral styles must stay distinguishable in the selector")
	}

	if !strings.Contains(strings.ToLower(personaDescriptions[model.PersonaNeutral]), "english technical artifacts") {
		t.Fatalf("the regionless style must state that technical artifacts are English: %q",
			personaDescriptions[model.PersonaNeutral])
	}
}

// TestRenderPersonaShowsEveryManagedDescription keeps the selector itself
// honest: each selectable style's description must actually reach the rendered
// screen.
func TestRenderPersonaShowsEveryManagedDescription(t *testing.T) {
	for _, persona := range PersonaOptions() {
		out := RenderPersona(persona, 0)
		if !strings.Contains(out, personaDescriptions[persona]) {
			t.Fatalf("RenderPersona(%q) does not show its description %q; output:\n%s",
				persona, personaDescriptions[persona], out)
		}
	}
}

func TestRenderPersonaDescribesStyleAxis(t *testing.T) {
	out := RenderPersona(model.PersonaGentle, 0)
	for _, want := range []string{
		"gentle",
		"neutral",
		"custom",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("RenderPersona() missing style option %q; output:\n%s", want, out)
		}
	}
	if strings.Contains(out, "gentleman-neutral-artifacts") {
		t.Fatalf("RenderPersona() still mentions the removed hybrid persona; output:\n%s", out)
	}
}

// TestReviewPersonaLabelKeepsThePersonaID guards the confirm-before-write
// screen: the reader must be able to see the exact persona value that will be
// written to state.json, not only its prose description.
func TestReviewPersonaLabelKeepsThePersonaID(t *testing.T) {
	for _, persona := range []model.PersonaID{
		model.PersonaGentle,
		model.PersonaNeutral,
	} {
		label := reviewPersonaLabel(persona)
		if !strings.Contains(label, string(persona)) {
			t.Fatalf("reviewPersonaLabel(%q) = %q, must contain the persona ID", persona, label)
		}
		if !strings.Contains(label, personaDescriptions[persona]) {
			t.Fatalf("reviewPersonaLabel(%q) = %q, must contain the description", persona, label)
		}
	}
}
