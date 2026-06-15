package persona

import "strings"

// regionalVoiceDirectives lists the exact bullet lines that hardcode a regional
// reply voice in the shipped persona and output-style assets.
//
// These lines — and only these — are what the two-axis model replaces: the
// composed language directive supplies the reply voice for the region the user
// actually selected, so a hardcoded Rioplatense line would contradict a user who
// picked Mexican or Colombian Spanish.
//
// Everything else in the assets' language sections is region-AGNOSTIC contract:
// reply-language anchoring to the latest user request, no drift from persona
// wording, and the English no-code-switching rules. Those are compatible with any
// region and are guarded upstream by the asset language-contract tests. Removing
// the whole section, as an earlier revision of this transform did, would silently
// drop them.
//
// The `## Persona Scope` guard ("Never inject Rioplatense slang, voseo, ... into
// generated code") also names voseo but governs generated ARTIFACTS rather than
// reply voice, so it is deliberately absent from this list.
var regionalVoiceDirectives = []string{
	"- When replying to the user in Spanish, use warm natural Rioplatense Spanish (voseo) without overloading the reply with slang.\n",
}

// stripRegionalVoiceDirective removes the hardcoded regional reply directive
// from the given markdown content and returns the result. Content that carries
// no such directive is returned unchanged, which makes the transform a no-op for
// assets whose voice content already lives elsewhere, and idempotent on repeated
// sync runs.
//
// This is a pure string transform — no file I/O.
func stripRegionalVoiceDirective(content string) string {
	for _, directive := range regionalVoiceDirectives {
		content = strings.ReplaceAll(content, directive, "")
	}
	return content
}
