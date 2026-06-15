package persona

import (
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/assets"
)

// TestStripRegionalVoiceDirectiveRemovesOnlyTheRegionalLine verifies the core
// contract: the transform removes the hardcoded regional reply directive and
// NOTHING else. Every other line of the language section is region-agnostic
// contract (reply-language anchoring, no drift, no code-switching) that the
// assets deliberately ship and that the two-axis model does not replace.
func TestStripRegionalVoiceDirectiveRemovesOnlyTheRegionalLine(t *testing.T) {
	input := strings.Join([]string{
		"## Language",
		"",
		"- Always match the user's current language in your reply.",
		"- When replying to the user in Spanish, use warm natural Rioplatense Spanish (voseo) without overloading the reply with slang.",
		"- Do not switch languages unless the user does, asks you to, or you are quoting/translating content.",
		"",
		"## Tone",
		"",
		"Passionate and direct.",
		"",
	}, "\n")

	got := stripRegionalVoiceDirective(input)

	if strings.Contains(got, "Rioplatense Spanish (voseo) without overloading") {
		t.Errorf("stripRegionalVoiceDirective() left the regional reply directive; got:\n%s", got)
	}
	for _, survivor := range []string{
		"## Language",
		"- Always match the user's current language in your reply.",
		"- Do not switch languages unless the user does, asks you to, or you are quoting/translating content.",
		"## Tone",
		"Passionate and direct.",
	} {
		if !strings.Contains(got, survivor) {
			t.Errorf("stripRegionalVoiceDirective() removed %q, which is not regional; got:\n%s", survivor, got)
		}
	}
}

// TestStripRegionalVoiceDirectivePreservesArtifactsGuard is the anti-overreach
// test. The `## Persona Scope` artifacts guard also names Rioplatense and voseo,
// but it governs generated artifacts, not reply voice. Removing it would drop a
// real protection, so a naive "delete any line mentioning voseo" transform must
// fail here.
func TestStripRegionalVoiceDirectivePreservesArtifactsGuard(t *testing.T) {
	guard := "- Never inject Rioplatense slang, voseo, or persona stylistic emphasis (CAPS, exclamations, rhetorical questions) into generated code, UI strings, or any task artifact."
	input := "## Persona Scope\n\n" + guard + "\n\n## Language\n\n- When replying to the user in Spanish, use warm natural Rioplatense Spanish (voseo) without overloading the reply with slang.\n"

	got := stripRegionalVoiceDirective(input)

	if !strings.Contains(got, guard) {
		t.Errorf("stripRegionalVoiceDirective() removed the Persona Scope artifacts guard; got:\n%s", got)
	}
	if strings.Contains(got, "When replying to the user in Spanish") {
		t.Errorf("stripRegionalVoiceDirective() left the regional reply directive; got:\n%s", got)
	}
}

// TestStripRegionalVoiceDirectiveIsNoOpWhenAbsent verifies content without a
// regional directive round-trips byte-identical. Claude's persona asset is in
// this state upstream: its voice content already moved into the output style.
func TestStripRegionalVoiceDirectiveIsNoOpWhenAbsent(t *testing.T) {
	input := "## Rules\n\n- Use conventional commits.\n\n## Expertise\n\nArchitecture.\n"

	if got := stripRegionalVoiceDirective(input); got != input {
		t.Errorf("stripRegionalVoiceDirective() modified content with no regional directive\ngot:  %q\nwant: %q", got, input)
	}
}

// TestStripRegionalVoiceDirectiveIsIdempotent guards the sync path: running the
// transform twice must equal running it once.
func TestStripRegionalVoiceDirectiveIsIdempotent(t *testing.T) {
	input := assets.MustRead("generic/persona-gentleman.md")

	once := stripRegionalVoiceDirective(input)
	twice := stripRegionalVoiceDirective(once)

	if once != twice {
		t.Errorf("stripRegionalVoiceDirective() is not idempotent:\n--- once ---\n%s\n--- twice ---\n%s", once, twice)
	}
}

// TestShippedAssetsCarryNoRegionalReplyDirective is the asset guard. After WU-5
// the shipped assets must not hardcode a regional reply voice — the composed
// directive supplies it for the region the user actually selected — while every
// region-agnostic language rule must still be present. The second half is the
// load-bearing part: an over-wide strip would silently delete guardrails the
// assets ship on purpose.
func TestShippedAssetsCarryNoRegionalReplyDirective(t *testing.T) {
	tests := []struct {
		path             string
		requiredContract []string
	}{
		{
			path: "generic/persona-gentleman.md",
			requiredContract: []string{
				"## Language",
				"Match the user's current language in your REPLY ONLY",
				"Do not switch languages unless the user does, asks you to, or you are quoting/translating content.",
			},
		},
		{
			path: "claude/output-style-gentleman.md",
			requiredContract: []string{
				"## Language Rules",
				"Always match the user's current language in your reply.",
				"Determine the reply language from the latest actual user request",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			content := assets.MustRead(tt.path)

			if strings.Contains(content, "use warm natural Rioplatense Spanish (voseo) without overloading") {
				t.Errorf("%s still hardcodes the Rioplatense reply directive; it must come from composeLanguageDirective", tt.path)
			}
			for _, required := range tt.requiredContract {
				if !strings.Contains(content, required) {
					t.Errorf("%s lost region-agnostic language contract %q — the strip must not widen beyond the regional directive", tt.path, required)
				}
			}
		})
	}
}
