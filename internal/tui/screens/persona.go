package screens

import (
	"strings"

	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
	"github.com/gentleman-programming/gentle-ai/v2/internal/tui/styles"
)

func PersonaOptions() []model.PersonaID {
	return []model.PersonaID{model.PersonaGentle, model.PersonaNeutral, model.PersonaCustom}
}

var personaDescriptions = map[model.PersonaID]string{
	// Descriptions describe the STYLE axis only — the reply language and region
	// are chosen on the next screen, so no entry here may promise a regional
	// voice. They also must not reuse the word "neutral" as a descriptor: issue
	// #833 is precisely that the word named two unrelated axes at once, and
	// TestPersonaDescriptionsNeverReuseNeutral pins that.
	model.PersonaGentle:  "Teaching-first mentor tone; reply language chosen on the next screen",
	model.PersonaNeutral: "No marked conversation tone; English technical artifacts",
	model.PersonaCustom:  "Do not install a managed persona; choose themes/logo on the next screens",
}

func RenderPersona(selected model.PersonaID, cursor int) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Choose your Persona"))
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render("Your own Gentleman! teaches before it solves."))
	b.WriteString("\n\n")

	for idx, persona := range PersonaOptions() {
		isSelected := persona == selected
		focused := idx == cursor
		b.WriteString(renderRadio(string(persona), isSelected, focused))
		b.WriteString(styles.SubtextStyle.Render("    " + personaDescriptions[persona]))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(renderOptions([]string{"Back"}, cursor-len(PersonaOptions())))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))

	return b.String()
}
