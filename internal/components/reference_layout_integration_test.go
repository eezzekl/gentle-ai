package components_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/agents"
	"github.com/gentleman-programming/gentle-ai/v2/internal/components/engram"
	"github.com/gentleman-programming/gentle-ai/v2/internal/components/persona"
	"github.com/gentleman-programming/gentle-ai/v2/internal/components/sdd"
	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

// rootBudgetLimit mirrors the character count at which Antigravity silently
// truncates an auto-loaded rules file. The whole stub+references layout exists to
// keep the assembled root under it.
const rootBudgetLimit = 12000

// assembleSharedRoot runs the three components that write ~/.gemini/GEMINI.md in
// the same order the sync pipeline uses (persona → sdd → engram). That order is
// load-bearing: persona must land first, and engram appends last.
func assembleSharedRoot(t *testing.T, home string, adapter agents.Adapter, selected []model.AgentID) {
	t.Helper()

	if _, err := persona.Inject(home, adapter, model.PersonaGentleman); err != nil {
		t.Fatalf("persona.Inject(%s) error = %v", adapter.Agent(), err)
	}
	if _, err := sdd.Inject(home, adapter, "", sdd.InjectOptions{SelectedAgents: selected}); err != nil {
		t.Fatalf("sdd.Inject(%s) error = %v", adapter.Agent(), err)
	}
	if _, err := engram.Inject(home, adapter); err != nil {
		t.Fatalf("engram.Inject(%s) error = %v", adapter.Agent(), err)
	}
}

// pinSharedReferenceHome pins home for both components that own a reference file,
// so the split applies instead of silently falling back to the inline body.
func pinSharedReferenceHome(t *testing.T, home string) {
	t.Helper()
	sdd.SetUserHomeDirForTest(t, home)
	engram.SetUserHomeDirForTest(t, home)
	engram.SetLookPathForTest(t, "/opt/homebrew/bin/engram", "")
}

func sharedRootPath(home string) string {
	return filepath.Join(home, ".gemini", "GEMINI.md")
}

func referencePaths(home string) []string {
	return []string{
		filepath.Join(home, ".gemini", "references", "engram-protocol.md"),
		filepath.Join(home, ".gemini", "references", "sdd-orchestrator.md"),
	}
}

// personaContentSample returns the longest single line the persona component
// writes for an adapter, taken from a clean injection of persona alone. Asserting
// against real rendered content keeps the byte-budget test honest without pinning
// a hand-copied phrase that silently rots when the asset is rewritten.
func personaContentSample(t *testing.T, adapter agents.Adapter) string {
	t.Helper()

	home := t.TempDir()
	if _, err := persona.Inject(home, adapter, model.PersonaGentleman); err != nil {
		t.Fatalf("persona.Inject(%s) error = %v", adapter.Agent(), err)
	}

	longest := ""
	for _, line := range strings.Split(readFileString(t, sharedRootPath(home)), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "<!--") {
			continue
		}
		if len(line) > len(longest) {
			longest = line
		}
	}
	if len(longest) < 40 {
		t.Fatalf("persona injection for %s produced no substantial content line", adapter.Agent())
	}
	return longest
}

// legacyInlineBody builds a pre-split inline body of roughly size bytes, opening
// with the heading the migration assertions look for. The real assets are not
// used on purpose: the fixture must stay a fixed, oversized legacy shape even
// when the shipped assets are rewritten upstream.
func legacyInlineBody(heading string, size int) string {
	var b strings.Builder
	b.WriteString(heading + "\n\n")
	filler := "Legacy inline orchestration prose that used to live in the root prompt file.\n"
	for b.Len() < size {
		b.WriteString(filler)
	}
	return b.String()
}

// headerOf returns the leading YAML frontmatter block of a prompt file, or ""
// when the file has none.
func headerOf(content string) string {
	if !strings.HasPrefix(content, "---\n") {
		return ""
	}
	end := strings.Index(content[4:], "\n---\n")
	if end < 0 {
		return ""
	}
	return content[:4+end+len("\n---\n")]
}

func readFileString(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return string(data)
}

