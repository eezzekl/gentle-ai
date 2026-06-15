package persona

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

func TestInjectGentlemanNeutralArtifactsUsesGentlemanConversationWithArtifactBoundary(t *testing.T) {
	home := t.TempDir()

	// PersonaGentlemanNeutralArtifacts is a migration alias. Argentina region +
	// artifactsInEnglish=true reproduces the original behavior it named:
	// Rioplatense conversation with English technical artifacts.
	result, err := Inject(home, opencodeAdapter(), model.PersonaGentlemanNeutralArtifacts, model.RegionArgentina, true)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() changed = false")
	}

	content, err := os.ReadFile(filepath.Join(home, ".config", "opencode", "AGENTS.md"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(content)
	for _, want := range []string{
		"Rioplatense",
		"Generated technical artifacts default to English",
		"Public/contextual comments follow the target context language",
		"If the selected reply language is English, every part of the direct reply must be English",
		"Prompts starting with or dominated by hi, hello, hey, or similar English greetings are English prompts",
		// Composed by ComposeLanguageDirective — the region-specific half of the
		// contract, alongside the region-agnostic rules the asset still ships.
		"All generated artifacts",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("installed persona missing %q; content:\n%s", want, text)
		}
	}
}
