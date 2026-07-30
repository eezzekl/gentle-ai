package model

import (
	"strings"
	"testing"
)

// TestComposeLanguageDirectiveCuratedRegions verifies that each curated region
// produces a directive containing both the expected language phrase and the
// correct artifacts clause.
func TestComposeLanguageDirectiveCuratedRegions(t *testing.T) {
	tests := []struct {
		name               string
		region             RegionID
		artifactsInEnglish bool
		wantLangPhrase     string
		wantArtifactPhrase string
	}{
		{
			name:               "argentina artifacts-in-english",
			region:             RegionArgentina,
			artifactsInEnglish: true,
			wantLangPhrase:     "Rioplatense",
			wantArtifactPhrase: "English",
		},
		{
			name:               "argentina artifacts-in-user-language",
			region:             RegionArgentina,
			artifactsInEnglish: false,
			wantLangPhrase:     "Rioplatense",
			wantArtifactPhrase: "language",
		},
		{
			name:               "mexico artifacts-in-english",
			region:             RegionMexico,
			artifactsInEnglish: true,
			wantLangPhrase:     "Mexico",
			wantArtifactPhrase: "English",
		},
		{
			name:               "mexico artifacts-in-user-language",
			region:             RegionMexico,
			artifactsInEnglish: false,
			wantLangPhrase:     "Mexico",
			wantArtifactPhrase: "language",
		},
		{
			name:               "colombia artifacts-in-english",
			region:             RegionColombia,
			artifactsInEnglish: true,
			wantLangPhrase:     "Colombia",
			wantArtifactPhrase: "English",
		},
		{
			name:               "colombia artifacts-in-user-language",
			region:             RegionColombia,
			artifactsInEnglish: false,
			wantLangPhrase:     "Colombia",
			wantArtifactPhrase: "language",
		},
		{
			name:               "spain artifacts-in-english",
			region:             RegionSpain,
			artifactsInEnglish: true,
			wantLangPhrase:     "Spain",
			wantArtifactPhrase: "English",
		},
		{
			name:               "spain artifacts-in-user-language",
			region:             RegionSpain,
			artifactsInEnglish: false,
			wantLangPhrase:     "Spain",
			wantArtifactPhrase: "language",
		},
		{
			name:               "chile artifacts-in-english",
			region:             RegionChile,
			artifactsInEnglish: true,
			wantLangPhrase:     "Chile",
			wantArtifactPhrase: "English",
		},
		{
			name:               "chile artifacts-in-user-language",
			region:             RegionChile,
			artifactsInEnglish: false,
			wantLangPhrase:     "Chile",
			wantArtifactPhrase: "language",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComposeLanguageDirective(tt.region, tt.artifactsInEnglish)
			if !strings.Contains(got, tt.wantLangPhrase) {
				t.Errorf("ComposeLanguageDirective(%q, %v) = %q\nwant string containing %q",
					tt.region, tt.artifactsInEnglish, got, tt.wantLangPhrase)
			}
			if !strings.Contains(got, tt.wantArtifactPhrase) {
				t.Errorf("ComposeLanguageDirective(%q, %v) = %q\nwant string containing %q",
					tt.region, tt.artifactsInEnglish, got, tt.wantArtifactPhrase)
			}
		})
	}
}

// TestComposeLanguageDirectiveUserLanguage verifies the user-language sentinel
// produces a follow-user directive with no forced region name.
func TestComposeLanguageDirectiveUserLanguage(t *testing.T) {
	got := ComposeLanguageDirective(RegionUserLanguage, true)

	// Must contain follow-user phrasing.
	if !strings.Contains(got, "language the user") {
		t.Errorf("user-language directive missing follow-user phrase; got %q", got)
	}
	// Must NOT contain any curated region name.
	for _, r := range []string{"Rioplatense", "Mexico", "Colombia", "Spain", "Chile"} {
		if strings.Contains(got, r) {
			t.Errorf("user-language directive must not contain %q; got %q", r, got)
		}
	}
	// Artifacts clause present.
	if !strings.Contains(got, "English") {
		t.Errorf("user-language directive with artifactsInEnglish=true missing English clause; got %q", got)
	}
}

