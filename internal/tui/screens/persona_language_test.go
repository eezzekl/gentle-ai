package screens

import (
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

func TestLanguageRegionOptionsAreFiveCuratedPlusUserLangPlusFreeText(t *testing.T) {
	got := LanguageRegionOptions()
	want := []model.RegionID{
		model.RegionArgentina,
		model.RegionMexico,
		model.RegionColombia,
		model.RegionSpain,
		model.RegionChile,
		model.RegionUserLanguage,
		RegionFreeTextSentinel,
	}
	if len(got) != len(want) {
		t.Fatalf("LanguageRegionOptions() = %v (len %d), want len %d", got, len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("LanguageRegionOptions()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestRenderPersonaLanguageShowsRegionsUserLangAndFreeText(t *testing.T) {
	sel := model.Selection{Region: string(model.RegionArgentina), ArtifactsInEnglish: true}
	out := RenderPersonaLanguage(sel, 0, "")
	for _, want := range []string{
		"Argentino (rioplatense, voseo)", // curated gentilicio-first label
		"Mexicano (tuteo)",
		"Idioma del usuario", // user-language option
		"Otro",               // free-text option
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("RenderPersonaLanguage() missing %q; output:\n%s", want, out)
		}
	}
}

func TestRenderPersonaLanguageCheckboxReflectsArtifactsFlag(t *testing.T) {
	on := RenderPersonaLanguage(model.Selection{ArtifactsInEnglish: true}, 0, "")
	off := RenderPersonaLanguage(model.Selection{ArtifactsInEnglish: false}, 0, "")
	if !strings.Contains(on, "[x]") {
		t.Fatalf("checkbox ON state should render a checked box; output:\n%s", on)
	}
	if !strings.Contains(off, "[ ]") {
		t.Fatalf("checkbox OFF state should render an unchecked box; output:\n%s", off)
	}
}

func TestRenderPersonaLanguageEchoesFreeTextValue(t *testing.T) {
	out := RenderPersonaLanguage(model.Selection{Region: "klingon"}, 6, "klingon")
	if !strings.Contains(out, "klingon") {
		t.Fatalf("RenderPersonaLanguage() should echo the free-text value; output:\n%s", out)
	}
}