// TestAssembledRootStaysUnderByteBudget is the Root Assembly Byte Budget
// requirement measured end to end: persona inline, both large bodies out of the
// root, and the whole file under 12,000 characters for both adapters.
func TestAssembledRootStaysUnderByteBudget(t *testing.T) {
	for name, adapter := range map[string]agents.Adapter{
		"gemini":      geminiAdapter(),
		"antigravity": antigravityAdapter(),
	} {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			pinSharedReferenceHome(t, home)

			assembleSharedRoot(t, home, adapter, []model.AgentID{adapter.Agent()})

			root := readFileString(t, sharedRootPath(home))
			if len(root) >= rootBudgetLimit {
				t.Fatalf("assembled root is %d bytes, must stay under %d", len(root), rootBudgetLimit)
			}

			// Persona must survive inline — it is the content the budget exists to
			// protect. The check is on content, not on the marker: gemini-cli writes
			// the Gentleman asset unwrapped while antigravity wraps it in a managed
			// section, and the requirement is about the content reaching the model.
			if sample := personaContentSample(t, adapter); !strings.Contains(root, sample) {
				t.Fatalf("assembled root lost the persona content (looked for %q); got:\n%s", sample, root)
			}
			// Neither large body may be inline.
			if strings.Contains(root, "### PROACTIVE SAVE TRIGGERS") {
				t.Fatalf("assembled root still inlines the engram protocol body")
			}
			if strings.Contains(root, "## Agent Teams Orchestrator") {
				t.Fatalf("assembled root still inlines the SDD orchestrator body")
			}
			// Both reference files must exist.
			for _, path := range referencePaths(home) {
				if _, err := os.Stat(path); err != nil {
					t.Fatalf("reference file %q missing: %v", path, err)
				}
			}
		})
	}
}