// TestComposeLanguageDirectiveFreeText verifies that an unrecognized region string
// is injected verbatim into the directive.
func TestComposeLanguageDirectiveFreeText(t *testing.T) {
	freeText := RegionID("yucateco")
	got := ComposeLanguageDirective(freeText, true)

	if !strings.Contains(got, "yucateco") {
		t.Errorf("free-text region not injected verbatim; got %q", got)
	}
	if !strings.Contains(got, "English") {
		t.Errorf("free-text directive missing English artifacts clause; got %q", got)
	}
}

// TestComposeLanguageDirectiveArtifactsClauseDiffers verifies that
// artifactsInEnglish=true and artifactsInEnglish=false produce distinct directives.
func TestComposeLanguageDirectiveArtifactsClauseDiffers(t *testing.T) {
	english := ComposeLanguageDirective(RegionArgentina, true)
	inLang := ComposeLanguageDirective(RegionArgentina, false)

	if english == inLang {
		t.Errorf("artifactsInEnglish=true and =false produced identical directives:\n%q", english)
	}
}

// TestComposeLanguageDirectiveArgentinaTrueMatchesProdText verifies that the
// argentina+artifactsInEnglish=true directive reproduces today's gentleman language
// behavior: Rioplatense/voseo + English artifacts.
func TestComposeLanguageDirectiveArgentinaTrueMatchesProdText(t *testing.T) {
	got := ComposeLanguageDirective(RegionArgentina, true)

	// The directive must communicate both the Rioplatense/voseo style and
	// the requirement that code artifacts stay in English.
	requiredPhrases := []string{"voseo", "English"}
	for _, phrase := range requiredPhrases {
		if !strings.Contains(got, phrase) {
			t.Errorf("argentina+true directive missing %q; got:\n%q", phrase, got)
		}
	}
}

// ─── Decision 5: neutral is regionless ───────────────────────────────────────

// TestComposeLanguageDirectiveEmptyRegionEmitsNoVoiceClause pins the regionless
// case. `neutral` carries no region by definition, so the directive must supply
// the artifacts contract and nothing else.
//
// Before Decision 5 an empty region fell through to the free-text branch and
// emitted the literal string "Reply in the following language/region: ." — a
// dangling period injected into every un-migrated install.
func TestComposeLanguageDirectiveEmptyRegionEmitsNoVoiceClause(t *testing.T) {
	for _, artifactsInEnglish := range []bool{true, false} {
		got := ComposeLanguageDirective("", artifactsInEnglish)

		if strings.Contains(got, "Reply in") {
			t.Errorf("ComposeLanguageDirective(\"\", %v) emitted a reply-voice clause; got:\n%s", artifactsInEnglish, got)
		}
		if strings.Contains(got, "language/region: .") {
			t.Errorf("ComposeLanguageDirective(\"\", %v) emitted the malformed free-text clause; got:\n%s", artifactsInEnglish, got)
		}
		if !strings.Contains(got, "generated artifacts") && !strings.Contains(got, "Generated artifacts") {
			t.Errorf("ComposeLanguageDirective(\"\", %v) dropped the artifacts clause; got:\n%s", artifactsInEnglish, got)
		}
	}
}

// TestComposeLanguageDirectiveEmptyRegionDiffersFromUserLanguage keeps the two
// regionless-looking cases distinct. `user-language` is a real selection a
// gentle user makes ("follow how I write"); an empty region is the absence of
// the axis entirely. Collapsing them would make `neutral` claim a reply-voice
// instruction it never selected.
func TestComposeLanguageDirectiveEmptyRegionDiffersFromUserLanguage(t *testing.T) {
	empty := ComposeLanguageDirective("", true)
	userLang := ComposeLanguageDirective(RegionUserLanguage, true)

	if empty == userLang {
		t.Fatalf("empty region and user-language must compose different directives; both were:\n%s", empty)
	}
	if !strings.Contains(userLang, "Reply in the language the user writes in") {
		t.Errorf("user-language lost its reply clause; got:\n%s", userLang)
	}
}

// TestComposeLanguageDirectiveFreeTextStillWorks guards against the empty-region
// fix swallowing the free-text branch it was carved out of.
func TestComposeLanguageDirectiveFreeTextStillWorks(t *testing.T) {
	got := ComposeLanguageDirective("yucateco", true)

	if !strings.Contains(got, "yucateco") {
		t.Errorf("free-text region not injected verbatim; got:\n%s", got)
	}
}
