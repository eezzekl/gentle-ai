package persona

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/agents"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/antigravity"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/codex"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/cursor"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/gemini"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/kilocode"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/kiro"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/qwen"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/vscode"
	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

// sharedPromptFileAdapters are the agents whose persona lands in a prompt file
// that other components also write into. Every one of them shares the file with
// the managed sdd-orchestrator section, so none of them may have its persona
// written as the whole file body.
//
// The three strategies differ only in wrapping: FileReplace writes plain
// markdown, InstructionsFile (VS Code) and SteeringFile (Kiro) prepend YAML
// frontmatter. sdd.Inject routes all three through injectFileAppend.
func sharedPromptFileAdapters() map[string]agents.Adapter {
	return map[string]agents.Adapter{
		"gemini":   gemini.NewAdapter(),
		"cursor":   cursor.NewAdapter(),
		"codex":    codex.NewAdapter(),
		"qwen":     qwen.NewAdapter(),
		"kilocode": kilocode.NewAdapter(),
		"vscode":   vscode.NewAdapter(),
		"kiro":     kiro.NewAdapter(),
	}
}

// pinAgentConfigEnv points the adapters that resolve their prompt path through
// platform config env vars at the test home, so VS Code does not write into the
// real user profile.
func pinAgentConfigEnv(t *testing.T, home string) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
}

// frontmatterAdapters are the subset that wrap their prompt file in YAML
// frontmatter, which must survive marker-bound persona injection intact.
func frontmatterAdapters() map[string]agents.Adapter {
	return map[string]agents.Adapter{
		"vscode": vscode.NewAdapter(),
		"kiro":   kiro.NewAdapter(),
	}
}

// gentlemanPersonas covers both Gentleman variants, since
// isGentlemanConversationPersona treats them identically.
func gentlemanPersonas() []model.PersonaID {
	return []model.PersonaID{model.PersonaGentleman, model.PersonaGentlemanNeutralArtifacts}
}

func writePromptFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
}

func readPromptFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return string(data)
}

// managedNeighbours is the pair of sections other components own in the same
// prompt file. Since the reference-file split these two blocks are the ONLY
// pointer to ~/.gemini/references/, so losing them silently disconnects the
// engram protocol and the whole SDD orchestrator contract.
const managedNeighbours = "<!-- gentle-ai:sdd-orchestrator -->\nRead the orchestrator reference file first.\n<!-- /gentle-ai:sdd-orchestrator -->\n\n" +
	"<!-- gentle-ai:engram-protocol -->\nRead the engram reference file first.\n<!-- /gentle-ai:engram-protocol -->\n"

// TestInjectPreservesManagedSectionsForGentleman is the regression guard for the
// stub wipe. The old preservation path bailed out for the Gentleman persona, so
// persona.Inject wrote the raw asset as the entire file and destroyed every
// managed section another component had written. Persona now owns one
// marker-bound section like every other component, which is what this pins.
func TestInjectPreservesManagedSectionsForGentleman(t *testing.T) {
	for name, adapter := range sharedPromptFileAdapters() {
		for _, personaID := range gentlemanPersonas() {
			t.Run(name+"/"+string(personaID), func(t *testing.T) {
				home := t.TempDir()
				pinAgentConfigEnv(t, home)
				promptPath := adapter.SystemPromptFile(home)
				writePromptFile(t, promptPath, managedNeighbours)

				if _, err := Inject(home, adapter, personaID); err != nil {
					t.Fatalf("Inject(%s, %s) error = %v", name, personaID, err)
				}

				got := readPromptFile(t, promptPath)
				for _, marker := range []string{
					"<!-- gentle-ai:sdd-orchestrator -->",
					"<!-- /gentle-ai:sdd-orchestrator -->",
					"<!-- gentle-ai:engram-protocol -->",
					"<!-- /gentle-ai:engram-protocol -->",
				} {
					if !strings.Contains(got, marker) {
						t.Fatalf("persona injection destroyed %q; got:\n%s", marker, got)
					}
				}
				if !strings.Contains(got, "Read the orchestrator reference file first.") ||
					!strings.Contains(got, "Read the engram reference file first.") {
					t.Fatalf("persona injection destroyed managed section bodies; got:\n%s", got)
				}
			})
		}
	}
}

// TestInjectWritesMarkerBoundPersonaForGentleman pins the shape that makes the
// preservation possible at all: without markers around the persona content, a
// later sync cannot tell managed persona output from user-authored text, so it
// can only guess — and guessing is what destroyed the neighbours.
func TestInjectWritesMarkerBoundPersonaForGentleman(t *testing.T) {
	for name, adapter := range sharedPromptFileAdapters() {
		for _, personaID := range gentlemanPersonas() {
			t.Run(name+"/"+string(personaID), func(t *testing.T) {
				home := t.TempDir()
				pinAgentConfigEnv(t, home)

				if _, err := Inject(home, adapter, personaID); err != nil {
					t.Fatalf("Inject(%s, %s) error = %v", name, personaID, err)
				}

				got := readPromptFile(t, adapter.SystemPromptFile(home))
				if !strings.Contains(got, "<!-- gentle-ai:persona -->") ||
					!strings.Contains(got, "<!-- /gentle-ai:persona -->") {
					t.Fatalf("persona content is not marker-bound on a fresh install; got:\n%s", got)
				}
				if n := strings.Count(got, "<!-- gentle-ai:persona -->"); n != 1 {
					t.Fatalf("persona marker appears %d times, want 1; got:\n%s", n, got)
				}
			})
		}
	}
}

