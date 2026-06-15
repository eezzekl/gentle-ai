package model

import "fmt"

// ComposeLanguageDirective returns a language instruction section for inclusion
// in the channel that carries an agent's reply voice. It is the single source of
// truth for all regional voice injection — no per-region asset files exist; this
// function composes the directive at build time.
//
// The section is headed "## Reply Language and Region" rather than "## Language"
// on purpose: the shipped assets keep their own "## Language" / "## Language
// Rules" sections, which hold the REGION-AGNOSTIC contract (anchoring the reply
// language to the latest user request, no drift, no code-switching). Reusing that
// header would emit two identically named H2 sections into one document. The two
// are complementary — the static section says HOW to follow the user's language,
// this one says WHICH regional voice to use.
//
// Region behavior:
//   - Curated region IDs (argentina, mexico, colombia, spain, chile) →
//     compose a "reply in X Spanish (style)" directive.
//   - RegionUserLanguage → "reply in the language the user writes in".
//   - Unrecognized string → treat as free text, inject verbatim.
//
// ArtifactsInEnglish:
//   - true  → "All generated artifacts (code, comments, identifiers) in English."
//   - false → "Generated artifacts may be in the selected language."
func ComposeLanguageDirective(region RegionID, artifactsInEnglish bool) string {
	langClause := composeLanguageClause(region)
	artifactsClause := composeArtifactsClause(artifactsInEnglish)
	return fmt.Sprintf("## Reply Language and Region\n\n%s\n\n%s", langClause, artifactsClause)
}

// composeLanguageClause builds the reply-language portion of the directive.
func composeLanguageClause(region RegionID) string {
	switch region {
	case RegionArgentina:
		return "Reply in Rioplatense Spanish (voseo, Argentine idioms). " +
			"Use voseo forms (vos, tus, etc.) naturally throughout your replies."
	case RegionMexico:
		return "Reply in Mexican Spanish (tuteo, Mexico Spanish). " +
			"Use natural Mexican idioms and tuteo forms."
	case RegionColombia:
		return "Reply in Colombian Spanish (paisa region). " +
			"Use natural Colombian/paisa idioms and tuteo forms."
	case RegionSpain:
		return "Reply in Castilian Spanish (Spain). " +
			"Use vosotros forms and Peninsular Spanish idioms."
	case RegionChile:
		return "Reply in Chilean Spanish. " +
			"Use natural Chilean idioms and speech patterns."
	case RegionUserLanguage:
		return "Reply in the language the user writes in. " +
			"Match the user's language and register for all conversational replies."
	default:
		// Free text: inject the raw string verbatim so the user's custom
		// region description is preserved exactly.
		return fmt.Sprintf("Reply in the following language/region: %s.", string(region))
	}
}

// composeArtifactsClause builds the artifacts-language portion of the directive.
func composeArtifactsClause(artifactsInEnglish bool) string {
	if artifactsInEnglish {
		return "All generated artifacts (code, comments, identifiers, UI copy, " +
			"commit messages, documentation) must be in English, " +
			"regardless of the conversational language."
	}
	return "Generated artifacts (code, comments, identifiers, UI copy) may be " +
		"in the selected language when explicitly requested by the user or project context."
}