// TestLegacyInlineRootMigratesToReferenceLayout is the Migration requirement: a
// pre-split install carries both bodies inline in a ~54KB root. Syncing over it
// must convert to stub+references, keep persona first, and leave content this
// tool does not own completely alone.
func TestLegacyInlineRootMigratesToReferenceLayout(t *testing.T) {
	home := t.TempDir()
	pinSharedReferenceHome(t, home)

	adapter := antigravityAdapter()
	rootPath := sharedRootPath(home)

	// Build the legacy layout the way a pre-split install left it: persona first,
	// then both full bodies inline under their markers, plus a block this tool
	// has never owned.
	foreign := "<!-- somebody-elses-tool -->\nHand-written user rules that gentle-ai must not touch.\n<!-- /somebody-elses-tool -->\n"
	legacy := "<!-- gentle-ai:persona -->\nlegacy persona body\n<!-- /gentle-ai:persona -->\n\n" +
		"<!-- gentle-ai:sdd-orchestrator -->\n" + legacyInlineBody("## Agent Teams Orchestrator", 27_000) + "\n<!-- /gentle-ai:sdd-orchestrator -->\n\n" +
		"<!-- gentle-ai:engram-protocol -->\n" + legacyInlineBody("### PROACTIVE SAVE TRIGGERS", 27_000) + "\n<!-- /gentle-ai:engram-protocol -->\n\n" +
		foreign
	if err := os.MkdirAll(filepath.Dir(rootPath), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	if err := os.WriteFile(rootPath, []byte(legacy), 0o644); err != nil {
		t.Fatalf("WriteFile(legacy root) error = %v", err)
	}
	if len(legacy) < 30_000 {
		t.Fatalf("legacy fixture is only %d bytes; it must be large enough to prove the truncation problem", len(legacy))
	}

	assembleSharedRoot(t, home, adapter, []model.AgentID{model.AgentAntigravity})

	root := readFileString(t, rootPath)

	if len(root) >= rootBudgetLimit {
		t.Fatalf("migrated root is %d bytes, must stay under %d", len(root), rootBudgetLimit)
	}
	if strings.Contains(root, "## Agent Teams Orchestrator") {
		t.Fatalf("migration left the SDD orchestrator body inline")
	}
	if strings.Contains(root, "### PROACTIVE SAVE TRIGGERS") {
		t.Fatalf("migration left the engram protocol body inline")
	}
	if !strings.Contains(root, foreign) {
		t.Fatalf("migration modified content gentle-ai does not own; got:\n%s", root)
	}

	// Persona-first ordering must survive the migration.
	personaIdx := strings.Index(root, "<!-- gentle-ai:persona -->")
	sddIdx := strings.Index(root, "<!-- gentle-ai:sdd-orchestrator -->")
	engramIdx := strings.Index(root, "<!-- gentle-ai:engram-protocol -->")
	if personaIdx < 0 || sddIdx < 0 || engramIdx < 0 {
		t.Fatalf("migrated root is missing a managed marker (persona=%d sdd=%d engram=%d)", personaIdx, sddIdx, engramIdx)
	}
	if personaIdx > sddIdx || personaIdx > engramIdx {
		t.Fatalf("persona is no longer first (persona=%d sdd=%d engram=%d)", personaIdx, sddIdx, engramIdx)
	}

	// Each managed marker must appear exactly once — migration replaces, never duplicates.
	for _, marker := range []string{"<!-- gentle-ai:persona -->", "<!-- gentle-ai:sdd-orchestrator -->", "<!-- gentle-ai:engram-protocol -->"} {
		if n := strings.Count(root, marker); n != 1 {
			t.Fatalf("marker %q appears %d times after migration, want 1", marker, n)
		}
	}
}

// TestSecondSyncOverReferenceLayoutIsByteIdentical is the Idempotency
// requirement at assembly level: no oscillation across a full second pass.
func TestSecondSyncOverReferenceLayoutIsByteIdentical(t *testing.T) {
	home := t.TempDir()
	pinSharedReferenceHome(t, home)

	adapter := antigravityAdapter()
	selected := []model.AgentID{model.AgentAntigravity}

	assembleSharedRoot(t, home, adapter, selected)

	first := map[string]string{sharedRootPath(home): readFileString(t, sharedRootPath(home))}
	for _, path := range referencePaths(home) {
		first[path] = readFileString(t, path)
	}

	assembleSharedRoot(t, home, adapter, selected)

	for path, want := range first {
		if got := readFileString(t, path); got != want {
			t.Fatalf("second pass changed %q", path)
		}
	}
}

// TestSyncOrderConvergence is design.md D3: gemini and antigravity write the same
// root and the same reference directory, so the result must not depend on which
// agent synced last. Without this, two machines with the same config drift.
func TestSyncOrderConvergence(t *testing.T) {
	selected := []model.AgentID{model.AgentGeminiCLI, model.AgentAntigravity}

	run := func(t *testing.T, order []agents.Adapter) map[string]string {
		t.Helper()
		home := t.TempDir()
		pinSharedReferenceHome(t, home)
		for _, adapter := range order {
			assembleSharedRoot(t, home, adapter, selected)
		}

		out := map[string]string{"root": readFileString(t, sharedRootPath(home))}
		for _, path := range referencePaths(home) {
			out[filepath.Base(path)] = readFileString(t, path)
		}
		return out
	}

	geminiFirst := run(t, []agents.Adapter{geminiAdapter(), antigravityAdapter()})
	antigravityFirst := run(t, []agents.Adapter{antigravityAdapter(), geminiAdapter()})

	for key, want := range geminiFirst {
		if got := antigravityFirst[key]; got != want {
			t.Fatalf("%s differs by sync order (gemini-first vs antigravity-first)", key)
		}
	}
}

// TestSharedPromptFileConvergesAcrossComponentOrder pins the invariant that keeps
// the persona and SDD frontmatter constants honest. VS Code and Kiro read a YAML
// header, and BOTH components seed it when the file is empty — so if the two
// headers ever drift apart, the file's content starts depending on which
// component ran first. Comparing rendered bytes catches that drift; comparing the
// unexported constants could not.
func TestSharedPromptFileConvergesAcrossComponentOrder(t *testing.T) {
	for name, adapter := range map[string]agents.Adapter{
		"vscode": vscodeAdapter(),
		"kiro":   kiroAdapter(),
	} {
		t.Run(name, func(t *testing.T) {
			injectPersona := func(home string) {
				if _, err := persona.Inject(home, adapter, model.PersonaGentleman); err != nil {
					t.Fatalf("persona.Inject(%s) error = %v", name, err)
				}
			}
			injectSDD := func(home string) {
				if _, err := sdd.Inject(home, adapter, ""); err != nil {
					t.Fatalf("sdd.Inject(%s) error = %v", name, err)
				}
			}

			run := func(steps ...func(string)) string {
				home := t.TempDir()
				t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
				t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
				for _, step := range steps {
					step(home)
				}
				return readFileString(t, adapter.SystemPromptFile(home))
			}

			// Block order is NOT asserted: the pipeline fixes persona before SDD at
			// internal/cli/sync.go and that order is load-bearing, so "SDD first" is a
			// state the product never produces. What must hold either way is that
			// whichever component seeds the header, the header is the same one and
			// there is exactly one of it — that is the drift this test exists to catch.
			for label, got := range map[string]string{
				"persona first": run(injectPersona, injectSDD),
				"sdd first":     run(injectSDD, injectPersona),
			} {
				if !strings.HasPrefix(got, "---\n") {
					t.Fatalf("%s (%s) does not open with its YAML header; got:\n%s", name, label, got)
				}
				if n := strings.Count(got, "\n---\n"); n != 1 {
					t.Fatalf("%s (%s) has %d header delimiters after the opener, want 1; got:\n%s", name, label, n, got)
				}
				if !strings.Contains(got, "<!-- gentle-ai:persona -->") {
					t.Fatalf("%s (%s) lost the persona section; got:\n%s", name, label, got)
				}
				if !strings.Contains(got, "<!-- gentle-ai:sdd-orchestrator -->") {
					t.Fatalf("%s (%s) lost the sdd-orchestrator section; got:\n%s", name, label, got)
				}
			}

			// The header must be byte-identical no matter which component wrote it:
			// persona and sdd hold separate copies of the constant, and a drift there
			// would silently change the file's metadata depending on run order.
			personaHeader := run(injectPersona)
			sddHeader := run(injectSDD)
			if h1, h2 := headerOf(personaHeader), headerOf(sddHeader); h1 != h2 {
				t.Fatalf("%s YAML header drifted between components\n--- persona ---\n%s\n--- sdd ---\n%s", name, h1, h2)
			}
		})
	}
}