// TestInjectPersonaIsIdempotentForGentleman guards against the marker-bound
// rewrite duplicating the section or oscillating across syncs.
func TestInjectPersonaIsIdempotentForGentleman(t *testing.T) {
	for name, adapter := range sharedPromptFileAdapters() {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			pinAgentConfigEnv(t, home)
			promptPath := adapter.SystemPromptFile(home)
			writePromptFile(t, promptPath, managedNeighbours)

			if _, err := Inject(home, adapter, model.PersonaGentleman); err != nil {
				t.Fatalf("Inject(%s) first pass error = %v", name, err)
			}
			first := readPromptFile(t, promptPath)

			if _, err := Inject(home, adapter, model.PersonaGentleman); err != nil {
				t.Fatalf("Inject(%s) second pass error = %v", name, err)
			}
			if second := readPromptFile(t, promptPath); second != first {
				t.Fatalf("second persona injection changed %s:\n%s", name, second)
			}
		})
	}
}

// TestInjectPersonaPreservesUserAuthoredContent is the safety half: healing the
// file must not become a licence to delete text the user wrote. Only gentle-ai's
// own legacy output may be replaced.
func TestInjectPersonaPreservesUserAuthoredContent(t *testing.T) {
	adapter := gemini.NewAdapter()
	home := t.TempDir()
	promptPath := adapter.SystemPromptFile(home)

	userRules := "# My own rules\n\nAlways answer in metric units.\n"
	writePromptFile(t, promptPath, userRules+"\n"+managedNeighbours)

	if _, err := Inject(home, adapter, model.PersonaGentleman); err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	got := readPromptFile(t, promptPath)
	if !strings.Contains(got, "Always answer in metric units.") {
		t.Fatalf("persona injection destroyed user-authored content; got:\n%s", got)
	}
}

// TestInjectSharedRootConvergesAcrossAdapterOrder is the design D3 guarantee at
// the persona level: gemini-cli (FileReplace) and antigravity (AppendToFile)
// write the same ~/.gemini/GEMINI.md, so the result must not depend on which one
// ran last.
func TestInjectSharedRootConvergesAcrossAdapterOrder(t *testing.T) {
	run := func(t *testing.T, order []agents.Adapter) string {
		t.Helper()
		home := t.TempDir()
		promptPath := order[0].SystemPromptFile(home)
		writePromptFile(t, promptPath, managedNeighbours)
		for _, adapter := range order {
			if _, err := Inject(home, adapter, model.PersonaGentleman); err != nil {
				t.Fatalf("Inject(%s) error = %v", adapter.Agent(), err)
			}
		}
		return readPromptFile(t, promptPath)
	}

	geminiFirst := run(t, []agents.Adapter{gemini.NewAdapter(), antigravity.NewAdapter()})
	antigravityFirst := run(t, []agents.Adapter{antigravity.NewAdapter(), gemini.NewAdapter()})

	if geminiFirst != antigravityFirst {
		t.Fatalf("shared root differs by adapter order\n--- gemini first ---\n%s\n--- antigravity first ---\n%s", geminiFirst, antigravityFirst)
	}
}

// TestInjectKeepsFrontmatterIntactAndUnduplicated covers the wrapped strategies.
// VS Code and Kiro read a YAML frontmatter header, and both the persona and the
// SDD component seed it. Marker-bound persona injection must leave exactly one
// header at the top of the file — a second one further down is not metadata, it
// is a visible `---` divider in the loaded prompt.
func TestInjectKeepsFrontmatterIntactAndUnduplicated(t *testing.T) {
	for name, adapter := range frontmatterAdapters() {
		for _, personaID := range gentlemanPersonas() {
			t.Run(name+"/"+string(personaID), func(t *testing.T) {
				home := t.TempDir()
				pinAgentConfigEnv(t, home)
				promptPath := adapter.SystemPromptFile(home)

				if _, err := Inject(home, adapter, personaID); err != nil {
					t.Fatalf("Inject(%s) error = %v", name, err)
				}

				got := readPromptFile(t, promptPath)
				if !strings.HasPrefix(got, "---\n") {
					t.Fatalf("%s prompt file does not open with YAML frontmatter; got:\n%s", name, got)
				}
				if n := strings.Count(got, "\n---\n"); n != 1 {
					t.Fatalf("%s prompt file has %d frontmatter delimiters after the opener, want 1; got:\n%s", name, n, got)
				}
				if !strings.Contains(got, "<!-- gentle-ai:persona -->") {
					t.Fatalf("%s persona content is not marker-bound; got:\n%s", name, got)
				}
			})
		}
	}
}

// TestInjectPreservesExistingFrontmatter is the migration half: an install that
// already has a frontmatter header must keep that exact header rather than
// gaining a second one when persona is re-injected.
func TestInjectPreservesExistingFrontmatter(t *testing.T) {
	for name, adapter := range frontmatterAdapters() {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			pinAgentConfigEnv(t, home)
			promptPath := adapter.SystemPromptFile(home)

			// Seed the file the way the SDD component leaves it when it runs first.
			seeded := "---\ninclusion: always\n---\n\n" + managedNeighbours
			writePromptFile(t, promptPath, seeded)

			if _, err := Inject(home, adapter, model.PersonaGentleman); err != nil {
				t.Fatalf("Inject(%s) error = %v", name, err)
			}

			got := readPromptFile(t, promptPath)
			if !strings.HasPrefix(got, "---\ninclusion: always\n---\n") {
				t.Fatalf("%s lost or replaced the existing frontmatter; got:\n%s", name, got)
			}
			if n := strings.Count(got, "\n---\n"); n != 1 {
				t.Fatalf("%s duplicated the frontmatter (%d delimiters); got:\n%s", name, n, got)
			}
			if !strings.Contains(got, "Read the orchestrator reference file first.") {
				t.Fatalf("%s persona injection destroyed the managed neighbour; got:\n%s", name, got)
			}
		})
	}
}
