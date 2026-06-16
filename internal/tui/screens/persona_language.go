package screens

import (
	"strings"

	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
	"github.com/gentleman-programming/gentle-ai/v2/internal/tui/styles"
)

// RegionFreeTextSentinel marks the "Otro… (texto libre)" option in the region
// radio list. It is NOT a stored region value: when chosen, the typed free text
// is written to Selection.Region verbatim.
const RegionFreeTextSentinel model.RegionID = "__free_text__"

// LanguageRegionOptions returns the ordered region radio list for the persona
// language screen: 5 curated regions, the user-language option, and the
// free-text sentinel.
func LanguageRegionOptions() []model.RegionID {
	return []model.RegionID{
		model.RegionArgentina,
		model.RegionMexico,
		model.RegionColombia,
		model.RegionSpain,
		model.RegionChile,
		model.RegionUserLanguage,
		RegionFreeTextSentinel,
	}
}

// regionRadioLabel resolves the Spanish, gentilicio-first label for a region
// option. The free-text sentinel renders the typed value (or a placeholder).
func regionRadioLabel(option model.RegionID, freeText string) string {
	if option == RegionFreeTextSentinel {
		if strings.TrimSpace(freeText) == "" {
			return "Otro… (texto libre)"
		}
		return "Otro… (texto libre): " + freeText
	}
	if label, ok := model.RegionMap[option]; ok {
		return label
	}
	return string(option)
}

// regionIsSelected reports whether the given option is the currently selected
// region. The free-text sentinel is selected when Selection.Region holds a value
// that is not one of the curated/user-language IDs.
func regionIsSelected(option model.RegionID, selected string) bool {
	if option == RegionFreeTextSentinel {
		if selected == "" {
			return false
		}
		for _, opt := range LanguageRegionOptions() {
			if opt == RegionFreeTextSentinel {
				continue
			}
			if string(opt) == selected {
				return false
			}
		}
		return true
	}
	return string(option) == selected
}

// RenderPersonaLanguage renders the language/region screen: region radios, the
// artifacts-in-English checkbox, and a Back option. The cursor spans region
// options [0..N), then the checkbox at N, then Back at N+1.
func RenderPersonaLanguage(selection model.Selection, cursor int, freeText string) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Choose your Language / Region"))
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render("How the assistant speaks to you — independent from the persona style."))
	b.WriteString("\n\n")

	options := LanguageRegionOptions()
	for idx, option := range options {
		focused := idx == cursor
		selected := regionIsSelected(option, selection.Region)
		b.WriteString(renderRadio(regionRadioLabel(option, freeText), selected, focused))
	}

	b.WriteString("\n")
	checkboxIdx := len(options)
	b.WriteString(renderCheckbox(
		"Generated artifacts (code, comments, identifiers) in English",
		selection.ArtifactsInEnglish,
		cursor == checkboxIdx,
	))

	b.WriteString("\n")
	b.WriteString(renderOptions([]string{"Back"}, cursor-checkboxIdx-1))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select/toggle • type for free text • esc: back"))

	return b.String()
}
